package system

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	redigo "github.com/gomodule/redigo/redis"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	repo "github.com/seakee/go-api/app/repository/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	passkeyActionRegister = "register"
	passkeyActionLogin    = "login"
	passkeyActionReauth   = "reauth"
)

type PasskeyOptionsResult struct {
	ChallengeID string `json:"challenge_id"`
	Options     any    `json:"options"`
}

type PasskeyItem struct {
	ID          uint       `json:"id"`
	DisplayName string     `json:"display_name"`
	AAGUID      string     `json:"aaguid,omitempty"`
	Transports  []string   `json:"transports,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PasskeyCredential describes credential payload produced by browser WebAuthn APIs.
type PasskeyCredential struct {
	ID       string         `json:"id"`
	RawID    string         `json:"raw_id"`
	Type     string         `json:"type"`
	Response map[string]any `json:"response"`
}

type passkeyWebAuthnUser struct {
	id          []byte
	name        string
	displayName string
	credentials []webauthn.Credential
}

func (u passkeyWebAuthnUser) WebAuthnID() []byte {
	return u.id
}

func (u passkeyWebAuthnUser) WebAuthnName() string {
	return u.name
}

func (u passkeyWebAuthnUser) WebAuthnDisplayName() string {
	return u.displayName
}

func (u passkeyWebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (a authService) BeginPasskeyRegistration(ctx context.Context, userID uint, reauthTicket, displayName string) (result PasskeyOptionsResult, errCode int, err error) {
	if err = a.ensurePasskeyConfig(); err != nil {
		return result, e.ERROR, err
	}
	if _, errCode, err = a.consumeValidatedReauthTicket(ctx, reauthTicket, userID); errCode != e.SUCCESS {
		return result, errCode, err
	}

	user, errCode, err := a.activeUserByID(ctx, userID)
	if errCode != e.SUCCESS {
		return result, errCode, err
	}

	passkeys, err := a.passkeyRepo.ListByUserID(ctx, userID)
	if err != nil {
		return result, e.ERROR, err
	}

	waUser := buildPasskeyWebAuthnUser(user, passkeys, displayName)
	wa, err := a.getWebAuthn()
	if err != nil {
		return result, e.ERROR, err
	}

	options, session, err := wa.BeginRegistration(
		waUser,
		passkeyRegistrationOptions(a.config.WebAuthn.UserVerification)...,
	)
	if err != nil {
		return result, e.PasskeyRegistrationFailed, err
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return result, e.ERROR, err
	}

	challengeID, err := a.generatePasskeyChallenge(ctx, passkeyChallenge{
		Action:      passkeyActionRegister,
		UserID:      userID,
		SessionData: sessionData,
		DisplayName: waUser.displayName,
	})
	if err != nil {
		return result, e.ERROR, err
	}

	result = PasskeyOptionsResult{
		ChallengeID: challengeID,
		Options:     options,
	}

	return result, e.SUCCESS, nil
}

func (a authService) FinishPasskeyRegistration(ctx context.Context, userID uint, challengeID string, credential PasskeyCredential) (result PasskeyItem, errCode int, err error) {
	challengeID = strings.TrimSpace(challengeID)
	if challengeID == "" {
		return result, e.InvalidPasskeyChallenge, nil
	}

	challenge, err := a.parsePasskeyChallenge(ctx, challengeID)
	if err != nil {
		if errors.Is(err, redigo.ErrNil) {
			return result, e.PasskeyChallengeExpired, nil
		}
		return result, e.ERROR, err
	}
	defer func() {
		if consumeErr := a.consumePasskeyChallenge(ctx, challengeID); consumeErr != nil && a.logger != nil {
			a.logger.Error(ctx, "consume passkey registration challenge failed", zap.Error(consumeErr))
		}
	}()

	if challenge == nil || challenge.Action != passkeyActionRegister || challenge.UserID != userID {
		return result, e.InvalidPasskeyChallenge, nil
	}

	session, err := parsePasskeySessionData(challenge.SessionData)
	if err != nil {
		return result, e.InvalidPasskeyChallenge, err
	}
	if passkeySessionExpired(session) {
		return result, e.PasskeyChallengeExpired, nil
	}

	user, err := a.userRepo.DetailByID(ctx, userID)
	if err != nil {
		return result, e.ERROR, err
	}
	if user == nil || user.Status != 1 {
		return result, e.UserNotFound, nil
	}

	passkeys, err := a.passkeyRepo.ListByUserID(ctx, userID)
	if err != nil {
		return result, e.ERROR, err
	}

	waUser := buildPasskeyWebAuthnUser(user, passkeys, challenge.DisplayName)
	wa, err := a.getWebAuthn()
	if err != nil {
		return result, e.ERROR, err
	}

	payload, err := marshalPasskeyCredentialPayload(credential)
	if err != nil {
		return result, e.InvalidParams, err
	}

	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(payload))
	if err != nil {
		return result, e.PasskeyVerificationFailed, err
	}

	verifiedCredential, err := wa.CreateCredential(waUser, *session, parsed)
	if err != nil {
		return result, e.PasskeyVerificationFailed, err
	}

	credentialID := encodeBytesToBase64URL(verifiedCredential.ID)
	existingPasskey, err := a.passkeyRepo.DetailByCredentialID(ctx, credentialID)
	if err != nil {
		return result, e.ERROR, err
	}
	if existingPasskey != nil {
		return result, e.PasskeyCredentialExists, nil
	}

	credentialJSON, err := json.Marshal(verifiedCredential)
	if err != nil {
		return result, e.PasskeyRegistrationFailed, err
	}

	transportsJSON, err := json.Marshal(authenticatorTransportsToStrings(verifiedCredential.Transport))
	if err != nil {
		return result, e.PasskeyRegistrationFailed, err
	}

	passkey := &systemModel.UserPasskey{
		UserID:              userID,
		CredentialID:        credentialID,
		CredentialPublicKey: string(credentialJSON),
		SignCount:           verifiedCredential.Authenticator.SignCount,
		AAGUID:              encodeAAGUID(verifiedCredential.Authenticator.AAGUID),
		TransportsJSON:      string(transportsJSON),
		UserHandle:          encodeBytesToBase64URL(session.UserID),
		DisplayName:         firstNonEmpty(strings.TrimSpace(challenge.DisplayName), defaultPasskeyDisplayName(user)),
	}

	passkey, err = a.passkeyRepo.Create(ctx, passkey)
	if err != nil {
		return result, e.PasskeyRegistrationFailed, err
	}

	return mapPasskeyItem(*passkey), e.SUCCESS, nil
}

func (a authService) BeginPasskeyLogin(ctx context.Context) (result PasskeyOptionsResult, errCode int, err error) {
	if err = a.ensurePasskeyConfig(); err != nil {
		return result, e.ERROR, err
	}

	wa, err := a.getWebAuthn()
	if err != nil {
		return result, e.ERROR, err
	}

	options, session, err := wa.BeginDiscoverableLogin(passkeyDiscoverableLoginOptions(a.config.WebAuthn.UserVerification)...)
	if err != nil {
		return result, e.PasskeyLoginFailed, err
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return result, e.ERROR, err
	}

	challengeID, err := a.generatePasskeyChallenge(ctx, passkeyChallenge{
		Action:      passkeyActionLogin,
		SessionData: sessionData,
	})
	if err != nil {
		return result, e.ERROR, err
	}

	result = PasskeyOptionsResult{
		ChallengeID: challengeID,
		Options:     options,
	}

	return result, e.SUCCESS, nil
}

func (a authService) BeginPasskeyReauth(ctx context.Context, userID uint) (result PasskeyOptionsResult, errCode int, err error) {
	if err = a.ensurePasskeyConfig(); err != nil {
		return result, e.ERROR, err
	}

	user, errCode, err := a.activeUserByID(ctx, userID)
	if errCode != e.SUCCESS {
		return result, errCode, err
	}

	passkeys, err := a.passkeyRepo.ListByUserID(ctx, userID)
	if err != nil {
		return result, e.ERROR, err
	}
	if len(passkeys) == 0 {
		return result, e.PasskeyCredentialNotFound, nil
	}

	waUser := buildPasskeyWebAuthnUser(user, passkeys, "")
	wa, err := a.getWebAuthn()
	if err != nil {
		return result, e.ERROR, err
	}

	options, session, err := wa.BeginLogin(
		waUser,
		passkeyDiscoverableLoginOptions(a.config.WebAuthn.UserVerification)...,
	)
	if err != nil {
		return result, e.PasskeyVerificationFailed, err
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return result, e.ERROR, err
	}

	challengeID, err := a.generatePasskeyChallenge(ctx, passkeyChallenge{
		Action:      passkeyActionReauth,
		UserID:      userID,
		SessionData: sessionData,
	})
	if err != nil {
		return result, e.ERROR, err
	}

	result = PasskeyOptionsResult{
		ChallengeID: challengeID,
		Options:     options,
	}

	return result, e.SUCCESS, nil
}

func (a authService) FinishPasskeyLogin(ctx context.Context, challengeID string, credential PasskeyCredential) (token AccessToken, errCode int, err error) {
	challengeID = strings.TrimSpace(challengeID)
	if challengeID == "" {
		return token, e.InvalidPasskeyChallenge, nil
	}

	challenge, err := a.parsePasskeyChallenge(ctx, challengeID)
	if err != nil {
		if errors.Is(err, redigo.ErrNil) {
			return token, e.PasskeyChallengeExpired, nil
		}
		return token, e.ERROR, err
	}
	defer func() {
		if consumeErr := a.consumePasskeyChallenge(ctx, challengeID); consumeErr != nil && a.logger != nil {
			a.logger.Error(ctx, "consume passkey login challenge failed", zap.Error(consumeErr))
		}
	}()

	if challenge == nil || challenge.Action != passkeyActionLogin {
		return token, e.InvalidPasskeyChallenge, nil
	}

	session, err := parsePasskeySessionData(challenge.SessionData)
	if err != nil {
		return token, e.InvalidPasskeyChallenge, err
	}
	if passkeySessionExpired(session) {
		return token, e.PasskeyChallengeExpired, nil
	}

	payload, err := marshalPasskeyCredentialPayload(credential)
	if err != nil {
		return token, e.InvalidParams, err
	}

	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(payload))
	if err != nil {
		return token, e.PasskeyVerificationFailed, err
	}

	user, passkey, waUser, errCode, err := a.resolvePasskeyLoginUser(ctx, parsed.RawID, parsed.Response.UserHandle)
	if errCode != e.SUCCESS {
		return token, errCode, err
	}

	wa, err := a.getWebAuthn()
	if err != nil {
		return token, e.ERROR, err
	}

	_, verifiedCredential, err := wa.ValidatePasskeyLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
		return waUser, nil
	}, *session, parsed)
	if err != nil {
		return token, e.PasskeyVerificationFailed, err
	}

	credentialID := encodeBytesToBase64URL(verifiedCredential.ID)
	if passkey == nil || passkey.UserID != user.ID {
		return token, e.PasskeyCredentialNotFound, nil
	}
	if credentialID != passkey.CredentialID {
		return token, e.PasskeyVerificationFailed, nil
	}

	if err = a.passkeyRepo.UpdateSignCount(ctx, passkey.ID, verifiedCredential.Authenticator.SignCount); err != nil {
		return token, e.ERROR, err
	}
	if err = a.passkeyRepo.UpdateLastUsedAt(ctx, passkey.ID, time.Now()); err != nil {
		return token, e.ERROR, err
	}

	tokenString, err := a.generateToken(user.UserName, user.ID)
	if err != nil {
		return token, e.ERROR, err
	}

	token = AccessToken{
		Token:    tokenString,
		ExpireIn: a.config.TokenExpireIn,
	}

	return token, e.SUCCESS, nil
}

func (a authService) FinishPasskeyReauth(ctx context.Context, userID uint, challengeID string, credential PasskeyCredential) (result ReauthResult, errCode int, err error) {
	challengeID = strings.TrimSpace(challengeID)
	if challengeID == "" {
		return result, e.InvalidPasskeyChallenge, nil
	}

	challenge, err := a.parsePasskeyChallenge(ctx, challengeID)
	if err != nil {
		if errors.Is(err, redigo.ErrNil) {
			return result, e.PasskeyChallengeExpired, nil
		}
		return result, e.ERROR, err
	}
	defer func() {
		if consumeErr := a.consumePasskeyChallenge(ctx, challengeID); consumeErr != nil && a.logger != nil {
			a.logger.Error(ctx, "consume passkey reauth challenge failed", zap.Error(consumeErr))
		}
	}()

	if challenge == nil || challenge.Action != passkeyActionReauth || challenge.UserID != userID {
		return result, e.InvalidPasskeyChallenge, nil
	}

	session, err := parsePasskeySessionData(challenge.SessionData)
	if err != nil {
		return result, e.InvalidPasskeyChallenge, err
	}
	if passkeySessionExpired(session) {
		return result, e.PasskeyChallengeExpired, nil
	}

	user, errCode, err := a.activeUserByID(ctx, userID)
	if errCode != e.SUCCESS {
		return result, errCode, err
	}

	passkeys, err := a.passkeyRepo.ListByUserID(ctx, userID)
	if err != nil {
		return result, e.ERROR, err
	}
	if len(passkeys) == 0 {
		return result, e.PasskeyCredentialNotFound, nil
	}

	waUser := buildPasskeyWebAuthnUser(user, passkeys, "")
	wa, err := a.getWebAuthn()
	if err != nil {
		return result, e.ERROR, err
	}

	payload, err := marshalPasskeyCredentialPayload(credential)
	if err != nil {
		return result, e.InvalidParams, err
	}

	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(payload))
	if err != nil {
		return result, e.PasskeyVerificationFailed, err
	}

	verifiedCredential, err := wa.ValidateLogin(waUser, *session, parsed)
	if err != nil {
		return result, e.PasskeyVerificationFailed, err
	}

	credentialID := encodeBytesToBase64URL(verifiedCredential.ID)
	passkey := findPasskeyByCredentialID(passkeys, credentialID)
	if passkey == nil {
		return result, e.PasskeyCredentialNotFound, nil
	}

	if err = a.passkeyRepo.UpdateSignCount(ctx, passkey.ID, verifiedCredential.Authenticator.SignCount); err != nil {
		return result, e.ERROR, err
	}
	if err = a.passkeyRepo.UpdateLastUsedAt(ctx, passkey.ID, time.Now()); err != nil {
		return result, e.ERROR, err
	}

	return a.issueHighRiskReauthTicket(ctx, userID)
}

func (a authService) ListPasskeys(ctx context.Context, userID uint) (list []PasskeyItem, errCode int, err error) {
	if userID == 0 {
		return nil, e.InvalidParams, nil
	}

	user, err := a.userRepo.DetailByID(ctx, userID)
	if err != nil {
		return nil, e.ERROR, err
	}
	if user == nil || user.Status != 1 {
		return nil, e.UserNotFound, nil
	}

	passkeys, err := a.passkeyRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, e.ERROR, err
	}

	list = make([]PasskeyItem, 0, len(passkeys))
	for _, passkey := range passkeys {
		list = append(list, mapPasskeyItem(passkey))
	}

	return list, e.SUCCESS, nil
}

func (a authService) DeletePasskey(ctx context.Context, userID, passkeyID uint, reauthTicketCode string) (errCode int, err error) {
	if passkeyID == 0 {
		return e.PasskeyCanNotBeNull, nil
	}
	if _, errCode, err = a.validateReauthTicket(ctx, reauthTicketCode, userID); errCode != e.SUCCESS {
		return errCode, err
	}

	passkey, err := a.passkeyRepo.DetailByIDAndUserID(ctx, passkeyID, userID)
	if err != nil {
		return e.ERROR, err
	}
	if passkey == nil {
		return e.PasskeyCredentialNotFound, nil
	}

	loginMethodCount, err := a.loginMethodCount(ctx, userID)
	if err != nil {
		return e.ERROR, err
	}
	if loginMethodCount <= 1 {
		return e.LastLoginMethodCannotBeRemoved, nil
	}

	if err = a.passkeyRepo.DeleteByIDAndUserID(ctx, passkeyID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return e.PasskeyCredentialNotFound, nil
		}
		return e.ERROR, err
	}
	if err = a.consumeReauthTicket(ctx, reauthTicketCode); err != nil && a.logger != nil {
		a.logger.Error(ctx, "failed to consume reauth ticket after deleting passkey", zap.Error(err))
	}

	return e.SUCCESS, nil
}

func (a authService) loginMethodCount(ctx context.Context, userID uint) (int, error) {
	return countLoginMethods(ctx, a.userRepo, a.identityRepo, a.passkeyRepo, userID)
}

func countLoginMethods(ctx context.Context, userRepo repo.UserRepo, identityRepo repo.UserIdentityRepo, passkeyRepo repo.UserPasskeyRepo, userID uint) (int, error) {
	currentUser, err := userRepo.DetailByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if currentUser == nil {
		return 0, nil
	}

	loginMethodCount := 0
	if strings.TrimSpace(currentUser.Password) != "" {
		loginMethodCount++
	}

	if identityRepo != nil {
		identityList, err := identityRepo.ListByUserID(ctx, userID)
		if err != nil {
			return 0, err
		}
		loginMethodCount += len(identityList)
	}

	if passkeyRepo != nil {
		passkeyCount, err := passkeyRepo.CountByUserID(ctx, userID)
		if err != nil {
			return 0, err
		}
		loginMethodCount += int(passkeyCount)
	}

	return loginMethodCount, nil
}

func (a authService) ensurePasskeyConfig() error {
	if strings.TrimSpace(a.config.WebAuthn.RPID) == "" {
		return errors.New("admin.webAuthn.rp_id is empty")
	}
	if len(a.config.WebAuthn.RPOrigins) == 0 {
		return errors.New("admin.webAuthn.rp_origins is empty")
	}
	return nil
}

func (a authService) getWebAuthn() (*webauthn.WebAuthn, error) {
	timeout := time.Second * time.Duration(a.config.WebAuthn.ChallengeExpireIn)

	return webauthn.New(&webauthn.Config{
		RPID:          strings.TrimSpace(a.config.WebAuthn.RPID),
		RPDisplayName: firstNonEmpty(strings.TrimSpace(a.config.WebAuthn.RPDisplayName), "Dudu Admin"),
		RPOrigins:     a.config.WebAuthn.RPOrigins,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce: true,
				Timeout: timeout,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce: true,
				Timeout: timeout,
			},
		},
	})
}

func passkeySessionExpired(session *webauthn.SessionData) bool {
	if session == nil || session.Expires.IsZero() {
		return false
	}

	return time.Now().After(session.Expires)
}

func passkeyRegistrationOptions(userVerification string) []webauthn.RegistrationOption {
	return []webauthn.RegistrationOption{
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			UserVerification: parseUserVerificationRequirement(userVerification),
		}),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	}
}

func passkeyDiscoverableLoginOptions(userVerification string) []webauthn.LoginOption {
	return []webauthn.LoginOption{
		webauthn.WithUserVerification(parseUserVerificationRequirement(userVerification)),
	}
}

func findPasskeyByCredentialID(passkeys []systemModel.UserPasskey, credentialID string) *systemModel.UserPasskey {
	for i := range passkeys {
		if passkeys[i].CredentialID == credentialID {
			return &passkeys[i]
		}
	}

	return nil
}

func (a authService) resolvePasskeyLoginUser(ctx context.Context, rawID, userHandle []byte) (user *systemModel.User, passkey *systemModel.UserPasskey, waUser passkeyWebAuthnUser, errCode int, err error) {
	if len(rawID) == 0 {
		return nil, nil, waUser, e.PasskeyVerificationFailed, errors.New("empty passkey credential id")
	}

	passkey, err = a.passkeyRepo.DetailByCredentialID(ctx, encodeBytesToBase64URL(rawID))
	if err != nil {
		return nil, nil, waUser, e.ERROR, err
	}
	if passkey == nil {
		return nil, nil, waUser, e.PasskeyCredentialNotFound, nil
	}

	userID, err := decodePasskeyUserHandleBinary(userHandle)
	if err != nil {
		return nil, nil, waUser, e.PasskeyVerificationFailed, err
	}
	if passkey.UserID != userID {
		return nil, nil, waUser, e.PasskeyVerificationFailed, nil
	}

	user, err = a.userRepo.DetailByID(ctx, userID)
	if err != nil {
		return nil, nil, waUser, e.ERROR, err
	}
	if user == nil || user.Status != 1 {
		return nil, nil, waUser, e.UserNotFound, nil
	}

	passkeys, err := a.passkeyRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, waUser, e.ERROR, err
	}
	if len(passkeys) == 0 {
		return nil, nil, waUser, e.PasskeyCredentialNotFound, nil
	}

	waUser = buildPasskeyWebAuthnUser(user, passkeys, "")
	if len(waUser.credentials) == 0 {
		return nil, nil, waUser, e.PasskeyCredentialNotFound, nil
	}

	return user, passkey, waUser, e.SUCCESS, nil
}

func buildPasskeyWebAuthnUser(user *systemModel.User, passkeys []systemModel.UserPasskey, displayName string) passkeyWebAuthnUser {
	return passkeyWebAuthnUser{
		id:          encodePasskeyUserHandleBinary(user.ID),
		name:        passkeyUserName(user),
		displayName: firstNonEmpty(strings.TrimSpace(displayName), defaultPasskeyDisplayName(user)),
		credentials: parsePasskeyCredentials(passkeys),
	}
}

func parsePasskeyCredentials(passkeys []systemModel.UserPasskey) []webauthn.Credential {
	credentials := make([]webauthn.Credential, 0, len(passkeys))
	for _, passkey := range passkeys {
		credential, ok := parseStoredCredential(passkey)
		if !ok {
			continue
		}
		credentials = append(credentials, credential)
	}
	return credentials
}

func parseStoredCredential(passkey systemModel.UserPasskey) (webauthn.Credential, bool) {
	credential := webauthn.Credential{}

	if strings.TrimSpace(passkey.CredentialPublicKey) != "" {
		if err := json.Unmarshal([]byte(passkey.CredentialPublicKey), &credential); err == nil && len(credential.ID) > 0 {
			return credential, true
		}
	}

	credentialIDBytes, err := decodeBase64URL(passkey.CredentialID)
	if err != nil || len(credentialIDBytes) == 0 {
		return webauthn.Credential{}, false
	}

	publicKey := []byte{}
	if strings.TrimSpace(passkey.CredentialPublicKey) != "" {
		if decoded, decodeErr := decodeBase64URL(passkey.CredentialPublicKey); decodeErr == nil {
			publicKey = decoded
		}
	}
	if len(publicKey) == 0 {
		return webauthn.Credential{}, false
	}

	var transports []string
	if strings.TrimSpace(passkey.TransportsJSON) != "" {
		_ = json.Unmarshal([]byte(passkey.TransportsJSON), &transports)
	}

	credential = webauthn.Credential{
		ID:        credentialIDBytes,
		PublicKey: publicKey,
		Transport: stringToAuthenticatorTransports(transports),
		Authenticator: webauthn.Authenticator{
			SignCount: passkey.SignCount,
			AAGUID:    decodeAAGUID(passkey.AAGUID),
		},
	}

	return credential, true
}

func parsePasskeySessionData(data json.RawMessage) (*webauthn.SessionData, error) {
	if len(data) == 0 {
		return nil, errors.New("empty passkey session data")
	}

	session := &webauthn.SessionData{}
	if err := json.Unmarshal(data, session); err != nil {
		return nil, err
	}

	return session, nil
}

func marshalPasskeyCredentialPayload(credential PasskeyCredential) ([]byte, error) {
	payload := map[string]any{
		"id":   strings.TrimSpace(credential.ID),
		"type": firstNonEmpty(strings.TrimSpace(credential.Type), string(protocol.PublicKeyCredentialType)),
	}

	rawID := firstNonEmpty(strings.TrimSpace(credential.RawID), strings.TrimSpace(credential.ID))
	if rawID != "" {
		payload["rawId"] = rawID
	}

	response := normalizePasskeyResponseMap(credential.Response)
	if len(response) == 0 {
		return nil, errors.New("empty passkey credential response")
	}
	payload["response"] = response

	return json.Marshal(payload)
}

func normalizePasskeyResponseMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}

	result := make(map[string]any, len(input))
	for key, value := range input {
		normalizedKey := passkeyProtocolJSONKey(key)
		result[normalizedKey] = normalizePasskeyResponseValue(value)
	}

	return result
}

func normalizePasskeyResponseValue(value any) any {
	switch raw := value.(type) {
	case map[string]any:
		return normalizePasskeyResponseMap(raw)
	case []any:
		result := make([]any, 0, len(raw))
		for _, item := range raw {
			result = append(result, normalizePasskeyResponseValue(item))
		}
		return result
	default:
		return raw
	}
}

func passkeyProtocolJSONKey(key string) string {
	switch strings.TrimSpace(key) {
	case "raw_id":
		return "rawId"
	case "client_data_json":
		return "clientDataJSON"
	case "attestation_object":
		return "attestationObject"
	case "authenticator_data":
		return "authenticatorData"
	case "client_extension_results":
		return "clientExtensionResults"
	case "authenticator_attachment":
		return "authenticatorAttachment"
	case "public_key":
		return "publicKey"
	case "public_key_algorithm":
		return "publicKeyAlgorithm"
	case "user_handle":
		return "userHandle"
	default:
		return key
	}
}

func parseUserVerificationRequirement(value string) protocol.UserVerificationRequirement {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(protocol.VerificationRequired):
		return protocol.VerificationRequired
	case string(protocol.VerificationDiscouraged):
		return protocol.VerificationDiscouraged
	default:
		return protocol.VerificationPreferred
	}
}

func mapPasskeyItem(passkey systemModel.UserPasskey) PasskeyItem {
	transports := make([]string, 0)
	if strings.TrimSpace(passkey.TransportsJSON) != "" {
		_ = json.Unmarshal([]byte(passkey.TransportsJSON), &transports)
	}

	return PasskeyItem{
		ID:          passkey.ID,
		DisplayName: passkey.DisplayName,
		AAGUID:      passkey.AAGUID,
		Transports:  transports,
		LastUsedAt:  passkey.LastUsedAt,
		CreatedAt:   passkey.CreatedAt,
	}
}

func authenticatorTransportsToStrings(transports []protocol.AuthenticatorTransport) []string {
	result := make([]string, 0, len(transports))
	for _, transport := range transports {
		value := strings.TrimSpace(string(transport))
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func stringToAuthenticatorTransports(transports []string) []protocol.AuthenticatorTransport {
	result := make([]protocol.AuthenticatorTransport, 0, len(transports))
	for _, transport := range transports {
		value := strings.TrimSpace(transport)
		if value != "" {
			result = append(result, protocol.AuthenticatorTransport(value))
		}
	}
	return result
}

func encodeAAGUID(aaguid []byte) string {
	if len(aaguid) == 0 {
		return ""
	}
	return hex.EncodeToString(aaguid)
}

func decodeAAGUID(aaguid string) []byte {
	aaguid = strings.TrimSpace(aaguid)
	if aaguid == "" {
		return nil
	}

	data, err := hex.DecodeString(aaguid)
	if err == nil {
		return data
	}

	data, err = decodeBase64URL(aaguid)
	if err == nil {
		return data
	}

	return nil
}

func encodeBytesToBase64URL(value []byte) string {
	if len(value) == 0 {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(value)
}

func decodeBase64URL(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("base64 value is empty")
	}

	data, err := base64.RawURLEncoding.DecodeString(value)
	if err == nil {
		return data, nil
	}

	return base64.StdEncoding.DecodeString(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func encodePasskeyUserHandleBinary(userID uint) []byte {
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, uint64(userID))
	return buffer
}

func decodePasskeyUserHandleBinary(userHandle []byte) (uint, error) {
	if len(userHandle) != 8 {
		return 0, errors.New("invalid passkey user handle")
	}

	return uint(binary.BigEndian.Uint64(userHandle)), nil
}

func passkeyUserName(user *systemModel.User) string {
	if user == nil {
		return ""
	}
	return firstNonEmpty(strings.TrimSpace(user.Email), strings.TrimSpace(user.Phone), strings.TrimSpace(user.UserName))
}

func defaultPasskeyDisplayName(user *systemModel.User) string {
	name := firstNonEmpty(strings.TrimSpace(user.UserName), passkeyUserName(user))
	if name == "" {
		return "Passkey"
	}
	return name + " Passkey"
}

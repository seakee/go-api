package system

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/seakee/go-api/app/config"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"gorm.io/gorm"
)

func TestAuthService_DeletePasskey(t *testing.T) {
	t.Run("empty passkey id returns explicit error", func(t *testing.T) {
		svc := &authService{}

		errCode, err := svc.DeletePasskey(context.Background(), 1, 0, "")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.PasskeyCanNotBeNull {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.PasskeyCanNotBeNull)
		}
	})

	t.Run("missing reauth ticket returns explicit error", func(t *testing.T) {
		svc := &authService{}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11, "")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.ReauthTicketCanNotBeNull {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.ReauthTicketCanNotBeNull)
		}
	})

	t.Run("invalid reauth ticket returns explicit error", func(t *testing.T) {
		svc := &authService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				if code != "reauth-code" {
					t.Fatalf("parseReauthTicket() code = %s, want reauth-code", code)
				}
				return &reauthTicket{UserID: 7, Action: "invalid"}, nil
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					t.Fatalf("DetailByIDAndUserID() called unexpectedly")
					return nil, nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.InvalidReauthTicket {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.InvalidReauthTicket)
		}
	})

	t.Run("reject delete when passkey is last login method", func(t *testing.T) {
		deleteCalled := false
		consumed := false

		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				if code != "reauth-code" {
					t.Fatalf("parseReauthTicket() code = %s, want reauth-code", code)
				}
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			consumeReauthTicketFn: func(ctx context.Context, code string) error {
				consumed = true
				return nil
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.LastLoginMethodCannotBeRemoved {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.LastLoginMethodCannotBeRemoved)
		}
		if deleteCalled {
			t.Fatalf("DeletePasskey() should not call DeleteByIDAndUserID()")
		}
		if consumed {
			t.Fatalf("DeletePasskey() should not consume reauth ticket on failure")
		}
	})

	t.Run("delete passkey succeeds when other login method exists", func(t *testing.T) {
		deleteCalled := false
		consumed := false

		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1, Password: "hashed"}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				if code != "reauth-code" {
					t.Fatalf("parseReauthTicket() code = %s, want reauth-code", code)
				}
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			consumeReauthTicketFn: func(ctx context.Context, code string) error {
				if code != "reauth-code" {
					t.Fatalf("consumeReauthTicket() code = %s, want reauth-code", code)
				}
				consumed = true
				return nil
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if !deleteCalled {
			t.Fatalf("DeletePasskey() should call DeleteByIDAndUserID()")
		}
		if !consumed {
			t.Fatalf("DeletePasskey() should consume reauth ticket on success")
		}
	})

	t.Run("delete passkey succeeds when another passkey remains", func(t *testing.T) {
		deleteCalled := false
		consumed := false

		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 2, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				if code != "reauth-code" {
					t.Fatalf("parseReauthTicket() code = %s, want reauth-code", code)
				}
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			consumeReauthTicketFn: func(ctx context.Context, code string) error {
				consumed = true
				return nil
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if !deleteCalled {
			t.Fatalf("DeletePasskey() should call DeleteByIDAndUserID()")
		}
		if !consumed {
			t.Fatalf("DeletePasskey() should consume reauth ticket on success")
		}
	})

	t.Run("delete maps record not found to passkey not found", func(t *testing.T) {
		consumed := false
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1, Password: "hashed"}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					return gorm.ErrRecordNotFound
				},
			},
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				if code != "reauth-code" {
					t.Fatalf("parseReauthTicket() code = %s, want reauth-code", code)
				}
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			consumeReauthTicketFn: func(ctx context.Context, code string) error {
				consumed = true
				return nil
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.PasskeyCredentialNotFound {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.PasskeyCredentialNotFound)
		}
		if consumed {
			t.Fatalf("DeletePasskey() should not consume reauth ticket when delete fails")
		}
	})
}

func TestAuthService_ListPasskeys(t *testing.T) {
	t.Run("returns user not found when parent user is missing", func(t *testing.T) {
		listCalled := false
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return nil, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserPasskey, error) {
					listCalled = true
					return nil, nil
				},
			},
		}

		list, errCode, err := svc.ListPasskeys(context.Background(), 9)
		if err != nil {
			t.Fatalf("ListPasskeys() error = %v, want nil", err)
		}
		if errCode != e.UserNotFound {
			t.Fatalf("ListPasskeys() errCode = %d, want %d", errCode, e.UserNotFound)
		}
		if list != nil {
			t.Fatalf("ListPasskeys() list = %+v, want nil", list)
		}
		if listCalled {
			t.Fatalf("ListPasskeys() should not call ListByUserID() when user is missing")
		}
	})
}

func TestAuthService_loginMethodCount(t *testing.T) {
	t.Run("counts password identity and passkey methods", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Password: "hashed"}, nil
				},
			},
			identityRepo: &mockUserIdentityRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserIdentity, error) {
					return []systemModel.UserIdentity{{Model: gorm.Model{ID: 1}, UserID: userID}}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
			},
		}

		count, err := svc.loginMethodCount(context.Background(), 9)
		if err != nil {
			t.Fatalf("loginMethodCount() error = %v", err)
		}
		if count != 3 {
			t.Fatalf("loginMethodCount() count = %d, want 3", count)
		}
	})

	t.Run("counts each identity and passkey credential individually", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Password: "hashed"}, nil
				},
			},
			identityRepo: &mockUserIdentityRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserIdentity, error) {
					return []systemModel.UserIdentity{
						{Model: gorm.Model{ID: 1}, UserID: userID},
						{Model: gorm.Model{ID: 2}, UserID: userID},
					}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 2, nil
				},
			},
		}

		count, err := svc.loginMethodCount(context.Background(), 9)
		if err != nil {
			t.Fatalf("loginMethodCount() error = %v", err)
		}
		if count != 5 {
			t.Fatalf("loginMethodCount() count = %d, want 5", count)
		}
	})

	t.Run("returns zero when user does not exist", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return nil, nil
				},
			},
		}

		count, err := svc.loginMethodCount(context.Background(), 9)
		if err != nil {
			t.Fatalf("loginMethodCount() error = %v", err)
		}
		if count != 0 {
			t.Fatalf("loginMethodCount() count = %d, want 0", count)
		}
	})

	t.Run("returns identity repo error", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}}, nil
				},
			},
			identityRepo: &mockUserIdentityRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserIdentity, error) {
					return nil, errors.New("list identities failed")
				},
			},
		}

		_, err := svc.loginMethodCount(context.Background(), 9)
		if err == nil {
			t.Fatalf("loginMethodCount() error = nil, want non-nil")
		}
	})
}

func TestParseStoredCredential(t *testing.T) {
	t.Run("parses json credential first", func(t *testing.T) {
		stored := webauthn.Credential{
			ID:        []byte{1, 2, 3},
			PublicKey: []byte{4, 5, 6},
			Transport: []protocol.AuthenticatorTransport{protocol.AuthenticatorTransport("usb")},
			Authenticator: webauthn.Authenticator{
				SignCount: 8,
				AAGUID:    []byte{9, 10},
			},
		}
		raw, err := json.Marshal(stored)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		credential, ok := parseStoredCredential(systemModel.UserPasskey{
			CredentialPublicKey: string(raw),
		})
		if !ok {
			t.Fatalf("parseStoredCredential() ok = false, want true")
		}
		if string(credential.ID) != string(stored.ID) {
			t.Fatalf("parseStoredCredential() credential.ID = %v, want %v", credential.ID, stored.ID)
		}
		if string(credential.PublicKey) != string(stored.PublicKey) {
			t.Fatalf("parseStoredCredential() credential.PublicKey = %v, want %v", credential.PublicKey, stored.PublicKey)
		}
		if credential.Authenticator.SignCount != stored.Authenticator.SignCount {
			t.Fatalf("parseStoredCredential() credential.Authenticator.SignCount = %d, want %d",
				credential.Authenticator.SignCount, stored.Authenticator.SignCount)
		}
	})

	t.Run("falls back to legacy credential fields", func(t *testing.T) {
		credential, ok := parseStoredCredential(systemModel.UserPasskey{
			CredentialID:        encodeBytesToBase64URL([]byte{11, 12}),
			CredentialPublicKey: encodeBytesToBase64URL([]byte{13, 14}),
			SignCount:           10,
			AAGUID:              encodeAAGUID([]byte{15, 16}),
			TransportsJSON:      `["usb","internal"]`,
		})
		if !ok {
			t.Fatalf("parseStoredCredential() ok = false, want true")
		}
		if string(credential.ID) != string([]byte{11, 12}) {
			t.Fatalf("parseStoredCredential() credential.ID = %v, want %v", credential.ID, []byte{11, 12})
		}
		if string(credential.PublicKey) != string([]byte{13, 14}) {
			t.Fatalf("parseStoredCredential() credential.PublicKey = %v, want %v", credential.PublicKey, []byte{13, 14})
		}
		if credential.Authenticator.SignCount != 10 {
			t.Fatalf("parseStoredCredential() credential.Authenticator.SignCount = %d, want 10", credential.Authenticator.SignCount)
		}
		if string(credential.Authenticator.AAGUID) != string([]byte{15, 16}) {
			t.Fatalf("parseStoredCredential() credential.Authenticator.AAGUID = %v, want %v", credential.Authenticator.AAGUID, []byte{15, 16})
		}
		if len(credential.Transport) != 2 {
			t.Fatalf("parseStoredCredential() len(credential.Transport) = %d, want 2", len(credential.Transport))
		}
	})

	t.Run("returns false when credential id is invalid", func(t *testing.T) {
		_, ok := parseStoredCredential(systemModel.UserPasskey{
			CredentialID:        "invalid***",
			CredentialPublicKey: "",
		})
		if ok {
			t.Fatalf("parseStoredCredential() ok = true, want false")
		}
	})

	t.Run("returns false when legacy public key is invalid", func(t *testing.T) {
		_, ok := parseStoredCredential(systemModel.UserPasskey{
			CredentialID:        encodeBytesToBase64URL([]byte{1, 2}),
			CredentialPublicKey: "invalid***",
		})
		if ok {
			t.Fatalf("parseStoredCredential() ok = true, want false")
		}
	})
}

func TestMarshalPasskeyCredentialPayload(t *testing.T) {
	credential := PasskeyCredential{
		ID: "credential-id",
		Response: map[string]any{
			"client_data_json":   "client-data",
			"authenticator_data": "auth-data",
			"user_handle":        "user-handle",
			"nested": map[string]any{
				"public_key":           "pk",
				"public_key_algorithm": -7,
			},
			"list": []any{
				map[string]any{"attestation_object": "attestation"},
			},
		},
	}

	raw, err := marshalPasskeyCredentialPayload(credential)
	if err != nil {
		t.Fatalf("marshalPasskeyCredentialPayload() error = %v", err)
	}

	var payload map[string]any
	if err = json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload["type"] != string(protocol.PublicKeyCredentialType) {
		t.Fatalf("payload.type = %v, want %s", payload["type"], protocol.PublicKeyCredentialType)
	}
	if payload["rawId"] != "credential-id" {
		t.Fatalf("payload.rawId = %v, want credential-id", payload["rawId"])
	}

	response, ok := payload["response"].(map[string]any)
	if !ok {
		t.Fatalf("payload.response type = %T, want map[string]any", payload["response"])
	}
	if _, has := response["client_data_json"]; has {
		t.Fatalf("response should not keep snake_case key client_data_json")
	}
	if response["clientDataJSON"] != "client-data" {
		t.Fatalf("response.clientDataJSON = %v, want client-data", response["clientDataJSON"])
	}
	if response["authenticatorData"] != "auth-data" {
		t.Fatalf("response.authenticatorData = %v, want auth-data", response["authenticatorData"])
	}
	if response["userHandle"] != "user-handle" {
		t.Fatalf("response.userHandle = %v, want user-handle", response["userHandle"])
	}

	nested, ok := response["nested"].(map[string]any)
	if !ok {
		t.Fatalf("response.nested type = %T, want map[string]any", response["nested"])
	}
	if nested["publicKey"] != "pk" {
		t.Fatalf("response.nested.publicKey = %v, want pk", nested["publicKey"])
	}
	if nested["publicKeyAlgorithm"] != float64(-7) {
		t.Fatalf("response.nested.publicKeyAlgorithm = %v, want -7", nested["publicKeyAlgorithm"])
	}

	list, ok := response["list"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("response.list = %v, want one element array", response["list"])
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		t.Fatalf("response.list[0] type = %T, want map[string]any", list[0])
	}
	if item["attestationObject"] != "attestation" {
		t.Fatalf("response.list[0].attestationObject = %v, want attestation", item["attestationObject"])
	}
}

func TestMarshalPasskeyCredentialPayload_EmptyResponse(t *testing.T) {
	_, err := marshalPasskeyCredentialPayload(PasskeyCredential{
		ID:       "credential-id",
		Response: map[string]any{},
	})
	if err == nil {
		t.Fatalf("marshalPasskeyCredentialPayload() error = nil, want non-nil")
	}
}

func TestAuthService_getWebAuthn_UsesChallengeExpireTimeout(t *testing.T) {
	svc := &authService{
		config: config.AdminConfig{
			WebAuthn: config.WebAuthnConfig{
				RPID:              "localhost",
				RPDisplayName:     "Dudu Admin",
				RPOrigins:         []string{"http://localhost:3000"},
				ChallengeExpireIn: 180,
			},
		},
	}

	wa, err := svc.getWebAuthn()
	if err != nil {
		t.Fatalf("getWebAuthn() error = %v", err)
	}

	if !wa.Config.Timeouts.Login.Enforce {
		t.Fatalf("getWebAuthn() login timeout should be enforced")
	}
	if !wa.Config.Timeouts.Registration.Enforce {
		t.Fatalf("getWebAuthn() registration timeout should be enforced")
	}
	if wa.Config.Timeouts.Login.Timeout != 180*time.Second {
		t.Fatalf("getWebAuthn() login timeout = %s, want %s", wa.Config.Timeouts.Login.Timeout, 180*time.Second)
	}
	if wa.Config.Timeouts.Registration.Timeout != 180*time.Second {
		t.Fatalf("getWebAuthn() registration timeout = %s, want %s", wa.Config.Timeouts.Registration.Timeout, 180*time.Second)
	}
}

func TestPasskeyRegistrationOptions_RequireResidentKey(t *testing.T) {
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          "localhost",
		RPDisplayName: "Dudu Admin",
		RPOrigins:     []string{"http://localhost:3000"},
	})
	if err != nil {
		t.Fatalf("webauthn.New() error = %v", err)
	}

	creation, _, err := wa.BeginRegistration(passkeyWebAuthnUser{
		id:          []byte{1},
		name:        "admin@example.com",
		displayName: "Admin Passkey",
	}, passkeyRegistrationOptions("preferred")...)
	if err != nil {
		t.Fatalf("BeginRegistration() error = %v", err)
	}

	if creation.Response.AuthenticatorSelection.ResidentKey != protocol.ResidentKeyRequirementRequired {
		t.Fatalf("resident key requirement = %s, want %s", creation.Response.AuthenticatorSelection.ResidentKey, protocol.ResidentKeyRequirementRequired)
	}
	if creation.Response.AuthenticatorSelection.RequireResidentKey == nil {
		t.Fatalf("require resident key = nil, want non-nil")
	}
	if !*creation.Response.AuthenticatorSelection.RequireResidentKey {
		t.Fatalf("require resident key = false, want true")
	}
}

func TestPasskeyDiscoverableLoginOptions(t *testing.T) {
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          "localhost",
		RPDisplayName: "Dudu Admin",
		RPOrigins:     []string{"http://localhost:3000"},
	})
	if err != nil {
		t.Fatalf("webauthn.New() error = %v", err)
	}

	assertion, session, err := wa.BeginDiscoverableLogin(passkeyDiscoverableLoginOptions("required")...)
	if err != nil {
		t.Fatalf("BeginDiscoverableLogin() error = %v", err)
	}

	if assertion.Response.UserVerification != protocol.VerificationRequired {
		t.Fatalf("user verification = %s, want %s", assertion.Response.UserVerification, protocol.VerificationRequired)
	}
	if len(assertion.Response.AllowedCredentials) != 0 {
		t.Fatalf("allowed credentials len = %d, want 0", len(assertion.Response.AllowedCredentials))
	}
	if len(session.UserID) != 0 {
		t.Fatalf("session user id = %v, want empty", session.UserID)
	}
}

func TestResolvePasskeyLoginUser(t *testing.T) {
	storedCredential := webauthn.Credential{
		ID:        []byte{1, 2, 3},
		PublicKey: []byte{4, 5, 6},
		Authenticator: webauthn.Authenticator{
			SignCount: 9,
			AAGUID:    []byte{7, 8},
		},
	}
	storedRaw, err := json.Marshal(storedCredential)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	t.Run("resolves user and passkey from raw id and user handle", func(t *testing.T) {
		credentialID := encodeBytesToBase64URL(storedCredential.ID)
		passkeyRecord := &systemModel.UserPasskey{
			Model:               gorm.Model{ID: 11},
			UserID:              7,
			CredentialID:        credentialID,
			CredentialPublicKey: string(storedRaw),
			UserHandle:          encodeBytesToBase64URL(encodePasskeyUserHandleBinary(7)),
		}

		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					if id != 7 {
						t.Fatalf("DetailByID() id = %d, want 7", id)
					}
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1, Email: "admin@example.com"}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByCredentialIDFunc: func(ctx context.Context, credentialID string) (*systemModel.UserPasskey, error) {
					if credentialID != encodeBytesToBase64URL(storedCredential.ID) {
						t.Fatalf("DetailByCredentialID() credentialID = %s, want %s", credentialID, encodeBytesToBase64URL(storedCredential.ID))
					}
					return passkeyRecord, nil
				},
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserPasskey, error) {
					if userID != 7 {
						t.Fatalf("ListByUserID() userID = %d, want 7", userID)
					}
					return []systemModel.UserPasskey{*passkeyRecord}, nil
				},
			},
		}

		user, passkey, waUser, errCode, err := svc.resolvePasskeyLoginUser(context.Background(), storedCredential.ID, encodePasskeyUserHandleBinary(7))
		if err != nil {
			t.Fatalf("resolvePasskeyLoginUser() error = %v", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("resolvePasskeyLoginUser() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if user == nil || user.ID != 7 {
			t.Fatalf("resolvePasskeyLoginUser() user = %+v, want user id 7", user)
		}
		if passkey == nil || passkey.ID != 11 {
			t.Fatalf("resolvePasskeyLoginUser() passkey = %+v, want passkey id 11", passkey)
		}
		if len(waUser.credentials) != 1 {
			t.Fatalf("resolvePasskeyLoginUser() credentials len = %d, want 1", len(waUser.credentials))
		}
	})

	t.Run("rejects invalid user handle", func(t *testing.T) {
		svc := &authService{
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByCredentialIDFunc: func(ctx context.Context, credentialID string) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{
						Model:               gorm.Model{ID: 11},
						UserID:              7,
						CredentialID:        encodeBytesToBase64URL(storedCredential.ID),
						CredentialPublicKey: string(storedRaw),
					}, nil
				},
			},
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					t.Fatalf("DetailByID() called unexpectedly")
					return nil, nil
				},
			},
		}

		_, _, _, errCode, err := svc.resolvePasskeyLoginUser(context.Background(), storedCredential.ID, []byte{1, 2, 3})
		if err == nil {
			t.Fatalf("resolvePasskeyLoginUser() error = nil, want non-nil")
		}
		if errCode != e.PasskeyVerificationFailed {
			t.Fatalf("resolvePasskeyLoginUser() errCode = %d, want %d", errCode, e.PasskeyVerificationFailed)
		}
	})
}

func TestDecodePasskeyUserHandleBinary(t *testing.T) {
	userID, err := decodePasskeyUserHandleBinary(encodePasskeyUserHandleBinary(9))
	if err != nil {
		t.Fatalf("decodePasskeyUserHandleBinary() error = %v", err)
	}
	if userID != 9 {
		t.Fatalf("decodePasskeyUserHandleBinary() userID = %d, want 9", userID)
	}

	if _, err = decodePasskeyUserHandleBinary([]byte{1, 2, 3}); err == nil {
		t.Fatalf("decodePasskeyUserHandleBinary() error = nil, want non-nil")
	}
}

func TestPasskeySessionExpired(t *testing.T) {
	if passkeySessionExpired(&webauthn.SessionData{}) {
		t.Fatalf("passkeySessionExpired() = true for zero expires, want false")
	}

	if passkeySessionExpired(&webauthn.SessionData{Expires: time.Now().Add(time.Minute)}) {
		t.Fatalf("passkeySessionExpired() = true for future expires, want false")
	}

	if !passkeySessionExpired(&webauthn.SessionData{Expires: time.Now().Add(-time.Minute)}) {
		t.Fatalf("passkeySessionExpired() = false for past expires, want true")
	}
}

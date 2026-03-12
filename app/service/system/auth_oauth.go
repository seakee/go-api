package system

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/seakee/go-api/app/pkg/e"
	pwd "github.com/seakee/go-api/app/pkg/password"
	repo "github.com/seakee/go-api/app/repository/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// tokenByOAuth completes the OAuth login flow and returns either an access token or a bind ticket.
func (a authService) tokenByOAuth(ctx context.Context, provider, code, state string) (token AccessToken, errCode int, err error) {
	if errCode = a.validateOAuthState(ctx, provider, state); errCode != e.SUCCESS {
		return token, errCode, nil
	}

	providerIdentity, err := a.fetchOAuthIdentity(ctx, provider, code)
	if err != nil {
		return token, e.ERROR, err
	}

	user, err := a.userRepo.GetOAuthUser(ctx, providerIdentity.Provider, providerIdentity.ProviderTenant, providerIdentity.ProviderSubject)
	if err != nil {
		return token, e.ERROR, err
	}
	if errCode = oauthLoginStatusErrCode(user); errCode != e.SUCCESS {
		return token, errCode, nil
	}

	if user == nil {
		token.BindTicket, err = a.generateBindTicket(ctx, bindTicket{
			Provider:        providerIdentity.Provider,
			ProviderTenant:  providerIdentity.ProviderTenant,
			ProviderSubject: providerIdentity.ProviderSubject,
			OAuthProfile:    providerIdentity.Profile,
		})
		if err != nil {
			return token, e.ERROR, err
		}

		token.OAuthProfile = providerIdentity.Profile
		token.SyncableFields = syncableFieldsForProfile(providerIdentity.Profile)
		return token, e.NeedBindOAuth, nil
	}

	accessTokenString, err := a.generateToken(user.UserName, user.ID)
	if err != nil {
		return token, e.ERROR, err
	}

	token.Token = accessTokenString
	token.ExpireIn = a.config.TokenExpireIn

	if a.identityRepo != nil {
		boundIdentity, detailErr := a.identityRepo.DetailByProvider(ctx, providerIdentity.Provider, providerIdentity.ProviderTenant, providerIdentity.ProviderSubject)
		if detailErr == nil && boundIdentity != nil {
			if updateErr := a.identityRepo.UpdateLastLogin(ctx, boundIdentity.ID, time.Now()); updateErr != nil && a.logger != nil {
				a.logger.Error(ctx, "update oauth identity last login failed", zap.Error(updateErr))
			}
		}
	}

	return token, e.SUCCESS, nil
}

// oauthProviderIdentity describes the verified third-party identity returned by the OAuth provider.
type oauthProviderIdentity struct {
	Provider        string
	ProviderTenant  string
	ProviderSubject string
	Profile         *OAuthProfile
}

// fetchOAuthIdentity exchanges the provider callback code for a normalized third-party identity.
func (a authService) fetchOAuthIdentity(ctx context.Context, provider, code string) (*oauthProviderIdentity, error) {
	switch strings.TrimSpace(provider) {
	case "feishu":
		accessToken, err := a.getFeishuUserAccessToken(ctx, code)
		if err != nil {
			return nil, err
		}

		unionID, profile, err := a.getFeishuUserProfile(ctx, accessToken)
		if err != nil {
			return nil, err
		}
		unionID = strings.TrimSpace(unionID)
		if unionID == "" {
			return nil, errors.New("empty feishu union id")
		}

		return &oauthProviderIdentity{
			Provider:        "feishu",
			ProviderTenant:  strings.TrimSpace(a.config.Oauth.Feishu.ClientID),
			ProviderSubject: unionID,
			Profile:         profile,
		}, nil
	case "wechat":
		accessToken, err := a.getWechatAccessToken(ctx)
		if err != nil {
			return nil, err
		}

		wechatUserID, err := a.getWechatUserID(ctx, accessToken, code)
		if err != nil {
			return nil, err
		}
		wechatUserID = strings.TrimSpace(wechatUserID)
		if wechatUserID == "" {
			return nil, errors.New("empty wechat userid")
		}

		profile, detailErr := a.getWechatUserProfile(ctx, accessToken, wechatUserID)
		if detailErr != nil && a.logger != nil {
			a.logger.Error(ctx, "failed to fetch wechat profile for oauth identity", zap.String("wechat_userid", wechatUserID), zap.Error(detailErr))
		}

		return &oauthProviderIdentity{
			Provider:        "wechat",
			ProviderTenant:  strings.TrimSpace(a.config.Oauth.Wechat.CorpID),
			ProviderSubject: wechatUserID,
			Profile:         profile,
		}, nil
	default:
		return nil, fmt.Errorf("%s type not support", provider)
	}
}

// validateOAuthState validates and consumes the short-lived OAuth state value.
func (a authService) validateOAuthState(ctx context.Context, provider, state string) int {
	oauthType, err := a.redis.GetStringWithContext(ctx, fmt.Sprintf(oauthStateKey, state))
	if err != nil || oauthType != provider {
		return e.InvalidOauthState
	}

	if _, err = a.redis.DelWithContext(ctx, fmt.Sprintf(oauthStateKey, state)); err != nil && a.logger != nil {
		a.logger.Error(ctx, "failed to delete oauth authorization state", zap.String("provider", provider), zap.Error(err))
	}

	return e.SUCCESS
}

// Reauth verifies local credentials before high-risk OAuth bind or unbind operations.
func (a authService) Reauth(ctx context.Context, identifier, password, challengeCode, totpCode string) (result ReauthResult, errCode int, err error) {
	challengeCode = strings.TrimSpace(challengeCode)
	totpCode = strings.TrimSpace(totpCode)
	if challengeCode != "" {
		if totpCode == "" {
			return result, e.TotpCodeCanNotBeNull, nil
		}

		var sc *safeCode
		if a.parseSafeCodeFn != nil {
			sc, err = a.parseSafeCodeFn(ctx, challengeCode)
		} else {
			sc, err = a.parseSafeCode(ctx, challengeCode)
		}
		if err != nil {
			return result, e.InvalidSafeCode, err
		}
		if sc == nil || sc.Action != "oauth_reauth" || sc.UserID == 0 {
			return result, e.InvalidSafeCode, nil
		}

		user, detailErr := a.userRepo.DetailByID(ctx, sc.UserID)
		if detailErr != nil {
			return result, e.ERROR, detailErr
		}
		if user == nil || user.Status != 1 {
			return result, e.UserNotFound, nil
		}
		if !user.TotpEnabled || !a.totp.VerifyTOTPCode(totpCode, user.TotpKey, 1) {
			return result, e.InvalidTotpCode, nil
		}

		result.ReauthTicket, err = a.generateReauthTicket(ctx, reauthTicket{
			UserID: user.ID,
			Action: "oauth_reauth",
		})
		if err != nil {
			return result, e.ERROR, err
		}

		return result, e.SUCCESS, nil
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" || password == "" {
		return result, e.IdentifierOrPasswordCanNotBeNull, nil
	}

	user, err := a.userRepo.DetailByIdentifier(ctx, identifier)
	if err != nil {
		return result, e.ERROR, err
	}
	if user == nil || user.Status != 1 {
		return result, e.UserNotFound, nil
	}

	matched, err := pwd.VerifyCredential(user.Password, password)
	if err != nil {
		return result, e.ERROR, err
	}
	if !matched {
		return result, e.IdentifierOrPasswordFail, nil
	}

	if user.TotpEnabled {
		result.SafeCode, err = a.generateSafeCode(ctx, safeCode{
			UserID: user.ID,
			Action: "oauth_reauth",
		})
		if err != nil {
			return result, e.ERROR, err
		}
		return result, e.NeedTfa, nil
	}

	result.ReauthTicket, err = a.generateReauthTicket(ctx, reauthTicket{
		UserID: user.ID,
		Action: "oauth_reauth",
	})
	if err != nil {
		return result, e.ERROR, err
	}

	return result, e.SUCCESS, nil
}

// ConfirmOAuthBind binds a verified third-party identity to the reauthenticated local account and completes login.
func (a authService) ConfirmOAuthBind(ctx context.Context, bindTicketCode, reauthTicketCode string, syncFields []string) (token AccessToken, errCode int, err error) {
	if strings.TrimSpace(bindTicketCode) == "" {
		return token, e.BindTicketCanNotBeNull, nil
	}

	if strings.TrimSpace(reauthTicketCode) == "" {
		return token, e.ReauthTicketCanNotBeNull, nil
	}

	bindTicket, err := a.parseBindTicket(ctx, bindTicketCode)
	if err != nil || bindTicket == nil {
		return token, e.InvalidBindTicket, err
	}

	reauthTicket, err := a.parseReauthTicket(ctx, reauthTicketCode)
	if err != nil || reauthTicket == nil || reauthTicket.Action != "oauth_reauth" || reauthTicket.UserID == 0 {
		return token, e.InvalidReauthTicket, err
	}

	syncUserName, syncAvatar, err := buildOAuthProfileSync(bindTicket.OAuthProfile, syncFields)
	if err != nil {
		return token, e.ERROR, err
	}

	err = a.userRepo.BindOAuthIdentity(ctx, repo.OAuthBindInput{
		UserID:          reauthTicket.UserID,
		Provider:        bindTicket.Provider,
		ProviderTenant:  bindTicket.ProviderTenant,
		ProviderSubject: bindTicket.ProviderSubject,
		DisplayName:     profileUserName(bindTicket.OAuthProfile),
		AvatarURL:       profileAvatar(bindTicket.OAuthProfile),
		RawProfile:      bindTicket.OAuthProfile,
		SyncUserName:    syncUserName,
		SyncAvatar:      syncAvatar,
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrOAuthIdentityConflict):
			return token, e.IdentifierConflict, nil
		case errors.Is(err, repo.ErrOAuthBindUserNotFound):
			return token, e.UserNotFound, nil
		default:
			return token, e.ERROR, err
		}
	}

	boundUser, detailErr := a.userRepo.DetailByID(ctx, reauthTicket.UserID)
	if detailErr != nil {
		return token, e.ERROR, detailErr
	}
	if boundUser == nil || boundUser.Status != 1 {
		return token, e.UserNotFound, nil
	}

	tokenString, tokenErr := a.generateToken(boundUser.UserName, boundUser.ID)
	if tokenErr != nil {
		return token, e.ERROR, tokenErr
	}

	token.Token = tokenString
	token.ExpireIn = a.config.TokenExpireIn

	if err = a.consumeBindTicket(ctx, bindTicketCode); err != nil && a.logger != nil {
		a.logger.Error(ctx, "failed to consume bind ticket", zap.Error(err))
	}

	if err = a.consumeReauthTicket(ctx, reauthTicketCode); err != nil && a.logger != nil {
		a.logger.Error(ctx, "failed to consume reauth ticket", zap.Error(err))
	}

	return token, e.SUCCESS, nil
}

// OAuthAccounts returns the third-party identities bound to the current user.
func (a authService) OAuthAccounts(ctx context.Context, userID uint) (accounts []OAuthAccount, errCode int, err error) {
	identityList, err := a.identityRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, e.ERROR, err
	}

	accounts = make([]OAuthAccount, 0, len(identityList))
	for _, userIdentity := range identityList {
		accounts = append(accounts, OAuthAccount{
			ID:             userIdentity.ID,
			Provider:       userIdentity.Provider,
			ProviderTenant: userIdentity.ProviderTenant,
			DisplayName:    userIdentity.DisplayName,
			AvatarURL:      userIdentity.AvatarURL,
			BoundAt:        userIdentity.BoundAt,
			LastLoginAt:    userIdentity.LastLoginAt,
		})
	}

	return accounts, e.SUCCESS, nil
}

// UnbindOAuth removes a specific third-party identity after reauthentication succeeds.
func (a authService) UnbindOAuth(ctx context.Context, userID, identityID uint, reauthTicketCode string) (errCode int, err error) {
	if identityID == 0 {
		return e.InvalidParams, nil
	}
	if strings.TrimSpace(reauthTicketCode) == "" {
		return e.ReauthTicketCanNotBeNull, nil
	}

	reauthTicket, err := a.parseReauthTicket(ctx, reauthTicketCode)
	if err != nil || reauthTicket == nil || reauthTicket.Action != "oauth_reauth" || reauthTicket.UserID != userID {
		return e.InvalidReauthTicket, err
	}

	boundIdentity, err := a.identityRepo.DetailByIDAndUserID(ctx, identityID, userID)
	if err != nil {
		return e.ERROR, err
	}
	if boundIdentity == nil {
		return e.OauthAccountNotBound, nil
	}

	identityList, err := a.identityRepo.ListByUserID(ctx, userID)
	if err != nil {
		return e.ERROR, err
	}

	currentUser, err := a.userRepo.DetailByID(ctx, userID)
	if err != nil {
		return e.ERROR, err
	}
	if currentUser == nil {
		return e.UserNotFound, nil
	}

	loginMethodCount := len(identityList)
	if strings.TrimSpace(currentUser.Password) != "" {
		loginMethodCount++
	}
	if loginMethodCount <= 1 {
		return e.LastLoginMethodCannotBeRemoved, nil
	}

	if err = a.identityRepo.DeleteByIDAndUserID(ctx, identityID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return e.OauthAccountNotBound, nil
		}
		return e.ERROR, err
	}

	if err = a.consumeReauthTicket(ctx, reauthTicketCode); err != nil && a.logger != nil {
		a.logger.Error(ctx, "failed to consume reauth ticket after unbind", zap.Error(err))
	}

	return e.SUCCESS, nil
}

func buildOAuthProfileSync(profile *OAuthProfile, syncFields []string) (userName, avatar string, err error) {
	if len(syncFields) == 0 || profile == nil {
		return "", "", nil
	}

	for _, field := range syncFields {
		switch strings.TrimSpace(field) {
		case "user_name":
			userName = strings.TrimSpace(profile.UserName)
		case "avatar":
			avatar = strings.TrimSpace(profile.Avatar)
		case "":
			continue
		default:
			return "", "", fmt.Errorf("unsupported oauth sync field: %s", field)
		}
	}

	return userName, avatar, nil
}

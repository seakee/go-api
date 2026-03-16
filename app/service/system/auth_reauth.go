package system

import (
	"context"
	"strings"

	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	pwd "github.com/seakee/go-api/app/pkg/password"
)

const (
	reauthMethodPasskey  = "passkey"
	reauthMethodPassword = "password"
)

// ReauthMethodsResult describes the available verification methods for sensitive operations.
type ReauthMethodsResult struct {
	DefaultMethod        string   `json:"default_method"`
	AvailableMethods     []string `json:"available_methods"`
	PasswordRequiresTotp bool     `json:"password_requires_totp"`
	TotpEnabled          bool     `json:"totp_enabled"`
	PasskeyCount         int64    `json:"passkey_count"`
}

func (a authService) ReauthMethods(ctx context.Context, userID uint) (result ReauthMethodsResult, errCode int, err error) {
	user, errCode, err := a.activeUserByID(ctx, userID)
	if errCode != e.SUCCESS {
		return result, errCode, err
	}

	passkeyCount, err := a.passkeyRepo.CountByUserID(ctx, userID)
	if err != nil {
		return result, e.ERROR, err
	}

	methods := make([]string, 0, 2)
	defaultMethod := ""
	if passkeyCount > 0 {
		methods = append(methods, reauthMethodPasskey)
		defaultMethod = reauthMethodPasskey
	}
	if strings.TrimSpace(user.Password) != "" {
		methods = append(methods, reauthMethodPassword)
		if defaultMethod == "" {
			defaultMethod = reauthMethodPassword
		}
	}

	result = ReauthMethodsResult{
		DefaultMethod:        defaultMethod,
		AvailableMethods:     methods,
		PasswordRequiresTotp: user.TotpEnabled,
		TotpEnabled:          user.TotpEnabled,
		PasskeyCount:         passkeyCount,
	}

	return result, e.SUCCESS, nil
}

func (a authService) ReauthByPassword(ctx context.Context, userID uint, password string) (result ReauthResult, errCode int, err error) {
	if strings.TrimSpace(password) == "" {
		return result, e.IdentifierOrPasswordCanNotBeNull, nil
	}

	user, errCode, err := a.activeUserByID(ctx, userID)
	if errCode != e.SUCCESS {
		return result, errCode, err
	}
	if strings.TrimSpace(user.Password) == "" {
		return result, e.IdentifierOrPasswordFail, nil
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
			Action: reauthActionHighRisk,
		})
		if err != nil {
			return result, e.ERROR, err
		}

		return result, e.NeedTfa, nil
	}

	return a.issueHighRiskReauthTicket(ctx, user.ID)
}

func (a authService) ReauthByTotp(ctx context.Context, userID uint, safeCodeCode, totpCode string) (result ReauthResult, errCode int, err error) {
	if strings.TrimSpace(totpCode) == "" {
		return result, e.TotpCodeCanNotBeNull, nil
	}

	sc, errCode, err := a.parseHighRiskSafeCode(ctx, safeCodeCode, userID)
	if errCode != e.SUCCESS {
		return result, errCode, err
	}

	user, errCode, err := a.activeUserByID(ctx, sc.UserID)
	if errCode != e.SUCCESS {
		return result, errCode, err
	}
	if !user.TotpEnabled || !a.totp.VerifyTOTPCode(totpCode, user.TotpKey, 1) {
		return result, e.InvalidTotpCode, nil
	}

	return a.issueHighRiskReauthTicket(ctx, user.ID)
}

func (a authService) issueHighRiskReauthTicket(ctx context.Context, userID uint) (result ReauthResult, errCode int, err error) {
	result.ReauthTicket, err = a.generateReauthTicket(ctx, reauthTicket{
		UserID: userID,
		Action: reauthActionHighRisk,
	})
	if err != nil {
		return result, e.ERROR, err
	}

	return result, e.SUCCESS, nil
}

func (a authService) activeUserByID(ctx context.Context, userID uint) (user *systemModel.User, errCode int, err error) {
	if userID == 0 {
		return nil, e.UserNotFound, nil
	}

	user, err = a.userRepo.DetailByID(ctx, userID)
	if err != nil {
		return nil, e.ERROR, err
	}
	if user == nil || user.Status != 1 {
		return nil, e.UserNotFound, nil
	}

	return user, e.SUCCESS, nil
}

func (a authService) parseHighRiskSafeCode(ctx context.Context, code string, userID uint) (sc *safeCode, errCode int, err error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, e.SafeCodeCanNotBeNull, nil
	}

	if a.parseSafeCodeFn != nil {
		sc, err = a.parseSafeCodeFn(ctx, code)
	} else {
		sc, err = a.parseSafeCode(ctx, code)
	}
	if err != nil {
		return nil, e.InvalidSafeCode, err
	}
	if sc == nil || sc.Action != reauthActionHighRisk || sc.UserID == 0 {
		return nil, e.InvalidSafeCode, nil
	}
	if userID != 0 && sc.UserID != userID {
		return nil, e.InvalidSafeCode, nil
	}

	return sc, e.SUCCESS, nil
}

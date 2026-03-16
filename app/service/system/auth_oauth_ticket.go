package system

import (
	"context"
	"strings"

	"github.com/seakee/go-api/app/pkg/e"
	"github.com/sk-pkg/util"
)

const reauthActionHighRisk = "high_risk_reauth"

type bindTicket struct {
	Provider        string        `json:"provider"`
	ProviderTenant  string        `json:"provider_tenant"`
	ProviderSubject string        `json:"provider_subject"`
	OAuthProfile    *OAuthProfile `json:"oauth_profile,omitempty"`
}

type reauthTicket struct {
	UserID uint   `json:"user_id"`
	Action string `json:"action"`
}

func (a authService) generateBindTicket(ctx context.Context, ticket bindTicket) (string, error) {
	if a.generateBindTicketFn != nil {
		return a.generateBindTicketFn(ctx, ticket)
	}

	code := util.RandUpStr(32)
	key := bindTicketPrefix + code
	if err := a.redis.SetJSONWithContext(ctx, key, ticket, a.config.SafeCodeExpireIn); err != nil {
		return "", err
	}

	return code, nil
}

func (a authService) parseBindTicket(ctx context.Context, code string) (*bindTicket, error) {
	if a.parseBindTicketFn != nil {
		return a.parseBindTicketFn(ctx, code)
	}

	var ticket *bindTicket
	if err := a.redis.GetJSONWithContext(ctx, bindTicketPrefix+code, &ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (a authService) consumeBindTicket(ctx context.Context, code string) error {
	if a.consumeBindTicketFn != nil {
		return a.consumeBindTicketFn(ctx, code)
	}

	_, err := a.redis.DelWithContext(ctx, bindTicketPrefix+code)
	return err
}

func (a authService) generateReauthTicket(ctx context.Context, ticket reauthTicket) (string, error) {
	if a.generateReauthTicketFn != nil {
		return a.generateReauthTicketFn(ctx, ticket)
	}

	code := util.RandUpStr(32)
	key := reauthTicketPrefix + code
	if err := a.redis.SetJSONWithContext(ctx, key, ticket, a.config.SafeCodeExpireIn); err != nil {
		return "", err
	}

	return code, nil
}

func (a authService) parseReauthTicket(ctx context.Context, code string) (*reauthTicket, error) {
	if a.parseReauthTicketFn != nil {
		return a.parseReauthTicketFn(ctx, code)
	}

	var ticket *reauthTicket
	if err := a.redis.GetJSONWithContext(ctx, reauthTicketPrefix+code, &ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (a authService) consumeReauthTicket(ctx context.Context, code string) error {
	if a.consumeReauthTicketFn != nil {
		return a.consumeReauthTicketFn(ctx, code)
	}

	_, err := a.redis.DelWithContext(ctx, reauthTicketPrefix+code)
	return err
}

func (a authService) validateReauthTicket(ctx context.Context, code string, userID uint) (*reauthTicket, int, error) {
	if strings.TrimSpace(code) == "" {
		return nil, e.ReauthTicketCanNotBeNull, nil
	}

	ticket, err := a.parseReauthTicket(ctx, code)
	if err != nil || ticket == nil || ticket.Action != reauthActionHighRisk || ticket.UserID == 0 {
		return nil, e.InvalidReauthTicket, err
	}
	if userID != 0 && ticket.UserID != userID {
		return nil, e.InvalidReauthTicket, nil
	}

	return ticket, e.SUCCESS, nil
}

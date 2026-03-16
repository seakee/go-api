package system

import (
	"context"
	"strings"

	"github.com/seakee/go-api/app/pkg/e"
)

func (s userService) parseReauthTicket(ctx context.Context, code string) (*reauthTicket, error) {
	if s.parseReauthTicketFn != nil {
		return s.parseReauthTicketFn(ctx, code)
	}
	if s.redis == nil {
		return nil, nil
	}

	var ticket *reauthTicket
	if err := s.redis.GetJSONWithContext(ctx, reauthTicketPrefix+code, &ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s userService) consumeReauthTicket(ctx context.Context, code string) error {
	if s.consumeReauthTicketFn != nil {
		return s.consumeReauthTicketFn(ctx, code)
	}
	if s.redis == nil {
		return nil
	}

	_, err := s.redis.DelWithContext(ctx, reauthTicketPrefix+code)
	return err
}

func (s userService) validateReauthTicket(ctx context.Context, code string, operatorUserID uint) (*reauthTicket, int, error) {
	if strings.TrimSpace(code) == "" {
		return nil, e.ReauthTicketCanNotBeNull, nil
	}

	ticket, err := s.parseReauthTicket(ctx, code)
	if err != nil || ticket == nil || ticket.Action != reauthActionHighRisk || ticket.UserID == 0 {
		return nil, e.InvalidReauthTicket, err
	}
	if operatorUserID != 0 && ticket.UserID != operatorUserID {
		return nil, e.InvalidReauthTicket, nil
	}

	return ticket, e.SUCCESS, nil
}

package system

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	redigo "github.com/gomodule/redigo/redis"
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

const takeReauthTicketLuaScript = `
local value = redis.call("GET", KEYS[1])
if not value then
	return nil
end

local ok, data = pcall(cjson.decode, value)
if not ok or not data then
	return nil
end

if data["action"] ~= ARGV[1] then
	return nil
end

local expected_user_id = tonumber(ARGV[2])
local ticket_user_id = tonumber(data["user_id"] or 0)
if not ticket_user_id or ticket_user_id == 0 then
	return nil
end

if expected_user_id ~= 0 and ticket_user_id ~= expected_user_id then
	return nil
end

redis.call("DEL", KEYS[1])
return value
`

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

func (a authService) takeReauthTicket(ctx context.Context, code string, userID uint, action string) (*reauthTicket, error) {
	if a.takeReauthTicketFn != nil {
		return a.takeReauthTicketFn(ctx, code, userID, action)
	}
	if a.redis == nil {
		return nil, redigo.ErrNil
	}

	reply, err := a.redis.LuaWithContext(
		ctx,
		1,
		takeReauthTicketLuaScript,
		[]string{
			a.redis.Prefix + reauthTicketPrefix + code,
			action,
			strconv.FormatUint(uint64(userID), 10),
		},
	)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, redigo.ErrNil
	}

	raw, err := redigo.Bytes(reply, nil)
	if err != nil {
		return nil, err
	}

	var ticket reauthTicket
	if err = json.Unmarshal(raw, &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (a authService) consumeValidatedReauthTicket(ctx context.Context, code string, userID uint) (*reauthTicket, int, error) {
	if strings.TrimSpace(code) == "" {
		return nil, e.ReauthTicketCanNotBeNull, nil
	}

	ticket, err := a.takeReauthTicket(ctx, code, userID, reauthActionHighRisk)
	if err != nil {
		if errors.Is(err, redigo.ErrNil) {
			return nil, e.InvalidReauthTicket, nil
		}
		return nil, e.InvalidReauthTicket, err
	}
	if ticket == nil || ticket.Action != reauthActionHighRisk || ticket.UserID == 0 {
		return nil, e.InvalidReauthTicket, nil
	}

	return ticket, e.SUCCESS, nil
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

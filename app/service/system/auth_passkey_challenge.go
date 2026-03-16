package system

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sk-pkg/util"
)

type passkeyChallenge struct {
	Action      string          `json:"action"`
	UserID      uint            `json:"user_id,omitempty"`
	ChallengeID string          `json:"challenge_id"`
	SessionData json.RawMessage `json:"session_data"`
	DisplayName string          `json:"display_name,omitempty"`
	CreatedAt   int64           `json:"created_at"`
}

func (a authService) generatePasskeyChallenge(ctx context.Context, challenge passkeyChallenge) (string, error) {
	if challenge.ChallengeID == "" {
		challenge.ChallengeID = util.RandUpStr(32)
	}
	if challenge.CreatedAt == 0 {
		challenge.CreatedAt = time.Now().Unix()
	}

	err := a.redis.SetJSONWithContext(
		ctx,
		util.SpliceStr(passkeyChallengePrefix, challenge.ChallengeID),
		challenge,
		a.config.WebAuthn.ChallengeExpireIn,
	)
	if err != nil {
		return "", err
	}

	return challenge.ChallengeID, nil
}

func (a authService) parsePasskeyChallenge(ctx context.Context, challengeID string) (*passkeyChallenge, error) {
	var challenge *passkeyChallenge
	err := a.redis.GetJSONWithContext(ctx, util.SpliceStr(passkeyChallengePrefix, challengeID), &challenge)
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

func (a authService) consumePasskeyChallenge(ctx context.Context, challengeID string) error {
	_, err := a.redis.DelWithContext(ctx, util.SpliceStr(passkeyChallengePrefix, challengeID))
	return err
}

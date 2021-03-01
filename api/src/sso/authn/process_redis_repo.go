package authn

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mredis"
)

// ProcessRedisRepo ...
type ProcessRedisRepo struct {
	mredis.SimpleKeyRedis
}

// NewAuthnProcessRedis ...
func NewAuthnProcessRedis(skr mredis.SimpleKeyRedis) ProcessRedisRepo {
	return ProcessRedisRepo{skr}
}

func (prr ProcessRedisRepo) key(loginChallenge string, tok string) string {
	return "authn_process:" + loginChallenge + ":" + tok
}

// Create ...
func (prr ProcessRedisRepo) Create(ctx context.Context, process *Process) error {
	value, err := json.Marshal(process)
	if err != nil {
		return merr.From(err).Desc("marshaling process")
	}
	// we keep in storage one hour the authentication process
	key := prr.key(process.LoginChallenge, process.AccessToken)
	return prr.SimpleKeyRedis.Set(ctx, key, value, time.Hour)
}

// Update ...
func (prr ProcessRedisRepo) Update(ctx context.Context, process Process) error {
	value, err := json.Marshal(process)
	if err != nil {
		return merr.From(err).Desc("marshaling process")
	}
	key := prr.key(process.LoginChallenge, process.AccessToken)
	lifetime := time.Until(time.Unix(process.ExpiresAt, 0))
	if err := prr.SimpleKeyRedis.Set(ctx, key, value, lifetime); err != nil {
		return merr.From(err).Desc("setting keep ttl")
	}
	return nil
}

// Get ...
func (prr ProcessRedisRepo) Get(ctx context.Context, loginChallenge string) (Process, error) {
	process := Process{}
	challengeKey := prr.key(loginChallenge, "*")
	values, err := prr.SimpleKeyRedis.MustFind(ctx, challengeKey)
	if err != nil {
		return process, err
	}
	value := values[0]
	if err := json.Unmarshal(value, &process); err != nil {
		return process, merr.From(err).Desc("unmarshaling authn process")
	}
	return process, nil
}

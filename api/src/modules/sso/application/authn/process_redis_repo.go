package authn

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type ProcessRedisRepo struct {
	repositories.SimpleKeyRedis
}

func NewAuthnProcessRedis(redConn *redis.Client) ProcessRedisRepo {
	return ProcessRedisRepo{repositories.NewSimpleKeyRedis(redConn)}
}

func (prr ProcessRedisRepo) key(loginChallenge string, tok string) string {
	return "authn_process:" + loginChallenge + ":" + tok
}

func (prr ProcessRedisRepo) Create(ctx context.Context, process *Process) error {
	value, err := json.Marshal(process)
	if err != nil {
		return merror.Transform(err).Describe("marshaling process")
	}
	// we keep in storage one hour the authentication process
	key := prr.key(process.LoginChallenge, process.AccessToken)
	return prr.SimpleKeyRedis.Set(ctx, key, value, time.Hour)
}

func (prr ProcessRedisRepo) Update(ctx context.Context, process Process) error {
	value, err := json.Marshal(process)
	if err != nil {
		return merror.Transform(err).Describe("marshaling process")
	}
	key := prr.key(process.LoginChallenge, process.AccessToken)
	lifetime := time.Until(time.Unix(process.ExpiresAt, 0))
	if err := prr.SimpleKeyRedis.Set(ctx, key, value, lifetime); err != nil {
		return merror.Transform(err).Describe("setting keep ttl")
	}
	return nil
}

func (prr ProcessRedisRepo) Get(ctx context.Context, loginChallenge string) (Process, error) {
	process := Process{}
	challengeKey := prr.key(loginChallenge, "*")
	values, err := prr.SimpleKeyRedis.MustFind(ctx, challengeKey)
	if err != nil {
		return process, err
	}
	value := values[0]
	if err := json.Unmarshal(value, &process); err != nil {
		return process, merror.Transform(err).Describe("unmarshaling authn process")
	}
	return process, nil
}

func (prr ProcessRedisRepo) GetByTok(ctx context.Context, tok string) (Process, error) {
	process := Process{}

	tokenKey := prr.key("*", tok)
	values, err := prr.SimpleKeyRedis.MustFind(ctx, tokenKey)
	if err != nil {
		return process, merror.Transform(err).Describe("getting token key")
	}
	value := values[0]
	if err := json.Unmarshal(value, &process); err != nil {
		return process, merror.Transform(err).Describe("unmarshaling authn process")
	}
	return process, nil
}

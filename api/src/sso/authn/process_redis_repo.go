package authn

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories"
)

// ProcessRedisRepo ...
type ProcessRedisRepo struct {
	cliID string
	repositories.SimpleKeyRedis
}

// NewAuthnProcessRedis ...
func NewAuthnProcessRedis(cliID string, redConn *redis.Client) ProcessRedisRepo {
	return ProcessRedisRepo{cliID, repositories.NewSimpleKeyRedis(redConn)}
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

// GetClaims ...
func (prr ProcessRedisRepo) GetClaims(ctx context.Context, tok string) (oidc.AccessClaims, error) {
	ac := oidc.AccessClaims{}

	tokenKey := prr.key("*", tok)
	values, err := prr.SimpleKeyRedis.MustFind(ctx, tokenKey)
	if err != nil {
		return ac, merr.From(err).Desc("getting token key")
	}
	process := Process{}
	value := values[0]
	if err := json.Unmarshal(value, &process); err != nil {
		return ac, merr.From(err).Desc("unmarshaling authn process")
	}

	// fill a claim structure with introspection
	ac = oidc.AccessClaims{
		Issuer: prr.cliID,
		// access token aren't usable externally
		Audiences: []string{prr.cliID},
		ClientID:  prr.cliID,

		ExpiresAt: process.ExpiresAt,
		IssuedAt:  process.IssuedAt,
		NotBefore: process.IssuedAt,

		Subject:    process.LoginChallenge,
		ACR:        process.CompleteAMRs.ToACR(),
		IdentityID: process.IdentityID,                 // potentially empty
		AccountID:  null.StringFrom(process.AccountID), // potentially empty

		Token: tok,
	}

	return ac, ac.Valid()
}

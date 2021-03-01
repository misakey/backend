package authz

import (
	"context"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mredis"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

type authnProcessInstropector struct {
	mredis.SimpleKeyRedis

	selfAudience string
}

// newAuthnProcessInstropector is the authnProcessInstropector structure constructor
func newAuthnProcessInstropector(
	tokenRepo interface{},
	selfAudience string) (a authnProcessInstropector, err error) {
	skr, ok := tokenRepo.(mredis.SimpleKeyRedis)
	if !ok {
		return a, merr.Internal().Desc("expect a mredis.SimpleKeyRedis as tokenRepo")
	}
	return authnProcessInstropector{skr, selfAudience}, nil
}

type authnProcessInstropection struct {
	LoginChallenge string          `json:"lgc"`
	SessionACR     oidc.ClassRef   `json:"sacr"`
	CompleteAMRs   oidc.MethodRefs `json:"camr"`
	IdentityID     string          `json:"mid"`
	AccountID      string          `json:"aid"`

	AccessToken string `json:"tok"`
	ExpiresAt   int64  `json:"exp"`
	IssuedAt    int64  `json:"iat"`
}

// GetBearerTok using Cookies ...
func (a authnProcessInstropector) GetBearerTok(ctx echo.Context) (tok string, fromCookie bool, err error) {
	tok, err = getBearerTokFromCookies(ctx, "authnaccesstoken", "authntokentype")
	if err != nil {
		return "", false, err
	}
	return tok, true, nil
}

// GetClaims using redis authn process session
func (a authnProcessInstropector) GetClaims(ctx context.Context, tok string) (oidc.AccessClaims, error) {
	ac := oidc.AccessClaims{}

	values, err := a.SimpleKeyRedis.MustFind(ctx, "authn_process:*:"+tok)
	if err != nil {
		return ac, merr.From(err).Desc("getting token key")
	}
	process := authnProcessInstropection{}
	value := values[0]
	if err := json.Unmarshal(value, &process); err != nil {
		return ac, merr.From(err).Desc("unmarshaling authn process")
	}

	// fill a claim structure with introspection
	ac = oidc.AccessClaims{
		Issuer: a.selfAudience,
		// access token aren't usable externally
		Audiences: []string{a.selfAudience},
		ClientID:  a.selfAudience,

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

// CheckClientID always return no error since the audience is set manually by
// the introspector code to audience attributes
func (a authnProcessInstropector) CheckClientID(_ context.Context, _ oidc.AccessClaims) error {
	return nil
}

// HandleErr
func (a authnProcessInstropector) HandleErr(eCtx echo.Context, next echo.HandlerFunc, err error) error {
	// not found authorization is possible during an authn process
	// some authn step will be required in this case
	// we don't raise any error regarding this
	if merr.IsANotFound(err) {
		return next(eCtx)
	}
	return merr.Unauthorized().Ori(merr.OriHeaders).Desc(err.Error())
}

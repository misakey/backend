package oidc

import (
	"context"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/volatiletech/null"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// AppRole contains a role owned for a given application.
type AppRole struct {
	ApplicationID string
	RoleLabel     string
}

// Claims declared to format our Access JWTs
// Implement https://godoc.org/github.com/dgrijalva/jwt-go#Claims interface
type AccessClaims struct {
	Issuer    string   `json:"iss"` // Service which distributed the token
	Audiences []string `json:"aud"` // Audiences which should answer to the tooken
	ClientID  string   `json:"cli"` // SSO client ID which generated the token

	ExpiresAt int64 `json:"exp"` // Expiry time
	IssuedAt  int64 `json:"iat"` // Issuing time
	NotBefore int64 `json:"nbf"` // Time before use

	Subject string   `json:"sub"` // Subject (owner) bound to the token
	Scope   string   `json:"sco"` // Scope beared by the token
	ACR     ClassRef `json:"acr"` // Authentication Class Reference

	Token      string      `json:"tok"` // Raw Access Token
	IdentityID string      `json:"mid"` // Misakey ID - Identity bound to the token
	AccountID  null.String `json:"aid"` // Account (nullable) bound to the token

	JWT string `json:"-"` // Raw JWT Token
}

// Valid : all required validation are today done on hydra side
func (c AccessClaims) Valid() error {
	now := clock.Now().Unix()
	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.

	// Verify if the token expired
	if !verifyExp(c.ExpiresAt, now) {
		return merror.Unauthorized().Describe("token expired")
	}

	// Verify if token is already issued
	if !verifyIat(c.IssuedAt, now) {
		return merror.Unauthorized().Describe("token used before issued")
	}

	// We need to setup the NotBefore condition
	// At the moment the value is 0
	// Verify not before condition
	if !verifyNbf(c.NotBefore, now) {
		return merror.Unauthorized().Describe("token not valid yet")
	}

	if c.Subject == "" {
		return merror.Unauthorized().Describe("empty subject")
	}

	if c.IdentityID == "" {
		return merror.Unauthorized().Describe("empty mid")
	}

	return nil
}

// SetRawJWT in the access claims
func (c *AccessClaims) SetRawJWT(jwt string) {
	c.JWT = jwt
}

// SetAccesses returns ctx with AccessClaims set inside it using accessContextKey
func SetAccesses(ctx context.Context, acc *AccessClaims) context.Context {
	return context.WithValue(ctx, accessContextKey{}, acc)
}

// GetAccesses returns AccessClaims found inside current context using defined accessContextKey
// It return a nil pointer if no claims have been found
func GetAccesses(ctx context.Context) *AccessClaims {
	val := ctx.Value(accessContextKey{})
	if val == nil {
		return nil
	}
	var accesses *AccessClaims
	accesses, ok := val.(*AccessClaims)
	if !ok {
		return nil
	}
	return accesses
}

// SetAccessClaimsJWT override current in-context access claims JWT value
func SetAccessClaimsJWT(ctx context.Context, jwt string) context.Context {
	return SetAccesses(ctx, &AccessClaims{JWT: jwt})
}

// ValidAudience : check if the client is part of the audience
func (c AccessClaims) ValidAudience(expected string) error {
	if !verifyAud(expected, c.Audiences) {
		return merror.Unauthorized().Describe("client is not part of the audience")
	}
	return nil
}

// GetSignedToken transforms an AccessClaims structure into a JWT
func GetSignedToken(ac AccessClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ac)
	return token.SignedString([]byte(JWTStaticSignature))
}

// AcccountConnect return boolean corresponding to the presence of an Account ID in claims
// it doesn't mean the connected identity has an existing linked account but either the end-user
// is connected on the account or just the identity (different ACRs)
func (acc AccessClaims) AccountConnected() bool {
	return acc.AccountID.Valid
}

// ----- helpers

func verifyAud(aud string, cmp []string) bool {
	return aud == "" || contains(cmp, aud)
}

func verifyExp(exp int64, now int64) bool {

	return now <= exp
}

func verifyIat(iat int64, now int64) bool {
	return now >= iat
}

func verifyNbf(nbf int64, now int64) bool {
	return now >= nbf
}

// Contains tells whether a contains x.
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

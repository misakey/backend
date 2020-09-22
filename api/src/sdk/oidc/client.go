package oidc

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// Client implementing some Open ID Connect concepts as a Relying Party (a.k.a. Third Party).
type Client struct {
	id       string
	tokenURL string

	signer jose.Signer
}

// NewClient configured with a tokenURL and an encoded JWK:
// a base64 encoded string of the JSON Web Key (the public and private keypair) following https://tools.ietf.org/html/rfc7517
// based on this string this constructor instantiates a JWK Signer to be able to sign client information in jwt.
func NewClient(id, tokenURL, encodedJWK string) (*Client, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedJWK)
	if err != nil {
		return nil, merror.Transform(err).Describe("could not decode encoded client jwk")
	}
	jwk := jose.JSONWebKey{}
	if err := jwk.UnmarshalJSON(decoded); err != nil {
		return nil, merror.Transform(err).Describe("could not unmarshal client jwk")
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jwk}, // RS256 is the only encryption algorithm handled by hydra today -https://github.com/ory/hydra/issues/1638
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return nil, merror.Transform(err).Describe("could not create jose.Signer")
	}

	oidcCli := &Client{
		id:       id,
		tokenURL: tokenURL,
		signer:   signer,
	}
	return oidcCli, nil
}

// ID returns the immutable client id
func (cli *Client) ID() string {
	return cli.id
}

// Assert claims created on the fly using the jwk signer and oidc client information
func (cli *Client) Assert(ctx context.Context) string {
	// generate a uuid for the jti
	jti, err := uuid.NewRandom()
	if err != nil {
		logger.FromCtx(ctx).Err(err).Msgf("could not generate uuid")
		return ""
	}
	claims := jwt.Claims{
		Issuer:   cli.id,
		Subject:  cli.id,
		Audience: jwt.Audience{cli.tokenURL},
		Expiry:   jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ID:       jti.String(),
	}
	assertion, err := jwt.Signed(cli.signer).Claims(claims).CompactSerialize()
	if err != nil {
		logger.FromCtx(ctx).Err(err).Msgf("could not assert")
		return ""
	}
	return assertion
}

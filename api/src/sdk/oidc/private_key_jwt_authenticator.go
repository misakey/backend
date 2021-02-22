package oidc

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"
	mhttp "gitlab.misakey.dev/misakey/backend/api/src/sdk/rester/http"
)

const privateKeyJWTFormat = "application/x-www-form-urlencoded"
const jwtBearerType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

// PrivateKeyJWTAuthenticator allows the Client Authentication using private_key_jwt method: https://openid.net/specs/openid-connect-core-1_0.html#ClientAuthentication
// It has 2 possible mode described below.
type PrivateKeyJWTAuthenticator struct {
	oidcCli *Client

	// Token mode makes the authenticator generating a bearer token for the client performing a concrete client_credentials flow,
	// instead of simply embedding private_key_jwt parameter in the body as it is necessary on an exchange token.
	// The token is then used as a Bearer Token (Authorization Header) for authorizing the request, it is kept in memory with its expiry time and renewed when necessary.
	tokenMode            bool
	currentAuthorization authorization // current token in use
	tokenRester          rester.Client
}

// Authorization containing a access token and expiry time information
type authorization struct {
	sync.Mutex // protect against multi token renewal

	Token     string `json:"access_token"`
	ExpiresIn int    `json:"expires_in"`
	Expiry    time.Time
}

// NewPrivateKeyJWTAuthenticator returned, configured with the given OIDCClient
func NewPrivateKeyJWTAuthenticator(oidcCli *Client, options ...func(*PrivateKeyJWTAuthenticator)) *PrivateKeyJWTAuthenticator {
	authenticator := &PrivateKeyJWTAuthenticator{oidcCli: oidcCli}

	// auto instantiate the tokenRester using oidcCli information
	authenticator.tokenRester = mhttp.NewClient(
		oidcCli.tokenURL,
		true,
		mhttp.SetFormat(mhttp.MimeTypeURLEncodedForm),
		mhttp.IgnoreProtocol(),
		mhttp.IgnoreInsecureHTTPS(),
	)
	// apply potentially options
	for _, opt := range options {
		opt(authenticator)
	}
	return authenticator
}

// UsingTokenMode sets tokenMode to true
// Token mode make the authenticator generating a bearer token for the client performing a concrete client_credentials flow,
// instead of just embedding private_key_jwt parameter in the body as it is necessary on an exchange token.
// The token is then used as a Bearer Token (Authorization Header), kept in memory with its expiry time and renewed when necessary.
func UsingTokenMode() func(*PrivateKeyJWTAuthenticator) {
	return func(authenticator *PrivateKeyJWTAuthenticator) {
		authenticator.tokenMode = true
	}
}

// Set client authentication considering used method is OIDC private_key_jwt: https://openid.net/specs/openid-connect-core-1_0.html#ClientAuthentication
func (authenticator *PrivateKeyJWTAuthenticator) Set(ctx context.Context, req *http.Request) {
	// return directly if no oidcCli has been configured
	if authenticator.oidcCli == nil {
		return
	}

	if authenticator.tokenMode {
		// 1. check token expiry - renew token if necessary
		authenticator.currentAuthorization.Lock() // ensure we don't renew twice at the same moment
		if time.Now().After(authenticator.currentAuthorization.Expiry) {
			authenticator.renewCurrentToken(ctx)
		}
		authenticator.currentAuthorization.Unlock()

		// 2. set valid token in authorization header
		req.Header.Set("Authorization", "Bearer "+authenticator.currentAuthorization.Token)
	} else {
		// without token mode, we just put the auth in the current body
		authenticator.embedClientAssertionInBody(ctx, req)
		// force format to application/x-www-form-urlencoded since it is the only to perform private_key_jwt
		// we need to set it here because it might not be done previsouly if the body was empty
		req.Header.Set("Content-Type", privateKeyJWTFormat)
	}
}

func (authenticator *PrivateKeyJWTAuthenticator) getOIDCAuthParams(ctx context.Context, grantType string, scope string) url.Values {
	params := url.Values{}
	params.Add("client_assertion_type", jwtBearerType)
	params.Add("client_id", authenticator.oidcCli.ID())
	params.Add("client_assertion", authenticator.oidcCli.Assert(ctx))
	if grantType != "" {
		params.Add("grant_type", grantType)
	}
	if scope != "" {
		params.Add("scope", scope)
	}
	return params
}

func (authenticator *PrivateKeyJWTAuthenticator) embedClientAssertionInBody(ctx context.Context, req *http.Request) {
	// finally set auth on request by concatenating it on body
	buf := new(bytes.Buffer)
	// grantType is the oauth2.0 paramater to authenticate that may be necessary
	var grantType, scope string = "client_credentials", "application"
	if req.Body != nil {
		body, _ := ioutil.ReadAll(req.Body)
		// if a body was here, we add a & operator to concat it
		if len(body) > 0 {
			grantType = "" // in case of an existing body, the grant_type should be already here
			_, _ = buf.Write(body)
			_, _ = buf.Write([]byte("&"))
		}
	}
	authParams := authenticator.getOIDCAuthParams(ctx, grantType, scope)
	_, _ = buf.Write([]byte(authParams.Encode()))
	req.ContentLength = int64(buf.Len())
	req.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
}

func (authenticator *PrivateKeyJWTAuthenticator) renewCurrentToken(ctx context.Context) {
	var grantType, scope string = "client_credentials", "application"
	authParams := authenticator.getOIDCAuthParams(ctx, grantType, scope)
	// the route is empty because the base url has been configured in the constructor with the full token url
	if err := authenticator.tokenRester.Post(ctx, "", nil, authParams, &authenticator.currentAuthorization); err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not perform client credentials")
		return
	}
	// compute final token expiration time
	// remove 2 minute from exact expiry time to ensure the avoidance of a 401 - renewal will be performed before the expiration
	expiry := time.Now().Add(time.Second * time.Duration(authenticator.currentAuthorization.ExpiresIn)).Add(-time.Minute * 2)
	authenticator.currentAuthorization.Expiry = expiry
}

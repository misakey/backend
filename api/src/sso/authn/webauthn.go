package authn

import (
	"context"
	"strings"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/mwebauthn"
)

func assertWebauthnCredentials(ctx context.Context, exec boil.ContextExecutor, curIdentity identity.Identity) error {
	_, err := identity.GetWebauthnIdentity(ctx, exec, curIdentity)

	return err
}

func prepareWebauthn(
	ctx context.Context, as *Service, exec boil.ContextExecutor, redConn *redis.Client,
	curIdentity identity.Identity, currentACR oidc.ClassRef, step *Step,
	passwordReset bool,
) (*Step, error) {
	// if in a password reset flow, ask for an email code
	if passwordReset {
		if currentACR.LessThan(oidc.ACR1) {
			return prepareEmailedCode(ctx, as, exec, redConn, curIdentity, currentACR, step, false)
		}
		// if at the end of a password reset flow, ask for the new password
		if currentACR == oidc.ACR3 {
			return prepareResetPassword(step)
		}
	} else {
		// first start with a password
		if currentACR.LessThan(oidc.ACR2) {
			return preparePassword(ctx, as, exec, redConn, curIdentity, currentACR, step, false)
		}
	}

	// then it is time for webauthn
	step.MethodName = oidc.AMRWebauthn
	wid, err := identity.GetWebauthnIdentity(ctx, exec, curIdentity)
	if err != nil {
		return step, merr.From(err).Desc("getting webauthn identity")
	}
	options, sessionData, err := as.WebauthnHandler.BeginLogin(&wid)
	if err != nil {
		return step, merr.From(err).Desc("beginning login")
	}

	if err := step.RawJSONMetadata.Marshal(options); err != nil {
		return step, merr.From(err).Desc("marshalling webauthn options")
	}

	// store session data in redis
	if err := mwebauthn.StoreSession(redConn, sessionData, curIdentity.ID, options.Response.Challenge.String()); err != nil {
		return step, merr.From(err).Desc("storing session")
	}

	return step, nil
}

func (as *Service) assertWebauthn(
	ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client,
	curIdentity identity.Identity, assertion Step) error {

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(assertion.RawJSONMetadata.String()))
	if err != nil {
		return merr.From(err).Desc("parsing credentials")
	}

	sessionData, err := mwebauthn.GetSession(redConn, curIdentity.ID, parsedResponse.Response.CollectedClientData.Challenge)
	if err != nil {
		return merr.From(err).Desc("getting session")
	}

	wid, err := identity.GetWebauthnIdentity(ctx, exec, curIdentity)
	if err != nil {
		return merr.From(err).Desc("getting webauthn identity")
	}

	_, err = as.WebauthnHandler.ValidateLogin(&wid, sessionData, parsedResponse)
	if err != nil {
		return merr.From(err).Desc("validating login")
	}

	return nil
}

package authn

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/pquerna/otp/totp"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/mtotp"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

func assertTOTPSecret(ctx context.Context, exec boil.ContextExecutor, curIdentity identity.Identity) error {
	if !mtotp.SecretExist(ctx, exec, curIdentity.ID) {
		return merr.Forbidden()
	}
	return nil
}

func prepareTOTP(
	ctx context.Context, as *Service, exec boil.ContextExecutor, redConn *redis.Client,
	curIdentity identity.Identity, currentACR oidc.ClassRef, step *Step,
	passwordReset bool,
) (*Step, error) {
	if passwordReset {
		// if in a password reset flow, ask for an email code
		if currentACR.LessThan(oidc.ACR1) {
			return prepareEmailedCode(ctx, as, exec, redConn, curIdentity, currentACR, step, false)
		}
		// if at the end of a password reset flow, ask for the new password
		if currentACR == oidc.ACR2 {
			return prepareResetPassword(step)
		}
	} else {
		// first ask for a password
		if currentACR.LessThan(oidc.ACR2) {
			return preparePassword(ctx, as, exec, redConn, curIdentity, currentACR, step, false)
		}
	}

	// then it is time for TOTP
	step.MethodName = oidc.AMRTOTP

	return step, nil
}

type totpAssertion struct {
	Code         string `json:"code"`
	RecoveryCode string `json:"recovery_code"`
}

func (as *Service) assertTOTP(
	ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client,
	curIdentity identity.Identity, assertion Step) error {

	mods := []qm.QueryMod{
		sqlboiler.TotpSecretWhere.IdentityID.EQ(curIdentity.ID),
	}
	secret, err := sqlboiler.TotpSecrets(mods...).One(ctx, exec)
	if err != nil {
		return merr.From(err).Desc("getting secret")
	}

	var content totpAssertion
	if err := assertion.RawJSONMetadata.Unmarshal(&content); err != nil {
		return merr.Forbidden().Ori(merr.OriBody).
			Desc(err.Error()).Add("metadata", merr.DVMalformed)
	}

	// if there is any, we validate the code
	// and if there is not, there may be a recovery code
	if content.Code != "" {
		if !totp.Validate(content.Code, secret.Secret) {
			return merr.Forbidden().Ori(merr.OriBody).Add("metadata", merr.DVInvalid)
		}
	} else if content.RecoveryCode != "" {
		if err := mtotp.CheckAndDeleteRecoveryCode(ctx, exec, secret.ID, content.RecoveryCode); err != nil {
			return err
		}
	} else {
		return merr.Forbidden().Ori(merr.OriBody).
			Desc(err.Error()).Add("metadata", merr.DVMalformed)
	}

	return nil
}

package authn

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn/argon2"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// AssertPasswordExistence ...
func (as *Service) AssertPasswordExistence(ctx context.Context, identity identity.Identity) error {
	if identity.AccountID.String == "" {
		return merr.Conflict().Desc("identity has no linked account").
			Add("identity_id", merr.DVConflict).
			Add("account_id", merr.DVRequired)
	}
	return nil
}

// preparePassword step by setting password hash information
func (as *Service) preparePassword(ctx context.Context, exec boil.ContextExecutor, curIdentity identity.Identity, step *Step) error {
	step.MethodName = oidc.AMRPrehashedPassword
	account, err := identity.GetAccount(ctx, exec, curIdentity.AccountID.String)
	if err != nil {
		return err
	}
	params, err := argon2.DecodeParams(account.Password)
	if err != nil {
		return err
	}
	if err := step.RawJSONMetadata.Marshal(params); err != nil {
		return merr.From(err).Desc("marshaling password metadata")
	}
	return nil
}

func (as *Service) assertPassword(
	ctx context.Context, exec boil.ContextExecutor,
	curIdentity identity.Identity, assertion Step,
) error {
	// transform metadata into argon2 password metadata structure
	pwdMetadata, err := argon2.ToMetadata(assertion.RawJSONMetadata)
	if err != nil {
		return merr.Forbidden().Ori(merr.OriBody).
			Desc(err.Error()).Add("metadata", merr.DVMalformed)
	}

	if curIdentity.AccountID.String == "" {
		return merr.Forbidden().Desc("identity has no linked account").
			Add("account_id", merr.DVRequired)
	}

	// retrieve the account
	account, err := identity.GetAccount(ctx, exec, curIdentity.AccountID.String)
	if err != nil {
		return err
	}

	// matches password
	pwdIsValid, err := pwdMetadata.Matches(account.Password)
	if err != nil {
		return err
	}
	if !pwdIsValid {
		return merr.Forbidden().Desc("invalid password").
			Ori(merr.OriBody).Add("metadata", merr.DVInvalid)
	}
	return nil
}

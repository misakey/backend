package authn

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn/argon2"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// assertPasswordExistence by checking the account id validity
// return a conflict error is not valid
func assertPasswordExistence(ctx context.Context, identity identity.Identity) error {
	if identity.AccountID.String == "" {
		return merr.Conflict().Desc("identity has no linked account").
			Add("identity_id", merr.DVConflict).
			Add("account_id", merr.DVRequired)
	}
	return nil
}

// preparePassword step by setting password hash information
func preparePassword(
	ctx context.Context, as *Service, exec boil.ContextExecutor,
	curIdentity identity.Identity, currentACR oidc.ClassRef, step *Step,
) (*Step, error) {
	step.MethodName = oidc.AMRPrehashedPassword
	// if the identity does not have an account, to prepare the password impossible
	// ask then for an account creation
	if err := assertPasswordExistence(ctx, curIdentity); err != nil {
		return requireAccount(ctx, as, exec, curIdentity, currentACR, step)
	}

	account, err := identity.GetAccount(ctx, exec, curIdentity.AccountID.String)
	if err != nil {
		return nil, err
	}
	params, err := argon2.DecodeParams(account.Password)
	if err != nil {
		return nil, err
	}
	if err := step.RawJSONMetadata.Marshal(params); err != nil {
		return nil, merr.From(err).Desc("marshaling password metadata")
	}
	return step, nil
}

func requireAccount(
	ctx context.Context, as *Service, exec boil.ContextExecutor,
	curIdentity identity.Identity, currentACR oidc.ClassRef, step *Step,
) (*Step, error) {
	// if the ACR of the current authn process is less than 1, return an emailed code step to upgrade it
	// because before setting an account, it is required to claim the email in the authn process
	if currentACR.LessThan(oidc.ACR1) {
		return prepareEmailedCode(ctx, as, exec, curIdentity, currentACR, step)
	}

	// otherwise, ask for account creation
	step.MethodName = oidc.AMRAccountCreation
	step.RawJSONMetadata = nil
	return step, nil
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

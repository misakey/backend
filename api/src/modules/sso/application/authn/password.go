package authn

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn/argon2"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// AssertPasswordExistence
func (as *Service) AssertPasswordExistence(ctx context.Context, identity domain.Identity) error {
	if identity.AccountID.String == "" {
		return merror.Conflict().Describe("identity has no linked account").
			Detail("identity_id", merror.DVConflict).
			Detail("account_id", merror.DVRequired)
	}
	return nil
}

// prepare password step by setting password hash information
func (as *Service) preparePassword(ctx context.Context, identity domain.Identity, step *Step) error {
	step.MethodName = oidc.AMRPrehashedPassword
	account, err := as.accountService.Get(ctx, identity.AccountID.String)
	if err != nil {
		return err
	}
	params, err := argon2.DecodeParams(account.Password)
	if err != nil {
		return err
	}
	if err := step.RawJSONMetadata.Marshal(params); err != nil {
		return merror.Transform(err).Describe("marshaling password metadata")
	}
	return nil
}

func (as *Service) assertPassword(ctx context.Context, identity domain.Identity, assertion Step) error {
	// transform metadata into argon2 password metadata structure
	pwdMetadata, err := argon2.ToMetadata(assertion.RawJSONMetadata)
	if err != nil {
		return merror.Forbidden().From(merror.OriBody).
			Describe(err.Error()).Detail("metadata", merror.DVMalformed)
	}

	if identity.AccountID.String == "" {
		return merror.Forbidden().Describe("identity has no linked account").
			Detail("account_id", merror.DVRequired)
	}

	// retrieve the account
	account, err := as.accountService.Get(ctx, identity.AccountID.String)
	if err != nil {
		return err
	}

	// matches password
	pwdIsValid, err := pwdMetadata.Matches(account.Password)
	if err != nil {
		return err
	}
	if !pwdIsValid {
		return merror.Forbidden().Describe("invalid password").
			From(merror.OriBody).Detail("metadata", merror.DVInvalid)
	}
	return nil
}

package authn

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn/argon2"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// AssertPasswordExistence
func (as *Service) AssertPasswordExistence(ctx context.Context, identityID string) error {
	// retrieve the identity to then get the account
	identity, err := as.identityService.Get(ctx, identityID)
	if err != nil {
		return merror.Transform(err).Describe("get identity")
	}

	if identity.AccountID.String == "" {
		return merror.Conflict().Describe("identity has no linked account").
			Detail("identity_id", merror.DVConflict).
			Detail("account_id", merror.DVRequired)
	}
	return nil
}

func (as *Service) assertPassword(ctx context.Context, assertion authn.Step) error {
	// transform metadata into argon2 password metadata structure
	pwdMetadata, err := argon2.ToMetadata(assertion.RawJSONMetadata)
	if err != nil {
		return merror.Forbidden().From(merror.OriBody).
			Describe(err.Error()).Detail("metadata", merror.DVMalformed)
	}

	// retrieve the identity
	identity, err := as.identityService.Get(ctx, assertion.IdentityID)
	if err != nil {
		return err
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

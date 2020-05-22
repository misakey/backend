package application

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

// CreateIdentityCommand orders:
// - the assurance of an identifier
// - a new identity (unconfirmed) linking the accoutn and the identifier
// - an identity confirmation flow that start the process of the identifier claiming
type CreateIdentityCommand struct {
	AccountID  string `json:"account_id"`
	Identifier struct {
		Value string `json:"value"`
		Kind  string `json:"kind"`
	} `json:"identifier"`
	DisplayName   string `json:"display_name"`
	Notifications string `json:"notifications"`
}

// Validate the CreateIdentityCommand
func (cmd CreateIdentityCommand) Validate() error {
	// we need to validate identifier separatly
	// because ozzo does not support nested struct validation
	if err := validation.ValidateStruct(&cmd.Identifier,
		validation.Field(&cmd.Identifier.Value, validation.Required, is.Email.Error("only emails are supported")),
		validation.Field(&cmd.Identifier.Kind, validation.Required, validation.In("email").Error("only emails are supported")),
	); err != nil {
		return err
	}

	if err := validation.ValidateStruct(&cmd,
		validation.Field(&cmd.AccountID, validation.Required, is.UUIDv4),
		// for the moment we only accept emails
		validation.Field(&cmd.DisplayName, validation.Required),
		validation.Field(&cmd.Notifications, validation.In("minimal", "moderate", "frequent")),
	); err != nil {
		return err
	}

	return nil
}

type CreateIdentityView struct {
	Identity domain.Identity `json:"identity"`
}

func (sso SSOService) CreateIdentity(ctx context.Context, cmd CreateIdentityCommand) (CreateIdentityView, error) {
	view := CreateIdentityView{}

	if !sso.displayNameFormat.Match([]byte(cmd.DisplayName)) {
		return view, merror.BadRequest().
			Describef("require %s matching", sso.displayNameFormat.String()).
			Detail("display_name", merror.DVInvalid)
	}

	// 1. ensure create the Identifier does exist
	identifier := domain.Identifier{
		Kind:  cmd.Identifier.Kind,
		Value: cmd.Identifier.Value,
	}
	if err := sso.identifierService.EnsureIdentifierExistence(ctx, &identifier); err != nil {
		return view, err
	}

	// 2. get the Account
	// TODO: check accesses
	if _, err := sso.accountService.Get(ctx, cmd.AccountID); err != nil {
		return view, merror.Conflict().Describe("could not find account").Detail("account_id", merror.DVNotFound)
	}

	// 3. create the Identity
	// TODO: manage the isAuthable potential conflict
	identity := domain.Identity{
		AccountID:     cmd.AccountID,
		IdentifierID:  identifier.ID,
		IsAuthable:    true,
		DisplayName:   cmd.DisplayName,
		Notifications: cmd.Notifications,
		Confirmed:     false,
	}
	if err := sso.identityService.Create(ctx, &identity); err != nil {
		return view, err
	}

	// 4. init the Identity Proofing process
	if err := sso.identityService.InitEmailCodeProofing(identity); err != nil {
		return view, err
	}

	view.Identity = identity
	return view, nil
}

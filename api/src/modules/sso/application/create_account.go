package application

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

// CreateAccountCommand orders:
// - the assurance of an identifier
// - a new account entity
// - a new identity (unconfirmed) linking the accoutn and the identifier
// - an identity confirmation flow that start the process of the identifier claiming
type CreateAccountCommand struct {
	Identifier struct {
		Value string `json:"value"`
		Kind  string `json:"kind"`
	} `json:"identifier"`
	DisplayName   string `json:"display_name"`
	Notifications string `json:"notifications"`
}

// Validate the CreateAccountCommand
func (cmd CreateAccountCommand) Validate() error {
	// we need to validate identifier separatly
	// because ozzo does not support nested struct validation
	if err := validation.ValidateStruct(&cmd.Identifier,
		validation.Field(&cmd.Identifier.Value, validation.Required, is.Email.Error("only emails are supported")),
		validation.Field(&cmd.Identifier.Kind, validation.Required, validation.In("email").Error("only emails are supported")),
	); err != nil {
		return err
	}

	if err := validation.ValidateStruct(&cmd,
		// for the moment we only accept emails
		validation.Field(&cmd.DisplayName, validation.Required),
		validation.Field(&cmd.Notifications, validation.Required, validation.In("minimal", "moderate", "frequent")),
	); err != nil {
		return err
	}

	return nil
}

type CreateAccountView struct {
	Account  domain.Account  `json:"account"`
	Identity domain.Identity `json:"identity"`
}

func (sso SSOService) CreateAccount(ctx context.Context, cmd CreateAccountCommand) (CreateAccountView, error) {
	view := CreateAccountView{}

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

	// 2. create the Account - we just need an ID
	account := domain.Account{}
	if err := sso.accountService.Create(ctx, &account); err != nil {
		return view, err
	}
	// 3. create the Identity
	identity := domain.Identity{
		AccountID:     account.ID,
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

	view.Account = account
	view.Identity = identity
	return view, nil
}

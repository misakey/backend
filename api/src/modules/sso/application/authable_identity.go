package application

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
)

// IdentityAuthableCmd orders:
// - the assurance of an identifier matching the received value
// - a new account if not authable identity linked to such identifier is found
// - a new identity (authable & unconfirmed) linking both previous entities
// - a init of confirmation code authencation method for the identity
type IdentityAuthableCmd struct {
	LoginChallenge string `json:"login_challenge"`
	Identifier     struct {
		Value string `json:"value"`
	} `json:"identifier"`
}

// Validate the IdentityAuthableCmd
func (cmd IdentityAuthableCmd) Validate() error {
	// validate nested structure separately
	if err := validation.ValidateStruct(&cmd.Identifier,
		validation.Field(&cmd.Identifier.Value, validation.Required, is.Email.Error("only emails are supported")),
	); err != nil {
		return err
	}

	if err := validation.ValidateStruct(&cmd,
		validation.Field(&cmd.LoginChallenge, validation.Required),
	); err != nil {
		return err
	}

	return nil
}

type IdentityAuthableView struct {
	Identity struct {
		DisplayName string      `json:"display_name"`
		AvatarURL   null.String `json:"avatar_url"`
	} `json:"identity"`
	AuthnStep struct {
		IdentityID string       `json:"identity_id"`
		MethodName authn.Method `json:"method_name"`
	} `json:"authn_step"`
}

// RequireIdentityAuthable for an auth flow.
// This method is used to retrieve information about the authable identity attached to an identifier value.
// The identifier value is set by the end-user on the interface and we receive it here.
// The function returns information about the Account & Identity that corresponds to the identifier.
// It creates is needed the trio identifier/account/identity.
// If an identity is created during this process, an confirmation code auth method is started
// This method will exceptionnaly both proof the identity and confirm the login flow within the auth flow.
func (sso SSOService) RequireIdentityAuthable(ctx context.Context, cmd IdentityAuthableCmd) (IdentityAuthableView, error) {
	var err error
	view := IdentityAuthableView{}

	// 0. check the login challenge exists
	_, err = sso.authFlowService.LoginGetContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return view, err
	}

	// 1. ensure create the Identifier does exist
	identifier := domain.Identifier{
		Kind:  domain.EmailIdentifier,
		Value: cmd.Identifier.Value,
	}
	if err := sso.identifierService.RequireIdentifier(ctx, &identifier); err != nil {
		return view, err
	}

	// 2. check if an identity exist for the identifier
	identityNotFound := func(err error) bool { return err != nil && merror.HasCode(err, merror.NotFoundCode) }
	var identity domain.Identity
	identity, err = sso.identityService.GetAuthableByIdentifierID(ctx, identifier.ID)
	if err != nil && !identityNotFound(err) {
		return view, err
	}

	// 3. create an account and an identity if nothing was found
	// or just retrieve the corresponding account
	if identityNotFound(err) {
		// a. create the Identity without account
		identity = domain.Identity{
			IdentifierID: identifier.ID,
			DisplayName:  cmd.Identifier.Value,
			IsAuthable:   true,
			Confirmed:    false,
		}
		if err := sso.identityService.Create(ctx, &identity); err != nil {
			return view, err
		}
	}

	// bind identity information on view
	view.Identity.DisplayName = identity.DisplayName
	view.Identity.AvatarURL = identity.AvatarURL
	view.AuthnStep.IdentityID = identity.ID

	// NOTE: if condition logic is commented because no password exists today
	// 4. if the identity has no linked account, we automatically init a emailed code authentication step
	// if identity.AccountID.IsZero() {
	// we ignore conflict error - if a code already exist, we still want to return authable identity information
	err = sso.authenticationService.CreateEmailedCode(ctx, identity.ID)
	if err != nil && !merror.HasCode(err, merror.ConflictCode) {
		return view, err
	}
	view.AuthnStep.MethodName = authn.EmailedCodeMethod
	// }
	return view, nil
}

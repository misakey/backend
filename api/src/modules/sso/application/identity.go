package application

import (
	"context"
	"io"
	"path/filepath"

	"github.com/volatiletech/sqlboiler/types"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn/argon2"
)

type IdentityQuery struct {
	IdentityID string
}

func (query IdentityQuery) Validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.IdentityID, v.Required, is.UUIDv4.Error("identity id must be an uuid v4")),
	)
}

type IdentityView struct {
	domain.Identity
}

func (sso SSOService) GetIdentity(ctx context.Context, query IdentityQuery) (IdentityView, error) {
	view := IdentityView{}
	var err error

	// verify identity access
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return view, merror.Forbidden()
	}

	if acc.Subject != query.IdentityID {
		return view, merror.Forbidden()
	}

	// set the view on identity retrieval
	view.Identity, err = sso.identityService.Get(ctx, query.IdentityID)
	if err != nil {
		return view, err
	}
	return view, err
}

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
	if err := v.ValidateStruct(&cmd.Identifier,
		v.Field(&cmd.Identifier.Value, v.Required, is.Email.Error("only emails are supported")),
	); err != nil {
		return err
	}

	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.LoginChallenge, v.Required),
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
		IdentityID string          `json:"identity_id"`
		MethodName authn.MethodRef `json:"method_name"`
		Metadata   *types.JSON     `json:"metadata"`
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
	if identity.AccountID.IsZero() {
		view.AuthnStep.MethodName = authn.AMREmailedCode
		// we ignore the conflict error code - if a code already exist, we still want to return authable identity information
		err = sso.authenticationService.CreateEmailedCode(ctx, identity.ID)
		if err != nil && !merror.HasCode(err, merror.ConflictCode) {
			return view, err
		}
	} else {
		view.AuthnStep.MethodName = authn.AMRPrehashedPassword
		account, err := sso.accountService.Get(ctx, identity.AccountID.String)
		if err != nil {
			return view, err
		}
		params, err := argon2.DecodeParams(account.Password)
		if err != nil {
			return view, err
		}
		view.AuthnStep.Metadata = &types.JSON{}
		if err := view.AuthnStep.Metadata.Marshal(params); err != nil {
			return view, err
		}
	}
	return view, nil
}

// PartialUpdateIdentityCmd
type PartialUpdateIdentityCmd struct {
	IdentityID    string
	DisplayName   string `json:"display_name"`
	Notifications string `json:"notifications"`
}

// Validate the IdentityAuthableCmd
func (cmd PartialUpdateIdentityCmd) Validate() error {
	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Notifications, v.In("minimal", "moderate", "frequent")),
		v.Field(&cmd.DisplayName, v.Length(3, 21)),
	); err != nil {
		return err
	}
	return nil
}

// PartialUpdateIdentity to change its display name or avatar
func (sso *SSOService) PartialUpdateIdentity(ctx context.Context, cmd PartialUpdateIdentityCmd) error {
	acc := ajwt.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.Subject != cmd.IdentityID {
		return merror.Forbidden()
	}

	identity, err := sso.identityService.Get(ctx, cmd.IdentityID)
	if err != nil {
		return err
	}

	if cmd.DisplayName != "" {
		identity.DisplayName = cmd.DisplayName
	}

	if cmd.Notifications != "" {
		identity.Notifications = cmd.Notifications
	}

	return sso.identityService.Update(ctx, &identity)
}

// UploadAvatarCmd
type UploadAvatarCmd struct {
	IdentityID string
	Data       io.Reader
	Extension  string
}

// Validate the UploadAvatarCmd
func (cmd UploadAvatarCmd) Validate() error {
	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Extension, v.Required),
	); err != nil {
		return err
	}
	return nil
}

func (sso *SSOService) UploadAvatar(ctx context.Context, cmd UploadAvatarCmd) error {
	acc := ajwt.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.Subject != cmd.IdentityID {
		return merror.Forbidden()
	}

	// get avatar's corresponding user
	identity, err := sso.identityService.Get(ctx, cmd.IdentityID)
	if err != nil {
		return err
	}

	// first remove existing avatar if it does exist
	if !identity.AvatarURL.IsZero() {
		avatarToDel := &domain.AvatarFile{
			Filename: filepath.Base(identity.AvatarURL.String),
		}
		if err := sso.identityService.DeleteAvatar(ctx, avatarToDel); err != nil {
			return err
		}
	}

	avatar := domain.AvatarFile{}
	// generate an UUID to use as a filename
	filename, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	avatar.Filename = filename.String() + avatar.Extension
	avatar.Extension = cmd.Extension
	avatar.Data = cmd.Data

	// upload the avatar to storage
	url, err := sso.identityService.UploadAvatar(ctx, &avatar)
	if err != nil {
		return err
	}

	// then save into user account the new avatar uri
	identity.AvatarURL = null.StringFrom(url)
	return sso.identityService.Update(ctx, &identity)
}

// DeleteAvatarCmd
type DeleteAvatarCmd struct {
	IdentityID string
}

// Validate the DeleteAvatarCmd
func (cmd DeleteAvatarCmd) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
	)
}

// DeleteAvatar for a given identity
func (sso *SSOService) DeleteAvatar(ctx context.Context, cmd DeleteAvatarCmd) error {
	avatar := domain.AvatarFile{}

	acc := ajwt.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.Subject != cmd.IdentityID {
		return merror.Forbidden()
	}

	// get identity
	identity, err := sso.identityService.Get(ctx, cmd.IdentityID)
	if err != nil {
		return err
	}

	// check that the identity has an avatar to delet
	if identity.AvatarURL.IsZero() {
		return merror.Conflict().Describe("avatar is not set").Detail("identity_id", merror.DVConflict)
	}

	// delete avatar
	avatar.Filename = filepath.Base(identity.AvatarURL.String)
	err = sso.identityService.DeleteAvatar(ctx, &avatar)
	if err != nil {
		return err
	}

	// update identity with empty avatar url field
	identity.AvatarURL = null.NewString("", false)

	return sso.identityService.Update(ctx, &identity)
}

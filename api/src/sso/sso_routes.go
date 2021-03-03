package sso

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oauth"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
)

func bindRoutes(
	router *echo.Echo,
	selfOIDCHandlers request.HandlerFactory,
	anyOIDCHandlers request.HandlerFactory,
	authnProcessHandlers request.HandlerFactory,
	ss *application.SSOService,
	oauthCodeFlow oauth.AuthorizationCodeFlow,
) {
	// ACCOUNT ROUTES
	accountPath := router.Group("/accounts")
	accountPath.GET(selfOIDCHandlers.NewACR2(
		"/:id/backup",
		func() request.Request { return &application.BackupQuery{} },
		ss.GetBackup,
		request.ResponseOK,
	))
	accountPath.PUT(selfOIDCHandlers.NewACR2(
		"/:id/deprecated/backup",
		func() request.Request { return &application.BackupUpdateCmd{} },
		ss.UpdateBackup,
		request.ResponseOK,
	))
	accountPath.GET(selfOIDCHandlers.NewPublic(
		"/:id/pwd-params",
		func() request.Request { return &application.PwdParamsQuery{} },
		ss.GetAccountPwdParams,
		request.ResponseOK,
	))
	accountPath.PUT(selfOIDCHandlers.NewACR2(
		"/:id/password",
		func() request.Request { return &application.ChangePasswordCmd{} },
		ss.ChangePassword,
		request.ResponseNoContent,
	))
	accountPath.GET(selfOIDCHandlers.NewACR2(
		"/:id/crypto/actions",
		func() request.Request { return &application.ListCryptoActionsQuery{} },
		ss.ListCryptoActions,
		request.ResponseOK,
	))
	accountPath.GET(selfOIDCHandlers.NewACR2(
		"/:account-id/crypto/actions/:action-id",
		func() request.Request { return &application.GetCryptoActionQuery{} },
		ss.GetCryptoAction,
		request.ResponseOK,
	))
	accountPath.DELETE(selfOIDCHandlers.NewACR2(
		"/:account-id/crypto/actions/:action-id",
		func() request.Request { return &application.DeleteCryptoActionQuery{} },
		ss.DeleteCryptoAction,
		request.ResponseNoContent,
	))
	// CRYPTO ROUTES
	cryptoPath := router.Group("/crypto")
	cryptoPath.POST(selfOIDCHandlers.NewACR2(
		"/migration/v2",
		func() request.Request { return &application.MigrateToSecretStorageQuery{} },
		ss.MigrateToSecretStorage,
		request.ResponseNoContent,
	))
	secretStoragePath := cryptoPath.Group("/secret-storage")
	secretStoragePath.GET(selfOIDCHandlers.NewACR2(
		"",
		func() request.Request { return &request.EmptyQuery{} },
		ss.GetSecretStorage,
		request.ResponseOK,
	))
	secretStoragePath.POST(selfOIDCHandlers.NewACR2(
		"/asym-keys",
		func() request.Request { return &crypto.SecretStorageAsymKey{} },
		ss.CreateSecretStorageAsymKey,
		request.ResponseOK,
	))
	secretStoragePath.DELETE(selfOIDCHandlers.NewACR2(
		"/asym-keys",
		func() request.Request { return &application.DeleteAsymKeysCmd{} },
		ss.DeleteAsymKeys,
		request.ResponseNoContent,
	))
	secretStoragePath.PUT(selfOIDCHandlers.NewACR2(
		"/box-key-shares/:box-id",
		func() request.Request { return &crypto.SecretStorageBoxKeyShare{} },
		ss.CreateSecretStorageBoxKeyShare,
		request.ResponseOK,
	))
	secretStoragePath.DELETE(selfOIDCHandlers.NewACR2(
		"/box-key-shares",
		func() request.Request { return &application.DeleteBoxKeySharesCmd{} },
		ss.DeleteBoxKeyShares,
		request.ResponseNoContent,
	))
	rootKeySharePath := cryptoPath.Group("/root-key-shares")
	rootKeySharePath.POST(selfOIDCHandlers.NewACR2(
		"",
		func() request.Request { return &application.RootKeyShareCreateCmd{} },
		ss.CreateRootKeyShare,
		request.ResponseCreated,
	))
	rootKeySharePath.GET(selfOIDCHandlers.NewACR2(
		"/:other-share-hash",
		func() request.Request { return &application.RootKeyShareQuery{} },
		ss.GetRootKeyShare,
		request.ResponseOK,
	))
	// IDENTITIES ROUTES
	identityPath := router.Group("/identities")
	identityPath.GET(selfOIDCHandlers.NewACR1(
		"/:id",
		func() request.Request { return &application.IdentityQuery{} },
		ss.GetIdentity,
		request.ResponseOK,
	))
	identityPath.PATCH(selfOIDCHandlers.NewACR1(
		"/:id",
		func() request.Request { return &application.PartialUpdateIdentityCmd{} },
		ss.PartialUpdateIdentity,
		request.ResponseNoContent,
	))
	identityPath.PUT(selfOIDCHandlers.NewACR1(
		"/:id/avatar",
		func() request.Request { return &application.UploadAvatarCmd{} },
		ss.UploadAvatar,
		request.ResponseNoContent,
	))
	identityPath.DELETE(selfOIDCHandlers.NewACR1(
		"/:id/avatar",
		func() request.Request { return &application.DeleteAvatarCmd{} },
		ss.DeleteAvatar,
		request.ResponseNoContent,
	))
	identityPath.POST(selfOIDCHandlers.NewACR2(
		"/:id/coupons",
		func() request.Request { return &application.AttachCouponCmd{} },
		ss.AttachCoupon,
		request.ResponseNoContent,
	))
	identityPath.HEAD(selfOIDCHandlers.NewACR1(
		"/:id/notifications",
		func() request.Request { return &application.IdentityNotifCountQuery{} },
		ss.CountIdentityNotification,
		request.ResponseNoContent,
		func(ctx echo.Context, count interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(count.(int)))
			return nil
		},
	))
	identityPath.GET(selfOIDCHandlers.NewACR1(
		"/:id/notifications",
		func() request.Request { return &application.IdentityNotifListQuery{} },
		ss.ListIdentityNotification,
		request.ResponseOK,
	))
	identityPath.PUT(selfOIDCHandlers.NewACR1(
		"/:id/notifications/acknowledgement",
		func() request.Request { return &application.IdentityNotifAckCmd{} },
		ss.AckIdentityNotification,
		request.ResponseNoContent,
	))
	identityPath.GET(selfOIDCHandlers.NewACR1(
		"/:id/organizations",
		func() request.Request { return &application.OrgListQuery{} },
		ss.ListIdentityOrgs,
		request.ResponseOK,
	))
	identityPath.GET(selfOIDCHandlers.NewPublic(
		"/:id/profile",
		func() request.Request { return &application.ProfileQuery{} },
		ss.GetProfile,
		request.ResponseOK,
	))
	identityPath.PATCH(selfOIDCHandlers.NewACR1(
		"/:id/profile/config",
		func() request.Request { return &application.ConfigProfileCmd{} },
		ss.SetProfileConfig,
		request.ResponseNoContent,
	))
	identityPath.GET(selfOIDCHandlers.NewACR1(
		"/:id/profile/config",
		func() request.Request { return &application.ConfigProfileQuery{} },
		ss.GetProfileConfig,
		request.ResponseOK,
	))
	identityPath.GET(selfOIDCHandlers.NewACR2(
		"/pubkey",
		func() request.Request { return &application.IdentityPubkeyByIdentifierQuery{} },
		ss.GetIdentityPubkeyByIdentifier,
		request.ResponseOK,
	))
	identityPath.GET(selfOIDCHandlers.NewACR2(
		"/:id/webauthn-credentials/create",
		func() request.Request { return &application.BeginWebAuthnRegistrationQuery{} },
		ss.BeginWebAuthnRegistration,
		request.ResponseOK,
	))
	identityPath.POST(selfOIDCHandlers.NewACR2(
		"/:id/webauthn-credentials/create",
		func() request.Request { return &application.FinishWebAuthnRegistrationQuery{} },
		ss.FinishWebAuthnRegistration,
		request.ResponseOK,
	))
	identityPath.GET(selfOIDCHandlers.NewACR2(
		"/:id/totp/enroll",
		func() request.Request { return &application.BeginTOTPEnrollmentQuery{} },
		ss.BeginTOTPEnrollment,
		request.ResponseOK,
	))
	identityPath.POST(selfOIDCHandlers.NewACR2(
		"/:id/totp/enroll",
		func() request.Request { return &application.FinishTOTPEnrollmentQuery{} },
		ss.FinishTOTPEnrollment,
		request.ResponseOK,
	))
	identityPath.POST(selfOIDCHandlers.NewACR3(
		"/:id/totp/recovery-codes",
		func() request.Request { return &application.RegenerateRecoveryCodesQuery{} },
		ss.RegenerateRecoveryCodes,
		request.ResponseOK,
	))
	identityPath.DELETE(selfOIDCHandlers.NewACR2(
		"/:id/totp",
		func() request.Request { return &application.DeleteSecretQuery{} },
		ss.DeleteSecret,
		request.ResponseNoContent,
	))

	// ORGANIZATIONS
	orgPath := router.Group("/organizations")
	orgPath.POST(selfOIDCHandlers.NewACR2(
		"",
		func() request.Request { return &application.OrgCreateCmd{} },
		ss.CreateOrg,
		request.ResponseCreated,
	))
	orgPath.PUT(selfOIDCHandlers.NewACR2(
		"/:id/secret",
		func() request.Request { return &application.GenerateSecretCmd{} },
		ss.GenerateSecret,
		request.ResponseOK,
	))

	orgPath.GET(selfOIDCHandlers.NewPublic(
		"/:id/public",
		func() request.Request { return &application.GetOrgPublicRequest{} },
		ss.GetOrgPublic,
		request.ResponseOK,
	))

	// DATATAGS
	orgPath.POST(anyOIDCHandlers.NewACR2(
		"/:id/datatags",
		func() request.Request { return &application.CreateDatatagCmd{} },
		ss.CreateDatatag,
		request.ResponseCreated,
	))
	orgPath.GET(anyOIDCHandlers.NewACR2(
		"/:id/datatags",
		func() request.Request { return &application.ListDatatagsCmd{} },
		ss.ListDatatags,
		request.ResponseOK,
	))
	orgPath.PATCH(anyOIDCHandlers.NewACR2(
		"/:id/datatags/:did",
		func() request.Request { return &application.PatchDatatagCmd{} },
		ss.PatchDatatag,
		request.ResponseNoContent,
	))

	// WEBAUTHN CREDENTIALS
	webauthnCredentialPath := router.Group("/webauthn-credentials")
	webauthnCredentialPath.GET(selfOIDCHandlers.NewACR2(
		"",
		func() request.Request { return &application.ListCredentialsQuery{} },
		ss.ListCredentials,
		request.ResponseOK,
	))
	webauthnCredentialPath.DELETE(selfOIDCHandlers.NewACR2(
		"/:id",
		func() request.Request { return &application.DeleteCredentialQuery{} },
		ss.DeleteCredential,
		request.ResponseNoContent,
	))

	// authn-steps creation
	// NOTE: /auth/authn-steps would be better
	router.POST(authnProcessHandlers.NewPublic(
		"/authn-steps",
		func() request.Request { return &application.AuthenticationStepCmd{} },
		ss.InitAuthnStep,
		request.ResponseNoContent,
	))

	// AUTH
	authPath := router.Group("/auth")
	// login flow
	authPath.GET(authnProcessHandlers.NewPublic(
		"/login",
		func() request.Request { return &application.LoginInitCmd{} },
		ss.LoginInit,
		request.ResponseRedirectFound,
	))
	authPath.GET(authnProcessHandlers.NewPublic(
		"/login/info",
		func() request.Request { return &application.LoginInfoQuery{} },
		ss.LoginInfo,
		request.ResponseOK,
	))
	authPath.POST(authnProcessHandlers.NewOptional(
		"/login/authn-step",
		func() request.Request { return &application.LoginAuthnStepCmd{} },
		ss.AssertAuthnStep,
		request.ResponseOK,
		ss.CleanAuthnCookie, // make obsolete previous cookie
		func(c echo.Context, stepViewInt interface{}) error {
			stepView, ok := stepViewInt.(application.LoginAuthnStepView)
			if !ok {
				return merr.Internal().Desc("expect application.LoginAuthnStepView type")
			}
			// set process access token info into the cookies
			authz.SetCookie(c, "authnaccesstoken", stepView.ForCookies.AccessToken, stepView.ForCookies.ExpirationDate)
			authz.SetCookie(c, "authntokentype", "bearer", stepView.ForCookies.ExpirationDate)
			return nil
		},
	))
	// identity in auth
	authPath.PUT(authnProcessHandlers.NewOptional(
		"/identities",
		func() request.Request { return &application.RequireIdentityCmd{} },
		ss.RequireIdentity,
		request.ResponseOK,
	))

	// consent flow
	authPath.GET(authnProcessHandlers.NewPublic(
		"/consent",
		func() request.Request { return &application.ConsentInitCmd{} },
		ss.InitConsent,
		request.ResponseRedirectFound,
	))
	authPath.GET(authnProcessHandlers.NewPublic(
		"/consent/info",
		func() request.Request { return &application.ConsentInfoQuery{} },
		ss.GetConsentInfo,
		request.ResponseOK,
	))
	authPath.POST(authnProcessHandlers.NewPublic(
		"/consent",
		func() request.Request { return &application.ConsentAcceptCmd{} },
		ss.AcceptConsent,
		request.ResponseOK,
	))
	// exchange token
	authPath.GET("/callback", func(ctx echo.Context) error {
		// remove cookie related to login flow/authn process authorization since it is the end
		_ = ss.CleanAuthnCookie(ctx, nil)
		oauthCodeFlow.ExchangeToken(ctx)
		return nil
	})
	// backup routes during the auth flow
	// (required for migration)
	authPath.GET(authnProcessHandlers.NewACR2(
		"/backup",
		func() request.Request { return &application.GetBackupQuery{} },
		ss.GetBackupDuringAuth,
		request.ResponseOK,
	))

	// secret storage routes during the auth flow
	authPath.GET(authnProcessHandlers.NewACR2(
		"/secret-storage",
		func() request.Request { return &application.GetSecretStorageQuery{} },
		ss.GetSecretStorageDuringAuth,
		request.ResponseOK,
	))
	authPath.POST(authnProcessHandlers.NewACR2(
		"/crypto/migration/v2",
		func() request.Request { return &application.MigrateToSecretStorageQuery{} },
		ss.MigrateToSecretStorage,
		request.ResponseNoContent,
	))
	authPath.POST(authnProcessHandlers.NewACR2(
		"/account-root-key-shares",
		func() request.Request { return &application.RootKeyShareCreateCmd{} },
		ss.CreateRootKeyShare,
		request.ResponseCreated,
	))
	// following routes allows audience of non-misakey oidc tokens
	authPath.POST(anyOIDCHandlers.NewACR1(
		"/logout",
		nil, // no request data required
		ss.Logout,
		request.ResponseNoContent,
		ss.CleanOIDCCookie,
	))
	// reset the auth flow using the login_challenge
	authPath.GET(authnProcessHandlers.NewPublic(
		"/reset",
		func() request.Request { return &application.FlowResetCmd{} },
		ss.ResetFlow,
		request.ResponseRedirectFound,
		ss.CleanOIDCCookie,
	))
	// user info
	authPath.GET(selfOIDCHandlers.NewOptional(
		"/userinfo",
		func() request.Request { return &application.GetUserInfoCmd{} },
		ss.GetUserInfo,
		request.ResponseOK,
	))
}

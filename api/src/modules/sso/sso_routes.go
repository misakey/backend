package sso

import (
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oauth"
)

func bindRoutes(
	router *echo.Echo,
	oidcHandlers request.HandlerFactory,
	authnProcessHandlers request.HandlerFactory,
	extOIDCHandlers request.HandlerFactory,
	ss *application.SSOService,
	oauthCodeFlow oauth.AuthorizationCodeFlow,
) {
	// ACCOUNT ROUTES
	accountPath := router.Group("/accounts")
	accountPath.GET(oidcHandlers.NewACR2(
		"/:id/backup",
		func() request.Request { return &application.BackupQuery{} },
		ss.GetBackup,
		request.ResponseOK,
	))
	accountPath.PUT(oidcHandlers.NewACR2(
		"/:id/backup",
		func() request.Request { return &application.BackupUpdateCmd{} },
		ss.UpdateBackup,
		request.ResponseOK,
	))
	accountPath.GET(oidcHandlers.NewPublic(
		"/:id/pwd-params",
		func() request.Request { return &application.PwdParamsQuery{} },
		ss.GetAccountPwdParams,
		request.ResponseOK,
	))
	accountPath.PUT(oidcHandlers.NewACR2(
		"/:id/password",
		func() request.Request { return &application.ChangePasswordCmd{} },
		ss.ChangePassword,
		request.ResponseNoContent,
	))
	accountPath.GET(oidcHandlers.NewACR2(
		"/:id/crypto/actions",
		func() request.Request { return &application.ListCryptoActionsQuery{} },
		ss.ListCryptoActions,
		request.ResponseOK,
	))
	accountPath.DELETE(oidcHandlers.NewACR2(
		"/:id/crypto/actions",
		func() request.Request { return &application.DeleteCryptoActionsCmd{} },
		ss.DeleteCryptoActionsUntil,
		request.ResponseNoContent,
	))
	// IDENTITIES ROUTES
	identityPath := router.Group("/identities")
	identityPath.GET(oidcHandlers.NewACR1(
		"/:id",
		func() request.Request { return &application.IdentityQuery{} },
		ss.GetIdentity,
		request.ResponseOK,
	))
	identityPath.PATCH(oidcHandlers.NewACR1(
		"/:id",
		func() request.Request { return &application.PartialUpdateIdentityCmd{} },
		ss.PartialUpdateIdentity,
		request.ResponseNoContent,
	))
	identityPath.PUT(oidcHandlers.NewACR1(
		"/:id/avatar",
		func() request.Request { return &application.UploadAvatarCmd{} },
		ss.UploadAvatar,
		request.ResponseNoContent,
	))
	identityPath.DELETE(oidcHandlers.NewACR1(
		"/:id/avatar",
		func() request.Request { return &application.DeleteAvatarCmd{} },
		ss.DeleteAvatar,
		request.ResponseNoContent,
	))
	identityPath.GET(oidcHandlers.NewPublic(
		"/:id/profile",
		func() request.Request { return &application.ProfileQuery{} },
		ss.GetProfile,
		request.ResponseOK,
	))
	identityPath.PATCH(oidcHandlers.NewACR1(
		"/:id/profile/config",
		func() request.Request { return &application.ConfigProfileCmd{} },
		ss.SetProfileConfig,
		request.ResponseNoContent,
	))
	identityPath.GET(oidcHandlers.NewACR1(
		"/:id/profile/config",
		func() request.Request { return &application.ConfigProfileQuery{} },
		ss.GetProfileConfig,
		request.ResponseOK,
	))
	identityPath.POST(oidcHandlers.NewACR2(
		"/:id/coupons",
		func() request.Request { return &application.AttachCouponCmd{} },
		ss.AttachCoupon,
		request.ResponseNoContent,
	))
	identityPath.GET(oidcHandlers.NewACR2(
		"/pubkey",
		func() request.Request { return &application.IdentityPubkeyByIdentifierQuery{} },
		ss.GetIdentityPubkeyByIdentifier,
		request.ResponseOK,
	))
	// NOTE: part of the auth flow - the path would be clearer with /auth/identities
	identityPath.PUT(authnProcessHandlers.NewOptional(
		"/authable",
		func() request.Request { return &application.IdentityAuthableCmd{} },
		ss.RequireAuthableIdentity,
		request.ResponseOK,
	))

	// BACKUP KEY SHARES ROUTES
	backupKeySharePath := router.Group("/backup-key-shares")
	backupKeySharePath.POST(oidcHandlers.NewACR2(
		"",
		func() request.Request { return &application.BackupKeyShareCreateCmd{} },
		ss.CreateBackupKeyShare,
		request.ResponseCreated,
	))
	backupKeySharePath.GET(oidcHandlers.NewACR2(
		"/:other-share-hash",
		func() request.Request { return &application.BackupKeyShareQuery{} },
		ss.GetBackupKeyShare,
		request.ResponseOK,
	))

	// BACKUP KEY ARCHIVES ROUTES
	backupArchivePath := router.Group("/backup-archives")
	backupArchivePath.GET(oidcHandlers.NewACR2(
		"",
		nil,
		ss.ListBackupArchives,
		request.ResponseOK,
	))
	backupArchivePath.GET(oidcHandlers.NewACR2(
		"/:id/data",
		func() request.Request { return &application.BackupArchiveDataQuery{} },
		ss.GetBackupArchiveData,
		request.ResponseOK,
	))
	backupArchivePath.DELETE(oidcHandlers.NewACR2(
		"/:id",
		func() request.Request { return &application.BackupArchiveDeleteCmd{} },
		ss.DeleteBackupArchive,
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
		oauthCodeFlow.ExchangeToken(ctx)
		return nil
	})
	// backup routes during the auth flow
	authPath.GET(authnProcessHandlers.NewACR2(
		"/backup",
		func() request.Request { return &application.GetBackupQuery{} },
		ss.GetBackupDuringAuth,
		request.ResponseOK,
	))
	authPath.POST(authnProcessHandlers.NewACR2(
		"/backup-key-shares",
		func() request.Request { return &application.BackupKeyShareCreateCmd{} },
		ss.CreateBackupKeyShare,
		request.ResponseCreated,
	))

	// following routes allows audience of non-misakey oidc tokens
	authPath.POST(extOIDCHandlers.NewACR1(
		"/logout",
		nil, // no request data required
		ss.Logout,
		request.ResponseNoContent,
		ss.CleanCookie,
	))
}

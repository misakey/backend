package sso

import (
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/backuparchive"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oauth"
)

func initRoutes(
	router *echo.Echo,
	authnProcessAuthzMidlw echo.MiddlewareFunc,
	oidcAuthzMidlw echo.MiddlewareFunc,
	extOIDCAuthzMidlw echo.MiddlewareFunc,
	ssoService application.SSOService,
	oauthCodeFlow oauth.AuthorizationCodeFlow,
	backupArchiveService backuparchive.BackupArchiveService,
) {
	// init entrypoints
	accountHTTP := entrypoints.NewAccountHTTP(ssoService)
	authFlowHTTP := entrypoints.NewAuthFlowHTTP(ssoService)
	authnHTTP := entrypoints.NewAuthnHTTP(ssoService)
	identityHTTP := entrypoints.NewIdentityHTTP(ssoService)
	backupKeyShareHTTP := entrypoints.NewBackupKeyShareHTTP(ssoService)
	backupArchiveHTTP := entrypoints.NewBackupArchiveHTTP(ssoService)
	cryptoActionHTTP := entrypoints.NewCryptoActionsHTTP(ssoService)

	routes := router.Group("")
	routes.POST("/authn-steps", authnHTTP.InitAuthnStep)

	accountRoutes := router.Group("/accounts")
	accountRoutes.GET("/:id/backup", accountHTTP.GetBackup, oidcAuthzMidlw)
	accountRoutes.PUT("/:id/backup", accountHTTP.UpdateBackup, oidcAuthzMidlw)
	accountRoutes.GET("/:id/pwd-params", accountHTTP.GetPwdParams)
	accountRoutes.PUT("/:id/password", accountHTTP.ChangePassword, oidcAuthzMidlw)
	accountRoutes.GET("/:id/crypto/actions", cryptoActionHTTP.ListCryptoActions, oidcAuthzMidlw)
	accountRoutes.DELETE("/:id/crypto/actions", cryptoActionHTTP.DeleteCryptoActions, oidcAuthzMidlw)

	authRoutes := router.Group("/auth")
	authRoutes.GET("/login", authFlowHTTP.LoginInit)
	authRoutes.GET("/login/info", authFlowHTTP.LoginInfo)
	authRoutes.POST("/login/authn-step", authFlowHTTP.LoginAuthnStep, authnProcessAuthzMidlw)
	authRoutes.POST("/logout", authFlowHTTP.Logout, extOIDCAuthzMidlw)
	authRoutes.GET("/consent", authFlowHTTP.ConsentInit)
	authRoutes.GET("/consent/info", authFlowHTTP.ConsentInfo)
	authRoutes.POST("/consent", authFlowHTTP.ConsentAccept)
	authRoutes.GET("/callback", func(ctx echo.Context) error {
		oauthCodeFlow.ExchangeToken(ctx.Response().Writer, ctx.Request())
		return nil
	})
	authRoutes.GET("/backup", authFlowHTTP.GetBackup, authnProcessAuthzMidlw)
	authRoutes.POST("/backup-key-shares", backupKeyShareHTTP.CreateBackupKeyShare, authnProcessAuthzMidlw)

	identityRoutes := router.Group("/identities")
	identityRoutes.GET("/:id", identityHTTP.GetIdentity, oidcAuthzMidlw)
	identityRoutes.PATCH("/:id", identityHTTP.PartiallyUpdateIdentity, oidcAuthzMidlw)
	identityRoutes.PUT("/:id/avatar", identityHTTP.UploadAvatar, oidcAuthzMidlw)
	identityRoutes.DELETE("/:id/avatar", identityHTTP.DeleteAvatar, oidcAuthzMidlw)
	identityRoutes.PUT("/authable", identityHTTP.RequireAuthableIdentity, authnProcessAuthzMidlw)
	identityRoutes.POST("/:id/coupons", identityHTTP.AttachCoupon, oidcAuthzMidlw)

	backupKeyShareRoutes := router.Group("/backup-key-shares")
	backupKeyShareRoutes.GET("/:other-share-hash", backupKeyShareHTTP.GetBackupKeyShare, oidcAuthzMidlw)
	backupKeyShareRoutes.POST("", backupKeyShareHTTP.CreateBackupKeyShare, oidcAuthzMidlw)

	backupArchiveRoutes := router.Group("/backup-archives")
	backupArchiveRoutes.GET("", backupArchiveHTTP.ListBackupArchives, oidcAuthzMidlw)
	backupArchiveRoutes.GET("/:id/data", backupArchiveHTTP.GetArchiveData, oidcAuthzMidlw)
	backupArchiveRoutes.DELETE("/:id", backupArchiveHTTP.DeleteArchive, oidcAuthzMidlw)
}

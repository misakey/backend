package sso

import (
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oauth"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func initRoutes(
	router *echo.Echo,
	authzMidlw echo.MiddlewareFunc,
	ssoService application.SSOService,
	oauthCodeFlow oauth.AuthorizationCodeFlow,
) {
	// init entrypoints
	accountHTTP := entrypoints.NewAccountHTTP(ssoService)
	authFlowHTTP := entrypoints.NewAuthFlowHTTP(ssoService)
	authnHTTP := entrypoints.NewAuthnHTTP(ssoService)
	identityHTTP := entrypoints.NewIdentityHTTP(ssoService)

	routes := router.Group("")
	routes.POST("/authn-steps", authnHTTP.InitAuthnStep)

	accountRoutes := router.Group("/accounts")
	accountRoutes.GET("/:id/backup", accountHTTP.GetBackup, authzMidlw)
	accountRoutes.PUT("/:id/backup", accountHTTP.UpdateBackup, authzMidlw)
	accountRoutes.GET("/:id/pwd-params", accountHTTP.GetPwdParams)
	// TODO: check ACR2 for change password
	accountRoutes.PUT("/:id/password", accountHTTP.ChangePassword, authzMidlw)
	//TODO: add acr1 for password reset
	accountRoutes.PUT("/:id/password/reset", accountHTTP.ResetPassword, authzMidlw)

	authRoutes := router.Group("/auth")
	authRoutes.GET("/login", authFlowHTTP.LoginInit)
	authRoutes.GET("/login/info", authFlowHTTP.LoginInfo)
	authRoutes.POST("/login/authn-step", authFlowHTTP.LoginAuthnStep)
	authRoutes.POST("/logout", authFlowHTTP.Logout, authzMidlw)
	authRoutes.GET("/consent", authFlowHTTP.ConsentInit)
	authRoutes.GET("/callback", func(ctx echo.Context) error {
		oauthCodeFlow.ExchangeToken(ctx.Response().Writer, ctx.Request())
		return nil
	})

	identityRoutes := router.Group("/identities")
	identityRoutes.GET("/:id", identityHTTP.GetIdentity, authzMidlw)
	identityRoutes.POST("/:id/account", identityHTTP.CreateAccount, authzMidlw)
	identityRoutes.PUT("/authable", identityHTTP.RequireAuthableIdentity)
}

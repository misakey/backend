package sso

import (
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oauth"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func initRoutes(
	router *echo.Echo,
	ssoService application.SSOService,
	oauthCodeFlow oauth.AuthorizationCodeFlow,
) {
	// init entrypoints
	identityHTTP := entrypoints.NewIdentityHTTP(ssoService)
	authFlowHTTP := entrypoints.NewAuthFlowHTTP(ssoService)
	authnHTTP := entrypoints.NewAuthnHTTP(ssoService)

	routes := router.Group("")
	routes.POST("/authn-steps", authnHTTP.InitAuthnStep)

	authRoutes := router.Group("/auth")
	authRoutes.GET("/login", authFlowHTTP.LoginInit)
	authRoutes.GET("/login/info", authFlowHTTP.LoginInfo)
	// TODO14: add a limit req on gateway to this endpoint
	authRoutes.POST("/login/authn-step", authFlowHTTP.LoginAuthnStep)
	authRoutes.GET("/consent", authFlowHTTP.ConsentInit)
	authRoutes.GET("/callback", func(ctx echo.Context) error {
		oauthCodeFlow.ExchangeToken(ctx.Response().Writer, ctx.Request())
		return nil
	})

	identityRoutes := router.Group("/identities")
	identityRoutes.PUT("/authable", identityHTTP.RequireAuthableIdentity)
	// identityRoutes.POST("/:id/assertions", identityHTTP.AssertIdentity)
	// identityRoutes.POST("/identities", identityHTTP.CreateIdentity)
}

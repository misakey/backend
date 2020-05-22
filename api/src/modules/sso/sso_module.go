package sso

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oidc"
	"gitlab.misakey.dev/misakey/msk-sdk-go/rester/http"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories"
)

func InitModule(router *echo.Echo, dbConn *sql.DB) {
	// init self authenticator for hydra rester
	selfAuth, err := oidc.NewClient(
		viper.GetString("authflow.self_client_id"),
		viper.GetString("authflow.token_url"),
		viper.GetString("authflow.self_encoded_jwk"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not create the oidc client")
	}

	// init resters
	publicHydraJSON := http.NewClient(viper.GetString("hydra.public_endpoint"), viper.GetBool("hydra.secure"))
	adminHydraJSON := http.NewClient(viper.GetString("hydra.admin_endpoint"), viper.GetBool("hydra.secure"))
	publicHydraFORM := http.NewClient(
		viper.GetString("hydra.public_endpoint"),
		viper.GetBool("hydra.secure"),
		http.SetFormat(http.URLENCODED_FORM_MIME_TYPE),
		http.SetAuthenticator(oidc.NewPrivateKeyJWTAuthenticator(selfAuth)),
	)
	adminHydraFORM := http.NewClient(
		viper.GetString("hydra.admin_endpoint"),
		viper.GetBool("hydra.secure"),
		http.SetFormat(http.URLENCODED_FORM_MIME_TYPE),
	)

	// init repositories
	accountRepo := repositories.NewAccountSQLBoiler(dbConn)
	identifierRepo := repositories.NewIdentifierSQLBoiler(dbConn)
	identityRepo := repositories.NewIdentitySQLBoiler(dbConn)
	identityProofRepo := repositories.NewIdentityProofMemory()
	hydraRepo := repositories.NewHydraHTTP(publicHydraJSON, publicHydraFORM, adminHydraJSON, adminHydraFORM)

	// init services
	accountService := account.NewAccountService(accountRepo)
	identifierService := identifier.NewIdentifierService(identifierRepo)
	identityService := identity.NewIdentityService(identityRepo, identityProofRepo)
	ssoService := application.NewSSOService(
		accountService,
		identityService,
		identifierService,
	)
	authFlowService := authflow.NewHandler(hydraRepo, viper.GetString("authflow.login_page_url"))

	// init presenters
	accountHTTP := entrypoints.NewAccountHTTP(ssoService)
	identityHTTP := entrypoints.NewIdentityHTTP(ssoService)
	authFlowHTTP := entrypoints.NewAuthFlowHTTP(authFlowService)

	authRoutes := router.Group("/auth")
	authRoutes.GET("/login", authFlowHTTP.InitLogin)

	accountRoutes := router.Group("/accounts")
	accountRoutes.POST("", accountHTTP.Create)

	identityRoutes := router.Group("/identities")
	identityRoutes.POST("/:id/assertions", identityHTTP.AssertIdentity)
	identityRoutes.POST("/identities", identityHTTP.CreateIdentity)
}

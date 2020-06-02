package sso

import (
	"database/sql"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oauth"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oidc"
	"gitlab.misakey.dev/misakey/msk-sdk-go/rester/http"

	"gitlab.misakey.dev/misakey/backend/api/src/adaptor/email"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
)

func InitModule(router *echo.Echo, dbConn *sql.DB) {
	initConfig()

	// init self authenticator for hydra rester
	selfAuth, err := oidc.NewClient(
		viper.GetString("authflow.self_client_id"),
		viper.GetString("authflow.hydra_token_url"),
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
	authnStepRepo := repositories.NewAuthenticationStepSQLBoiler(dbConn)
	hydraRepo := repositories.NewHydraHTTP(publicHydraJSON, publicHydraFORM, adminHydraJSON, adminHydraFORM)
	templateRepo := email.NewTemplateFileSystem(viper.GetString("mail.templates"))
	var emailRepo email.Sender
	env := os.Getenv("ENV")
	if env == "development" {
		emailRepo = email.NewLogMailer()
	} else if env == "production" {
		emailRepo = email.NewMailerAmazonSES(viper.GetString("aws.ses_region"))
	} else {
		log.Fatal().Msg("wrong env value")
	}
	emailRenderer, err := email.NewEmailRenderer(
		templateRepo,
		[]string{
			"code_html", "code_txt",
		},
		viper.GetString("mail.from"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("email renderer")
	}

	// init services
	accountService := account.NewAccountService(accountRepo)
	identifierService := identifier.NewIdentifierService(identifierRepo)
	identityService := identity.NewIdentityService(identityRepo)
	authFlowService := authflow.NewAuthFlowService(
		hydraRepo,
		viper.GetString("authflow.login_page_url"),
		viper.GetString("authflow.consent_page_url"),
	)
	authenticationService := authn.NewService(
		authnStepRepo,
		identifierService,
		identityService,
		emailRenderer, emailRepo,
	)
	ssoService := application.NewSSOService(
		accountService,
		identityService,
		identifierService,
		authFlowService,
		authenticationService,
	)
	oauthCodeFlow, err := oauth.NewAuthorizationCodeFlow(
		viper.GetString("authflow.self_client_id"),
		viper.GetString("authflow.auth_url"),
		viper.GetString("authflow.code_redirect_url"),
		publicHydraFORM,
		viper.GetString("authflow.hydra_token_url"),
		viper.GetString("authflow.token_redirect_url"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("oauth authorization code flow")
	}

	// init authorization middleware
	authzMidlw := authz.NewTokenIntrospectionMidlw(
		viper.GetString("authflow.self_client_id"),
		adminHydraFORM,
	)

	// bind all routes to the router
	initRoutes(router, authzMidlw, ssoService, *oauthCodeFlow)
}

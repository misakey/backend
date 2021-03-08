package sso

import (
	"fmt"
	"os"

	"github.com/duo-labs/webauthn/webauthn"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/db"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mredis"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oauth"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester/http"

	"gitlab.misakey.dev/misakey/backend/api/src/notifications/email"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// InitModule ...
func InitModule(router *echo.Echo) Process {
	initConfig()

	// init db connections
	dbConn, err := db.NewPSQLConn(
		os.Getenv("DSN_SSO"),
		viper.GetInt("sql.max_open_connections"),
		viper.GetInt("sql.max_idle_connections"),
		viper.GetDuration("sql.conn_max_lifetime"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to db")
	}
	// init db connections
	boxDBConn, err := db.NewPSQLConn(
		os.Getenv("DSN_BOX"),
		viper.GetInt("sql.max_open_connections"),
		viper.GetInt("sql.max_idle_connections"),
		viper.GetDuration("sql.conn_max_lifetime"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to db")
	}

	// init redis connection
	redConn := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.address"), viper.GetString("redis.port")),
		Password: "",
		DB:       0,
	})
	if _, err := redConn.Ping().Result(); err != nil {
		log.Fatal().Err(err).Msg("could not connect to redis")
	}

	// assign self client id to a variable since used many times
	selfCliID := viper.GetString("authflow.self_client_id")

	// init simple key redis components
	simpleKeyRedis := mredis.NewSimpleKeyRedis(redConn)

	// init self authenticator for hydra rester
	selfAuth, err := oidc.NewClient(
		selfCliID,
		viper.GetString("authflow.hydra_token_url"),
		viper.GetString("authflow.self_encoded_jwk"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not create the oidc client")
	}

	// init webauthn handler
	webauthnHandler, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "Misakey",
		RPID:          viper.GetString("server.domain"),
		RPOrigin:      viper.GetString("authflow.home_page_url"),
		RPIcon:        "https://static.misakey.com/ssoClientsLogo/misakey.png",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("could not init webauthn handler")
	}

	// init hydra resters
	publicURL, adminURL := viper.GetString("hydra.public_endpoint"), viper.GetString("hydra.admin_endpoint")
	secure := viper.GetBool("hydra.secure")

	publicHydraJSON := http.NewClient(publicURL, secure)
	adminHydraJSON := http.NewClient(adminURL, secure)
	adminHydraFORM := http.NewClient(adminURL, secure, http.SetFormat(http.MimeTypeURLEncodedForm))
	protectedPublicHydraFORM := http.NewClient(publicURL, secure, http.SetFormat(http.MimeTypeURLEncodedForm), http.SetAuthenticator(oidc.NewPrivateKeyJWTAuthenticator(selfAuth)))

	// init repositories
	authnSessionRepo := authn.NewAuthnSessionRedis(simpleKeyRedis)
	authnProcessRepo := authn.NewAuthnProcessRedis(simpleKeyRedis)
	hydraRepo := authflow.NewHydraHTTP(publicHydraJSON, adminHydraJSON, adminHydraFORM, protectedPublicHydraFORM)
	templateRepo := email.NewTemplateFileSystem(viper.GetString("mail.templates"))
	var emailRepo email.Sender
	var avatarRepo identity.AvatarRepo
	env := os.Getenv("ENV")
	if env == "development" {
		emailRepo = email.NewLogMailer()
		avatarRepo = identity.NewAvatarFileSystem(viper.GetString("server.avatars"), viper.GetString("server.avatar_url"))
	} else if env == "production" {
		emailRepo = email.NewMailerAmazonSES(viper.GetString("aws.ses_region"), viper.GetString("aws.ses_configuration_set"))
		avatarRepo, err = identity.NewAvatarAmazonS3(viper.GetString("aws.s3_region"), viper.GetString("aws.user_content_bucket"))
		if err != nil {
			log.Fatal().Msg("could not initiate AWS S3 avatar bucket connection")
		}
	} else {
		log.Fatal().Msg("unknown ENV value (should be production|development)")
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
	identityService := identity.NewService(avatarRepo, dbConn)
	authFlowService := authflow.NewService(
		identityService, hydraRepo,
		viper.GetString("authflow.home_page_url"),
		viper.GetString("authflow.login_page_url"),
		viper.GetString("authflow.consent_page_url"),
		selfCliID,
	)
	authenticationService := authn.NewService(
		authnSessionRepo, authnProcessRepo,
		emailRenderer, emailRepo,
		webauthnHandler, viper.GetString("authflow.app_name"),
	)
	backupKeyShareService := crypto.NewBackupKeyShareService(simpleKeyRedis, viper.GetDuration("backup_key_share.expiration"))
	ssoService := application.NewSSOService(
		identityService,
		authFlowService,
		authenticationService,
		backupKeyShareService,
		viper.GetDuration("root_key_share.expiration"),
		selfCliID,

		boxDBConn, dbConn,
		redConn,
	)
	oauthCodeFlow, err := oauth.NewAuthorizationCodeFlow(
		selfCliID,
		redConn,
		viper.GetString("authflow.auth_url"),
		viper.GetString("authflow.code_redirect_url"),
		protectedPublicHydraFORM,
		viper.GetString("authflow.hydra_token_url"),
		viper.GetString("authflow.token_redirect_url"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("oauth authorization code flow")
	}

	// init authorization middlewares
	selfOnly := true
	selfOIDCAuthzMidlw := authz.NewTokenIntrospector("hydra", selfCliID, selfOnly, adminHydraFORM, redConn)
	selfOIDCHandlerFactory := request.NewHandlerFactory(selfOIDCAuthzMidlw)

	anyOIDCAuthzMidlw := authz.NewTokenIntrospector("hydra", selfCliID, !selfOnly, adminHydraFORM, redConn)
	anyOIDCHandlerFactory := request.NewHandlerFactory(anyOIDCAuthzMidlw)

	// NOTE: authnProcessIntrospector is by-design a self client id introspector
	authnProcessAuthzMidlw := authz.NewTokenIntrospector("authn_process", selfCliID, !selfOnly, simpleKeyRedis, redConn)
	authnProcessHandlerFactory := request.NewHandlerFactory(authnProcessAuthzMidlw)

	// bind all routes to the router
	bindRoutes(
		router,
		selfOIDCHandlerFactory,
		anyOIDCHandlerFactory,
		authnProcessHandlerFactory,
		&ssoService,
		*oauthCodeFlow,
	)

	// bind static assets for avatars only if configuration has been set up
	avatarLocation := viper.GetString("server.avatars")
	if len(avatarLocation) > 0 {
		router.Static("/avatars", avatarLocation)
	}

	return Process{
		IdentityIntraProcess:     identity.NewIntraprocessHelper(dbConn, redConn),
		CryptoActionIntraProcess: crypto.NewIntraprocessHelper(dbConn, redConn),
		SSOService:               &ssoService,
	}
}

// Process ...
type Process struct {
	SSOService               *application.SSOService
	IdentityIntraProcess     *identity.IntraprocessHelper
	CryptoActionIntraProcess *crypto.IntraprocessHelper
}

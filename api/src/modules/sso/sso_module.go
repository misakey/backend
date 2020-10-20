package sso

import (
	"fmt"
	"os"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"gitlab.misakey.dev/misakey/backend/api/src/adaptor/email"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/backuparchive"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/backupkeyshare"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/coupon"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/cryptoaction"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/db"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oauth"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester/http"
)

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

	// init redis connection
	redConn := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.address"), viper.GetString("redis.port")),
		Password: "",
		DB:       0,
	})
	if _, err := redConn.Ping().Result(); err != nil {
		log.Fatal().Err(err).Msg("could not connect to redis")
	}

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
	publicHydraJSON := http.NewClient(viper.GetString("hydra.public_endpoint"), viper.GetBool("hydra.secure"), http.SetAuthenticator(&oidc.BearerTokenAuthenticator{}))
	adminHydraJSON := http.NewClient(viper.GetString("hydra.admin_endpoint"), viper.GetBool("hydra.secure"), http.SetAuthenticator(&oidc.BearerTokenAuthenticator{}))
	publicHydraFORM := http.NewClient(
		viper.GetString("hydra.public_endpoint"),
		viper.GetBool("hydra.secure"),
		http.SetFormat(http.MimeTypeURLEncodedForm),
		http.SetAuthenticator(oidc.NewPrivateKeyJWTAuthenticator(selfAuth)),
	)
	adminHydraFORM := http.NewClient(
		viper.GetString("hydra.admin_endpoint"),
		viper.GetBool("hydra.secure"),
		http.SetFormat(http.MimeTypeURLEncodedForm),
		http.SetAuthenticator(&oidc.BearerTokenAuthenticator{}),
	)

	// init repositories
	accountRepo := repositories.NewAccountSQLBoiler(dbConn)
	identifierRepo := repositories.NewIdentifierSQLBoiler(dbConn)
	identityRepo := repositories.NewIdentitySQLBoiler(dbConn)
	authnStepRepo := authn.NewAuthnStepSQLBoiler(dbConn)
	backupArchiveRepo := repositories.NewBackupArchiveSQLBoiler(dbConn)
	usedCouponRepo := repositories.NewUsedCouponSQLBoiler(dbConn)
	cryptoActionRepo := repositories.NewCryptoActionSQLBoiler(dbConn)
	backupKeyRepo := backupkeyshare.NewRedisRepo(redConn, viper.GetDuration("backup_key_share.expiration"))
	authnSessionRepo := authn.NewAuthnSessionRedis(redConn)
	authnProcessRepo := authn.NewAuthnProcessRedis(viper.GetString("authflow.self_client_id"), redConn)
	hydraRepo := authflow.NewHydraHTTP(publicHydraJSON, publicHydraFORM, adminHydraJSON, adminHydraFORM)
	templateRepo := email.NewTemplateFileSystem(viper.GetString("mail.templates"))
	var emailRepo email.Sender
	var avatarRepo identity.AvatarRepo
	env := os.Getenv("ENV")
	if env == "development" {
		emailRepo = email.NewLogMailer()
		avatarRepo = repositories.NewAvatarFileSystem(viper.GetString("server.avatars"), viper.GetString("server.avatar_url"))
	} else if env == "production" {
		emailRepo = email.NewMailerAmazonSES(viper.GetString("aws.ses_region"), viper.GetString("aws.ses_configuration_set"),)
		avatarRepo, err = repositories.NewAvatarAmazonS3(viper.GetString("aws.s3_region"), viper.GetString("aws.user_content_bucket"))
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
	accountService := account.NewAccountService(accountRepo)
	identifierService := identifier.NewIdentifierService(identifierRepo)
	identityService := identity.NewIdentityService(identityRepo, avatarRepo, identifierService)
	backupArchiveService := backuparchive.NewBackupArchiveService(backupArchiveRepo)
	usedCouponService := coupon.NewUsedCouponService(usedCouponRepo)
	cryptoActionService := cryptoaction.NewCryptoActionService(cryptoActionRepo)
	authFlowService := authflow.NewAuthFlowService(
		identityService, hydraRepo,
		viper.GetString("authflow.login_page_url"),
		viper.GetString("authflow.consent_page_url"),
		viper.GetString("authflow.self_client_id"),
	)
	authenticationService := authn.NewService(
		authnStepRepo, authnSessionRepo, authnProcessRepo,
		identifierService,
		identityService,
		accountService,
		emailRenderer, emailRepo,
	)
	backupKeyShareService := backupkeyshare.NewBackupKeyShareService(backupKeyRepo)
	ssoService := application.NewSSOService(
		accountService,
		identityService,
		identifierService,
		authFlowService,
		authenticationService,
		backupKeyShareService,
		backupArchiveService,
		usedCouponService,
		cryptoActionService,
	)
	oauthCodeFlow, err := oauth.NewAuthorizationCodeFlow(
		viper.GetString("authflow.self_client_id"),
		redConn,
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
	oidcAuthzMidlw := authz.NewOIDCIntrospector(
		viper.GetString("authflow.self_client_id"),
		true,
		adminHydraFORM,
		redConn,
		true,
	)

	extOIDCAuthzMidlw := authz.NewOIDCIntrospector(
		viper.GetString("authflow.self_client_id"),
		false,
		adminHydraFORM,
		redConn,
		false,
	)

	authnProcessAuthzMidlw := authz.NewAuthnProcessIntrospector(viper.GetString("authflow.self_client_id"), authnProcessRepo)

	oidcHandlerFactory := request.NewHandlerFactory(oidcAuthzMidlw)
	authnProcessHandlerFactory := request.NewHandlerFactory(authnProcessAuthzMidlw)
	extOIDCHandlerFactory := request.NewHandlerFactory(extOIDCAuthzMidlw)

	// bind all routes to the router
	bindRoutes(
		router,
		oidcHandlerFactory,
		authnProcessHandlerFactory,
		extOIDCHandlerFactory,
		&ssoService,
		*oauthCodeFlow,
	)
	// bind static assets for avatars only if configuration has been set up
	avatarLocation := viper.GetString("server.avatars")
	if len(avatarLocation) > 0 {
		router.Static("/avatars", avatarLocation)
	}

	return Process{
		SSOService:   &ssoService,
		IdentityRepo: entrypoints.NewIdentityIntraprocess(identityService),
	}
}

type Process struct {
	SSOService   *application.SSOService
	IdentityRepo entrypoints.IdentityIntraprocess
}

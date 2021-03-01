package box

import (
	"fmt"
	"os"

	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/db"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/application"
	bentrypoints "gitlab.misakey.dev/misakey/backend/api/src/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester/http"
)

// InitModule ...
func InitModule(
	router *echo.Echo,
	identityRepo external.IdentityRepo,
	cryptoRepo external.CryptoRepo,
) Process {
	// init the box module configuration
	initConfig()

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

	ssoDBConn, err := db.NewPSQLConn(
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

	var filesRepo files.FileStorageRepo
	env := os.Getenv("ENV")
	if env == "development" {
		filesRepo = files.NewFileSystem(viper.GetString("server.encrypted_files"))
	} else if env == "production" {
		filesRepo = files.NewFileAmazonS3(
			viper.GetString("aws.s3_region"),
			viper.GetString("aws.encrypted_files_bucket"),
		)
	} else {
		log.Fatal().Msg("unknown ENV value (should be production|development)")
	}

	// assign self client id to a variable since used many times
	selfCliID := viper.GetString("authflow.self_client_id")

	boxService := application.NewBoxApplication(boxDBConn, ssoDBConn, redConn, filesRepo, selfCliID, identityRepo, cryptoRepo)
	wsHandler := bentrypoints.NewWebsocketHandler(viper.GetStringSlice("websockets.allowed_origins"), &boxService)

	adminHydraFORM := http.NewClient(
		viper.GetString("hydra.admin_endpoint"),
		viper.GetBool("hydra.secure"),
		http.SetFormat(http.MimeTypeURLEncodedForm),
		http.SetAuthenticator(&oidc.BearerTokenAuthenticator{}),
	)

	// init authorization middleware
	selfOnly := true
	selfOIDCAuthzMidlw := authz.NewTokenIntrospector("hydra", selfCliID, selfOnly, adminHydraFORM, redConn)
	selfOIDCHandlerFactory := request.NewHandlerFactory(selfOIDCAuthzMidlw)

	anyOIDCAuthzMidlw := authz.NewTokenIntrospector("hydra", selfCliID, !selfOnly, adminHydraFORM, redConn)

	// bind all routes to the router
	bindRoutes(
		router,
		&boxService,
		wsHandler,
		selfOIDCHandlerFactory,
		anyOIDCAuthzMidlw,
	)

	return Process{
		BoxService: &boxService,
	}
}

// Process ...
type Process struct {
	BoxService *application.BoxApplication
}

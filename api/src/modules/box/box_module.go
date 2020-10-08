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

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/application"
	bentrypoints "gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester/http"
)

func InitModule(router *echo.Echo) Process {
	// init the box module configuration
	initConfig()

	// init db connections
	dbConn, err := db.NewPSQLConn(
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

	boxService := application.NewBoxApplication(dbConn, redConn, filesRepo)
	wsHandler := bentrypoints.NewWebsocketHandler(viper.GetStringSlice("websockets.allowed_origins"), &boxService)
	quotumIntraprocess := bentrypoints.NewQuotumIntraprocess(boxService)

	adminHydraFORM := http.NewClient(
		viper.GetString("hydra.admin_endpoint"),
		viper.GetBool("hydra.secure"),
		http.SetFormat(http.URLENCODED_FORM_MIME_TYPE),
		http.SetAuthenticator(&oidc.BearerTokenAuthenticator{}),
	)

	// init authorization middleware
	authzMidlw := authz.NewOIDCIntrospector(
		viper.GetString("authflow.self_client_id"),
		true,
		adminHydraFORM,
	)

	bindRoutes(
		router,
		&boxService,
		wsHandler,
		request.NewHandlerFactory(authzMidlw),
		authzMidlw,
	)
	return Process{
		BoxService:         &boxService,
		QuotumIntraprocess: &quotumIntraprocess,
	}
}

type Process struct {
	BoxService         *application.BoxApplication
	QuotumIntraprocess bentrypoints.QuotaIntraprocessInterface
}

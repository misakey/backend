package box

import (
	"fmt"
	"os"

	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/msk-sdk-go/db"
	"gitlab.misakey.dev/misakey/msk-sdk-go/rester/http"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories"
)

type handler struct {
	repo repositories.RealWorld
}

func InitModule(router *echo.Echo, identityIntraprocess entrypoints.IdentityIntraprocessInterface) {
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

	adminHydraFORM := http.NewClient(
		viper.GetString("hydra.admin_endpoint"),
		viper.GetBool("hydra.secure"),
		http.SetFormat(http.URLENCODED_FORM_MIME_TYPE),
	)

	// init authorization middleware
	authzMidlw := authz.NewTokenIntrospectionMidlw(
		viper.GetString("authflow.self_client_id"),
		adminHydraFORM,
	)

	var repo repositories.RealWorld
	env := os.Getenv("ENV")
	if env == "development" {
		repo = repositories.NewRealWorld(
			dbConn,
			identityIntraprocess,
			repositories.NewBoxFileSystem(viper.GetString("server.encrypted_files")),
			repositories.NewEventsCountRedis(
				redConn,
			),
		)
	} else if env == "production" {
		repo = repositories.NewRealWorld(
			dbConn,
			identityIntraprocess,
			repositories.NewBoxFileAmazonS3(
				viper.GetString("aws.s3_region"),
				viper.GetString("aws.encrypted_files_bucket"),
			),
			repositories.NewEventsCountRedis(
				redConn,
			),
		)
	} else {
		log.Fatal().Msg("unknown ENV value (should be production|development)")
	}

	h := handler{repo: repo}
	bindRoutes(router, h, authzMidlw)
}

package box

import (
	"context"
	"database/sql"
	"io"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/msk-sdk-go/db"
	"gitlab.misakey.dev/misakey/msk-sdk-go/rester/http"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
)

type handler struct {
	db           *sql.DB
	identityRepo entrypoints.IdentityIntraprocessInterface
	files        fileRepo
}

type fileRepo interface {
	Upload(context.Context, string, string, io.Reader) error
	Download(context.Context, string, string) ([]byte, error)
	Delete(context.Context, string, string) error
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

	var files fileRepo
	env := os.Getenv("ENV")
	if env == "development" {
		files = repositories.NewBoxFileSystem(viper.GetString("server.encrypted_files"))
	} else if env == "production" {
		files, err = repositories.NewBoxFileAmazonS3(viper.GetString("aws.s3_region"), viper.GetString("aws.encrypted_files_bucket"))
		if err != nil {
			log.Fatal().Msg("could not initiate AWS S3 avatar bucket connection")
		}
	} else {
		log.Fatal().Msg("unknown ENV value (should be production|development)")
	}

	h := handler{
		db:           dbConn,
		identityRepo: identityIntraprocess,
		files:        files,
	}

	bindRoutes(router, h, authzMidlw)
}

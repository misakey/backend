package box

import (
	"database/sql"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/msk-sdk-go/db"
	"gitlab.misakey.dev/misakey/msk-sdk-go/rester/http"
)

type handler struct {
	DB              *sql.DB
	IdentityService entrypoints.IdentityIntraprocessInterface
}

func InitModule(router *echo.Echo, identityIntraprocess entrypoints.IdentityIntraprocessInterface) {
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

	h := handler{
		DB:              dbConn,
		IdentityService: identityIntraprocess,
	}

	bindRoutes(router, h, authzMidlw)
}

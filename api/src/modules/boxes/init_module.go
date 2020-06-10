package boxes

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/msk-sdk-go/rester/http"
)

type handler struct {
	DB              *sql.DB
	IdentityService entrypoints.IdentityIntraprocessInterface
}

func InitModule(router *echo.Echo, dbConn *sql.DB, identityIntraprocess entrypoints.IdentityIntraprocessInterface) {
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

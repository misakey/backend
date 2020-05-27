package boxes

import (
	"database/sql"

	"github.com/labstack/echo/v4"
)

type handler struct {
	DB *sql.DB
}

func InitModule(router *echo.Echo, dbConn *sql.DB) {
	h := handler{dbConn}

	bindRoutes(router, h)
}

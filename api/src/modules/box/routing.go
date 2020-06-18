package box

import (
	"github.com/labstack/echo/v4"
)

func bindRoutes(router *echo.Echo, h handler, authzMidlw echo.MiddlewareFunc) {
	boxRouter := router.Group("/boxes", authzMidlw)
	boxRouter.GET("/:id", h.getBox)
	boxRouter.HEAD("", h.countBoxes)
	boxRouter.GET("", h.listBoxes)
	boxRouter.POST("", h.CreateBox)
	boxRouter.GET("/:id/events", h.listEvents)
	boxRouter.POST("/:id/events", h.postEvent)
}

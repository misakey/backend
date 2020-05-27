package boxes

import "github.com/labstack/echo/v4"

func bindRoutes(router *echo.Echo, h handler) {
	boxRouter := router.Group("/boxes")
	boxRouter.POST("", h.CreateBox)
	boxRouter.GET("/:id/events", h.listEvents)
	boxRouter.POST("/:id/events", h.postEvent)
}

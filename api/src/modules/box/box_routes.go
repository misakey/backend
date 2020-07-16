package box

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
)

func bindRoutes(router *echo.Echo, bs application.BoxApplication, authzMidlw echo.MiddlewareFunc) {
	boxRouter := router.Group("/boxes", authzMidlw)

	getBox := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.ReadBoxRequest{} },
		bs.ReadBox,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id", getBox)

	countBoxes := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.CountBoxesRequest{} },
		bs.CountBoxes,
		entrypoints.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	)
	boxRouter.HEAD("", countBoxes)

	listBoxes := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.ListBoxesRequest{} },
		bs.ListBoxes,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("", listBoxes)

	createBox := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.CreateBoxRequest{} },
		bs.CreateBox,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("", createBox)

	listEvents := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.ListEventsRequest{} },
		bs.ListEvents,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id/events", listEvents)

	postEvents := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.CreateEventRequest{} },
		bs.CreateEvent,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("/:id/events", postEvents)

	uploadEncryptedFile := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.UploadEncryptedFileRequest{} },
		bs.UploadEncryptedFile,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("/:bid/encrypted-files", uploadEncryptedFile)

	downloadEncryptedFile := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.DownloadEncryptedFileRequest{} },
		bs.DownloadEncryptedFile,
		entrypoints.ResponseBlob,
	)
	boxRouter.GET("/:bid/encrypted-files/:eid", downloadEncryptedFile)

	newEventsCount := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.AckNewEventsCountRequest{} },
		bs.AckNewEventsCount,
		entrypoints.ResponseNoContent,
	)
	boxRouter.PUT("/:id/new-events-count/ack", newEventsCount)

	keyShareRouter := router.Group("/box-key-shares", authzMidlw)
	createKeyShare := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.CreateKeyShareRequest{} },
		bs.CreateKeyShare,
		entrypoints.ResponseCreated,
	)
	keyShareRouter.POST("", createKeyShare)

	getKeyShare := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.GetKeyShareRequest{} },
		bs.GetKeyShare,
		entrypoints.ResponseOK,
	)
	keyShareRouter.GET("/:other-share-hash", getKeyShare)
}

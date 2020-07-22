package box

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
)

func bindRoutes(router *echo.Echo, bs application.BoxApplication, authzMidlw echo.MiddlewareFunc) {
	// Boxes
	boxRouter := router.Group("/boxes")

	getBox := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.ReadBoxRequest{} },
		true,
		bs.ReadBox,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id", getBox, authzMidlw)

	getBoxPublic := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.ReadBoxPublicRequest{} },
		false,
		bs.ReadBoxPublic,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id/public", getBoxPublic)

	countBoxes := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.CountBoxesRequest{} },
		true,
		bs.CountBoxes,
		entrypoints.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	)
	boxRouter.HEAD("", countBoxes, authzMidlw)

	listBoxes := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.ListBoxesRequest{} },
		true,
		bs.ListBoxes,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("", listBoxes, authzMidlw)

	createBox := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.CreateBoxRequest{} },
		true,
		bs.CreateBox,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("", createBox, authzMidlw)

	listEvents := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.ListEventsRequest{} },
		true,
		bs.ListEvents,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id/events", listEvents, authzMidlw)

	postEvents := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.CreateEventRequest{} },
		true,
		bs.CreateEvent,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("/:id/events", postEvents, authzMidlw)

	uploadEncryptedFile := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.UploadEncryptedFileRequest{} },
		true,
		bs.UploadEncryptedFile,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("/:bid/encrypted-files", uploadEncryptedFile, authzMidlw)

	downloadEncryptedFile := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.DownloadEncryptedFileRequest{} },
		true,
		bs.DownloadEncryptedFile,
		entrypoints.ResponseBlob,
	)
	boxRouter.GET("/:bid/encrypted-files/:eid", downloadEncryptedFile, authzMidlw)

	newEventsCount := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.AckNewEventsCountRequest{} },
		true,
		bs.AckNewEventsCount,
		entrypoints.ResponseNoContent,
	)
	boxRouter.PUT("/:id/new-events-count/ack", newEventsCount, authzMidlw)

	// Key Shares
	keyShareRouter := router.Group("/box-key-shares", authzMidlw)

	createKeyShare := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.CreateKeyShareRequest{} },
		true,
		bs.CreateKeyShare,
		entrypoints.ResponseCreated,
	)
	keyShareRouter.POST("", createKeyShare)

	getKeyShare := entrypoints.NewHTTPEntrypoint(
		func() entrypoints.Request { return &application.GetKeyShareRequest{} },
		true,
		bs.GetKeyShare,
		entrypoints.ResponseOK,
	)
	keyShareRouter.GET("/:other-share-hash", getKeyShare)
}

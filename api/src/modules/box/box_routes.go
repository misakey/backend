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

	// ----------------------
	// Boxes related routes

	createBox := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.CreateBoxRequest{} },
		bs.CreateBox,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("", createBox, authzMidlw)

	getBox := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.ReadBoxRequest{} },
		bs.ReadBox,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id", getBox, authzMidlw)

	getBoxPublic := entrypoints.NewPublicHTTP(
		func() entrypoints.Request { return &application.ReadBoxPublicRequest{} },
		bs.ReadBoxPublic,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id/public", getBoxPublic)

	countBoxes := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.CountBoxesRequest{} },
		bs.CountBoxes,
		entrypoints.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	)
	boxRouter.HEAD("/joined", countBoxes, authzMidlw)
	boxRouter.HEAD("", countBoxes, authzMidlw)

	listBoxes := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.ListBoxesRequest{} },
		bs.ListBoxes,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/joined", listBoxes, authzMidlw)
	boxRouter.GET("", listBoxes, authzMidlw)

	listBoxMembers := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.ListBoxMembersRequest{} },
		bs.ListBoxMembers,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id/members", listBoxMembers, authzMidlw)

	deleteBox := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.DeleteBoxRequest{} },
		bs.DeleteBox,
		entrypoints.ResponseNoContent,
	)
	boxRouter.DELETE("/:id", deleteBox, authzMidlw)

	// ----------------------
	// Access related routes
	listAccesses := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.ListAccessesRequest{} },
		bs.ListAccesses,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id/accesses", listAccesses, authzMidlw)

	// ----------------------
	// Events related routes

	listEvents := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.ListEventsRequest{} },
		bs.ListEvents,
		entrypoints.ResponseOK,
	)
	boxRouter.GET("/:id/events", listEvents, authzMidlw)

	countEvents := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.CountEventsRequest{} },
		bs.CountEvents,
		entrypoints.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	)
	boxRouter.HEAD("/:id/events", countEvents, authzMidlw)

	postEvents := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.CreateEventRequest{} },
		bs.CreateEvent,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("/:id/events", postEvents, authzMidlw)

	uploadEncryptedFile := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.UploadEncryptedFileRequest{} },
		bs.UploadEncryptedFile,
		entrypoints.ResponseCreated,
	)
	boxRouter.POST("/:bid/encrypted-files", uploadEncryptedFile, authzMidlw)

	newEventsCount := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.AckNewEventsCountRequest{} },
		bs.AckNewEventsCount,
		entrypoints.ResponseNoContent,
	)
	boxRouter.PUT("/:id/new-events-count/ack", newEventsCount, authzMidlw)

	// ----------------------
	// Box Key Shares related routes

	keyShareRouter := router.Group("/box-key-shares", authzMidlw)

	createKeyShare := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.CreateKeyShareRequest{} },
		bs.CreateKeyShare,
		entrypoints.ResponseCreated,
	)
	keyShareRouter.POST("", createKeyShare)

	getKeyShare := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.GetKeyShareRequest{} },
		bs.GetKeyShare,
		entrypoints.ResponseOK,
	)
	keyShareRouter.GET("/:other-share-hash", getKeyShare)

	// ----------------------
	// Saved Files related routes

	savedFileRouter := router.Group("/saved-files", authzMidlw)

	createSavedFile := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.CreateSavedFileRequest{} },
		bs.CreateSavedFile,
		entrypoints.ResponseCreated,
	)
	savedFileRouter.POST("", createSavedFile)

	deleteSavedFile := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.DeleteSavedFileRequest{} },
		bs.DeleteSavedFile,
		entrypoints.ResponseNoContent,
	)
	savedFileRouter.DELETE("/:id", deleteSavedFile)

	listSavedFiles := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.ListSavedFilesRequest{} },
		bs.ListSavedFiles,
		entrypoints.ResponseOK,
	)
	savedFileRouter.GET("", listSavedFiles)

	// ----------------------
	// Encrypted Files related routes
	encryptedFileRouter := router.Group("/encrypted-files", authzMidlw)

	downloadEncryptedFile := entrypoints.NewProtectedHTTP(
		func() entrypoints.Request { return &application.DownloadEncryptedFileRequest{} },
		bs.DownloadEncryptedFile,
		entrypoints.ResponseBlob,
	)
	encryptedFileRouter.GET("/:id", downloadEncryptedFile, authzMidlw)
}

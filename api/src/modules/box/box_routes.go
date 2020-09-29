package box

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
)

func bindRoutes(router *echo.Echo, bs *application.BoxApplication, wsh entrypoints.WebsocketHandler, authzMidlw echo.MiddlewareFunc) {
	// Boxes
	boxRouter := router.Group("/boxes")

	// ----------------------
	// Boxes related routes

	createBox := request.NewProtectedHTTP(
		func() request.Request { return &application.CreateBoxRequest{} },
		bs.CreateBox,
		request.ResponseCreated,
	)
	boxRouter.POST("", createBox, authzMidlw)

	getBox := request.NewProtectedHTTP(
		func() request.Request { return &application.ReadBoxRequest{} },
		bs.ReadBox,
		request.ResponseOK,
	)
	boxRouter.GET("/:id", getBox, authzMidlw)

	getBoxPublic := request.NewPublicHTTP(
		func() request.Request { return &application.ReadBoxPublicRequest{} },
		bs.ReadBoxPublic,
		request.ResponseOK,
	)
	boxRouter.GET("/:id/public", getBoxPublic)

	countBoxes := request.NewProtectedHTTP(
		func() request.Request { return &application.CountBoxesRequest{} },
		bs.CountBoxes,
		request.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	)
	boxRouter.HEAD("/joined", countBoxes, authzMidlw)

	listBoxes := request.NewProtectedHTTP(
		func() request.Request { return &application.ListBoxesRequest{} },
		bs.ListBoxes,
		request.ResponseOK,
	)
	boxRouter.GET("/joined", listBoxes, authzMidlw)

	listBoxMembers := request.NewProtectedHTTP(
		func() request.Request { return &application.ListBoxMembersRequest{} },
		bs.ListBoxMembers,
		request.ResponseOK,
	)
	boxRouter.GET("/:id/members", listBoxMembers, authzMidlw)

	deleteBox := request.NewProtectedHTTP(
		func() request.Request { return &application.DeleteBoxRequest{} },
		bs.DeleteBox,
		request.ResponseNoContent,
	)
	boxRouter.DELETE("/:id", deleteBox, authzMidlw)

	// ----------------------
	// Access related routes
	listAccesses := request.NewProtectedHTTP(
		func() request.Request { return &application.ListAccessesRequest{} },
		bs.ListAccesses,
		request.ResponseOK,
	)
	boxRouter.GET("/:id/accesses", listAccesses, authzMidlw)

	// ----------------------
	// Events related routes

	listEvents := request.NewProtectedHTTP(
		func() request.Request { return &application.ListEventsRequest{} },
		bs.ListEvents,
		request.ResponseOK,
	)
	boxRouter.GET("/:id/events", listEvents, authzMidlw)

	boxRouter.GET("/:id/events/ws", wsh.ListEventsWS, authzMidlw)

	countEvents := request.NewProtectedHTTP(
		func() request.Request { return &application.CountEventsRequest{} },
		bs.CountEvents,
		request.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	)
	boxRouter.HEAD("/:id/events", countEvents, authzMidlw)

	postEvents := request.NewProtectedHTTP(
		func() request.Request { return &application.CreateEventRequest{} },
		bs.CreateEvent,
		request.ResponseCreated,
	)
	boxRouter.POST("/:id/events", postEvents, authzMidlw)

	batchPostEvents := request.NewProtectedHTTP(
		func() request.Request { return &application.BatchCreateEventRequest{} },
		bs.BatchCreateEvent,
		request.ResponseCreated,
	)
	boxRouter.POST("/:id/batch-events", batchPostEvents, authzMidlw)

	uploadEncryptedFile := request.NewProtectedHTTP(
		func() request.Request { return &application.UploadEncryptedFileRequest{} },
		bs.UploadEncryptedFile,
		request.ResponseCreated,
	)
	boxRouter.POST("/:bid/encrypted-files", uploadEncryptedFile, authzMidlw)

	newEventsCount := request.NewProtectedHTTP(
		func() request.Request { return &application.AckNewEventsCountRequest{} },
		bs.AckNewEventsCount,
		request.ResponseNoContent,
	)
	boxRouter.PUT("/:id/new-events-count/ack", newEventsCount, authzMidlw)

	// ----------------------
	// Box Key Shares related routes

	keyShareRouter := router.Group("/box-key-shares", authzMidlw)

	createKeyShare := request.NewProtectedHTTP(
		func() request.Request { return &application.CreateKeyShareRequest{} },
		bs.CreateKeyShare,
		request.ResponseCreated,
	)
	keyShareRouter.POST("", createKeyShare)

	getKeyShare := request.NewProtectedHTTP(
		func() request.Request { return &application.GetKeyShareRequest{} },
		bs.GetKeyShare,
		request.ResponseOK,
	)
	keyShareRouter.GET("/:other-share-hash", getKeyShare)

	// ----------------------
	// Saved Files related routes

	savedFileRouter := router.Group("/saved-files", authzMidlw)

	createSavedFile := request.NewProtectedHTTP(
		func() request.Request { return &application.CreateSavedFileRequest{} },
		bs.CreateSavedFile,
		request.ResponseCreated,
	)
	savedFileRouter.POST("", createSavedFile)

	deleteSavedFile := request.NewProtectedHTTP(
		func() request.Request { return &application.DeleteSavedFileRequest{} },
		bs.DeleteSavedFile,
		request.ResponseNoContent,
	)
	savedFileRouter.DELETE("/:id", deleteSavedFile)

	listSavedFiles := request.NewProtectedHTTP(
		func() request.Request { return &application.ListSavedFilesRequest{} },
		bs.ListSavedFiles,
		request.ResponseOK,
	)
	savedFileRouter.GET("", listSavedFiles)

	// ----------------------
	// Encrypted Files related routes
	encryptedFileRouter := router.Group("/encrypted-files", authzMidlw)

	downloadEncryptedFile := request.NewProtectedHTTP(
		func() request.Request { return &application.DownloadEncryptedFileRequest{} },
		bs.DownloadEncryptedFile,
		request.ResponseBlob,
	)
	encryptedFileRouter.GET("/:id", downloadEncryptedFile, authzMidlw)

	// ----------------------
	// box-users related routes
	boxUserRouter := router.Group("/box-users", authzMidlw)

	listStorageQuota := request.NewProtectedHTTP(
		func() request.Request { return &application.ListUserStorageQuotaRequest{} },
		bs.ListUserStorageQuota,
		request.ResponseOK,
	)
	boxUserRouter.GET("/:id/storage-quota", listStorageQuota, authzMidlw)

	getUserVaultSpace := request.NewProtectedHTTP(
		func() request.Request { return &application.GetVaultUsedSpaceRequest{} },
		bs.GetVaultUsedSpace,
		request.ResponseOK,
	)
	boxUserRouter.GET("/:id/vault-used-space", getUserVaultSpace, authzMidlw)

	// ----------------------
	// box-used-space related routes
	boxUsedSpacesRouter := router.Group("/box-used-spaces", authzMidlw)

	listBoxUsedSpace := request.NewProtectedHTTP(
		func() request.Request { return &application.ListBoxUsedSpaceRequest{} },
		bs.ListBoxUsedSpace,
		request.ResponseOK,
	)
	boxUsedSpacesRouter.GET("", listBoxUsedSpace, authzMidlw)
}

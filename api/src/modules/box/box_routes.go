package box

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/application"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

func bindRoutes(
	router *echo.Echo,
	bs *application.BoxApplication,
	wsh entrypoints.WebsocketHandler,
	oidcHandlerFactory request.HandlerFactory,
	authzMidlw echo.MiddlewareFunc,
) {
	// ----------------------
	// Boxes related routes
	boxPath := router.Group("/boxes")
	boxPath.POST(oidcHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.CreateBoxRequest{} },
		bs.CreateBox,
		request.ResponseCreated,
	))
	boxPath.GET(oidcHandlerFactory.NewACR1(
		"/:id",
		func() request.Request { return &application.ReadBoxRequest{} },
		bs.ReadBox,
		request.ResponseOK,
	))
	boxPath.GET(oidcHandlerFactory.NewPublic(
		"/:id/public",
		func() request.Request { return &application.ReadBoxPublicRequest{} },
		bs.ReadBoxPublic,
		request.ResponseOK,
	))
	boxPath.HEAD(oidcHandlerFactory.NewACR2(
		"/joined",
		func() request.Request { return &application.CountBoxesRequest{} },
		bs.CountBoxes,
		request.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	))
	boxPath.GET(oidcHandlerFactory.NewACR2(
		"/joined",
		func() request.Request { return &application.ListBoxesRequest{} },
		bs.ListBoxes,
		request.ResponseOK,
	))
	boxPath.GET(oidcHandlerFactory.NewACR1(
		"/:id/members",
		func() request.Request { return &application.ListBoxMembersRequest{} },
		bs.ListBoxMembers,
		request.ResponseOK,
	))
	boxPath.DELETE(oidcHandlerFactory.NewACR2(
		"/:id",
		func() request.Request { return &application.DeleteBoxRequest{} },
		bs.DeleteBox,
		request.ResponseNoContent,
	))

	// ----------------------
	// Access related routes
	boxPath.GET(oidcHandlerFactory.NewACR2(
		"/:id/accesses",
		func() request.Request { return &application.ListAccessesRequest{} },
		bs.ListAccesses,
		request.ResponseOK,
	))

	// ----------------------
	// Events related routes
	boxPath.GET(oidcHandlerFactory.NewACR1(
		"/:id/events",
		func() request.Request { return &application.ListEventsRequest{} },
		bs.ListEvents,
		request.ResponseOK,
	))

	boxPath.HEAD(oidcHandlerFactory.NewACR1(
		"/:id/events",
		func() request.Request { return &application.CountEventsRequest{} },
		bs.CountEvents,
		request.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	))
	boxPath.POST(oidcHandlerFactory.NewACR1(
		"/:id/events",
		func() request.Request { return &application.CreateEventRequest{} },
		bs.CreateEvent,
		request.ResponseCreated,
	))
	boxPath.POST(oidcHandlerFactory.NewACR2(
		"/:id/batch-events",
		func() request.Request { return &application.BatchCreateEventRequest{} },
		bs.BatchCreateEvent,
		request.ResponseCreated,
	))
	boxPath.POST(oidcHandlerFactory.NewACR1(
		"/:bid/encrypted-files",
		func() request.Request { return &application.UploadEncryptedFileRequest{} },
		bs.UploadEncryptedFile,
		request.ResponseCreated,
	))
	boxPath.PUT(oidcHandlerFactory.NewACR1(
		"/:id/new-events-count/ack",
		func() request.Request { return &application.AckNewEventsCountRequest{} },
		bs.AckNewEventsCount,
		request.ResponseNoContent,
	))

	// ----------------------
	// Box Users related routes
	boxUsersPath := router.Group("/box-users")
	boxUsersPath.GET("/:id/ws", wsh.BoxUsersWS, authzMidlw)

	// ----------------------
	// Box Key Shares related routes
	keySharePath := router.Group("/box-key-shares")
	keySharePath.POST(oidcHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.CreateKeyShareRequest{} },
		bs.CreateKeyShare,
		request.ResponseCreated,
	))
	keySharePath.GET(oidcHandlerFactory.NewACR1(
		"/:other-share-hash",
		func() request.Request { return &application.GetKeyShareRequest{} },
		bs.GetKeyShare,
		request.ResponseOK,
	))

	// ----------------------
	// Saved Files related routes
	savedFilePath := router.Group("/saved-files")
	savedFilePath.POST(oidcHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.CreateSavedFileRequest{} },
		bs.CreateSavedFile,
		request.ResponseCreated,
	))
	savedFilePath.DELETE(oidcHandlerFactory.NewACR2(
		"/:id",
		func() request.Request { return &application.DeleteSavedFileRequest{} },
		bs.DeleteSavedFile,
		request.ResponseNoContent,
	))
	savedFilePath.GET(oidcHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.ListSavedFilesRequest{} },
		bs.ListSavedFiles,
		request.ResponseOK,
	))

	// ----------------------
	// Encrypted Files related routes
	encryptedFilePath := router.Group("/encrypted-files")
	encryptedFilePath.GET(oidcHandlerFactory.NewACR1(
		"/:id",
		func() request.Request { return &application.DownloadEncryptedFileRequest{} },
		bs.DownloadEncryptedFile,
		request.ResponseStream,
	))

	// ----------------------
	// box-users related routes
	boxUserPath := router.Group("/box-users")
	boxUserPath.GET(oidcHandlerFactory.NewACR1(
		"/:id/storage-quota",
		func() request.Request { return &application.ListUserStorageQuotaRequest{} },
		bs.ListUserStorageQuota,
		request.ResponseOK,
	))
	boxUserPath.GET(oidcHandlerFactory.NewACR1(
		"/:id/vault-used-space",
		func() request.Request { return &application.GetVaultUsedSpaceRequest{} },
		bs.GetVaultUsedSpace,
		request.ResponseOK,
	))

	// ----------------------
	// box-used-space related routes
	boxUsedSpacesPath := router.Group("/box-used-spaces")
	boxUsedSpacesPath.GET(oidcHandlerFactory.NewACR1(
		"",
		func() request.Request { return &application.ListBoxUsedSpaceRequest{} },
		bs.ListBoxUsedSpace,
		request.ResponseOK,
	))
}

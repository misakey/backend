package box

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/box/application"
	"gitlab.misakey.dev/misakey/backend/api/src/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

func bindRoutes(
	router *echo.Echo,
	app *application.BoxApplication,
	wsh entrypoints.WebsocketHandler,
	selfOIDCHandlerFactory request.HandlerFactory,
	anyOIDCHandlerFactory request.HandlerFactory,
) {
	// ----------------------
	// Boxes related routes
	boxPath := router.Group("/boxes")
	boxPath.POST(selfOIDCHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.CreateBoxRequest{} },
		app.CreateBox,
		request.ResponseCreated,
	))
	boxPath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:id",
		func() request.Request { return &application.GetBoxRequest{} },
		app.GetBox,
		request.ResponseOK,
	))
	boxPath.GET(selfOIDCHandlerFactory.NewPublic(
		"/:id/public",
		func() request.Request { return &application.GetBoxPublicRequest{} },
		app.GetBoxPublic,
		request.ResponseOK,
	))
	boxPath.HEAD(selfOIDCHandlerFactory.NewACR2(
		"/joined",
		func() request.Request { return &application.CountBoxesRequest{} },
		app.CountBoxes,
		request.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	))
	boxPath.GET(selfOIDCHandlerFactory.NewACR2(
		"/joined",
		func() request.Request { return &application.ListBoxesRequest{} },
		app.ListBoxes,
		request.ResponseOK,
	))
	boxPath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:id/members",
		func() request.Request { return &application.ListBoxMembersRequest{} },
		app.ListBoxMembers,
		request.ResponseOK,
	))
	boxPath.DELETE(selfOIDCHandlerFactory.NewACR2(
		"/:id",
		func() request.Request { return &application.DeleteBoxRequest{} },
		app.DeleteBox,
		request.ResponseNoContent,
	))

	// organizations routes
	orgPath := router.Group("/organizations")
	orgPath.POST(anyOIDCHandlerFactory.NewACR2(
		"/:oid/boxes",
		func() request.Request { return &application.CreateOrgBoxRequest{} },
		app.CreateOrgBox,
		request.ResponseCreated,
	))
	orgPath.GET(anyOIDCHandlerFactory.NewACR2(
		"/:oid/boxes/:id",
		func() request.Request { return &application.GetOrgBoxRequest{} },
		app.GetOrgBox,
		request.ResponseOK,
	))

	// ----------------------
	// Access related routes
	boxPath.GET(selfOIDCHandlerFactory.NewACR2(
		"/:id/accesses",
		func() request.Request { return &application.ListAccessesRequest{} },
		app.ListAccesses,
		request.ResponseOK,
	))

	// ----------------------
	// Events related routes
	boxPath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:id/events",
		func() request.Request { return &application.ListEventsRequest{} },
		app.ListEvents,
		request.ResponseOK,
	))

	boxPath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:id/files",
		func() request.Request { return &application.ListBoxFilesRequest{} },
		app.ListBoxFiles,
		request.ResponseOK,
	))

	boxPath.HEAD(selfOIDCHandlerFactory.NewACR1(
		"/:id/files",
		func() request.Request { return &application.CountBoxFilesRequest{} },
		app.CountBoxFiles,
		request.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	))

	boxPath.HEAD(selfOIDCHandlerFactory.NewACR1(
		"/:id/events",
		func() request.Request { return &application.CountEventsRequest{} },
		app.CountEvents,
		request.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	))
	boxPath.POST(anyOIDCHandlerFactory.NewACR1(
		"/:id/events",
		func() request.Request { return &application.CreateEventRequest{} },
		app.CreateEvent,
		request.ResponseCreated,
	))
	boxPath.POST(selfOIDCHandlerFactory.NewACR2(
		"/:id/batch-events",
		func() request.Request { return &application.BatchCreateEventRequest{} },
		app.BatchCreateEvent,
		request.ResponseCreated,
	))
	boxPath.POST(anyOIDCHandlerFactory.NewACR1(
		"/:bid/encrypted-files",
		func() request.Request { return &application.UploadEncryptedFileRequest{} },
		app.UploadEncryptedFile,
		request.ResponseCreated,
	))
	boxPath.PUT(selfOIDCHandlerFactory.NewACR1(
		"/:id/new-events-count/ack",
		func() request.Request { return &application.AckNewEventsCountRequest{} },
		app.AckNewEventsCount,
		request.ResponseNoContent,
	))

	// ----------------------
	// Box Users related routes
	boxUsersPath := router.Group("/box-users")
	boxUsersPath.GET("/:id/ws", wsh.BoxUsersWS, anyOIDCHandlerFactory.GetAuthzMdlw())
	boxUsersPath.PUT(selfOIDCHandlerFactory.NewACR1(
		"/:id/boxes/:bid/settings",
		func() request.Request { return &application.UpdateBoxSettingsRequest{} },
		app.UpdateBoxSettings,
		request.ResponseNoContent,
	))

	boxUsersPath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:id/boxes/:bid/settings",
		func() request.Request { return &application.GetBoxSettingsRequest{} },
		app.GetBoxSettings,
		request.ResponseOK,
	))

	boxUsersPath.POST(selfOIDCHandlerFactory.NewACR2(
		"/:id/contact",
		func() request.Request { return &application.BoxUserContactRequest{} },
		app.BoxUserContact,
		request.ResponseCreated,
	))

	// ----------------------
	// Box Key Shares related routes
	keySharePath := router.Group("/box-key-shares")
	keySharePath.POST(selfOIDCHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.CreateKeyShareRequest{} },
		app.CreateKeyShare,
		request.ResponseCreated,
	))
	keySharePath.GET(selfOIDCHandlerFactory.NewACR2(
		"/encrypted-invitation-key-share",
		func() request.Request { return &application.GetEncryptedKeyShareRequest{} },
		app.GetEncryptedKeyShare,
		request.ResponseOK,
	))
	keySharePath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:other-share-hash",
		func() request.Request { return &application.GetKeyShareRequest{} },
		app.GetKeyShare,
		request.ResponseOK,
	))

	// ----------------------
	// Saved Files related routes
	savedFilePath := router.Group("/saved-files")
	savedFilePath.POST(selfOIDCHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.CreateSavedFileRequest{} },
		app.CreateSavedFile,
		request.ResponseCreated,
	))
	savedFilePath.DELETE(selfOIDCHandlerFactory.NewACR2(
		"/:id",
		func() request.Request { return &application.DeleteSavedFileRequest{} },
		app.DeleteSavedFile,
		request.ResponseNoContent,
	))
	savedFilePath.GET(selfOIDCHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.ListSavedFilesRequest{} },
		app.ListSavedFiles,
		request.ResponseOK,
	))
	savedFilePath.HEAD(selfOIDCHandlerFactory.NewACR2(
		"",
		func() request.Request { return &application.CountSavedFilesRequest{} },
		app.CountSavedFiles,
		request.ResponseNoContent,
		func(ctx echo.Context, data interface{}) error {
			ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(data.(int)))
			return nil
		},
	))

	// ----------------------
	// Encrypted Files related routes
	encryptedFilePath := router.Group("/encrypted-files")
	encryptedFilePath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:id",
		func() request.Request { return &application.DownloadEncryptedFileRequest{} },
		app.DownloadEncryptedFile,
		request.ResponseStream,
	))

	// ----------------------
	// box-users related routes
	boxUserPath := router.Group("/box-users")
	boxUserPath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:id/storage-quota",
		func() request.Request { return &application.ListUserStorageQuotaRequest{} },
		app.ListUserStorageQuota,
		request.ResponseOK,
	))
	boxUserPath.GET(selfOIDCHandlerFactory.NewACR1(
		"/:id/vault-used-space",
		func() request.Request { return &application.GetVaultUsedSpaceRequest{} },
		app.GetVaultUsedSpace,
		request.ResponseOK,
	))
	boxUserPath.POST(selfOIDCHandlerFactory.NewACR2(
		"/:id/saved-files",
		func() request.Request { return &application.UploadSavedFileRequest{} },
		app.UploadSavedFile,
		request.ResponseCreated,
	))

	// ----------------------
	// box-used-space related routes
	boxUsedSpacesPath := router.Group("/box-used-spaces")
	boxUsedSpacesPath.GET(selfOIDCHandlerFactory.NewACR1(
		"",
		func() request.Request { return &application.ListBoxUsedSpaceRequest{} },
		app.ListBoxUsedSpace,
		request.ResponseOK,
	))
}

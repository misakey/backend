package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type BackupArchiveHTTP struct {
	service *application.SSOService
}

func NewBackupArchiveHTTP(service *application.SSOService) BackupArchiveHTTP {
	return BackupArchiveHTTP{service: service}
}

// ListArchives handles GET /backup-archives
func (entrypoint BackupArchiveHTTP) ListBackupArchives(ctx echo.Context) error {
	archives, err := entrypoint.service.ListBackupArchives(ctx.Request().Context())
	if err != nil {
		return merror.Transform(err).Describe("listing identities")
	}
	return ctx.JSON(http.StatusOK, archives)
}

func (entrypoint BackupArchiveHTTP) GetArchiveData(ctx echo.Context) error {
	query := application.GetBackupArchiveDataQuery{
		ArchiveID: ctx.Param("id"),
	}

	if err := query.Validate(); err != nil {
		return merror.Transform(err)
	}

	data, err := entrypoint.service.GetBackupArchiveData(ctx.Request().Context(), query.ArchiveID)
	if err != nil {
		return merror.Transform(err).Describe("service get archive data")
	}
	return ctx.JSON(http.StatusOK, data)
}

func (entrypoint BackupArchiveHTTP) DeleteArchive(ctx echo.Context) error {
	query := application.DeleteBackupArchiveQuery{
		ArchiveID: ctx.Param("id"),
	}

	if err := ctx.Bind(&query); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	if err := query.Validate(); err != nil {
		return merror.Transform(err)
	}

	err := entrypoint.service.DeleteBackupArchive(ctx.Request().Context(), query.ArchiveID, query.Reason)
	if err != nil {
		return merror.Transform(err).Describe("deleting backup archive")
	}

	return ctx.NoContent(http.StatusNoContent)
}

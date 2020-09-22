package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type BackupKeyShareHTTP struct {
	service application.SSOService
}

func NewBackupKeyShareHTTP(service application.SSOService) BackupKeyShareHTTP {
	return BackupKeyShareHTTP{service: service}
}

// Handles POST /backup-key-shares - store backup key shares
func (entrypoint BackupKeyShareHTTP) CreateBackupKeyShare(ctx echo.Context) error {
	cmd := application.CreateBackupKeyShareCmd{}

	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := entrypoint.service.BackupKeyShareCreate(ctx.Request().Context(), cmd); err != nil {
		return merror.Transform(err).Describe("create backup key share").From(merror.OriBody)
	}

	return ctx.JSON(http.StatusCreated, cmd)
}

// Handles GET /backup-key-shares/:other-share-hash - retrieve a key share
// using the hash of the other share
func (entrypoint BackupKeyShareHTTP) GetBackupKeyShare(ctx echo.Context) error {
	query := application.BackupKeyShareQuery{
		OtherShareHash: ctx.Param("other-share-hash"),
	}

	if err := query.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	backupKeyShare, err := entrypoint.service.BackupKeyShareGet(ctx.Request().Context(), query)
	if err != nil {
		return merror.Transform(err).Describe("get backup key share").From(merror.OriBody)
	}
	return ctx.JSON(http.StatusOK, backupKeyShare)
}

package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
)

// BackupArchiveView ...
type BackupArchiveView struct {
	crypto.BackupArchive
}

// ListBackupArchives ...
func (sso *SSOService) ListBackupArchives(ctx context.Context, _ request.Request) (interface{}, error) {
	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merr.Forbidden()
	}

	archives, err := crypto.ListBackupArchives(ctx, sso.ssoDB, acc.AccountID.String)
	if err != nil {
		return nil, err
	}

	views := make([]BackupArchiveView, len(archives))
	for i, archive := range archives {
		views[i].BackupArchive = archive
	}

	return views, nil
}

// BackupArchiveDataQuery ...
type BackupArchiveDataQuery struct {
	archiveID string
}

// BindAndValidate ...
func (query *BackupArchiveDataQuery) BindAndValidate(eCtx echo.Context) error {
	query.archiveID = eCtx.Param("id")

	return v.ValidateStruct(query,
		v.Field(&query.archiveID, v.Required, is.UUIDv4.Error("archive id must be uuid v4 ")),
	)
}

// GetBackupArchiveData ...
func (sso *SSOService) GetBackupArchiveData(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*BackupArchiveDataQuery)
	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return "", merr.Forbidden()
	}

	archive, err := crypto.GetBackupArchive(ctx, sso.ssoDB, query.archiveID)
	if err != nil {
		return "", merr.From(err).Desc("retrieving archive")
	}

	if acc.AccountID.String != archive.AccountID {
		return "", merr.Forbidden()
	}

	if archive.DeletedAt.Valid || archive.RecoveredAt.Valid {
		return "", merr.Gone()
	}

	return archive.Data.String, nil
}

// BackupArchiveDeleteCmd ...
type BackupArchiveDeleteCmd struct {
	archiveID string
	Reason    string `json:"reason"`
}

// BindAndValidate ...
func (cmd *BackupArchiveDeleteCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriQuery)
	}
	cmd.archiveID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.Reason, v.Required, v.In("recovery", "deletion")),
		v.Field(&cmd.archiveID, v.Required, is.UUIDv4),
	)
}

// DeleteBackupArchive ...
func (sso *SSOService) DeleteBackupArchive(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*BackupArchiveDeleteCmd)
	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merr.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.ssoDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	archive, err := crypto.GetBackupArchiveMetadata(ctx, tr, cmd.archiveID)
	if err != nil {
		return nil, merr.From(err).Desc("retrieving archive metadata")
	}

	if acc.AccountID.String != archive.AccountID {
		err = merr.Forbidden()
		return nil, err
	}
	if archive.DeletedAt.Valid || archive.RecoveredAt.Valid {
		err = merr.Gone()
		return nil, err
	}

	err = crypto.DeleteBackupArchive(ctx, tr, cmd.archiveID, cmd.Reason)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}

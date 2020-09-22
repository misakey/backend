package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type BackupArchiveView struct {
	domain.BackupArchive
}

func (sso SSOService) ListBackupArchives(ctx context.Context) ([]BackupArchiveView, error) {
	acc := ajwt.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merror.Forbidden()
	}

	archives, err := sso.backupArchiveService.ListBackupArchives(ctx, acc.AccountID.String)
	if err != nil {
		return nil, err
	}

	views := make([]BackupArchiveView, len(archives))
	for i, archive := range archives {
		views[i].BackupArchive = archive
	}

	return views, nil
}

type GetBackupArchiveDataQuery struct {
	ArchiveID string
}

func (query GetBackupArchiveDataQuery) Validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.ArchiveID, v.Required, is.UUIDv4.Error("archive id must be uuid v4 ")),
	)
}

func (sso SSOService) GetBackupArchiveData(ctx context.Context, archiveID string) (string, error) {
	acc := ajwt.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return "", merror.Forbidden()
	}

	archive, err := sso.backupArchiveService.GetBackupArchive(ctx, archiveID)
	if err != nil {
		return "", merror.Transform(err).Describe("retrieving archive")
	}

	if acc.AccountID.String != archive.AccountID {
		return "", merror.Forbidden()
	}

	if archive.DeletedAt.Valid || archive.RecoveredAt.Valid {
		return "", merror.Gone()
	}

	return archive.Data.String, nil
}

type DeleteBackupArchiveQuery struct {
	ArchiveID string
	Reason    string `json:"reason"`
}

func (query DeleteBackupArchiveQuery) Validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.Reason, v.Required, v.In("recovery", "deletion")),
		v.Field(&query.ArchiveID, v.Required, is.UUIDv4),
	)
}

func (sso SSOService) DeleteBackupArchive(ctx context.Context, archiveID string, reason string) error {
	acc := ajwt.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return merror.Forbidden()
	}

	archive, err := sso.backupArchiveService.GetBackupArchiveMetadata(ctx, archiveID)
	if err != nil {
		return merror.Transform(err).Describe("retrieving archive metadata")
	}

	if acc.AccountID.String != archive.AccountID {
		return merror.Forbidden()
	}
	if archive.DeletedAt.Valid || archive.RecoveredAt.Valid {
		return merror.Gone()
	}

	return sso.backupArchiveService.DeleteBackupArchive(ctx, archiveID, reason)
}

package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initCreateBackupArchiveTable() {
	goose.AddMigration(upCreateBackupArchiveTable, downCreateBackupArchiveTable)
}

func upCreateBackupArchiveTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE backup_archive(
			id UUID PRIMARY KEY,
			account_id UUID REFERENCES account ON DELETE CASCADE NOT NULL,
			data TEXT,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			recovered_at timestamptz,
			deleted_at timestamptz
		)
	`)
	return err
}

func downCreateBackupArchiveTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE backup_archive;`)
	return err
}

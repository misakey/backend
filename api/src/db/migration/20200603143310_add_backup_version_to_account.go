package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(UpAddBackupVersionToAccount, DownAddBackupVersionToAccount)
}

func UpAddBackupVersionToAccount(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		ADD COLUMN backup_version INTEGER NOT NULL DEFAULT 1;
	`)
	return err
}

func DownAddBackupVersionToAccount(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		DROP COLUMN backup_verison;
	`)
	return err
}

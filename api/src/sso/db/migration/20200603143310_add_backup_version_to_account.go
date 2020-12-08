package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initAddBackupVersionToAccount() {
	goose.AddMigration(upAddBackupVersionToAccount, downAddBackupVersionToAccount)
}

func upAddBackupVersionToAccount(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		ADD COLUMN backup_version INTEGER NOT NULL DEFAULT 1;
	`)
	return err
}

func downAddBackupVersionToAccount(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		DROP COLUMN backup_version;
	`)
	return err
}

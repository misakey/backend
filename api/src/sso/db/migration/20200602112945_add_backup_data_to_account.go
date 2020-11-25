package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initAddBackupDataToAccount() {
	goose.AddMigration(upAddBackupDataToAccount, downAddBackupDataToAccount)
}

func upAddBackupDataToAccount(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		ADD COLUMN backup_data VARCHAR NOT NULL DEFAULT '';
`)
	return err
}

func downAddBackupDataToAccount(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		DROP COLUMN backup_data;
`)
	return err
}

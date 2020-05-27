package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(UpAddBackupDataToAccount, DownAddBackupDataToAccount)
}

func UpAddBackupDataToAccount(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		ADD COLUMN backup_data VARCHAR NOT NULL DEFAULT '';
`)
	return err
}

func DownAddBackupDataToAccount(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		DROP COLUMN backup_data;
`)
	return err
}

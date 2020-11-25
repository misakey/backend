package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initRenameKeyShareTable() {
	goose.AddMigration(upRenameKeyShareTable, downRenameKeyShareTable)
}

func upRenameKeyShareTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		ALTER TABLE key_share
		RENAME TO box_key_share;
	`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		ALTER TABLE box_key_share
		RENAME COLUMN invitation_hash TO other_share_hash;
	`); err != nil {
		return err
	}

	return nil
}

func downRenameKeyShareTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		ALTER TABLE box_key_share
		RENAME TO key_share;
	`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		ALTER TABLE box_key_share
		RENAME COLUMN other_share_hash TO invitation_hash;
	`); err != nil {
		return err
	}

	return nil
}

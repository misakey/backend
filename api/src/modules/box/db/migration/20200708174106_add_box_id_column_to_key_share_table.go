package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func initAddBoxIDColumnToKeyShareTable() {
	goose.AddMigration(upAddBoxIDColumnToKeyShareTable, downAddBoxIDColumnToKeyShareTable)
}

func upAddBoxIDColumnToKeyShareTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`
			DELETE FROM key_share;
	`); err != nil {
		return fmt.Errorf("flushing the current key-shares: %v", err)
	}
	if _, err := tx.Exec(`
		ALTER TABLE key_share
			ADD COLUMN box_id UUID NOT NULL,
			ADD COLUMN creator_id UUID NOT NULL;
	`); err != nil {
		return fmt.Errorf("adding box_id column: %v", err)
	}
	return nil
}

func downAddBoxIDColumnToKeyShareTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE key_share
			DROP COLUMN box_id,
			DROP COLUMN creator_id;
	`)
	return err
}

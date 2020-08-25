package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initAddColorToUserIdentity() {
	goose.AddMigration(upAddColorToUserIdentity, downAddColorToUserIdentity)
}

func upAddColorToUserIdentity(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE identity
	ADD COLUMN color VARCHAR(8) NOT NULL DEFAULT '';
`)
	return err
}

func downAddColorToUserIdentity(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE identity
		DROP COLUMN color;
	`)
	return err
}

package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initCreateKeyShareTable() {
	goose.AddMigration(upCreateKeyShareTable, downCreateKeyShareTable)
}

func upCreateKeyShareTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE key_share(
			invitation_hash VARCHAR(255) PRIMARY KEY NOT NULL,
			share VARCHAR(255) NOT NULL,
			created_at timestamptz NOT NULL
		);
	`)
	return err
}

func downCreateKeyShareTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE key_share;`)
	return err
}

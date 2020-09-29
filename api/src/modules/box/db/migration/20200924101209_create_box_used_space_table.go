package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initCreateBoxUsedSpaceTable() {
	goose.AddMigration(upCreateBoxUsedSpaceTable, downCreateBoxUsedSpaceTable)
}

func upCreateBoxUsedSpaceTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE box_used_space(
		id UUID PRIMARY KEY,
		box_id UUID UNIQUE NOT NULL,
		value BIGINT NOT NULL,
		created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	return err
}

func downCreateBoxUsedSpaceTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE box_used_space;`)
	return err
}

package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initAddEventsTable() {
	goose.AddMigration(upAddEventsTable, downAddEventsTable)
}

func upAddEventsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE event(
		id UUID PRIMARY KEY,
		box_id UUID NOT NULL,
		created_at timestamptz NOT NULL,
		sender_id UUID NOT NULL,
		type VARCHAR(127) NOT NULL,
		content JSON
);`)
	return err
}

func downAddEventsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE event;`)
	return err
}

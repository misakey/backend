package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initCreateSavedFileTable() {
	goose.AddMigration(upCreateSavedFileTable, downCreateSavedFileTable)
}

func upCreateSavedFileTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE saved_file(
		id UUID PRIMARY KEY,
		identity_id UUID NOT NULL,
		encrypted_file_id UUID REFERENCES encrypted_file NOT NULL,
		encrypted_metadata TEXT NOT NULL,
		key_fingerprint VARCHAR(255) NOT NULL,
		created_at timestamptz NOT NULL
	);`)
	return err
}

func downCreateSavedFileTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE saved_file;`)
	return err
}

package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initCreateEncryptedFileTable() {
	goose.AddMigration(upCreateEncryptedFileTable, downCreateEncryptedFileTable)
}

func upCreateEncryptedFileTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE encrypted_file(
		id UUID PRIMARY KEY,
		size BIGINT NOT NULL,
		created_at timestamptz NOT NULL
	);`)
	return err
}

func downCreateEncryptedFileTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE encrypted_file;`)
	return err
}

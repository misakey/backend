package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initCreateUniqueIndexOnSavedFileTable() {
	goose.AddMigration(upCreateUniqueIndexOnSavedFileTable, downCreateUniqueIndexOnSavedFileTable)
}

func upCreateUniqueIndexOnSavedFileTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE UNIQUE INDEX saved_file_encrypted_file_id_identity_id_idx
		ON saved_file (encrypted_file_id, identity_id);`)
	return err
}

func downCreateUniqueIndexOnSavedFileTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP INDEX saved_file_encrypted_file_id_identity_id_idx;`)
	return err
}

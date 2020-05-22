package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upCreateIdentifierTable, downCreateIdentifierTable)
}

func upCreateIdentifierTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE identifier(
		  id UUID PRIMARY KEY,
		  kind VARCHAR(32) NOT NULL,
		  value VARCHAR(255) UNIQUE NOT NULL
	);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX identifier_value_idx
          ON identifier (value);`)
	if err != nil {
		return err
	}
	return nil
}

func downCreateIdentifierTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE identifier;`)
	return err
}

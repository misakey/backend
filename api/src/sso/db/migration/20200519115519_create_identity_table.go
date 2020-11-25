package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initCreateIdentityTable() {
	goose.AddMigration(upCreateIdentityTable, downCreateIdentityTable)
}

func upCreateIdentityTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE identity(
		  id UUID PRIMARY KEY,
			account_id UUID REFERENCES account,
			identifier_id UUID NOT NULL REFERENCES identifier,
			is_authable BOOLEAN NOT NULL,
			display_name VARCHAR(255) NOT NULL,
			notifications VARCHAR(32) NOT NULL DEFAULT 'minimal',
			avatar_url VARCHAR(255),
			confirmed BOOLEAN NOT NULL DEFAULT False
	);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX identity_account_id_idx
          ON identity (account_id);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE UNIQUE INDEX identity_authable_identifier_idx
          ON identity (identifier_id, is_authable);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE UNIQUE INDEX identity_identifier_account_idex
          ON identity (account_id, identifier_id);`)
	if err != nil {
		return err
	}

	return nil
}

func downCreateIdentityTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE identity;`)
	return err
}

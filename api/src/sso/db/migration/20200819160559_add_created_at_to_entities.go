package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initCreatedAtToEntities() {
	goose.AddMigration(upAddCreatedAtToEntities, downAddCreatedAtToEntities)
}

func upAddCreatedAtToEntities(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW();
	`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE identity
		ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW();
	`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE identifier
		ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW();
	`)
	if err != nil {
		return err
	}
	return nil
}

func downAddCreatedAtToEntities(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE account
		DROP COLUMN created_at;
	`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE identity
		DROP COLUMN created_at;
	`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE identifier
		DROP COLUMN created_at;
	`)
	if err != nil {
		return err
	}
	return nil
}

package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initCreateDatatagTable() {
	goose.AddMigration(upCreateDatatagTable, downCreateDatatagTable)
}

func upCreateDatatagTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE datatag(
			id UUID PRIMARY KEY NOT NULL,
			name VARCHAR(255) NOT NULL,
			organization_id UUID NOT NULL REFERENCES organization ON DELETE CASCADE,
			created_at timestamptz NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX datatag_name_idx
          ON datatag (name);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX datatag_organization_id_idx
          ON datatag (organization_id);`)
	if err != nil {
		return err
	}

	return nil
}

func downCreateDatatagTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE datatag;`)
	return err
}

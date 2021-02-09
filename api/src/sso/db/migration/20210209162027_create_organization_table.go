package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initCreateOrganizationTable() {
	goose.AddMigration(upCreateOrganizationTable, downCreateOrganizationTable)
}

func upCreateOrganizationTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE organization(
		id UUID PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		domain VARCHAR(255),
		logo_url VARCHAR(255),
		creator_id UUID NOT NULL REFERENCES identity,
		created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX organization_creator_id_idx ON organization (creator_id);`)
	return err
}

func downCreateOrganizationTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE organization;`)
	return err
}

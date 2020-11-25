package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initCreateBoxSettingsTable() {
	goose.AddMigration(upCreateBoxSettingsTableGo, downCreateBoxSettingsTableGo)
}

func upCreateBoxSettingsTableGo(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE box_setting(
		id SERIAL PRIMARY KEY,
		box_id UUID NOT NULL,
		identity_id UUID NOT NULL,
		muted BOOLEAN NOT NULL,
		updated_at timestamptz NOT NULL,
		CONSTRAINT box_id_identity_id_idx UNIQUE (box_id, identity_id)
	);`)
	return err
}

func downCreateBoxSettingsTableGo(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE box_setting;`)
	return err
}

package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upCreateAuthenticateStepTable, downCreateAuthenticateStepTable)
}

func upCreateAuthenticateStepTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		CREATE TABLE authentication_step(
			id SERIAL PRIMARY KEY,
			identity_id UUID NOT NULL REFERENCES identity,
			method_name VARCHAR(32) NOT NULL,
			metadata JSON NOT NULL,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			complete_at timestamptz DEFAULT NULL
		)
	;`); err != nil {
		return fmt.Errorf("authentication_step table: %v", err)
	}
	if _, err := tx.Exec(`
		CREATE INDEX authentication_step_identity_id_idx
		ON authentication_step (identity_id)
	;`); err != nil {
		return fmt.Errorf("authentication_step_identity_id_idx table: %v", err)
	}
	return nil
}

func downCreateAuthenticateStepTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE authentication_step;`)
	return err
}

package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func initAddReferColumnOnBoxEvents() {
	goose.AddMigration(upAddReferColumnOnBoxEvents, downAddReferColumnOnBoxEvents)
}

func upAddReferColumnOnBoxEvents(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	if _, err := tx.Exec(`ALTER TABLE event
		ADD COLUMN referrer_id UUID REFERENCES event DEFAULT NULL,
		ALTER COLUMN content TYPE JSONB;
	`); err != nil {
		return fmt.Errorf("adding referrer_id column to event: %v", err)
	}
	return nil
}

func downAddReferColumnOnBoxEvents(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	if _, err := tx.Exec(`ALTER TABLE event
		DROP COLUMN referrer_id;
	`); err != nil {
		return fmt.Errorf("dropping referrer_id column from event: %v", err)
	}
	return nil
}

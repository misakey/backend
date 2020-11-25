package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func initCreateIdentityNotificationTable() {
	goose.AddMigration(upCreateIdentityNotificationTable, downCreateIdentityNotificationTable)
}

func upCreateIdentityNotificationTable(tx *sql.Tx) error {
	// create the identity_notification table
	if _, err := tx.Exec(`
		CREATE TABLE identity_notification(
			id SERIAL PRIMARY KEY,
			identity_id UUID NOT NULL REFERENCES identity ON DELETE CASCADE,
			type VARCHAR(32) NOT NULL,
			details JSONB,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			acknowledged_at timestamptz DEFAULT NULL
		)
	`); err != nil {
		return fmt.Errorf("creating identity_notification: %v", err)
	}
	return nil
}

func downCreateIdentityNotificationTable(tx *sql.Tx) error {
	// drop the identity_notification table
	if _, err := tx.Exec(`DROP TABLE identity_notification;`); err != nil {
		return fmt.Errorf("dropping identity_notification: %v", err)
	}
	return nil
}

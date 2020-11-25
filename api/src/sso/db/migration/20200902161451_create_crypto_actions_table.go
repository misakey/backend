package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initCreateCryptoActionsTable() {
	goose.AddMigration(upCreateCryptoActionsTable, downCreateCryptoActionsTable)
}

func upCreateCryptoActionsTable(tx *sql.Tx) error {
	// ON DELETE CASCADE & NOT NULL on account_id column are not linked
	_, err := tx.Exec(`
		CREATE TABLE crypto_action(
			id UUID PRIMARY KEY,
			account_id UUID REFERENCES account ON DELETE CASCADE NOT NULL,
			sender_identity_id UUID REFERENCES identity ON DELETE SET NULL,
			type VARCHAR(255) NOT NULL,
			box_id UUID,
			encryption_public_key VARCHAR(255) NOT NULL,
			encrypted TEXT NOT NULL,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func downCreateCryptoActionsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE crypto_action;`)
	return err
}

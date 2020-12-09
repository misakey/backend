package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initAddIdentityNonIdentifiedPubkey() {
	goose.AddMigration(upAddIdentityNonIdentifiedPubkey, downAddIdentityNonIdentifiedPubkey)
}

func upAddIdentityNonIdentifiedPubkey(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE identity
		ADD COLUMN non_identified_pubkey VARCHAR(255);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downAddIdentityNonIdentifiedPubkey(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE identity
		DROP COLUMN non_identified_pubkey;
	`)
	if err != nil {
		return err
	}
	return nil
}

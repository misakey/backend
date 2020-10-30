package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initAddIdentityPubkey() {
	goose.AddMigration(upAddIdentityPubkey, downAddIdentityPubkey)
}

func upAddIdentityPubkey(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE identity
		ADD COLUMN pubkey VARCHAR(255);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downAddIdentityPubkey(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE identity
		DROP COLUMN pubkey;
	`)
	if err != nil {
		return err
	}
	return nil
}

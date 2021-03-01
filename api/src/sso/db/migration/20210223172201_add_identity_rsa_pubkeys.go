package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initAddIdentityRsaPubkeys() {
	goose.AddMigration(upAddIdentityRsaPubkeys, downAddIdentityRsaPubkeys)
}

func upAddIdentityRsaPubkeys(tx *sql.Tx) error {
	_, err := tx.Exec(`
	  ALTER TABLE identity
		ADD COLUMN pubkey_aes_rsa VARCHAR(1023),
		ADD COLUMN non_identified_pubkey_aes_rsa VARCHAR(1023);
	`)
	return err
}

func downAddIdentityRsaPubkeys(tx *sql.Tx) error {
	_, err := tx.Exec(`
	  ALTER TABLE identity
		DROP COLUMN pubkey_aes_rsa,
		DROP COLUMN non_identified_pubkey_aes_rsa;
	`)
	return err
}

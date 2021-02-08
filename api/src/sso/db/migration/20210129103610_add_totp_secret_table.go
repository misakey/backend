package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initAddTOTPSecretTable() {
	goose.AddMigration(upAddTotpSecretTable, downAddTotpSecretTable)
}

func upAddTotpSecretTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE totp_secret(
				id SERIAL PRIMARY KEY,
				identity_id UUID UNIQUE NOT NULL REFERENCES identity ON DELETE CASCADE,
				secret VARCHAR(32) NOT NULL,
				backup VARCHAR(11)[],
				created_at timestamptz NOT NULL
        );`)
	return err
}

func downAddTotpSecretTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE totp_secret;`)

	return err
}

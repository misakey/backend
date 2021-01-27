package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initAddWebauthnCredentialTable() {
	goose.AddMigration(upAddWebauthnCredentialTable, downAddWebauthnCredentialTable)
}

func upAddWebauthnCredentialTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE webauthn_credential(
				id VARCHAR(127) PRIMARY KEY,
				name VARCHAR(127) NOT NULL,
				identity_id UUID REFERENCES identity ON DELETE CASCADE NOT NULL,
				public_key BYTEA NOT NULL,
				attestation_type VARCHAR(127) NOT NULL,
				aaguid BYTEA NOT NULL,
				sign_count INTEGER NOT NULL,
				clone_warning BOOLEAN NOT NULL,
				created_at timestamptz NOT NULL
        );`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX webauthn_credential_identity_id_idx
          ON webauthn_credential (identity_id);`)
	if err != nil {
		return err
	}

	return nil
}

func downAddWebauthnCredentialTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE webauthn_credential;`)

	return err
}

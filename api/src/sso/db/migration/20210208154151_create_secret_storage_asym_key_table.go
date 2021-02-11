package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func initCreateSecretStorageAsymKeyTable() {
	goose.AddMigration(upCreateSecretStorageAsymKeyTable, downCreateSecretStorageAsymKeyTable)
}

func upCreateSecretStorageAsymKeyTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE secret_storage_asym_key(
			id UUID PRIMARY KEY,
			public_key VARCHAR(255) NOT NULL,
			encrypted_secret_key VARCHAR(2047) NOT NULL,
			account_root_key_hash VARCHAR(255) NOT NULL REFERENCES secret_storage_account_root_key (key_hash) ON DELETE CASCADE,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT one_per_pubkey_per_root_key UNIQUE(public_key, account_root_key_hash)
		)
	`)
	if err != nil {
		return fmt.Errorf("creating table secret_storage_asym_key: %v", err)
	}
	return nil
}

func downCreateSecretStorageAsymKeyTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`DROP TABLE secret_storage_asym_key;`); err != nil {
		return fmt.Errorf("dropping table secret_storage_asym_key: %v", err)
	}
	return nil
}

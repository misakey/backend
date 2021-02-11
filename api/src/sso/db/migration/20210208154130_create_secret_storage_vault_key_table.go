package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func initCreateSecretStorageVaultKeyTable() {
	goose.AddMigration(upCreateSecretStorageVaultKeyTable, downCreateSecretStorageVaultKeyTable)
}

func upCreateSecretStorageVaultKeyTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE secret_storage_vault_key(
			key_hash VARCHAR(255) PRIMARY KEY,
			account_root_key_hash VARCHAR(255) NOT NULL REFERENCES secret_storage_account_root_key (key_hash) ON DELETE CASCADE,
			encrypted_key VARCHAR(1023) NOT NULL,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("creating table secret_storage_account_root_key: %v", err)
	}
	return nil
}

func downCreateSecretStorageVaultKeyTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`DROP TABLE secret_storage_vault_key;`); err != nil {
		return fmt.Errorf("dropping table secret_storage_vault_key: %v", err)
	}
	return nil
}

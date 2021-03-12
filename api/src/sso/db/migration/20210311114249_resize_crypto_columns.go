package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

var toResize = []struct {
	table        string
	column       string
	originalSize int
	newSize      int
}{
	{
		table:        "crypto_action",
		column:       "encryption_public_key",
		originalSize: 255,
		newSize:      1023,
	},
	{
		table:        "secret_storage_asym_key",
		column:       "public_key",
		originalSize: 255,
		newSize:      1023,
	},
	{
		table:        "secret_storage_asym_key",
		column:       "encrypted_secret_key",
		originalSize: 2047,
		newSize:      4095,
	},
	{
		table:        "secret_storage_box_key_share",
		column:       "encrypted_invitation_share",
		originalSize: 2047,
		newSize:      4095,
	},
}

func initResizeCryptoColumns() {
	goose.AddMigration(upResizeCryptoColumns, downResizeCryptoColumns)
}

func upResize(tx *sql.Tx, table string, column string, toSize int) error {
	_, err := tx.Exec(fmt.Sprintf(
		`ALTER TABLE %s
		ALTER COLUMN %s TYPE VARCHAR(%d);
		`,
		table, column, toSize,
	))
	if err != nil {
		return fmt.Errorf("resizing %s.%s: %v", table, column, err)
	}

	return nil
}

func downResize(tx *sql.Tx, table string, column string, toSize int) error {
	_, err := tx.Exec(fmt.Sprintf(`
		DELETE FROM %s
		WHERE length(%s) > %d;
		`,
		table, column, toSize,
	))
	if err != nil {
		return fmt.Errorf("deleting oversized values in %s.%s: %v", table, column, err)
	}

	_, err = tx.Exec(fmt.Sprintf(`
		ALTER TABLE %s
		ALTER COLUMN %s TYPE VARCHAR(%d);
		`,
		table, column, toSize,
	))
	if err != nil {
		return fmt.Errorf("altering type of %s.%s: %v", table, column, err)
	}

	return nil
}

func upResizeCryptoColumns(tx *sql.Tx) error {
	for _, each := range toResize {
		if err := upResize(tx, each.table, each.column, each.newSize); err != nil {
			return err
		}
	}

	return nil
}

func downResizeCryptoColumns(tx *sql.Tx) error {
	for _, each := range toResize {
		if err := downResize(tx, each.table, each.column, each.originalSize); err != nil {
			return err
		}
	}

	return nil
}

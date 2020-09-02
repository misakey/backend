package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initAddCouponAndIdentityLevel() {
	goose.AddMigration(upAddCouponAndIdentityLevel, downAddCouponAndIdentityLevel)
}

func upAddCouponAndIdentityLevel(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE identity
		ADD COLUMN level INTEGER NOT NULL DEFAULT 10;
	`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE TABLE used_coupon(
		id SERIAL PRIMARY KEY,
		identity_id UUID NOT NULL REFERENCES identity ON DELETE CASCADE,
		value VARCHAR(32) NOT NULL,
		created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	return err
}

func downAddCouponAndIdentityLevel(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE identity
		DROP COLUMN level;
	`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DROP TABLE used_coupon;`)
	return err
}

package migration

import (
	"os"

	_ "github.com/lib/pq" // import psql driver
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/db"
)

// Launch the migration
func Launch() {
	// /!\ Do not forget to add you new migration here
	initCreateAccountTable()
	initCreateIdentifierTable()
	initCreateIdentityTable()
	initCreateAuthenticateStepTable()
	initAddBackupDataToAccount()
	initAddBackupVersionToAccount()
	initCreateBackupArchiveTable()
	initCreatedAtToEntities()
	initAddColorToUserIdentity()
	initAddCouponAndIdentityLevel()
	initCreateCryptoActionsTable()
	initCreateIdentityProfileSharingConsentAndRemoveIdentityConfirmed()
	initAddIdentityPubkey()
	initCreateIdentityNotificationTable()
	initAddIdentityNonIdentifiedPubkey()

	db.StartMigration(os.Getenv("DSN_SSO"), os.Getenv("MIGRATION_DIR_SSO"))
}

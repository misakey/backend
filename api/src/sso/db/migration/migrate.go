package migration

import (
	"os"

	_ "github.com/lib/pq"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/db"
)

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

	db.StartMigration(os.Getenv("DSN_SSO"), os.Getenv("MIGRATION_DIR_SSO"))
}

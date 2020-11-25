package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/db/migration"
)

var ssoMigrateCmd = &cobra.Command{
	Use:   "sso-migrate",
	Short: "Migrate the sso postgresql database",
	Long:  `A Goose migrator to migrate the sso postgresql database.`,
	Run: func(cmd *cobra.Command, args []string) {
		migration.Launch()
	},
}

func init() {
	RootCmd.AddCommand(ssoMigrateCmd)
}

package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.misakey.dev/misakey/backend/api/src/box/db/migration"
)

var boxMigrateCmd = &cobra.Command{
	Use:   "box-migrate",
	Short: "Migrate the box postgresql database",
	Long:  `A Goose migrator to migrate the box postgresql database.`,
	Run: func(cmd *cobra.Command, args []string) {
		migration.Launch()
	},
}

func init() {
	RootCmd.AddCommand(boxMigrateCmd)
}

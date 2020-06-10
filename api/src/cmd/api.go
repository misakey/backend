package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/msk-sdk-go/bubble"
	"gitlab.misakey.dev/misakey/msk-sdk-go/db"
	"gitlab.misakey.dev/misakey/msk-sdk-go/echorouter"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/generic"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk"
)

var cfgFile string
var goose string
var env = os.Getenv("ENV")

func init() {
	cobra.OnInitialize()
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	RootCmd.PersistentFlags().StringVar(&goose, "goose", "up", "goose command")
}

var RootCmd = &cobra.Command{
	Use:   "api",
	Short: "Run the API",
	Long:  "This service is responsible for managing all the Misakey backend API",
	Run: func(cmd *cobra.Command, args []string) {
		initService()
	},
}

func initService() {
	// init logger
	log.Logger = logger.ZerologLogger()

	// add error needles to auto handle some specific errors on layers we use everywhere
	bubble.AddNeedle(bubble.PSQLNeedle{})
	bubble.AddNeedle(sdk.NewOzzoNeedle())
	bubble.AddNeedle(sdk.EchoNeedle{})
	bubble.Lock()

	initDefaultConfig()

	// init echo router using sdk call
	e := echorouter.New()
	e.HideBanner = true

	// init db connections
	dbConn, err := db.NewPSQLConn(
		os.Getenv("DATABASE_URL"),
		viper.GetInt("sql.max_open_connections"),
		viper.GetInt("sql.max_idle_connections"),
		viper.GetDuration("sql.conn_max_lifetime"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to db")
	}

	// init modules
	generic.InitModule(e)
	identityIntraprocess := sso.InitModule(e, dbConn)
	boxes.InitModule(e, dbConn, identityIntraprocess)

	// finally launch the echo server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", viper.GetInt("server.port"))))
}

func initDefaultConfig() {
	// always look for the configuration file in the /etc folder
	env := os.Getenv("ENV")
	if env == "development" {
		viper.SetConfigName("api-config.dev")
	} else {
		viper.SetConfigName("api-config")
	}
	viper.AddConfigPath("/etc/")

	// defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("sql.max_open_connections", 50)
	viper.SetDefault("sql.max_idle_connections", 2)
	viper.SetDefault("sql.conn_max_lifetime", "0m")

	// try reading in a config
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("could not read configuration")
	}
}

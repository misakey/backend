package cmd

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/msk-sdk-go/bubble"
	"gitlab.misakey.dev/misakey/msk-sdk-go/echorouter"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/generic"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/generic/pprof"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk"
)

var cfgFile string
var goose string

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
	pprof.Wrap(e)

	// init modules
	generic.InitModule(e)
	identityIntraprocess := sso.InitModule(e)
	box.InitModule(e, identityIntraprocess)

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

	// set defaults value for configuration
	// some of these fields are shared between modules.
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("hydra.secure", true)
	viper.SetDefault("sql.max_open_connections", 15)
	viper.SetDefault("sql.max_idle_connections", 15)
	viper.SetDefault("sql.conn_max_lifetime", "5m")

	// try reading in a config
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("could not read configuration")
	}
}

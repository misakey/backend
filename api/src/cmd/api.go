package cmd

import (
	"fmt"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/rs/zerolog/log"
	echopprof "github.com/sevenNt/echo-pprof"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.misakey.dev/misakey/msk-sdk-go/bubble"
	mecho "gitlab.misakey.dev/misakey/msk-sdk-go/echo"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/generic"
)

var cfgFile string
var goose string
var env = os.Getenv("ENV")

func init() {
	cobra.OnInitialize()
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	RootCmd.PersistentFlags().StringVar(&goose, "goose", "up", "goose command")
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var RootCmd = &cobra.Command{
	Use:   "auth",
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
	bubble.AddNeedle(bubble.ValidatorNeedle{})
	bubble.AddNeedle(bubble.EchoNeedle{})
	bubble.Lock()

	initDefaultConfig()

	// init echo framework with compressed HTTP responses, custom logger format and custom validator
	e := echo.New()
	e.Use(mecho.NewZerologLogger())
	e.Use(mecho.NewLogger())
	e.Use(middleware.Recover())
	echopprof.Wrap(e)

	genericPresenter := generic.NewGenericEcho()

	// Bind generic routes
	generic := e.Group("")
	generic.GET("/version", genericPresenter.GetVersion)

	// finally launch the echo server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", viper.GetInt("server.port"))))
}

func initDefaultConfig() {
	// always look for the configuration file in the /etc folder
	viper.AddConfigPath("/etc/")

	// try reading in a config
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("could not read configuration")
	}
}

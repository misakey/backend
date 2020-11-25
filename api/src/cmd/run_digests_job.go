package cmd

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/go-redis/redis/v7"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/config"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/db"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/notifications/email"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/notifications/jobs"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
)

var frequency string

var DigestsJobCmd = &cobra.Command{
	Use:   "digests-job",
	Short: "Run the digests job",
	Long:  "This job is responsible for notifying users about new events in the app.",
	Run: func(cmd *cobra.Command, args []string) {
		initDigestsJob()
	},
}

func initDigestsJob() {
	initDefaultDigestsConfig()

	// init logger
	log.Logger = logger.ZerologLogger(viper.GetString("log.level"))
	ctx := logger.SetLogger(context.Background(), &log.Logger)

	// init db connections
	ssoDBConn, err := db.NewPSQLConn(
		os.Getenv("DSN_SSO"),
		viper.GetInt("sql.max_open_connections"),
		viper.GetInt("sql.max_idle_connections"),
		viper.GetDuration("sql.conn_max_lifetime"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to db")
	}

	boxDBConn, err := db.NewPSQLConn(
		os.Getenv("DSN_BOX"),
		viper.GetInt("sql.max_open_connections"),
		viper.GetInt("sql.max_idle_connections"),
		viper.GetDuration("sql.conn_max_lifetime"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to db")
	}

	// init redis connection
	redConn := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.address"), viper.GetString("redis.port")),
		Password: "",
		DB:       0,
	})
	if _, err := redConn.Ping().Result(); err != nil {
		log.Fatal().Err(err).Msg("could not connect to redis")
	}

	templateRepo := email.NewTemplateFileSystem(viper.GetString("mail.templates"))
	var emailRepo email.Sender
	env := os.Getenv("ENV")
	if env == "development" {
		emailRepo = email.NewLogMailer()
	} else if env == "production" {
		emailRepo = email.NewMailerAmazonSES(viper.GetString("aws.ses_region"), viper.GetString("aws.ses_configuration_set"))
		if err != nil {
			log.Fatal().Msg("could not instantiate SES Mailer")
		}
	} else {
		log.Fatal().Msg("unknown ENV value (should be production|development)")
	}

	emailRenderer, err := email.NewEmailRenderer(
		templateRepo,
		[]string{
			"notification_html", "notification_txt",
			"notificationNoAccount_html", "notificationNoAccount_txt",
		},
		viper.GetString("mail.from"),
	)
	if err != nil {
		log.Fatal().Msg("could not instantiate email renderer")
	}

	// nil for avatar repo since digest job doesn't care about identity's avatars.
	identityMapper := events.NewIdentityMapper(identity.NewIntraprocessHelper(ssoDBConn))

	digestService, err := jobs.NewDigestJob(
		frequency, viper.GetString("digests.domain"),
		boxDBConn, redConn, identityMapper,
		emailRepo, emailRenderer,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not instantiate digest job")
	}

	if err := digestService.SendDigests(ctx); err != nil {
		log.Error().Err(err).Msg("could not send digests")
	}

}

func initDefaultDigestsConfig() {
	// always look for the configuration file in the /etc folder
	env := os.Getenv("ENV")
	if env == "development" {
		viper.SetConfigName("digests.dev")
	} else {
		viper.SetConfigName("digests")
	}
	viper.AddConfigPath("/etc/")

	// set defaults value for configuration
	// some of these fields are shared between modules.
	viper.SetDefault("log.level", "info")
	viper.SetDefault("sql.max_open_connections", 15)
	viper.SetDefault("sql.max_idle_connections", 15)
	viper.SetDefault("sql.conn_max_lifetime", "5m")

	// try reading in a config
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("could not read configuration")
	}

	mandatoryFields := []string{
		"mail.templates",
		"mail.from",
		"redis.address",
		"redis.port",
		"digests.domain",
	}
	config.FatalIfMissing("Digests", mandatoryFields)
	config.Print("Digests", []string{})
}

func init() {
	RootCmd.PersistentFlags().StringVar(&frequency, "frequency", "minimal", "frequency configuration")
	RootCmd.AddCommand(DigestsJobCmd)
}

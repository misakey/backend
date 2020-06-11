package sso

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/config"
)

func initConfig() {
	// set defaults value for configuration
	viper.SetDefault("hydra.secure", true)

	// handle missing mandatory fields
	mandatoryFields := []string{
		"authflow.self_client_id",
		"authflow.hydra_token_url",
		"authflow.self_encoded_jwk",
		"hydra.public_endpoint",
		"hydra.admin_endpoint",
		"authflow.auth_url",
		"authflow.code_redirect_url",
		"authflow.hydra_token_url",
		"authflow.token_redirect_url",
		"mail.templates",
		"mail.from",
	}
	switch os.Getenv("ENV") {
	case "production":
		mandatoryFields = append(mandatoryFields, []string{"aws.ses_region"}...)
		if os.Getenv("AWS_ACCESS_KEY") == "" {
			log.Warn().Msg("AWS_ACCESS_KEY not set")
		}
		if os.Getenv("AWS_SECRET_KEY") == "" {
			log.Warn().Msg("AWS_SECRET_KEY not set")
		}
	case "development":
		log.Info().Msg("{} Development mode is activated. {}")
	default:
		log.Fatal().Msg("unknown ENV value (should be production|development)")
	}
	config.FatalIfMissing("SSO", mandatoryFields)
	secretFields := []string{
		"authflow.self_encoded_jwk",
	}
	config.Print("SSO", secretFields)
}
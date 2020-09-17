package box

import (
	"os"

	"github.com/rs/zerolog/log"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/config"
)

func initConfig() {
	// handle missing mandatory fields
	mandatoryFields := []string{
		"authflow.self_client_id",
		"hydra.admin_endpoint",
		"redis.address",
		"redis.port",
		"websockets.allowed_origins",
	}
	switch os.Getenv("ENV") {
	case "production":
		mandatoryFields = append(mandatoryFields, []string{"aws.ses_region", "aws.s3_region", "aws.encrypted_files_bucket"}...)
		if os.Getenv("AWS_ACCESS_KEY") == "" {
			log.Warn().Msg("AWS_ACCESS_KEY not set")
		}
		if os.Getenv("AWS_SECRET_KEY") == "" {
			log.Warn().Msg("AWS_SECRET_KEY not set")
		}
	case "development":
		mandatoryFields = append(mandatoryFields, []string{"server.encrypted_files"}...)
		log.Info().Msg("{} Development mode is activated. {}")
	default:
		log.Fatal().Msg("unknown ENV value (should be production|development)")
	}

	// no secret fields so far
	secretFields := []string{"authflow.self_encoded_jwk"}
	config.FatalIfMissing("BOX", mandatoryFields)
	config.Print("BOX", secretFields)
}

package logger

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

type loggerCtxKey struct{}

// SetLogger returns ctx with Logger set inside it using loggerCtxKey
func SetLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, logger)
}

func FromCtx(ctx context.Context) *zerolog.Logger {
	return ctx.Value(loggerCtxKey{}).(*zerolog.Logger)
}

func ZerologLogger() zerolog.Logger {
	serviceName := filepath.Base(os.Args[0])
	serviceVersion := os.Getenv("VERSION")
	zerolog.TimeFieldFormat = time.RFC3339
	l := zerolog.
		New(os.Stdout).
		With().
		Timestamp().
		Str("service_name", serviceName).
		Str("service_version", serviceVersion).
		Caller().
		Logger()

	// do not log in json in development environment
	if os.Getenv("ENV") == "development" {
		l = l.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return l
}

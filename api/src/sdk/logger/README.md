# Logger Package

Logger package should be used by all our services to share same log format or storage.

Sharing a log format allows us to collect process them easily in order to build metrics and other features around it.

## Echo Logger Middlewares

We use two different loggers:
- The default [echo logger](https://godoc.org/github.com/labstack/echo#Logger), instantiated by the `NewLogger()` middleware. It logs:
  - the requests (automatically)
- A [zerolog logger](https://github.com/rs/zerolog), instantiated by the `NewZerologLogger()` middleware. It logs:
  - the errors (automatically, when a merr is returned)
  - the logs specified in the service with `logger.FromCtx(ctx)` (`logger` being `import "gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"`)

## Service Logger

When you are not in a request scope and have no access to derived `context.Context`, you can instantiate a zerolog logger with `logger.ZerologLogger()`. It will automatically add the `service_name` and `service_version` fields extracted from the running binary and `VERSION` environment variable.

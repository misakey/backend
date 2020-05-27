# API service

API is the Misakey backend service

:warning: This section is a work in progress.

## Environment variables

- `ENV`: `production` or `development`. This will change some behaviours like the way to send emails
- `AWS_ACCESS_KEY`: Only on `production`. Needed to send emails.
- `AWS_SECRET_KEY`: Only on `production`. Needed to send emails.

## Migrations

After running the migration with Goose,
here is the SQLBoiler command you should run to update the SQLBoiler model:

    $ # You should be located in 'api/'
    $ sqlboiler psql --config config/sqlboiler.toml --output src/sqlboiler

# Create a migration

1. Run in this folder the command `goose -dir=./migration/ create {migration_name}`
2. Change the `init()` function name to `init{migration_name}()`.
3. Add the call of this newly named function in the `migrate.go` file at the end of other migration calls.

# Run the migrations

Run the `docker-compose up --build -V` command to have a new database
with current local migrations run on it.

See `./docker-compose.yml` for more information about the database configuration
(port, secret, database name...).

# Generate models

Run in this folder `sqlboiler psql` to generate the model. They will be create in
the corresponding module repository folder.

-- NOTE: Concurrent indexes cannot be created using a transaction so we use .sql files
-- Go file migrations always run in transaction mode.
-- +goose Up
-- +goose NO TRANSACTION
-- SQL in this section is executed when the migration is applied.
CREATE INDEX CONCURRENTLY event_type_id_idx ON event (type, id);
CREATE INDEX CONCURRENTLY event_most_recent_idx ON event (created_at DESC);

-- +goose Down
-- +goose NO TRANSACTION
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX CONCURRENTLY event_type_id_idx;
DROP INDEX CONCURRENTLY event_most_recent_idx;
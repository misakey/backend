-- NOTE: Concurrent indexes cannot be created using a transaction so we use .sql files
-- Go file migrations always run in transaction mode.
-- +goose Up
-- +goose NO TRANSACTION
-- SQL in this section is executed when the migration is applied.
CREATE INDEX CONCURRENTLY event_referer_id_idx ON event (referer_id);
CREATE INDEX CONCURRENTLY event_box_id_idx ON event (box_id);
CREATE INDEX CONCURRENTLY event_sender_id_idx ON event (sender_id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX CONCURRENTLY event_referer_id_idx;
DROP INDEX CONCURRENTLY event_box_id_idx;
DROP INDEX CONCURRENTLY event_sender_id_idx;

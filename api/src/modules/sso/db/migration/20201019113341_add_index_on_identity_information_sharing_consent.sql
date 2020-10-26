-- NOTE: Concurrent indexes cannot be created using a transaction so we use .sql files
-- Go file migrations always run in transaction mode.
-- +goose Up
-- +goose NO TRANSACTION
-- SQL in this section is executed when the migration is applied.
CREATE INDEX CONCURRENTLY identity_profile_sharing_consent_identity_id_idx ON identity_profile_sharing_consent (identity_id);

-- +goose Down
-- +goose NO TRANSACTION
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX CONCURRENTLY identity_profile_sharing_consent_identity_id_idx;

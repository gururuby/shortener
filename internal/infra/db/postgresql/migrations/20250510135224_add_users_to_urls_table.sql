-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN user_id INT REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX ON urls(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls DROP COLUMN user_id;
-- +goose StatementEnd

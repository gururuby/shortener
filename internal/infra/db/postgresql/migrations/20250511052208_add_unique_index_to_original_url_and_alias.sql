-- +goose Up
-- +goose StatementBegin
DROP INDEX urls_original_url_idx;
CREATE UNIQUE INDEX ON urls(original_url,alias);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE UNIQUE INDEX ON urls (original_url);
DROP INDEX urls_original_url_alias_idx
-- +goose StatementEnd

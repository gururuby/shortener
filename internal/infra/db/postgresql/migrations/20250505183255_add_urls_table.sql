-- +goose Up
-- +goose StatementBegin
CREATE TABLE urls (
    uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    alias varchar(255) NOT NULL,
    original_url varchar(255) NOT NULL
);

CREATE UNIQUE INDEX ON urls (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE urls;
-- +goose StatementEnd

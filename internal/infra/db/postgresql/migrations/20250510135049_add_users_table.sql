-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY
                   );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd

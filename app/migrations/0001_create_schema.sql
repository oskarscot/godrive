-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    google_sub TEXT UNIQUE NOT NULL,
    email TEXT NOT NULL,
    picture TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd

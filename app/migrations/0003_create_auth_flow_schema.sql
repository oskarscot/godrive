-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS auth_flow (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash TEXT UNIQUE NOT NULL,
    expected_state TEXT UNIQUE NOT NULL,
    verifier TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE auth_flow;
-- +goose StatementEnd

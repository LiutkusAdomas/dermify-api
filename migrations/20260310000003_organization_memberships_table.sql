-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS organization_memberships (
    id         BIGSERIAL PRIMARY KEY,
    org_id     BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(20) NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member', 'viewer')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (org_id, user_id)
);

CREATE INDEX idx_org_memberships_org_id ON organization_memberships(org_id);
CREATE INDEX idx_org_memberships_user_id ON organization_memberships(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS organization_memberships;
-- +goose StatementEnd

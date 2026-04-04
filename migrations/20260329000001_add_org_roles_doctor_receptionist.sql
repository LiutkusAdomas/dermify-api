-- +goose Up
-- +goose StatementBegin
ALTER TABLE organization_memberships
    DROP CONSTRAINT IF EXISTS organization_memberships_role_check;

ALTER TABLE organization_memberships
    ADD CONSTRAINT organization_memberships_role_check
    CHECK (role IN ('admin', 'member', 'viewer', 'doctor', 'receptionist'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE organization_memberships
    DROP CONSTRAINT IF EXISTS organization_memberships_role_check;

ALTER TABLE organization_memberships
    ADD CONSTRAINT organization_memberships_role_check
    CHECK (role IN ('admin', 'member', 'viewer'));
-- +goose StatementEnd

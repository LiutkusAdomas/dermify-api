-- +goose Up
-- +goose StatementBegin
ALTER TABLE organizations
ADD COLUMN IF NOT EXISTS invite_from_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS invite_from_name VARCHAR(255);

ALTER TABLE organization_invitations
    DROP CONSTRAINT IF EXISTS organization_invitations_role_check;

ALTER TABLE organization_invitations
    ADD CONSTRAINT organization_invitations_role_check
    CHECK (role IN ('admin', 'member', 'viewer', 'doctor', 'receptionist'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE organization_invitations
    DROP CONSTRAINT IF EXISTS organization_invitations_role_check;

ALTER TABLE organization_invitations
    ADD CONSTRAINT organization_invitations_role_check
    CHECK (role IN ('admin', 'member', 'viewer'));

ALTER TABLE organizations
DROP COLUMN IF EXISTS invite_from_email,
DROP COLUMN IF EXISTS invite_from_name;
-- +goose StatementEnd

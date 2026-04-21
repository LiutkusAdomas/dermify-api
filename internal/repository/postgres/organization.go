package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresOrganizationRepository implements service.OrganizationRepository.
type PostgresOrganizationRepository struct {
	db *sql.DB
}

// NewPostgresOrganizationRepository creates a new PostgresOrganizationRepository.
func NewPostgresOrganizationRepository(db *sql.DB) *PostgresOrganizationRepository {
	return &PostgresOrganizationRepository{db: db}
}

// CreateWithMembership inserts an org and adds the creator as admin in a transaction.
func (r *PostgresOrganizationRepository) CreateWithMembership(ctx context.Context, org *domain.Organization, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // best-effort rollback

	err = tx.QueryRowContext(ctx,
		`INSERT INTO organizations (name, slug, description, timezone, invite_from_email, invite_from_name) VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		org.Name, org.Slug, org.Description, org.Timezone, org.InviteFromEmail, org.InviteFromName,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrOrgSlugExists
		}
		return fmt.Errorf("inserting organization: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO organization_memberships (org_id, user_id, role) VALUES ($1, $2, 'admin')`,
		org.ID, userID,
	)
	if err != nil {
		return fmt.Errorf("inserting membership: %w", err)
	}

	return tx.Commit()
}

// ListByUser returns all organizations a user belongs to, with roles.
func (r *PostgresOrganizationRepository) ListByUser(ctx context.Context, userID int64) ([]*domain.OrganizationWithRole, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT o.id, o.name, o.slug, COALESCE(o.description, ''), COALESCE(o.logo_url, ''), COALESCE(o.timezone, 'UTC'),
		        COALESCE(o.invite_from_email, ''), COALESCE(o.invite_from_name, ''),
		        o.created_at, o.updated_at, om.role
		 FROM organizations o
		 JOIN organization_memberships om ON om.org_id = o.id
		 WHERE om.user_id = $1
		 ORDER BY o.name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying user organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*domain.OrganizationWithRole
	for rows.Next() {
		var o domain.OrganizationWithRole
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.Description, &o.LogoURL, &o.Timezone, &o.InviteFromEmail, &o.InviteFromName,
			&o.CreatedAt, &o.UpdatedAt, &o.Role); err != nil {
			return nil, fmt.Errorf("scanning organization: %w", err)
		}
		orgs = append(orgs, &o)
	}
	return orgs, rows.Err()
}

// GetByID retrieves a single organization.
func (r *PostgresOrganizationRepository) GetByID(ctx context.Context, orgID int64) (*domain.Organization, error) {
	var o domain.Organization
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, COALESCE(description, ''), COALESCE(logo_url, ''), COALESCE(timezone, 'UTC'),
		        COALESCE(invite_from_email, ''), COALESCE(invite_from_name, ''), created_at, updated_at
		 FROM organizations WHERE id = $1`, orgID,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.Description, &o.LogoURL, &o.Timezone, &o.InviteFromEmail, &o.InviteFromName, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrOrgNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying organization: %w", err)
	}
	return &o, nil
}

// Update modifies organization fields.
func (r *PostgresOrganizationRepository) Update(ctx context.Context, orgID int64, name, description, logoURL, timezone, inviteFromEmail, inviteFromName *string) (*domain.Organization, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argIdx))
		args = append(args, *name)
		argIdx++
	}
	if description != nil {
		setClauses = append(setClauses, "description = $"+strconv.Itoa(argIdx))
		args = append(args, *description)
		argIdx++
	}
	if logoURL != nil {
		setClauses = append(setClauses, "logo_url = $"+strconv.Itoa(argIdx))
		args = append(args, *logoURL)
		argIdx++
	}
	if timezone != nil {
		setClauses = append(setClauses, "timezone = $"+strconv.Itoa(argIdx))
		args = append(args, *timezone)
		argIdx++
	}
	if inviteFromEmail != nil {
		setClauses = append(setClauses, "invite_from_email = $"+strconv.Itoa(argIdx))
		args = append(args, *inviteFromEmail)
		argIdx++
	}
	if inviteFromName != nil {
		setClauses = append(setClauses, "invite_from_name = $"+strconv.Itoa(argIdx))
		args = append(args, *inviteFromName)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil, service.ErrOrgNoFieldsToUpdate
	}

	setClauses = append(setClauses, "updated_at = $"+strconv.Itoa(argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, orgID)
	query := "UPDATE organizations SET " + strings.Join(setClauses, ", ") + " WHERE id = $" + strconv.Itoa(argIdx)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("updating organization: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, service.ErrOrgNotFound
	}

	return r.GetByID(ctx, orgID)
}

// GetMemberRole returns a user's role in an organization.
func (r *PostgresOrganizationRepository) GetMemberRole(ctx context.Context, orgID, userID int64) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		`SELECT role FROM organization_memberships WHERE org_id = $1 AND user_id = $2`,
		orgID, userID,
	).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", service.ErrOrgNotMember
	}
	if err != nil {
		return "", fmt.Errorf("querying membership: %w", err)
	}
	return role, nil
}

// ListMembers returns all members of an organization.
func (r *PostgresOrganizationRepository) ListMembers(ctx context.Context, orgID int64) ([]*domain.OrgMember, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT om.id, u.id, u.username, COALESCE(u.email, ''), u.email, om.role, u.must_change_password, om.created_at
		 FROM organization_memberships om
		 JOIN users u ON u.id = om.user_id
		 WHERE om.org_id = $1
		 ORDER BY om.created_at`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying members: %w", err)
	}
	defer rows.Close()

	var members []*domain.OrgMember
	for rows.Next() {
		var m domain.OrgMember
		if err := rows.Scan(&m.ID, &m.UserID, &m.FirstName, &m.LastName,
			&m.Email, &m.Role, &m.MustChangePassword, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, &m)
	}
	return members, rows.Err()
}

// UpdateMemberRole changes a member's role.
func (r *PostgresOrganizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID int64, role string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE organization_memberships SET role = $1 WHERE org_id = $2 AND user_id = $3`,
		role, orgID, userID,
	)
	if err != nil {
		return fmt.Errorf("updating member role: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrOrgMemberNotFound
	}
	return nil
}

// RemoveMember deletes a membership.
func (r *PostgresOrganizationRepository) RemoveMember(ctx context.Context, orgID, userID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM organization_memberships WHERE org_id = $1 AND user_id = $2`,
		orgID, userID,
	)
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrOrgMemberNotFound
	}
	return nil
}

// CountAdmins returns the number of admins in an organization (optionally excluding a user).
func (r *PostgresOrganizationRepository) CountAdmins(ctx context.Context, orgID int64, excludeUserID *int64) (int, error) {
	var count int
	if excludeUserID != nil {
		err := r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM organization_memberships WHERE org_id = $1 AND role = 'admin' AND user_id != $2`,
			orgID, *excludeUserID,
		).Scan(&count)
		return count, err
	}
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM organization_memberships WHERE org_id = $1 AND role = 'admin'`,
		orgID,
	).Scan(&count)
	return count, err
}

// CreateInvitation inserts an invitation record.
func (r *PostgresOrganizationRepository) CreateInvitation(ctx context.Context, orgID, invitedBy int64, email, role, token string, expiresAt time.Time) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO organization_invitations (org_id, invited_by, email, role, token, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5, 'pending', $6) RETURNING id`,
		orgID, invitedBy, email, role, token, expiresAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("creating invitation: %w", err)
	}
	return id, nil
}

// HasPendingInvitation checks if a pending invitation exists for this org+email.
func (r *PostgresOrganizationRepository) HasPendingInvitation(ctx context.Context, orgID int64, email string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM organization_invitations
		 WHERE org_id = $1 AND email = $2 AND status = 'pending' AND expires_at > NOW()`,
		orgID, email,
	).Scan(&count)
	return count > 0, err
}

// IsAlreadyMember checks if an email is already a member of the organization.
func (r *PostgresOrganizationRepository) IsAlreadyMember(ctx context.Context, orgID int64, email string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM organization_memberships om
		 JOIN users u ON u.id = om.user_id
		 WHERE om.org_id = $1 AND u.email = $2`,
		orgID, email,
	).Scan(&count)
	return count > 0, err
}

// GetOrgName returns the organization name.
func (r *PostgresOrganizationRepository) GetOrgName(ctx context.Context, orgID int64) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx,
		`SELECT name FROM organizations WHERE id = $1`, orgID,
	).Scan(&name)
	return name, err
}

// GetUserName returns a user's display name.
func (r *PostgresOrganizationRepository) GetUserName(ctx context.Context, userID int64) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx,
		`SELECT username FROM users WHERE id = $1`, userID,
	).Scan(&name)
	return name, err
}

// ListUserInvitations lists pending invitations for a user's email.
func (r *PostgresOrganizationRepository) ListUserInvitations(ctx context.Context, email string) ([]*domain.OrgInvitation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT i.id, i.org_id, o.name, i.email, i.role, i.status,
		        u.username, i.token, i.expires_at, i.created_at
		 FROM organization_invitations i
		 JOIN organizations o ON o.id = i.org_id
		 JOIN users u ON u.id = i.invited_by
		 WHERE i.email = $1 AND i.status = 'pending' AND i.expires_at > NOW()
		 ORDER BY i.created_at DESC`,
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("querying user invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*domain.OrgInvitation
	for rows.Next() {
		var inv domain.OrgInvitation
		if err := rows.Scan(&inv.ID, &inv.OrgID, &inv.OrgName, &inv.Email, &inv.Role,
			&inv.Status, &inv.InvitedBy, &inv.Token, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning invitation: %w", err)
		}
		invitations = append(invitations, &inv)
	}
	return invitations, rows.Err()
}

// GetInvitationByToken retrieves a pending invitation by token.
func (r *PostgresOrganizationRepository) GetInvitationByToken(ctx context.Context, token string) (*domain.OrgInvitation, error) {
	var inv domain.OrgInvitation
	err := r.db.QueryRowContext(ctx,
		`SELECT i.id, i.org_id, o.name, i.email, i.role, i.expires_at
		 FROM organization_invitations i
		 JOIN organizations o ON o.id = i.org_id
		 WHERE i.token = $1 AND i.status = 'pending' AND i.expires_at > NOW()`,
		token,
	).Scan(&inv.ID, &inv.OrgID, &inv.OrgName, &inv.Email, &inv.Role, &inv.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrInvitationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying invitation: %w", err)
	}
	return &inv, nil
}

// AcceptInvitation marks invitation as accepted and adds user to org in a transaction.
func (r *PostgresOrganizationRepository) AcceptInvitation(ctx context.Context, invID, orgID, userID int64, role string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // best-effort rollback

	_, err = tx.ExecContext(ctx,
		`UPDATE organization_invitations SET status = 'accepted' WHERE id = $1`, invID,
	)
	if err != nil {
		return fmt.Errorf("updating invitation: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO organization_memberships (org_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (org_id, user_id) DO NOTHING`,
		orgID, userID, role,
	)
	if err != nil {
		return fmt.Errorf("creating membership: %w", err)
	}

	return tx.Commit()
}

// ConfirmInvitation marks invitation accepted and provisions membership for existing invited user.
func (r *PostgresOrganizationRepository) ConfirmInvitation(ctx context.Context, orgID, invitationID int64, requirePasswordChange bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // best-effort rollback

	var (
		email string
		role  string
	)
	err = tx.QueryRowContext(ctx,
		`SELECT email, role
		 FROM organization_invitations
		 WHERE id = $1 AND org_id = $2 AND status = 'pending' AND expires_at > NOW()
		 FOR UPDATE`,
		invitationID, orgID,
	).Scan(&email, &role)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrInvitationNotFound
	}
	if err != nil {
		return fmt.Errorf("querying pending invitation: %w", err)
	}

	var userID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM users WHERE LOWER(email) = LOWER($1)`,
		email,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrInvitationUserNotFound
	}
	if err != nil {
		return fmt.Errorf("querying invited user: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE organization_invitations
		 SET status = 'accepted'
		 WHERE id = $1`,
		invitationID,
	)
	if err != nil {
		return fmt.Errorf("updating invitation status: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO organization_memberships (org_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (org_id, user_id) DO NOTHING`,
		orgID, userID, role,
	)
	if err != nil {
		return fmt.Errorf("creating organization membership: %w", err)
	}

	if requirePasswordChange {
		_, err = tx.ExecContext(ctx,
			`UPDATE users SET must_change_password = true, updated_at = NOW() WHERE id = $1`,
			userID,
		)
		if err != nil {
			return fmt.Errorf("enabling must_change_password: %w", err)
		}
	}

	return tx.Commit()
}

// DeclineInvitation marks an invitation as declined.
func (r *PostgresOrganizationRepository) DeclineInvitation(ctx context.Context, token, email string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE organization_invitations SET status = 'declined'
		 WHERE token = $1 AND email = $2 AND status = 'pending'`,
		token, email,
	)
	if err != nil {
		return fmt.Errorf("declining invitation: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrInvitationNotFound
	}
	return nil
}

// ListOrgInvitations lists all invitations for an organization.
func (r *PostgresOrganizationRepository) ListOrgInvitations(ctx context.Context, orgID int64) ([]*domain.OrgInvitation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT i.id, i.org_id, o.name, i.email, i.role, i.status,
		        u.username, EXISTS(SELECT 1 FROM users iu WHERE LOWER(iu.email) = LOWER(i.email)) AS has_account,
		        i.expires_at, i.created_at
		 FROM organization_invitations i
		 JOIN organizations o ON o.id = i.org_id
		 JOIN users u ON u.id = i.invited_by
		 WHERE i.org_id = $1
		 ORDER BY i.created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying org invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*domain.OrgInvitation
	for rows.Next() {
		var inv domain.OrgInvitation
		if err := rows.Scan(&inv.ID, &inv.OrgID, &inv.OrgName, &inv.Email, &inv.Role,
			&inv.Status, &inv.InvitedBy, &inv.HasAccount, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning invitation: %w", err)
		}
		invitations = append(invitations, &inv)
	}
	return invitations, rows.Err()
}

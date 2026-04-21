package domain

import "time"

// Organization represents a company or clinic.
type Organization struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	LogoURL         string    `json:"logo_url,omitempty"`
	Timezone        string    `json:"timezone"`
	InviteFromEmail string    `json:"invite_from_email,omitempty"`
	InviteFromName  string    `json:"invite_from_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// OrganizationWithRole pairs an organization with the user's role in it.
type OrganizationWithRole struct {
	Organization
	Role string `json:"role"`
}

// OrgMembership roles.
const (
	OrgRoleAdmin        = "admin"
	OrgRoleMember       = "member"
	OrgRoleViewer       = "viewer"
	OrgRoleDoctor       = "doctor"
	OrgRoleReceptionist = "receptionist"
)

// ValidOrgRole checks if a role string is a valid organization role.
func ValidOrgRole(role string) bool {
	switch role {
	case OrgRoleAdmin, OrgRoleMember, OrgRoleViewer, OrgRoleDoctor, OrgRoleReceptionist:
		return true
	}
	return false
}

// AllOrgRoles returns all valid organization roles.
func AllOrgRoles() []string {
	return []string{OrgRoleAdmin, OrgRoleMember, OrgRoleViewer, OrgRoleDoctor, OrgRoleReceptionist}
}

// OrgMember represents a user's membership in an organization.
type OrgMember struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
}

// Invitation status values.
const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusDeclined = "declined"
	InvitationStatusExpired  = "expired"
)

// OrgInvitation represents an invitation to join an organization.
type OrgInvitation struct {
	ID         int64     `json:"id"`
	OrgID      int64     `json:"org_id"`
	OrgName    string    `json:"org_name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	Status     string    `json:"status"`
	InvitedBy  string    `json:"invited_by"`
	HasAccount bool      `json:"has_account"`
	Token      string    `json:"token,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

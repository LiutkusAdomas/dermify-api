package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"

	"dermify-api/internal/domain"
	"dermify-api/internal/pkg/email"
)

// Sentinel errors for organization operations.
var (
	ErrOrgNotFound          = errors.New("organization not found")          //nolint:gochecknoglobals // sentinel error
	ErrOrgSlugExists        = errors.New("organization slug already exists") //nolint:gochecknoglobals // sentinel error
	ErrOrgNotMember         = errors.New("not a member of this organization") //nolint:gochecknoglobals // sentinel error
	ErrOrgNotAdmin          = errors.New("admin access required")            //nolint:gochecknoglobals // sentinel error
	ErrOrgMemberNotFound    = errors.New("membership not found")             //nolint:gochecknoglobals // sentinel error
	ErrOrgLastAdmin         = errors.New("cannot remove the last admin")     //nolint:gochecknoglobals // sentinel error
	ErrOrgNoFieldsToUpdate  = errors.New("no fields to update")             //nolint:gochecknoglobals // sentinel error
	ErrOrgAlreadyMember     = errors.New("user is already a member")        //nolint:gochecknoglobals // sentinel error
	ErrInvitationExists     = errors.New("pending invitation already exists") //nolint:gochecknoglobals // sentinel error
	ErrInvitationNotFound   = errors.New("invitation not found or expired")  //nolint:gochecknoglobals // sentinel error
	ErrInvitationWrongEmail = errors.New("invitation is for a different email") //nolint:gochecknoglobals // sentinel error
)

// OrganizationRepository defines the data access contract for organizations.
type OrganizationRepository interface {
	CreateWithMembership(ctx context.Context, org *domain.Organization, userID int64) error
	ListByUser(ctx context.Context, userID int64) ([]*domain.OrganizationWithRole, error)
	GetByID(ctx context.Context, orgID int64) (*domain.Organization, error)
	Update(ctx context.Context, orgID int64, name, description, logoURL *string) (*domain.Organization, error)
	GetMemberRole(ctx context.Context, orgID, userID int64) (string, error)
	ListMembers(ctx context.Context, orgID int64) ([]*domain.OrgMember, error)
	UpdateMemberRole(ctx context.Context, orgID, userID int64, role string) error
	RemoveMember(ctx context.Context, orgID, userID int64) error
	CountAdmins(ctx context.Context, orgID int64, excludeUserID *int64) (int, error)
	CreateInvitation(ctx context.Context, orgID, invitedBy int64, email, role, token string, expiresAt time.Time) (int64, error)
	HasPendingInvitation(ctx context.Context, orgID int64, email string) (bool, error)
	IsAlreadyMember(ctx context.Context, orgID int64, email string) (bool, error)
	GetOrgName(ctx context.Context, orgID int64) (string, error)
	GetUserName(ctx context.Context, userID int64) (string, error)
	ListUserInvitations(ctx context.Context, email string) ([]*domain.OrgInvitation, error)
	GetInvitationByToken(ctx context.Context, token string) (*domain.OrgInvitation, error)
	AcceptInvitation(ctx context.Context, invID, orgID, userID int64, role string) error
	DeclineInvitation(ctx context.Context, token, email string) error
	ListOrgInvitations(ctx context.Context, orgID int64) ([]*domain.OrgInvitation, error)
}

// OrganizationService handles organization business logic.
type OrganizationService struct {
	repo        OrganizationRepository
	emailClient *email.Client
}

// NewOrganizationService creates a new OrganizationService.
func NewOrganizationService(repo OrganizationRepository, emailClient *email.Client) *OrganizationService {
	return &OrganizationService{repo: repo, emailClient: emailClient}
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`) //nolint:gochecknoglobals // compile once

// GenerateSlug creates a URL-friendly slug from a name.
func GenerateSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`[\s]+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`-{2,}`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// ValidSlug checks if a slug is well-formed.
func ValidSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}

// Create creates a new organization and adds the creator as admin.
func (s *OrganizationService) Create(ctx context.Context, name, slug, description string, userID int64) (*domain.Organization, error) {
	if slug == "" {
		slug = GenerateSlug(name)
	}
	org := &domain.Organization{
		Name:        name,
		Slug:        slug,
		Description: description,
	}
	if err := s.repo.CreateWithMembership(ctx, org, userID); err != nil {
		return nil, err
	}
	return org, nil
}

// ListByUser returns all organizations a user belongs to.
func (s *OrganizationService) ListByUser(ctx context.Context, userID int64) ([]*domain.OrganizationWithRole, error) {
	return s.repo.ListByUser(ctx, userID)
}

// GetByID retrieves a single organization.
func (s *OrganizationService) GetByID(ctx context.Context, orgID int64) (*domain.Organization, error) {
	return s.repo.GetByID(ctx, orgID)
}

// Update modifies organization details.
func (s *OrganizationService) Update(ctx context.Context, orgID int64, name, description, logoURL *string) (*domain.Organization, error) {
	return s.repo.Update(ctx, orgID, name, description, logoURL)
}

// GetMemberRole returns a user's role in an organization.
func (s *OrganizationService) GetMemberRole(ctx context.Context, orgID, userID int64) (string, error) {
	return s.repo.GetMemberRole(ctx, orgID, userID)
}

// ListMembers returns all members of an organization.
func (s *OrganizationService) ListMembers(ctx context.Context, orgID int64) ([]*domain.OrgMember, error) {
	return s.repo.ListMembers(ctx, orgID)
}

// UpdateMemberRole changes a member's role, protecting the last admin.
func (s *OrganizationService) UpdateMemberRole(ctx context.Context, orgID, userID int64, role string) error {
	if role != domain.OrgRoleAdmin {
		currentRole, err := s.repo.GetMemberRole(ctx, orgID, userID)
		if err != nil {
			return err
		}
		if currentRole == domain.OrgRoleAdmin {
			count, err := s.repo.CountAdmins(ctx, orgID, &userID)
			if err != nil {
				return err
			}
			if count == 0 {
				return ErrOrgLastAdmin
			}
		}
	}
	return s.repo.UpdateMemberRole(ctx, orgID, userID, role)
}

// RemoveMember removes a member, protecting the last admin.
func (s *OrganizationService) RemoveMember(ctx context.Context, orgID, userID int64) error {
	currentRole, err := s.repo.GetMemberRole(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if currentRole == domain.OrgRoleAdmin {
		count, err := s.repo.CountAdmins(ctx, orgID, nil)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrOrgLastAdmin
		}
	}
	return s.repo.RemoveMember(ctx, orgID, userID)
}

func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// InviteUser creates an invitation and optionally sends an email.
func (s *OrganizationService) InviteUser(ctx context.Context, orgID, inviterID int64, targetEmail, role string) (*domain.OrgInvitation, error) {
	isMember, err := s.repo.IsAlreadyMember(ctx, orgID, targetEmail)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, ErrOrgAlreadyMember
	}

	hasPending, err := s.repo.HasPendingInvitation(ctx, orgID, targetEmail)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, ErrInvitationExists
	}

	token, err := generateInviteToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invID, err := s.repo.CreateInvitation(ctx, orgID, inviterID, targetEmail, role, token, expiresAt)
	if err != nil {
		return nil, err
	}

	orgName, _ := s.repo.GetOrgName(ctx, orgID)
	inviterName, _ := s.repo.GetUserName(ctx, inviterID)

	if s.emailClient != nil {
		go s.emailClient.SendInvitation(targetEmail, orgName, inviterName, token) //nolint:errcheck // async fire-and-forget
	}

	return &domain.OrgInvitation{
		ID:        invID,
		OrgID:     orgID,
		OrgName:   orgName,
		Email:     targetEmail,
		Role:      role,
		Status:    domain.InvitationStatusPending,
		InvitedBy: inviterName,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// ListUserInvitations returns pending invitations for a user's email.
func (s *OrganizationService) ListUserInvitations(ctx context.Context, email string) ([]*domain.OrgInvitation, error) {
	return s.repo.ListUserInvitations(ctx, email)
}

// AcceptInvitation accepts a pending invitation by token.
func (s *OrganizationService) AcceptInvitation(ctx context.Context, token string, userID int64, userEmail string) error {
	inv, err := s.repo.GetInvitationByToken(ctx, token)
	if err != nil {
		return err
	}
	if inv.Email != userEmail {
		return ErrInvitationWrongEmail
	}
	return s.repo.AcceptInvitation(ctx, inv.ID, inv.OrgID, userID, inv.Role)
}

// DeclineInvitation declines a pending invitation by token.
func (s *OrganizationService) DeclineInvitation(ctx context.Context, token, email string) error {
	return s.repo.DeclineInvitation(ctx, token, email)
}

// ListOrgInvitations lists all invitations for an organization.
func (s *OrganizationService) ListOrgInvitations(ctx context.Context, orgID int64) ([]*domain.OrgInvitation, error) {
	return s.repo.ListOrgInvitations(ctx, orgID)
}

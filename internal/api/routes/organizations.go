package routes

import (
	"dermify-api/config"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// OrgRoutes handles all organization-related routes.
type OrgRoutes struct {
	orgSvc  *service.OrganizationService
	authSvc *service.AuthService
	config  *config.Configuration
	metrics *metrics.Client
}

// NewOrgRoutes creates a new OrgRoutes instance.
func NewOrgRoutes(orgSvc *service.OrganizationService, authSvc *service.AuthService, cfg *config.Configuration, m *metrics.Client) *OrgRoutes {
	return &OrgRoutes{
		orgSvc:  orgSvc,
		authSvc: authSvc,
		config:  cfg,
		metrics: m,
	}
}

// RegisterRoutes registers all organization routes under the /orgs prefix.
func (or *OrgRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/orgs", func(r chi.Router) {
		r.Use(middleware.RequireAuth(or.config))

		r.Post("/", handlers.HandleCreateOrganization(or.orgSvc, or.authSvc, or.metrics))
		r.Get("/", handlers.HandleListUserOrganizations(or.orgSvc, or.metrics))

		r.Route("/{orgId}", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireOrgMember(or.orgSvc))
				r.Get("/", handlers.HandleGetOrganization(or.orgSvc, or.metrics))
				r.Get("/members", handlers.HandleListMembers(or.orgSvc, or.metrics))
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireOrgAdmin(or.orgSvc))
				r.Put("/", handlers.HandleUpdateOrganization(or.orgSvc, or.metrics))
				r.Put("/members/{userId}/role", handlers.HandleUpdateMemberRole(or.orgSvc, or.metrics))
				r.Put("/members/{userId}/must-change-password", handlers.HandleSetMemberMustChangePassword(or.orgSvc, or.authSvc, or.metrics))
				r.Delete("/members/{userId}", handlers.HandleRemoveMember(or.orgSvc, or.metrics))
				r.Post("/invitations", handlers.HandleInviteUser(or.orgSvc, or.metrics))
				r.Get("/invitations", handlers.HandleListOrgInvitations(or.orgSvc, or.metrics))
				r.Post("/invitations/{invitationId}/confirm", handlers.HandleConfirmInvitation(or.orgSvc, or.metrics))
			})
		})
	})
}

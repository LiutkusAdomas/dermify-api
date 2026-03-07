package domain

// RoleDoctor is the role constant for doctor users.
const RoleDoctor = "doctor"

// RoleAdmin is the role constant for admin users.
const RoleAdmin = "admin"

// ValidRole checks whether the given role string is a recognized role.
func ValidRole(role string) bool {
	return role == RoleDoctor || role == RoleAdmin
}

// AllRoles returns all valid roles in the system.
func AllRoles() []string {
	return []string{RoleDoctor, RoleAdmin}
}

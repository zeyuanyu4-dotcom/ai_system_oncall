package constant

// User roles (global)
const (
	RoleNormalUser  = "normal_user"
	RoleDeveloper   = "developer"
	RoleTester      = "tester"
	RoleProjectAdmin = "project_admin"
	RoleSystemAdmin = "system_admin"
)

// Project roles
const (
	ProjectRoleMember   = "member"
	ProjectRoleDeveloper = "developer"
	ProjectRoleTester    = "tester"
	ProjectRoleAdmin     = "admin"
	ProjectRoleOwner     = "owner"
)

// User status
const (
	UserStatusEnabled  = 1
	UserStatusDisabled = 0
)

// IsAdmin checks if user is project admin or higher
func IsProjectAdmin(role string) bool {
	return role == ProjectRoleAdmin || role == ProjectRoleOwner
}

// IsSystemAdmin checks if user has system admin role
func IsSystemAdmin(globalRole string) bool {
	return globalRole == RoleSystemAdmin
}

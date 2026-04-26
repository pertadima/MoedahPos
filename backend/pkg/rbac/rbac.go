package rbac

// IsSuperAdmin returns true if the role name is "superadmin".
func IsSuperAdmin(roleName string) bool {
	return roleName == "superadmin"
}

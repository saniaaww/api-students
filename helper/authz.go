package helper

// PermissionSet menyimpan daftar permission yang dimiliki oleh setiap role.
type PermissionSet struct {
	byRole map[string]map[string]struct{}
}

// NewPermissionSet membuat PermissionSet baru dari data role dan permission.
func NewPermissionSet(
	roles map[string][]string,
) *PermissionSet {
	ps := &PermissionSet{
		byRole: make(map[string]map[string]struct{}),
	}

	for role, permissions := range roles {
		ps.byRole[role] = make(map[string]struct{})

		for _, permission := range permissions {
			ps.byRole[role][permission] = struct{}{}
		}
	}

	return ps
}

func (ps *PermissionSet) Can(role, permission string) bool {
	if ps == nil {
		return false
	}

	permissions, ok := ps.byRole[role]
	if !ok {
		return false
	}

	_, ok = permissions[permission]
	return ok
}

// PermissionsOf mengembalikan daftar permission
// yang dimiliki oleh sebuah role.
func (ps *PermissionSet) PermissionsOf(role string) []string {
	if ps == nil {
		return []string{}
	}

	permissions, ok := ps.byRole[role]
	if !ok {
		return []string{}
	}

	result := make([]string, 0, len(permissions))

	for permission := range permissions {
		result = append(result, permission)
	}

	return result
}

// KnownRoles mengembalikan daftar role
// yang dikenal oleh sistem.
func (ps *PermissionSet) KnownRoles() []string {
	if ps == nil {
		return []string{}
	}

	result := make([]string, 0, len(ps.byRole))

	for role := range ps.byRole {
		result = append(result, role)
	}

	return result
}

// IsKnownRole mengecek apakah sebuah role
// terdaftar dalam PermissionSet.
func (ps *PermissionSet) IsKnownRole(role string) bool {
	if ps == nil {
		return false
	}

	_, ok := ps.byRole[role]
	return ok
}
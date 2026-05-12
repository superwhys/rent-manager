package rbac

// CreateRoleParams 描述创建角色时需要满足的领域输入。
type CreateRoleParams struct {
	OrganizationID int
	Name           string
	IsDefaultOwner bool
}

// CreatePermissionGroupParams 描述创建权限组时需要满足的领域输入。
type CreatePermissionGroupParams struct {
	Code        string
	Name        string
	Description string
	SortOrder   int
}

// CreatePermissionParams 描述创建权限时需要满足的领域输入。
type CreatePermissionParams struct {
	Code        string
	APIPath     string
	Description string
}

// PermissionGroupPermissionsParams 描述权限组绑定权限时需要满足的领域输入。
type PermissionGroupPermissionsParams struct {
	PermissionGroupID int
	PermissionIDs     []int
}

// RolePermissionGroupsParams 描述角色绑定权限组时需要满足的领域输入。
type RolePermissionGroupsParams struct {
	OrganizationID     int
	RoleID             int
	PermissionGroupIDs []int
}

// OrganizationMemberRolesParams 描述组织成员绑定角色时需要满足的领域输入。
type OrganizationMemberRolesParams struct {
	OrganizationID int
	AccountID      int
	RoleIDs        []int
}

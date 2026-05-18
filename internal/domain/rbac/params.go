package rbac

// CreateRoleParams 描述创建角色时需要满足的领域输入。
type CreateRoleParams struct {
	AccountID      int
	Name           string
	IsDefaultOwner bool
}

// UpdateRoleParams 描述更新角色时允许变更的领域输入。
type UpdateRoleParams struct {
	ID             int
	AccountID      int
	Name           string
	IsDefaultOwner bool
	IsActive       bool
}

// ListRolesQuery 描述角色列表查询条件。
type ListRolesQuery struct {
	AccountID int
	IsActive  *bool
	Limit     int
	Offset    int
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
	Description string
}

// PermissionGroupPermissionsParams 描述权限组绑定权限时需要满足的领域输入。
type PermissionGroupPermissionsParams struct {
	PermissionGroupID int
	PermissionIDs     []int
}

// RolePermissionGroupsParams 描述角色绑定权限组时需要满足的领域输入。
type RolePermissionGroupsParams struct {
	AccountID          int
	RoleID             int
	PermissionGroupIDs []int
}

// UnitAccountRolesParams 描述单元授权账号角色时需要满足的领域输入。
type UnitAccountRolesParams struct {
	OwnerAccountID  int
	UnitID          int
	TargetAccountID int
	RoleIDs         []int
}

// ListUnitGrantsQuery 描述单元授权列表查询条件。
type ListUnitGrantsQuery struct {
	OwnerAccountID  int
	UnitID          int
	TargetAccountID int
	Status          string
	Limit           int
	Offset          int
}

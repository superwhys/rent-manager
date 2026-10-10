package rbac

type CreateRoleParams struct {
	AccountID      string
	Name           string
	IsDefaultOwner bool
}

type UpdateRoleParams struct {
	ID             int
	AccountID      string
	Name           string
	IsDefaultOwner bool
	IsActive       bool
}

type ListRolesQuery struct {
	AccountID string
	IsActive  *bool
	Limit     int
	Offset    int
}

type CreatePermissionGroupParams struct {
	Code        string
	Name        string
	Description string
	SortOrder   int
}

type PermissionGroupPermissionsParams struct {
	PermissionGroupID int
	PermissionCodes   []string
}

type RolePermissionGroupsParams struct {
	AccountID          string
	RoleID             int
	PermissionGroupIDs []int
}

type UnitAccountRolesParams struct {
	OwnerAccountID  string
	UnitID          int
	TargetAccountID string
	RoleIDs         []int
}

type ListUnitGrantsQuery struct {
	OwnerAccountID  string
	UnitID          int
	TargetAccountID string
	Status          string
	Limit           int
	Offset          int
}

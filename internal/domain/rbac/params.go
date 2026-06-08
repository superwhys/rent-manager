package rbac

type CreateRoleParams struct {
	AccountID      int
	Name           string
	IsDefaultOwner bool
}

type UpdateRoleParams struct {
	ID             int
	AccountID      int
	Name           string
	IsDefaultOwner bool
	IsActive       bool
}

type ListRolesQuery struct {
	AccountID int
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

type CreatePermissionParams struct {
	Code        string
	Description string
}

type PermissionGroupPermissionsParams struct {
	PermissionGroupID int
	PermissionIDs     []int
}

type RolePermissionGroupsParams struct {
	AccountID          int
	RoleID             int
	PermissionGroupIDs []int
}

type UnitAccountRolesParams struct {
	OwnerAccountID  int
	UnitID          int
	TargetAccountID int
	RoleIDs         []int
}

type ListUnitGrantsQuery struct {
	OwnerAccountID  int
	UnitID          int
	TargetAccountID int
	Status          string
	Limit           int
	Offset          int
}

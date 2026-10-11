package rbac

import domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"

// CreateRoleItem 表示批量创建角色中的单个角色参数。
type CreateRoleItem struct {
	Name           string `json:"name"             validate:"required"`
	IsDefaultOwner bool   `json:"is_default_owner"`
}

// CreateRolesCommand 表示批量创建角色命令。
type CreateRolesCommand struct {
	Roles []CreateRoleItem `json:"roles" validate:"required,min=1,dive"`
}

// CreatePermissionGroupItem 表示批量创建权限组中的单个权限组参数。
type CreatePermissionGroupItem struct {
	Code        string `json:"code"        validate:"required"`
	Name        string `json:"name"        validate:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// CreatePermissionGroupsCommand 表示批量创建权限组命令。
type CreatePermissionGroupsCommand struct {
	PermissionGroups []CreatePermissionGroupItem `json:"permission_groups" validate:"required,min=1,dive"`
}

// PermissionGroupPermissionsCommand 表示权限组绑定或解绑权限命令。
type PermissionGroupPermissionsCommand struct {
	PermissionGroupID int      `uri:"permission_group_id" validate:"required,min=1"`
	PermissionCodes   []string `                          validate:"required,min=1,dive,required" json:"permission_codes"`
}

// RolePermissionGroupsCommand 表示角色绑定或解绑权限组命令。
type RolePermissionGroupsCommand struct {
	RoleID             int   `uri:"role_id" validate:"required,min=1"`
	PermissionGroupIDs []int `              validate:"required,min=1,dive,min=1" json:"permission_group_ids"`
}

// RoleIDCommand 表示角色 ID 路径参数。
type RoleIDCommand struct {
	RoleID int `uri:"role_id" validate:"required,min=1"`
}

// UpdateRoleCommand 表示更新角色命令。
type UpdateRoleCommand struct {
	RoleID         int    `uri:"role_id" validate:"required,min=1"`
	Name           string `              validate:"required"       json:"name"`
	IsDefaultOwner bool   `                                        json:"is_default_owner"`
	IsActive       *bool  `                                        json:"is_active"`
}

// UnitAccountRolesCommand 表示单元账号角色授权命令。
type UnitAccountRolesCommand struct {
	UnitID    int    `uri:"unit_id"    validate:"required,min=1"`
	AccountID string `uri:"account_id" validate:"required,len=64,hexadecimal,lowercase"`
	RoleIDs   []int  `                 validate:"required,min=1,dive,min=1"             json:"role_ids"`
}

// UnitGrantsQuery 表示单元授权列表查询参数。
type UnitGrantsQuery struct {
	UnitID    int    `uri:"unit_id" validate:"required,min=1"`
	AccountID string `              validate:"omitempty,len=64,hexadecimal,lowercase" form:"account_id"`
	Status    string `                                                                form:"status"`
	Limit     int    `                                                                form:"limit"`
	Offset    int    `                                                                form:"offset"`
}

// RoleResponse 是角色响应。
type RoleResponse struct {
	ID                 int    `json:"id"`
	AccountID          string `json:"account_id"`
	Name               string `json:"name"`
	IsDefaultOwner     bool   `json:"is_default_owner"`
	IsActive           bool   `json:"is_active"`
	PermissionGroupIDs []int  `json:"permission_group_ids"`
}

// PermissionGroupResponse 是权限组响应。
type PermissionGroupResponse struct {
	ID              int      `json:"id"`
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Description     *string  `json:"description,omitempty"`
	SortOrder       int      `json:"sort_order"`
	IsActive        bool     `json:"is_active"`
	PermissionCodes []string `json:"permission_codes"`
}

// PermissionResponse 是权限响应。
type PermissionResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// UnitGrantResponse 是单元授权响应。
type UnitGrantResponse struct {
	ID                 int           `json:"id"`
	UnitID             int           `json:"unit_id"`
	AccountID          string        `json:"account_id"`
	RoleID             int           `json:"role_id"`
	GrantedByAccountID string        `json:"granted_by_account_id"`
	Status             string        `json:"status"`
	Role               *RoleResponse `json:"role,omitempty"`
}

// ToRoleResponse 将角色领域对象转换为响应对象。
func ToRoleResponse(role *domainrbac.Role) *RoleResponse {
	if role == nil {
		return nil
	}
	return &RoleResponse{
		ID:                 role.ID,
		AccountID:          role.AccountID,
		Name:               role.Name,
		IsDefaultOwner:     role.IsDefaultOwner,
		IsActive:           role.IsActive,
		PermissionGroupIDs: append([]int{}, role.PermissionGroupIDs...),
	}
}

// ToRoleListResponse 将角色列表转换为响应列表。
func ToRoleListResponse(roles []*domainrbac.Role) []*RoleResponse {
	resp := make([]*RoleResponse, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, ToRoleResponse(role))
	}
	return resp
}

// ToPermissionGroupResponse 将权限组领域对象转换为响应对象。
func ToPermissionGroupResponse(group *domainrbac.PermissionGroup) *PermissionGroupResponse {
	return &PermissionGroupResponse{
		ID:              group.ID,
		Code:            group.Code,
		Name:            group.Name,
		Description:     group.Description,
		SortOrder:       group.SortOrder,
		IsActive:        group.IsActive,
		PermissionCodes: append([]string{}, group.PermissionCodes...),
	}
}

// ToPermissionGroupListResponse 将权限组列表转换为响应列表。
func ToPermissionGroupListResponse(
	groups []*domainrbac.PermissionGroup,
) []*PermissionGroupResponse {
	resp := make([]*PermissionGroupResponse, 0, len(groups))
	for _, group := range groups {
		resp = append(resp, ToPermissionGroupResponse(group))
	}
	return resp
}

// ToPermissionResponse 将权限领域对象转换为响应对象。
func ToPermissionResponse(permission *domainrbac.Permission) *PermissionResponse {
	return &PermissionResponse{
		Code:        permission.Code,
		Description: permission.Description,
	}
}

// ToPermissionListResponse 将权限列表转换为响应列表。
func ToPermissionListResponse(permissions []*domainrbac.Permission) []*PermissionResponse {
	resp := make([]*PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		resp = append(resp, ToPermissionResponse(permission))
	}
	return resp
}

// ToUnitGrantResponse 将单元授权领域对象转换为响应对象。
func ToUnitGrantResponse(grant *domainrbac.UnitAccountRole) *UnitGrantResponse {
	return &UnitGrantResponse{
		ID:                 grant.ID,
		UnitID:             grant.UnitID,
		AccountID:          grant.AccountID,
		RoleID:             grant.RoleID,
		GrantedByAccountID: grant.GrantedByAccountID,
		Status:             grant.Status,
		Role:               ToRoleResponse(grant.Role),
	}
}

// ToUnitGrantListResponse 将单元授权列表转换为响应列表。
func ToUnitGrantListResponse(grants []*domainrbac.UnitAccountRole) []*UnitGrantResponse {
	resp := make([]*UnitGrantResponse, 0, len(grants))
	for _, grant := range grants {
		resp = append(resp, ToUnitGrantResponse(grant))
	}
	return resp
}

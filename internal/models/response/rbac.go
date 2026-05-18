package response

import domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"

type RoleResponse struct {
	ID             int    `json:"id"`
	AccountID      int    `json:"account_id"`
	Name           string `json:"name"`
	IsDefaultOwner bool   `json:"is_default_owner"`
	IsActive       bool   `json:"is_active"`
}

type PermissionGroupResponse struct {
	ID          int     `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	SortOrder   int     `json:"sort_order"`
	IsActive    bool    `json:"is_active"`
}

type PermissionResponse struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type UnitGrantResponse struct {
	ID                 int           `json:"id"`
	UnitID             int           `json:"unit_id"`
	AccountID          int           `json:"account_id"`
	RoleID             int           `json:"role_id"`
	GrantedByAccountID int           `json:"granted_by_account_id"`
	Status             string        `json:"status"`
	Role               *RoleResponse `json:"role,omitempty"`
}

func ToRoleResponse(role *domainrbac.Role) *RoleResponse {
	if role == nil {
		return nil
	}
	return &RoleResponse{
		ID:             role.ID,
		AccountID:      role.AccountID,
		Name:           role.Name,
		IsDefaultOwner: role.IsDefaultOwner,
		IsActive:       role.IsActive,
	}
}

func ToRoleListResponse(roles []*domainrbac.Role) []*RoleResponse {
	resp := make([]*RoleResponse, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, ToRoleResponse(role))
	}
	return resp
}

func ToPermissionGroupResponse(group *domainrbac.PermissionGroup) *PermissionGroupResponse {
	if group == nil {
		return nil
	}
	return &PermissionGroupResponse{
		ID:          group.ID,
		Code:        group.Code,
		Name:        group.Name,
		Description: group.Description,
		SortOrder:   group.SortOrder,
		IsActive:    group.IsActive,
	}
}

func ToPermissionGroupListResponse(groups []*domainrbac.PermissionGroup) []*PermissionGroupResponse {
	resp := make([]*PermissionGroupResponse, 0, len(groups))
	for _, group := range groups {
		resp = append(resp, ToPermissionGroupResponse(group))
	}
	return resp
}

func ToPermissionResponse(permission *domainrbac.Permission) *PermissionResponse {
	if permission == nil {
		return nil
	}
	return &PermissionResponse{
		ID:          permission.ID,
		Code:        permission.Code,
		Description: permission.Description,
		IsActive:    permission.IsActive,
	}
}

func ToPermissionListResponse(permissions []*domainrbac.Permission) []*PermissionResponse {
	resp := make([]*PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		resp = append(resp, ToPermissionResponse(permission))
	}
	return resp
}

func ToUnitGrantResponse(grant *domainrbac.UnitAccountRole) *UnitGrantResponse {
	if grant == nil {
		return nil
	}
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

func ToUnitGrantListResponse(grants []*domainrbac.UnitAccountRole) []*UnitGrantResponse {
	resp := make([]*UnitGrantResponse, 0, len(grants))
	for _, grant := range grants {
		resp = append(resp, ToUnitGrantResponse(grant))
	}
	return resp
}

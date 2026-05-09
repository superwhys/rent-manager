package response

import domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"

type RoleResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type PermissionResponse struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	APIPath     string `json:"api_path"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func ToRoleResponse(role *domainrbac.Role) *RoleResponse {
	if role == nil {
		return nil
	}
	return &RoleResponse{
		ID:       role.ID,
		Name:     role.Name,
		IsActive: role.IsActive,
	}
}

func ToRoleListResponse(roles []*domainrbac.Role) []*RoleResponse {
	resp := make([]*RoleResponse, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, ToRoleResponse(role))
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
		APIPath:     permission.APIPath,
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

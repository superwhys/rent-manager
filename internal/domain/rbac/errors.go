package rbac

import "errors"

var (
	// ErrRoleOrganizationRequired 表示角色所属组织不能为空。
	ErrRoleOrganizationRequired = errors.New("role organization is required")
	// ErrRoleNameRequired 表示角色名称不能为空。
	ErrRoleNameRequired = errors.New("role name is required")
	// ErrPermissionGroupCodeRequired 表示权限组代码不能为空。
	ErrPermissionGroupCodeRequired = errors.New("permission group code is required")
	// ErrPermissionGroupNameRequired 表示权限组名称不能为空。
	ErrPermissionGroupNameRequired = errors.New("permission group name is required")
	// ErrPermissionCodeRequired 表示权限代码不能为空。
	ErrPermissionCodeRequired = errors.New("permission code is required")
	// ErrPermissionAPIPathRequired 表示权限路径不能为空。
	ErrPermissionAPIPathRequired = errors.New("permission api path is required")
	// ErrPermissionDescriptionRequired 表示权限描述不能为空。
	ErrPermissionDescriptionRequired = errors.New("permission description is required")
	// ErrRoleNotInOrganization 表示角色不属于指定组织。
	ErrRoleNotInOrganization = errors.New("role is not in organization")
	// ErrAccountNotInOrganization 表示账号不属于指定组织。
	ErrAccountNotInOrganization = errors.New("account is not in organization")
	// ErrCrossOrganizationRoleAssignment 表示跨组织绑定角色。
	ErrCrossOrganizationRoleAssignment = errors.New("cannot assign roles across organizations")
)

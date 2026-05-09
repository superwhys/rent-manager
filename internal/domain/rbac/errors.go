package rbac

import "errors"

var (
	// ErrRoleNameRequired 表示角色名称不能为空。
	ErrRoleNameRequired = errors.New("role name is required")
	// ErrPermissionCodeRequired 表示权限代码不能为空。
	ErrPermissionCodeRequired = errors.New("permission code is required")
	// ErrPermissionAPIPathRequired 表示权限路径不能为空。
	ErrPermissionAPIPathRequired = errors.New("permission api path is required")
	// ErrPermissionDescriptionRequired 表示权限描述不能为空。
	ErrPermissionDescriptionRequired = errors.New("permission description is required")
)

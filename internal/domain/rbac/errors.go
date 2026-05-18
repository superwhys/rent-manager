package rbac

import "errors"

var (
	// ErrRoleAccountRequired 表示角色所属账号不能为空。
	ErrRoleAccountRequired = errors.New("role account is required")
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
	// ErrPermissionGroupAlreadyExists 表示权限组已经存在。
	ErrPermissionGroupAlreadyExists = errors.New("permission group already exists")
	// ErrPermissionAlreadyExists 表示权限已经存在。
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	// ErrPermissionGroupNotFound 表示权限组不存在。
	ErrPermissionGroupNotFound = errors.New("permission group not found")
	// ErrPermissionNotFound 表示权限不存在。
	ErrPermissionNotFound = errors.New("permission not found")
	// ErrRoleNotInAccount 表示角色不属于当前账号。
	ErrRoleNotInAccount = errors.New("role is not in account")
	// ErrUnitNotInAccount 表示单元不属于当前账号。
	ErrUnitNotInAccount = errors.New("unit is not in account")
	// ErrAccountNotFound 表示账号不存在。
	ErrAccountNotFound = errors.New("account not found")
)

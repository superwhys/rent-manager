package rbac

import "errors"

var (
	// ErrRoleAccountRequired 表示角色所属账号不能为空。
	ErrRoleAccountRequired = errors.New("角色所属账号不能为空")
	// ErrRoleNameRequired 表示角色名称不能为空。
	ErrRoleNameRequired = errors.New("角色名称不能为空")
	// ErrPermissionGroupCodeRequired 表示权限组代码不能为空。
	ErrPermissionGroupCodeRequired = errors.New("权限组编码不能为空")
	// ErrPermissionGroupNameRequired 表示权限组名称不能为空。
	ErrPermissionGroupNameRequired = errors.New("权限组名称不能为空")
	// ErrPermissionCodeRequired 表示权限代码不能为空。
	ErrPermissionCodeRequired = errors.New("权限编码不能为空")
	// ErrPermissionDescriptionRequired 表示权限描述不能为空。
	ErrPermissionDescriptionRequired = errors.New("权限描述不能为空")
	// ErrPermissionGroupAlreadyExists 表示权限组已经存在。
	ErrPermissionGroupAlreadyExists = errors.New("权限组已存在")
	// ErrPermissionAlreadyExists 表示权限已经存在。
	ErrPermissionAlreadyExists = errors.New("权限已存在")
	// ErrPermissionGroupNotFound 表示权限组不存在。
	ErrPermissionGroupNotFound = errors.New("权限组不存在")
	// ErrPermissionNotFound 表示权限不存在。
	ErrPermissionNotFound = errors.New("权限不存在")
	// ErrRoleNotInAccount 表示角色不属于当前账号。
	ErrRoleNotInAccount = errors.New("角色不属于当前账号")
	// ErrUnitNotInAccount 表示单元不属于当前账号。
	ErrUnitNotInAccount = errors.New("单元不属于当前账号")
	// ErrAccountNotFound 表示账号不存在。
	ErrAccountNotFound = errors.New("账号不存在")
	// ErrRoleNotFound 表示角色不存在。
	ErrRoleNotFound = errors.New("角色不存在")
	// ErrRoleAlreadyExists 表示角色名称已存在。
	ErrRoleAlreadyExists = errors.New("角色已存在")
	// ErrPermissionGroupIDsRequired 表示角色必须绑定权限组。
	ErrPermissionGroupIDsRequired = errors.New("权限组 ID 不能为空")
	// ErrRoleIDsRequired 表示授权必须包含角色。
	ErrRoleIDsRequired = errors.New("角色 ID 不能为空")
	// ErrUnitRequired 表示单元不能为空。
	ErrUnitRequired = errors.New("单元不能为空")
	// ErrTargetAccountRequired 表示被授权账号不能为空。
	ErrTargetAccountRequired = errors.New("被授权账号不能为空")
	// ErrCannotGrantToOwner 表示不能把自己的单元授权给自己。
	ErrCannotGrantToOwner = errors.New("不能将单元授权给拥有者自己")
	// ErrUnitGrantNotFound 表示单元授权记录不存在。
	ErrUnitGrantNotFound = errors.New("单元授权记录不存在")
	// ErrForbidden 表示没有权限访问目标资源。
	ErrForbidden = errors.New("无权访问")
	// ErrPermissionCodeEmpty 表示权限校验 code 不能为空。
	ErrPermissionCodeEmpty = errors.New("权限编码不能为空")
)

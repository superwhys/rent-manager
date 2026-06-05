package rbac

import "git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"

var (
	// ErrRoleAccountRequired 表示角色所属账号不能为空。
	ErrRoleAccountRequired = xerr.New(207001, "角色所属账号不能为空")
	// ErrRoleNameRequired 表示角色名称不能为空。
	ErrRoleNameRequired = xerr.New(207002, "角色名称不能为空")
	// ErrPermissionGroupCodeRequired 表示权限组代码不能为空。
	ErrPermissionGroupCodeRequired = xerr.New(207003, "权限组编码不能为空")
	// ErrPermissionGroupNameRequired 表示权限组名称不能为空。
	ErrPermissionGroupNameRequired = xerr.New(207004, "权限组名称不能为空")
	// ErrPermissionCodeRequired 表示权限代码不能为空。
	ErrPermissionCodeRequired = xerr.New(207005, "权限编码不能为空")
	// ErrPermissionDescriptionRequired 表示权限描述不能为空。
	ErrPermissionDescriptionRequired = xerr.New(207006, "权限描述不能为空")
	// ErrPermissionGroupAlreadyExists 表示权限组已经存在。
	ErrPermissionGroupAlreadyExists = xerr.New(207007, "权限组已存在")
	// ErrPermissionAlreadyExists 表示权限已经存在。
	ErrPermissionAlreadyExists = xerr.New(207008, "权限已存在")
	// ErrPermissionGroupNotFound 表示权限组不存在。
	ErrPermissionGroupNotFound = xerr.New(207009, "权限组不存在")
	// ErrPermissionNotFound 表示权限不存在。
	ErrPermissionNotFound = xerr.New(207010, "权限不存在")
	// ErrRoleNotInAccount 表示角色不属于当前账号。
	ErrRoleNotInAccount = xerr.New(207011, "角色不属于当前账号")
	// ErrUnitNotInAccount 表示单元不属于当前账号。
	ErrUnitNotInAccount = xerr.New(207012, "单元不属于当前账号")
	// ErrAccountNotFound 表示账号不存在。
	ErrAccountNotFound = xerr.New(207013, "账号不存在")
	// ErrRoleNotFound 表示角色不存在。
	ErrRoleNotFound = xerr.New(207014, "角色不存在")
	// ErrRoleAlreadyExists 表示角色名称已存在。
	ErrRoleAlreadyExists = xerr.New(207015, "角色已存在")
	// ErrPermissionGroupIDsRequired 表示角色必须绑定权限组。
	ErrPermissionGroupIDsRequired = xerr.New(207016, "权限组 ID 不能为空")
	// ErrRoleIDsRequired 表示授权必须包含角色。
	ErrRoleIDsRequired = xerr.New(207017, "角色 ID 不能为空")
	// ErrUnitRequired 表示单元不能为空。
	ErrUnitRequired = xerr.New(207018, "单元不能为空")
	// ErrTargetAccountRequired 表示被授权账号不能为空。
	ErrTargetAccountRequired = xerr.New(207019, "被授权账号不能为空")
	// ErrCannotGrantToOwner 表示不能把自己的单元授权给自己。
	ErrCannotGrantToOwner = xerr.New(207020, "不能将单元授权给拥有者自己")
	// ErrUnitGrantNotFound 表示单元授权记录不存在。
	ErrUnitGrantNotFound = xerr.New(207021, "单元授权记录不存在")
	// ErrForbidden 表示没有权限访问目标资源。
	ErrForbidden = xerr.New(207022, "无权访问")
	// ErrPermissionCodeEmpty 表示权限校验 code 不能为空。
	ErrPermissionCodeEmpty = xerr.New(207023, "权限编码不能为空")
)

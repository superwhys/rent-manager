package rbac

// CreateRoleParams 描述创建角色时需要满足的领域输入。
type CreateRoleParams struct {
	Name string
}

// CreatePermissionParams 描述创建权限时需要满足的领域输入。
type CreatePermissionParams struct {
	Code        string
	APIPath     string
	Description string
}

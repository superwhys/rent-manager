package account

import "errors"

var (
	// ErrOrganizationNameRequired 表示组织名称不能为空。
	ErrOrganizationNameRequired = errors.New("organization name is required")
	// ErrOrganizationIDRequired 表示组织 ID 不能为空。
	ErrOrganizationIDRequired = errors.New("organization id is required")
	// ErrAccountNotFound 表示账号不存在。
	ErrAccountNotFound = errors.New("account not found")
	// ErrOrganizationNotFound 表示组织不存在。
	ErrOrganizationNotFound = errors.New("organization not found")
	// ErrAccountNotInOrganization 表示账号不属于指定组织。
	ErrAccountNotInOrganization = errors.New("account is not in organization")
)

package account

// CreateOrganizationParams describes organization creation input.
type CreateOrganizationParams struct {
	Name string
}

// AssignAccountRolesParams describes account role updates inside its organization.
type AssignAccountRolesParams struct {
	OrganizationID int
	AccountID      int
	RoleIDs        []int
}

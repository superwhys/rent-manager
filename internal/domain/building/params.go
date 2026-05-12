package building

type CreateParams struct {
	OrganizationID      int
	Name                string
	Address             string
	City                string
	Province            string
	DefaultMeterItemIDs []int
}

type UpdateParams struct {
	ID                  int
	OrganizationID      int
	Name                string
	Address             string
	City                string
	Province            string
	DefaultMeterItemIDs []int
}

type ListQuery struct {
	OrganizationID int
	Name           string
	City           string
	Province       string
	Limit          int
	Offset         int
}

package unit

const (
	ListScopeOwned   = "owned"
	ListScopeGranted = "granted"
	ListScopeAll     = "all"

	AccessTypeOwned   = "owned"
	AccessTypeGranted = "granted"
)

type CreateParams struct {
	AccountID           string
	Name                string
	Address             string
	City                string
	Province            string
	DefaultMeterItemIDs []int
	Config              *UnitConfig
}

type UpdateParams struct {
	ID                  int
	AccountID           string
	Name                string
	Address             string
	City                string
	Province            string
	DefaultMeterItemIDs []int
	Config              *UnitConfig
}

type ListQuery struct {
	AccountID string
	Scope     string
}

package unit

const (
	ListScopeOwned   = "owned"
	ListScopeGranted = "granted"
	ListScopeAll     = "all"

	AccessTypeOwned   = "owned"
	AccessTypeGranted = "granted"
)

type CreateParams struct {
	AccountID           int
	Name                string
	Address             string
	City                string
	Province            string
	DefaultMeterItemIDs []int
}

type UpdateParams struct {
	ID                  int
	AccountID           int
	Name                string
	Address             string
	City                string
	Province            string
	DefaultMeterItemIDs []int
}

type ListQuery struct {
	AccountID int
	Scope     string
	Name      string
	City      string
	Province  string
	Limit     int
	Offset    int
}

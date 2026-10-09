package roomtype

type CreateParams struct {
	AccountID string
	Name      string
}

type UpdateParams struct {
	ID        int
	AccountID string
	Name      string
}

type ListQuery struct {
	AccountID string
	Name      string
	Limit     int
	Offset    int
}

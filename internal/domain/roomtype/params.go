package roomtype

type CreateParams struct {
	AccountID int
	Name      string
}

type UpdateParams struct {
	ID        int
	AccountID int
	Name      string
}

type ListQuery struct {
	AccountID int
	Name      string
	Limit     int
	Offset    int
}

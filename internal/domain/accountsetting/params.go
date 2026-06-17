package accountsetting

type CreateParams struct {
	AccountID int
	Key       string
	Value     string
}

type UpdateParams struct {
	ID        int
	AccountID int
	Key       string
	Value     string
}

type ListQuery struct {
	AccountID int
}

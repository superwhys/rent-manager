package accountsetting

type CreateParams struct {
	AccountID string
	Key       string
	Value     string
}

type UpdateParams struct {
	ID        int
	AccountID string
	Key       string
	Value     string
}

type ListQuery struct {
	AccountID string
}

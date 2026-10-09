package accountsetting

import (
	"strings"
	"time"
)

type AccountSetting struct {
	ID        int
	AccountID string
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewAccountSetting(params CreateParams) (*AccountSetting, error) {
	setting := &AccountSetting{
		AccountID: params.AccountID,
		Key:       normalizeKey(params.Key),
		Value:     strings.TrimSpace(params.Value),
	}
	if err := setting.Validate(); err != nil {
		return nil, err
	}
	return setting, nil
}

func (s *AccountSetting) Update(params UpdateParams) error {
	s.ID = params.ID
	s.AccountID = params.AccountID
	s.Key = normalizeKey(params.Key)
	s.Value = strings.TrimSpace(params.Value)
	return s.Validate()
}

func (s *AccountSetting) Validate() error {
	if strings.TrimSpace(s.AccountID) == "" {
		return ErrAccountRequired
	}
	if s.Key == "" {
		return ErrKeyRequired
	}
	if s.Value == "" {
		return ErrValueRequired
	}
	return nil
}

func normalizeKey(key string) string {
	return strings.TrimSpace(key)
}

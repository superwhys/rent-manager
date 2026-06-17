package accountsetting

import (
	"strconv"
	"strings"
	"time"
)

type AccountSetting struct {
	ID        int
	AccountID int
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
	if s.AccountID <= 0 {
		return ErrAccountRequired
	}
	if s.Key == "" {
		return ErrKeyRequired
	}
	if s.Value == "" {
		return ErrValueRequired
	}
	switch s.Key {
	case KeyMeterReadingDay:
		return validateMeterReadingDayValue(s.Value)
	default:
		return ErrUnsupportedKey
	}
}

func MeterReadingDayValue(day int) (string, error) {
	value := strconv.Itoa(day)
	if err := validateMeterReadingDayValue(value); err != nil {
		return "", err
	}
	return value, nil
}

func ParseMeterReadingDay(value string) (int, error) {
	value = strings.TrimSpace(value)
	day, err := strconv.Atoi(value)
	if err != nil {
		return 0, ErrInvalidMeterReadingDay
	}
	if day < MinMeterReadingDay || day > MaxMeterReadingDay {
		return 0, ErrInvalidMeterReadingDay
	}
	return day, nil
}

func validateMeterReadingDayValue(value string) error {
	_, err := ParseMeterReadingDay(value)
	return err
}

func normalizeKey(key string) string {
	return strings.TrimSpace(key)
}

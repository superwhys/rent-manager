package accountsetting

import (
	"errors"
	"testing"
)

const (
	testAccountID10 = "000000000000000000000000000000000000000000000000000000000000000a"
)

func TestNewAccountSettingAcceptsAnyKeyValue(t *testing.T) {
	setting, err := NewAccountSetting(CreateParams{
		AccountID: testAccountID10,
		Key:       "theme",
		Value:     "dark",
	})
	if err != nil {
		t.Fatalf("NewAccountSetting() error = %v", err)
	}
	if setting.AccountID != testAccountID10 || setting.Key != "theme" || setting.Value != "dark" {
		t.Fatalf("setting = %#v, want account=10 key=theme value=dark", setting)
	}
}

func TestNewAccountSettingTrimsKeyValue(t *testing.T) {
	setting, err := NewAccountSetting(CreateParams{
		AccountID: testAccountID10,
		Key:       " theme ",
		Value:     " dark ",
	})
	if err != nil {
		t.Fatalf("NewAccountSetting() error = %v", err)
	}
	if setting.Key != "theme" || setting.Value != "dark" {
		t.Fatalf("setting = %#v, want trimmed key/value", setting)
	}
}

func TestNewAccountSettingRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name   string
		params CreateParams
		want   error
	}{
		{
			name:   "account required",
			params: CreateParams{Key: "theme", Value: "dark"},
			want:   ErrAccountRequired,
		},
		{
			name:   "key required",
			params: CreateParams{AccountID: testAccountID10, Value: "dark"},
			want:   ErrKeyRequired,
		},
		{
			name:   "value required",
			params: CreateParams{AccountID: testAccountID10, Key: "theme"},
			want:   ErrValueRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAccountSetting(tt.params)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

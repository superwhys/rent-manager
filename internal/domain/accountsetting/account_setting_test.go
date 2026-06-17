package accountsetting

import (
	"errors"
	"testing"
)

func TestNewAccountSettingAcceptsMeterReadingDay(t *testing.T) {
	setting, err := NewAccountSetting(CreateParams{
		AccountID: 10,
		Key:       KeyMeterReadingDay,
		Value:     "20",
	})
	if err != nil {
		t.Fatalf("NewAccountSetting() error = %v", err)
	}
	if setting.AccountID != 10 || setting.Key != KeyMeterReadingDay || setting.Value != "20" {
		t.Fatalf("setting = %#v, want account=10 key=%q value=20", setting, KeyMeterReadingDay)
	}
}

func TestNewAccountSettingRejectsUnsupportedKey(t *testing.T) {
	_, err := NewAccountSetting(CreateParams{
		AccountID: 10,
		Key:       "unknown",
		Value:     "1",
	})
	if !errors.Is(err, ErrUnsupportedKey) {
		t.Fatalf("error = %v, want ErrUnsupportedKey", err)
	}
}

func TestNewAccountSettingRejectsInvalidMeterReadingDay(t *testing.T) {
	for _, value := range []string{"0", "32", "abc"} {
		t.Run(value, func(t *testing.T) {
			_, err := NewAccountSetting(CreateParams{
				AccountID: 10,
				Key:       KeyMeterReadingDay,
				Value:     value,
			})
			if !errors.Is(err, ErrInvalidMeterReadingDay) {
				t.Fatalf("error = %v, want ErrInvalidMeterReadingDay", err)
			}
		})
	}
}

func TestMeterReadingDayValue(t *testing.T) {
	value, err := MeterReadingDayValue(15)
	if err != nil {
		t.Fatalf("MeterReadingDayValue() error = %v", err)
	}
	if value != "15" {
		t.Fatalf("value = %q, want 15", value)
	}

	_, err = MeterReadingDayValue(32)
	if !errors.Is(err, ErrInvalidMeterReadingDay) {
		t.Fatalf("error = %v, want ErrInvalidMeterReadingDay", err)
	}
}

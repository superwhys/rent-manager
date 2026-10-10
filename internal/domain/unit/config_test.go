package unit

import (
	"errors"
	"testing"
)

const (
	testAccountID1 = "0000000000000000000000000000000000000000000000000000000000000001"
)

func TestDefaultUnitConfig(t *testing.T) {
	config := DefaultUnitConfig()
	if config.MeterReadingDay != DefaultMeterReadingDay {
		t.Fatalf("meter reading day = %d, want %d", config.MeterReadingDay, DefaultMeterReadingDay)
	}
}

func TestUnitConfigValidateMeterReadingDay(t *testing.T) {
	for _, day := range []int{1, 31} {
		t.Run("valid", func(t *testing.T) {
			config, err := NewUnitConfig(day)
			if err != nil {
				t.Fatalf("NewUnitConfig(%d) error = %v", day, err)
			}
			if config.MeterReadingDay != day {
				t.Fatalf("meter reading day = %d, want %d", config.MeterReadingDay, day)
			}
		})
	}

	for _, day := range []int{0, 32} {
		t.Run("invalid", func(t *testing.T) {
			_, err := NewUnitConfig(day)
			if !errors.Is(err, ErrInvalidMeterReadingDay) {
				t.Fatalf("error = %v, want ErrInvalidMeterReadingDay", err)
			}
		})
	}
}

func TestUnitConfigIgnoresRemovedBillingFields(t *testing.T) {
	tests := []string{
		`{"meter_reading_day":20,"billing_day":15,"billing_month":"current"}`,
		`{"meter_reading_day":20,"billing_day":-1,"billing_month":"next"}`,
		`{"meter_reading_day":20,"billing_day":32,"billing_month":"previous"}`,
		`{"meter_reading_day":20,"billing_day":"1","billing_month":1}`,
		`{"meter_reading_day":20,"billing_day":1.5,"billing_month":null}`,
	}
	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			config, err := ParseUnitConfigJSON(value)
			if err != nil {
				t.Fatalf("ParseUnitConfigJSON() error = %v", err)
			}
			if config.MeterReadingDay != 20 {
				t.Fatalf("meter reading day = %d, want 20", config.MeterReadingDay)
			}
			serialized, err := config.JSONString()
			if err != nil || serialized != `{"meter_reading_day":20}` {
				t.Fatalf("JSONString() = %q, %v; want meter_reading_day only", serialized, err)
			}
		})
	}
}

func TestUnitConfigJSONRoundTripDropsUnknownFields(t *testing.T) {
	config, err := ParseUnitConfigJSON(`{"meter_reading_day":20,"unknown":"ignored"}`)
	if err != nil {
		t.Fatalf("ParseUnitConfigJSON() error = %v", err)
	}
	if config.MeterReadingDay != 20 {
		t.Fatalf("meter reading day = %d, want 20", config.MeterReadingDay)
	}

	value, err := config.JSONString()
	if err != nil {
		t.Fatalf("JSONString() error = %v", err)
	}
	if value != `{"meter_reading_day":20}` {
		t.Fatalf("json = %q, want meter_reading_day only", value)
	}
}

func TestParseUnitConfigJSONRejectsInvalidValue(t *testing.T) {
	tests := []string{
		"",
		`null`,
		`[]`,
		`{}`,
		`{"meter_reading_day":0}`,
		`{"meter_reading_day":32}`,
		`{"meter_reading_day":"1"}`,
	}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			if _, err := ParseUnitConfigJSON(value); err == nil {
				t.Fatalf("ParseUnitConfigJSON(%q) error = nil, want error", value)
			}
		})
	}
}

func TestNewUnitUsesDefaultConfig(t *testing.T) {
	unit, err := NewUnit(CreateParams{
		AccountID: testAccountID1,
		Name:      "1栋",
		Address:   "地址",
		City:      "城市",
		Province:  "省份",
	})
	if err != nil {
		t.Fatalf("NewUnit() error = %v", err)
	}
	if unit.Config.MeterReadingDay != DefaultMeterReadingDay {
		t.Fatalf(
			"meter reading day = %d, want %d",
			unit.Config.MeterReadingDay,
			DefaultMeterReadingDay,
		)
	}
}

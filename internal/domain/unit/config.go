package unit

import (
	"encoding/json"
	"strings"
)

const (
	DefaultMeterReadingDay = 1
	MinMeterReadingDay     = 1
	MaxMeterReadingDay     = 31
)

// UnitConfig 描述单元扩展配置。
type UnitConfig struct {
	MeterReadingDay int `json:"meter_reading_day"`
}

// DefaultUnitConfig 返回单元扩展配置默认值。
func DefaultUnitConfig() UnitConfig {
	return UnitConfig{MeterReadingDay: DefaultMeterReadingDay}
}

func NewUnitConfig(meterReadingDay int) (UnitConfig, error) {
	config := UnitConfig{MeterReadingDay: meterReadingDay}
	if err := config.Validate(); err != nil {
		return UnitConfig{}, err
	}
	return config, nil
}

func ParseUnitConfigJSON(value string) (UnitConfig, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return UnitConfig{}, ErrInvalidJsonExtend
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(value), &raw); err != nil || raw == nil {
		return UnitConfig{}, ErrInvalidJsonExtend
	}

	var config UnitConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return UnitConfig{}, ErrInvalidJsonExtend
	}
	if err := config.Validate(); err != nil {
		return UnitConfig{}, err
	}
	return config, nil
}

func (c UnitConfig) Validate() error {
	if c.MeterReadingDay < MinMeterReadingDay || c.MeterReadingDay > MaxMeterReadingDay {
		return ErrInvalidMeterReadingDay
	}
	return nil
}

func (c UnitConfig) JSONString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return "", ErrInvalidJsonExtend
	}
	return string(data), nil
}

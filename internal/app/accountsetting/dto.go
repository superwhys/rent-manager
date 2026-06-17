package accountsetting

// UpdateSettingsCommand 是更新当前账号配置的请求体。
type UpdateSettingsCommand struct {
	MeterReadingDay int `json:"meter_reading_day" validate:"required,min=1,max=31"`
}

// Settings 表示当前账号配置视图。
type Settings struct {
	MeterReadingDay int
}

// SettingsResponse 是当前账号配置响应。
type SettingsResponse struct {
	MeterReadingDay int `json:"meter_reading_day"`
}

// ToSettingsResponse 将账号配置视图转换为响应。
func ToSettingsResponse(settings *Settings) SettingsResponse {
	if settings == nil {
		return SettingsResponse{}
	}
	return SettingsResponse{
		MeterReadingDay: settings.MeterReadingDay,
	}
}

package accountsetting

// UpdateSettingsCommand 是更新当前账号配置的请求体。
type UpdateSettingsCommand struct {
	Key   string `json:"key"   validate:"required,max=64"`
	Value string `json:"value" validate:"required,max=255"`
}

// Settings 表示当前账号配置视图。
type Settings map[string]string

// SettingsResponse 是当前账号配置响应。
type SettingsResponse map[string]string

// ToSettingsResponse 将账号配置视图转换为响应。
func ToSettingsResponse(settings *Settings) SettingsResponse {
	response := make(SettingsResponse, len(*settings))
	for key, value := range *settings {
		response[key] = value
	}
	return response
}

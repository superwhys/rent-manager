package accountsetting

import (
	"context"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainaccountsetting "github.com/superwhys/rent-manager/internal/domain/accountsetting"
)

// Service 提供当前账号配置用例。
type Service struct {
	repo domainaccountsetting.Repository
}

// NewService 创建账号配置服务。
func NewService(repo domainaccountsetting.Repository) *Service {
	return &Service{repo: repo}
}

// GetSettings 获取当前账号配置。
func (s *Service) GetSettings(ctx context.Context) (*Settings, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	settings := Settings{}
	records, err := s.repo.List(ctx, domainaccountsetting.ListQuery{AccountID: accountID})
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		settings[record.Key] = record.Value
	}
	return &settings, nil
}

// UpdateSettings 更新当前账号配置。
func (s *Service) UpdateSettings(
	ctx context.Context,
	cmd *UpdateSettingsCommand,
) (*Settings, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	setting, err := domainaccountsetting.NewAccountSetting(domainaccountsetting.CreateParams{
		AccountID: accountID,
		Key:       cmd.Key,
		Value:     cmd.Value,
	})
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}
	return s.GetSettings(ctx)
}

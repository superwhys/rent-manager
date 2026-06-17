package accountsetting

import (
	"context"

	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	domainaccountsetting "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/accountsetting"
)

// Service 提供当前账号配置用例。
type Service struct {
	repo domainaccountsetting.Repository
}

// NewService 创建账号配置服务。
func NewService(repo domainaccountsetting.Repository) *Service {
	return &Service{repo: repo}
}

// GetSettings 获取当前账号配置，缺失配置会使用默认值。
func (s *Service) GetSettings(ctx context.Context) (*Settings, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	settings := &Settings{MeterReadingDay: domainaccountsetting.DefaultMeterReadingDay}
	records, err := s.repo.List(ctx, domainaccountsetting.ListQuery{AccountID: accountID})
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record == nil {
			continue
		}
		if record.Key != domainaccountsetting.KeyMeterReadingDay {
			continue
		}
		day, err := domainaccountsetting.ParseMeterReadingDay(record.Value)
		if err != nil {
			return nil, err
		}
		settings.MeterReadingDay = day
	}
	return settings, nil
}

// UpdateSettings 更新当前账号配置。
func (s *Service) UpdateSettings(ctx context.Context, cmd *UpdateSettingsCommand) (*Settings, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	value, err := domainaccountsetting.MeterReadingDayValue(cmd.MeterReadingDay)
	if err != nil {
		return nil, err
	}
	setting, err := domainaccountsetting.NewAccountSetting(domainaccountsetting.CreateParams{
		AccountID: accountID,
		Key:       domainaccountsetting.KeyMeterReadingDay,
		Value:     value,
	})
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}
	return s.GetSettings(ctx)
}

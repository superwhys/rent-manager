package accountsetting

import (
	"context"
	"testing"

	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	domainaccountsetting "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/accountsetting"
)

func TestGetSettingsReturnsDefault(t *testing.T) {
	service := NewService(newFakeAccountSettingRepo())
	ctx := appcore.SetAccountIDToContext(context.Background(), 10)

	settings, err := service.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}
	if settings.MeterReadingDay != domainaccountsetting.DefaultMeterReadingDay {
		t.Fatalf("meter reading day = %d, want %d", settings.MeterReadingDay, domainaccountsetting.DefaultMeterReadingDay)
	}
}

func TestUpdateSettingsStoresMeterReadingDay(t *testing.T) {
	repo := newFakeAccountSettingRepo()
	service := NewService(repo)
	ctx := appcore.SetAccountIDToContext(context.Background(), 10)

	settings, err := service.UpdateSettings(ctx, &UpdateSettingsCommand{MeterReadingDay: 20})
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if settings.MeterReadingDay != 20 {
		t.Fatalf("meter reading day = %d, want 20", settings.MeterReadingDay)
	}
	stored := repo.settings[10][domainaccountsetting.KeyMeterReadingDay]
	if stored == nil {
		t.Fatalf("stored setting is nil")
	}
	if stored.AccountID != 10 || stored.Value != "20" {
		t.Fatalf("stored = %#v, want account=10 value=20", stored)
	}
}

func TestUpdateSettingsOverwritesExistingValue(t *testing.T) {
	repo := newFakeAccountSettingRepo()
	repo.seed(&domainaccountsetting.AccountSetting{
		AccountID: 10,
		Key:       domainaccountsetting.KeyMeterReadingDay,
		Value:     "15",
	})
	service := NewService(repo)
	ctx := appcore.SetAccountIDToContext(context.Background(), 10)

	settings, err := service.UpdateSettings(ctx, &UpdateSettingsCommand{MeterReadingDay: 25})
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if settings.MeterReadingDay != 25 {
		t.Fatalf("meter reading day = %d, want 25", settings.MeterReadingDay)
	}
	if len(repo.settings[10]) != 1 {
		t.Fatalf("settings count = %d, want 1", len(repo.settings[10]))
	}
	if repo.settings[10][domainaccountsetting.KeyMeterReadingDay].Value != "25" {
		t.Fatalf("stored value = %q, want 25", repo.settings[10][domainaccountsetting.KeyMeterReadingDay].Value)
	}
}

func TestSettingsAreScopedByAccount(t *testing.T) {
	repo := newFakeAccountSettingRepo()
	repo.seed(&domainaccountsetting.AccountSetting{
		AccountID: 20,
		Key:       domainaccountsetting.KeyMeterReadingDay,
		Value:     "22",
	})
	service := NewService(repo)
	ctx := appcore.SetAccountIDToContext(context.Background(), 10)

	settings, err := service.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}
	if settings.MeterReadingDay != domainaccountsetting.DefaultMeterReadingDay {
		t.Fatalf("meter reading day = %d, want default %d", settings.MeterReadingDay, domainaccountsetting.DefaultMeterReadingDay)
	}
	if _, err := service.UpdateSettings(ctx, &UpdateSettingsCommand{MeterReadingDay: 18}); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if repo.settings[20][domainaccountsetting.KeyMeterReadingDay].Value != "22" {
		t.Fatalf("account 20 value = %q, want 22", repo.settings[20][domainaccountsetting.KeyMeterReadingDay].Value)
	}
	if repo.settings[10][domainaccountsetting.KeyMeterReadingDay].Value != "18" {
		t.Fatalf("account 10 value = %q, want 18", repo.settings[10][domainaccountsetting.KeyMeterReadingDay].Value)
	}
}

type fakeAccountSettingRepo struct {
	settings map[int]map[string]*domainaccountsetting.AccountSetting
}

func newFakeAccountSettingRepo() *fakeAccountSettingRepo {
	return &fakeAccountSettingRepo{settings: make(map[int]map[string]*domainaccountsetting.AccountSetting)}
}

func (r *fakeAccountSettingRepo) seed(settings ...*domainaccountsetting.AccountSetting) {
	for _, setting := range settings {
		_, _ = r.Upsert(context.Background(), setting)
	}
}

func (r *fakeAccountSettingRepo) GetByKey(ctx context.Context, accountID int, key string) (*domainaccountsetting.AccountSetting, error) {
	if r.settings[accountID] == nil || r.settings[accountID][key] == nil {
		return nil, domainaccountsetting.ErrNotFound
	}
	return cloneAccountSetting(r.settings[accountID][key]), nil
}

func (r *fakeAccountSettingRepo) List(ctx context.Context, query domainaccountsetting.ListQuery) ([]*domainaccountsetting.AccountSetting, error) {
	settingsByKey := r.settings[query.AccountID]
	settings := make([]*domainaccountsetting.AccountSetting, 0, len(settingsByKey))
	for _, setting := range settingsByKey {
		settings = append(settings, cloneAccountSetting(setting))
	}
	return settings, nil
}

func (r *fakeAccountSettingRepo) Upsert(ctx context.Context, setting *domainaccountsetting.AccountSetting) (*domainaccountsetting.AccountSetting, error) {
	if r.settings[setting.AccountID] == nil {
		r.settings[setting.AccountID] = make(map[string]*domainaccountsetting.AccountSetting)
	}
	next := cloneAccountSetting(setting)
	if existing := r.settings[setting.AccountID][setting.Key]; existing != nil {
		next.ID = existing.ID
	}
	if next.ID == 0 {
		next.ID = len(r.settings[setting.AccountID]) + 1
	}
	r.settings[setting.AccountID][setting.Key] = next
	return cloneAccountSetting(next), nil
}

func cloneAccountSetting(setting *domainaccountsetting.AccountSetting) *domainaccountsetting.AccountSetting {
	if setting == nil {
		return nil
	}
	next := *setting
	return &next
}

var _ domainaccountsetting.Repository = (*fakeAccountSettingRepo)(nil)

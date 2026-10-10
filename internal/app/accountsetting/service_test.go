package accountsetting

import (
	"context"
	"testing"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainaccountsetting "github.com/superwhys/rent-manager/internal/domain/accountsetting"
)

const (
	testAccountID10 = "000000000000000000000000000000000000000000000000000000000000000a"
	testAccountID20 = "0000000000000000000000000000000000000000000000000000000000000014"
)

func TestGetSettingsReturnsAccountSettingsMap(t *testing.T) {
	repo := newFakeAccountSettingRepo()
	repo.seed(
		&domainaccountsetting.AccountSetting{
			AccountID: testAccountID10,
			Key:       "theme",
			Value:     "dark",
		},
		&domainaccountsetting.AccountSetting{
			AccountID: testAccountID10,
			Key:       "locale",
			Value:     "zh-CN",
		},
		&domainaccountsetting.AccountSetting{
			AccountID: testAccountID20,
			Key:       "theme",
			Value:     "light",
		},
	)
	service := NewService(repo)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	settings, err := service.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}
	if len(*settings) != 2 {
		t.Fatalf("settings count = %d, want 2", len(*settings))
	}
	if (*settings)["theme"] != "dark" || (*settings)["locale"] != "zh-CN" {
		t.Fatalf("settings = %#v, want account 10 settings", *settings)
	}
}

func TestUpdateSettingsStoresKeyValue(t *testing.T) {
	repo := newFakeAccountSettingRepo()
	service := NewService(repo)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	settings, err := service.UpdateSettings(
		ctx,
		&UpdateSettingsCommand{Key: "theme", Value: "dark"},
	)
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if (*settings)["theme"] != "dark" {
		t.Fatalf("settings = %#v, want theme=dark", *settings)
	}
	stored := repo.settings[testAccountID10]["theme"]
	if stored == nil {
		t.Fatalf("stored setting is nil")
	}
	if stored.AccountID != testAccountID10 || stored.Value != "dark" {
		t.Fatalf("stored = %#v, want account=10 value=dark", stored)
	}
}

func TestUpdateSettingsOverwritesExistingValue(t *testing.T) {
	repo := newFakeAccountSettingRepo()
	repo.seed(
		&domainaccountsetting.AccountSetting{
			AccountID: testAccountID10,
			Key:       "theme",
			Value:     "light",
		},
	)
	service := NewService(repo)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	settings, err := service.UpdateSettings(
		ctx,
		&UpdateSettingsCommand{Key: "theme", Value: "dark"},
	)
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if (*settings)["theme"] != "dark" {
		t.Fatalf("settings = %#v, want theme=dark", *settings)
	}
	if len(repo.settings[testAccountID10]) != 1 {
		t.Fatalf("settings count = %d, want 1", len(repo.settings[testAccountID10]))
	}
	if repo.settings[testAccountID10]["theme"].Value != "dark" {
		t.Fatalf("stored value = %q, want dark", repo.settings[testAccountID10]["theme"].Value)
	}
}

func TestSettingsAreScopedByAccount(t *testing.T) {
	repo := newFakeAccountSettingRepo()
	repo.seed(
		&domainaccountsetting.AccountSetting{
			AccountID: testAccountID20,
			Key:       "theme",
			Value:     "light",
		},
	)
	service := NewService(repo)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	settings, err := service.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}
	if len(*settings) != 0 {
		t.Fatalf("settings = %#v, want empty", *settings)
	}
	if _, err := service.UpdateSettings(
		ctx,
		&UpdateSettingsCommand{Key: "theme", Value: "dark"},
	); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if repo.settings[testAccountID20]["theme"].Value != "light" {
		t.Fatalf("account 20 value = %q, want light", repo.settings[testAccountID20]["theme"].Value)
	}
	if repo.settings[testAccountID10]["theme"].Value != "dark" {
		t.Fatalf("account 10 value = %q, want dark", repo.settings[testAccountID10]["theme"].Value)
	}
}

type fakeAccountSettingRepo struct {
	settings map[string]map[string]*domainaccountsetting.AccountSetting
}

func newFakeAccountSettingRepo() *fakeAccountSettingRepo {
	return &fakeAccountSettingRepo{
		settings: make(map[string]map[string]*domainaccountsetting.AccountSetting),
	}
}

func (r *fakeAccountSettingRepo) seed(settings ...*domainaccountsetting.AccountSetting) {
	for _, setting := range settings {
		_, _ = r.Upsert(context.Background(), setting)
	}
}

func (r *fakeAccountSettingRepo) GetByKey(
	ctx context.Context,
	accountID string,
	key string,
) (*domainaccountsetting.AccountSetting, error) {
	if r.settings[accountID] == nil || r.settings[accountID][key] == nil {
		return nil, domainaccountsetting.ErrNotFound
	}
	return cloneAccountSetting(r.settings[accountID][key]), nil
}

func (r *fakeAccountSettingRepo) List(
	ctx context.Context,
	query domainaccountsetting.ListQuery,
) ([]*domainaccountsetting.AccountSetting, error) {
	settingsByKey := r.settings[query.AccountID]
	settings := make([]*domainaccountsetting.AccountSetting, 0, len(settingsByKey))
	for _, setting := range settingsByKey {
		settings = append(settings, cloneAccountSetting(setting))
	}
	return settings, nil
}

func (r *fakeAccountSettingRepo) Upsert(
	ctx context.Context,
	setting *domainaccountsetting.AccountSetting,
) (*domainaccountsetting.AccountSetting, error) {
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

func cloneAccountSetting(
	setting *domainaccountsetting.AccountSetting,
) *domainaccountsetting.AccountSetting {
	if setting == nil {
		return nil
	}
	next := *setting
	return &next
}

var _ domainaccountsetting.Repository = (*fakeAccountSettingRepo)(nil)

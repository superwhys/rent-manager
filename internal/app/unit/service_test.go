package unit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

const (
	testAccountID10 = "000000000000000000000000000000000000000000000000000000000000000a"
)

func TestCreateUnitUsesDefaultConfig(t *testing.T) {
	service := newUnitTestService()
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	unit, err := service.CreateUnit(ctx, &CreateUnitCommand{
		Name:     "1栋",
		Address:  "地址",
		City:     "城市",
		Province: "省份",
	})
	if err != nil {
		t.Fatalf("CreateUnit() error = %v", err)
	}
	if unit.Config.MeterReadingDay != domainunit.DefaultMeterReadingDay {
		t.Fatalf(
			"meter reading day = %d, want %d",
			unit.Config.MeterReadingDay,
			domainunit.DefaultMeterReadingDay,
		)
	}
}

func TestCreateUnitStoresConfigFromJsonExtend(t *testing.T) {
	service := newUnitTestService()
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
	day := 20

	unit, err := service.CreateUnit(ctx, &CreateUnitCommand{
		Name:       "1栋",
		Address:    "地址",
		City:       "城市",
		Province:   "省份",
		JsonExtend: &UnitConfigCommand{MeterReadingDay: &day},
	})
	if err != nil {
		t.Fatalf("CreateUnit() error = %v", err)
	}
	if unit.Config.MeterReadingDay != 20 {
		t.Fatalf("meter reading day = %d, want 20", unit.Config.MeterReadingDay)
	}
}

func TestUpdateUnitPreservesConfigWhenJsonExtendOmitted(t *testing.T) {
	service := newUnitTestService()
	repo := service.repo.(*fakeUnitRepo)
	repo.seed(&domainunit.Unit{
		ID:        1,
		AccountID: testAccountID10,
		Name:      "1栋",
		Address:   "地址",
		City:      "城市",
		Province:  "省份",
		Config:    domainunit.UnitConfig{MeterReadingDay: 20},
	})
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	unit, err := service.UpdateUnit(ctx, &UpdateUnitCommand{
		UnitID:   1,
		Name:     "2栋",
		Address:  "新地址",
		City:     "城市",
		Province: "省份",
	})
	if err != nil {
		t.Fatalf("UpdateUnit() error = %v", err)
	}
	if unit.Config != (domainunit.UnitConfig{MeterReadingDay: 20}) {
		t.Fatalf("config = %+v, want meter reading day 20", unit.Config)
	}
}

func TestUpdateUnitMeterReadingConfig(t *testing.T) {
	meterReadingDay, invalidDay := 21, 32
	tests := []struct {
		name    string
		config  *UnitConfigCommand
		wantDay int
		wantErr error
	}{
		{name: "omitted config", wantDay: 20},
		{name: "empty config", config: &UnitConfigCommand{}, wantDay: 20},
		{name: "changes meter day", config: &UnitConfigCommand{MeterReadingDay: &meterReadingDay}, wantDay: 21},
		{name: "invalid meter day", config: &UnitConfigCommand{MeterReadingDay: &invalidDay}, wantErr: domainunit.ErrInvalidMeterReadingDay},
	}
	for _, owned := range []bool{true, false} {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				service := newUnitTestService()
				service.rbacService = &unitListRbacService{allowed: map[string]bool{domainrbac.PermissionUnitUpdate: true}}
				repo := service.repo.(*fakeUnitRepo)
				originalConfig := domainunit.UnitConfig{MeterReadingDay: 20}
				repo.seed(&domainunit.Unit{
					ID: 1, AccountID: testAccountID10, Name: "1栋", Address: "地址", City: "城市", Province: "省份",
					Config: originalConfig,
				})
				callerID := testAccountID10
				if !owned {
					callerID = "0000000000000000000000000000000000000000000000000000000000000014"
				}
				ctx := appcore.SetAccountIDToContext(context.Background(), callerID)
				unit, err := service.UpdateUnit(ctx, &UpdateUnitCommand{
					UnitID: 1, Name: "2栋", Address: "地址", City: "城市", Province: "省份", JsonExtend: tt.config,
				})
				if tt.wantErr != nil {
					if !errors.Is(err, tt.wantErr) || unit != nil {
						t.Fatalf("UpdateUnit() = %+v, %v; want nil, %v", unit, err, tt.wantErr)
					}
					if stored := repo.units[1]; stored.Config != originalConfig || stored.Name != "1栋" {
						t.Fatalf("rejected update changed stored unit: %+v", stored)
					}
					return
				}
				if err != nil {
					t.Fatalf("UpdateUnit() error = %v", err)
				}
				if unit.Config.MeterReadingDay != tt.wantDay {
					t.Fatalf("meter reading day = %d, want %d", unit.Config.MeterReadingDay, tt.wantDay)
				}
			})
		}
	}
}

func TestUnitResponseReturnsJsonExtendOnly(t *testing.T) {
	unit := &domainunit.Unit{
		ID:        1,
		AccountID: testAccountID10,
		Name:      "1栋",
		Address:   "地址",
		City:      "城市",
		Province:  "省份",
		Config:    domainunit.UnitConfig{MeterReadingDay: 20},
	}

	data, err := json.Marshal(ToUnitResponse(unit))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var response map[string]any
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if _, ok := response["meter_reading_day"]; ok {
		t.Fatalf("response has top-level meter_reading_day: %s", data)
	}
	jsonExtend, ok := response["json_extend"].(map[string]any)
	if !ok {
		t.Fatalf("json_extend = %#v, want object", response["json_extend"])
	}
	if jsonExtend["meter_reading_day"] != float64(20) {
		t.Fatalf("json_extend = %#v, want meter_reading_day=20", jsonExtend)
	}
	for _, field := range []string{"billing_day", "billing_month"} {
		if _, ok := jsonExtend[field]; ok {
			t.Fatalf("json_extend has removed field %s: %s", field, data)
		}
		if _, ok := response[field]; ok {
			t.Fatalf("response has removed top-level field %s: %s", field, data)
		}
	}
}

func newUnitTestService() *Service {
	return &Service{
		repo:        newFakeUnitRepo(),
		rbacService: &fakeUnitRbacService{},
	}
}

type fakeUnitRepo struct {
	units map[int]*domainunit.Unit
	next  int
}

func newFakeUnitRepo() *fakeUnitRepo {
	return &fakeUnitRepo{
		units: make(map[int]*domainunit.Unit),
		next:  1,
	}
}

func (r *fakeUnitRepo) seed(units ...*domainunit.Unit) {
	for _, unit := range units {
		next := cloneUnit(unit)
		r.units[next.ID] = next
		if next.ID >= r.next {
			r.next = next.ID + 1
		}
	}
}

func (r *fakeUnitRepo) Create(
	ctx context.Context,
	unit *domainunit.Unit,
) (*domainunit.Unit, error) {
	next := cloneUnit(unit)
	next.ID = r.next
	r.next++
	r.units[next.ID] = next
	return cloneUnit(next), nil
}

func (r *fakeUnitRepo) Get(
	ctx context.Context,
	accountID string,
	id int,
) (*domainunit.Unit, error) {
	unit := r.units[id]
	if unit == nil || unit.AccountID != accountID {
		return nil, domainunit.ErrNotFound
	}
	return cloneUnit(unit), nil
}

func (r *fakeUnitRepo) GetByID(ctx context.Context, id int) (*domainunit.Unit, error) {
	unit := r.units[id]
	if unit == nil {
		return nil, domainunit.ErrNotFound
	}
	return cloneUnit(unit), nil
}

func (r *fakeUnitRepo) List(
	ctx context.Context,
	query domainunit.ListQuery,
) ([]*domainunit.Unit, error) {
	units := make([]*domainunit.Unit, 0, len(r.units))
	for _, unit := range r.units {
		if unit.AccountID == query.AccountID {
			units = append(units, cloneUnit(unit))
		}
	}
	return units, nil
}

func (r *fakeUnitRepo) Update(
	ctx context.Context,
	unit *domainunit.Unit,
) (*domainunit.Unit, error) {
	if r.units[unit.ID] == nil {
		return nil, domainunit.ErrNotFound
	}
	r.units[unit.ID] = cloneUnit(unit)
	return cloneUnit(unit), nil
}

func (r *fakeUnitRepo) Delete(ctx context.Context, accountID string, id int) error {
	delete(r.units, id)
	return nil
}

func cloneUnit(unit *domainunit.Unit) *domainunit.Unit {
	if unit == nil {
		return nil
	}
	next := *unit
	next.DefaultMeterItemIDs = append([]int(nil), unit.DefaultMeterItemIDs...)
	return &next
}

type fakeUnitRbacService struct{}

func (s *fakeUnitRbacService) RequireUnitPermission(
	ctx context.Context,
	accountID string,
	unitID int,
	permissionCode string,
) error {
	return nil
}

var (
	_ domainunit.Repository   = (*fakeUnitRepo)(nil)
	_ domainrbac.IRbacService = (*fakeUnitRbacService)(nil)
)

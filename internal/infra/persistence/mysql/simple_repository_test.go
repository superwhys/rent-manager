package mysql

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domainaccountsetting "github.com/superwhys/rent-manager/internal/domain/accountsetting"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainroomtype "github.com/superwhys/rent-manager/internal/domain/roomtype"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestAccountSettingRepositoryRestoresDeletedSetting(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, other, _ := seedRBACUnit(t, db)
	repo := NewAccountSettingRepository(db)
	setting, err := repo.Upsert(ctx, &domainaccountsetting.AccountSetting{AccountID: owner, Key: "billing-day", Value: "5"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByKey(ctx, other, setting.Key); !errors.Is(err, domainaccountsetting.ErrNotFound) {
		t.Fatalf("跨账号读取配置错误 = %v", err)
	}
	deletedAt := time.Now().Add(-time.Hour)
	if err := db.Model(&models.AccountSetting{}).Where("id = ?", setting.ID).Updates(map[string]interface{}{"deleted_at": deletedAt, "updated_at": deletedAt}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByKey(ctx, owner, setting.Key); !errors.Is(err, domainaccountsetting.ErrNotFound) {
		t.Fatalf("已删除配置读取错误 = %v", err)
	}
	restored, err := repo.Upsert(ctx, &domainaccountsetting.AccountSetting{AccountID: owner, Key: setting.Key, Value: "10"})
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID != setting.ID || restored.Value != "10" || restored.DeletedAt != nil || !restored.UpdatedAt.After(deletedAt) {
		t.Fatalf("配置未原地恢复: %+v", restored)
	}
	listed, err := repo.List(ctx, domainaccountsetting.ListQuery{AccountID: owner})
	if err != nil || len(listed) != 1 || listed[0].ID != setting.ID {
		t.Fatalf("恢复后的配置列表 = %+v, %v", listed, err)
	}
}

func TestAccountSettingRepositoryPreservesAccountIdentity(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := NewAccountSettingRepository(db)
	owner := testAccountID("owner")
	setting, err := repo.Upsert(ctx, &domainaccountsetting.AccountSetting{AccountID: owner, Key: "billing-day", Value: "5"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByKey(ctx, strings.ToUpper(owner), setting.Key); !errors.Is(err, domainaccountsetting.ErrNotFound) {
		t.Fatalf("账号 ID 大小写不同不应读取到配置: %v", err)
	}
	stored, err := repo.GetByKey(ctx, owner, setting.Key)
	if err != nil || stored.AccountID != owner {
		t.Fatalf("完整账号 ID 未保留: %+v, %v", stored, err)
	}
}

func TestRoomTypeRepository(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, other, unit := seedRBACUnit(t, db)
	repo := NewRoomTypeRepository(db)
	roomType, err := repo.Create(ctx, &domainroomtype.RoomType{AccountID: owner, Name: "一室一厅"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, &domainroomtype.RoomType{AccountID: owner, Name: roomType.Name}); !errors.Is(err, domainroomtype.ErrNameAlreadyExists) {
		t.Fatalf("重复房型错误 = %v", err)
	}
	if _, err := repo.Get(ctx, other, roomType.ID); !errors.Is(err, domainroomtype.ErrNotFound) {
		t.Fatalf("跨账号读取房型错误 = %v", err)
	}
	roomType.Name = "两室一厅"
	updated, err := repo.Update(ctx, roomType)
	if err != nil || updated.Name != roomType.Name {
		t.Fatalf("更新房型 = %+v, %v", updated, err)
	}
	room := &models.Room{UnitID: unit.ID, RoomTypeID: roomType.ID, RoomNo: "101"}
	if err := db.Create(room).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, owner, roomType.ID); !errors.Is(err, domainroomtype.ErrRoomTypeHasRooms) {
		t.Fatalf("已绑定房型删除错误 = %v", err)
	}
	if err := db.Model(room).Update("deleted_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, owner, roomType.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, owner, roomType.ID); !errors.Is(err, domainroomtype.ErrNotFound) {
		t.Fatalf("已删除房型读取错误 = %v", err)
	}
}

func TestRepositoryContainsLiteralCharacters(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, _, _ := seedRBACUnit(t, db)
	roomTypes := NewRoomTypeRepository(db)
	meterItems := NewMeterItemRepository(db)
	for _, name := range []string{"前%后", "前_后", `前\后`, "前X后"} {
		if _, err := roomTypes.Create(ctx, &domainroomtype.RoomType{AccountID: owner, Name: name}); err != nil {
			t.Fatal(err)
		}
		if _, err := meterItems.Create(ctx, &domainmeteritem.MeterItem{AccountID: owner, Code: name, Unit: "度", CalcType: domainmeteritem.CalcTypeDifference}); err != nil {
			t.Fatal(err)
		}
	}
	for _, pattern := range []string{"%", "_", `\`} {
		t.Run(pattern, func(t *testing.T) {
			types, err := roomTypes.List(ctx, domainroomtype.ListQuery{AccountID: owner, Name: pattern})
			if err != nil || len(types) != 1 || types[0].Name != "前"+pattern+"后" {
				t.Fatalf("房型字面匹配 %q = %+v, %v", pattern, types, err)
			}
			items, err := meterItems.List(ctx, domainmeteritem.ListQuery{AccountID: owner, Code: pattern})
			if err != nil || len(items) != 1 || items[0].Code != "前"+pattern+"后" {
				t.Fatalf("抄表项字面匹配 %q = %+v, %v", pattern, items, err)
			}
		})
	}
}

func TestMeterItemRepositoryPreservesNullableRemark(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, other, _ := seedRBACUnit(t, db)
	repo := NewMeterItemRepository(db)
	remark := "保留这条备注"
	item, err := repo.Create(ctx, &domainmeteritem.MeterItem{
		AccountID: owner, Code: "electricity", Unit: "度", UnitPrice: 1.5, IsSystem: true, CalcType: domainmeteritem.CalcTypeDifference, Remark: &remark,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, item); !errors.Is(err, domainmeteritem.ErrCodeAlreadyExists) {
		t.Fatalf("重复抄表项错误 = %v", err)
	}
	if _, err := repo.Get(ctx, other, item.ID); !errors.Is(err, domainmeteritem.ErrNotFound) {
		t.Fatalf("跨账号读取抄表项错误 = %v", err)
	}
	item.Remark = nil
	item.UnitPrice = 0
	item.IsSystem = false
	updated, err := repo.Update(ctx, item)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Remark == nil || *updated.Remark != remark || updated.UnitPrice != 0 || !updated.IsSystem {
		t.Fatalf("抄表项 nil备注/零值/系统字段行为不正确: %+v", updated)
	}
	empty := ""
	item.Remark = &empty
	updated, err = repo.Update(ctx, item)
	if err != nil || updated.Remark == nil || *updated.Remark != "" {
		t.Fatalf("空备注写入 = %+v, %v", updated, err)
	}
}

func TestRepositoryFactoryTransaction(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	factory := NewRepositoryFactory(db)
	repos := factory.Repositories()
	rollbackErr := errors.New("回滚测试")
	accountID := testAccountID("owner")
	var unitID int
	create := func(tx Repositories) error {
		unit, err := tx.Unit.Create(ctx, &domainunit.Unit{
			AccountID: accountID, Name: "事务单元", Address: "地址", City: "深圳", Province: "广东", Config: domainunit.DefaultUnitConfig(),
		})
		if err != nil {
			return err
		}
		unitID = unit.ID
		_, err = tx.AccountSetting.Upsert(ctx, &domainaccountsetting.AccountSetting{AccountID: accountID, Key: "billing-day", Value: "10"})
		return err
	}
	if err := factory.WithTransaction(ctx, func(tx Repositories) error {
		if err := create(tx); err != nil {
			return err
		}
		return rollbackErr
	}); !errors.Is(err, rollbackErr) {
		t.Fatalf("事务回滚错误 = %v", err)
	}
	if _, err := repos.Unit.Get(ctx, accountID, unitID); !errors.Is(err, domainunit.ErrNotFound) {
		t.Fatalf("单元未随事务回滚: %v", err)
	}
	if _, err := repos.AccountSetting.GetByKey(ctx, accountID, "billing-day"); !errors.Is(err, domainaccountsetting.ErrNotFound) {
		t.Fatalf("配置未随事务回滚: %v", err)
	}
	if err := factory.WithTransaction(ctx, create); err != nil {
		t.Fatal(err)
	}
	if _, err := repos.Unit.Get(ctx, accountID, unitID); err != nil {
		t.Fatalf("事务提交后单元读取失败: %v", err)
	}
	setting, err := repos.AccountSetting.GetByKey(ctx, accountID, "billing-day")
	if err != nil || setting.Value != "10" {
		t.Fatalf("事务提交后配置 = %+v, %v", setting, err)
	}
}

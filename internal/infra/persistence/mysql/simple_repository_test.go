package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	domainaccount "github.com/superwhys/rent-manager/internal/domain/account"
	domainaccountsetting "github.com/superwhys/rent-manager/internal/domain/accountsetting"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainroomtype "github.com/superwhys/rent-manager/internal/domain/roomtype"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestAccountRepository(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := NewAccountRepository(db)
	account, err := repo.Create(ctx, &domainaccount.Account{
		Name: "测试账号", Email: "test@example.com", Password: "password", WechatOpenID: "test-openid", IsActive: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.IsActive || account.CreatedAt.IsZero() || account.UpdatedAt.IsZero() {
		t.Fatalf("账号默认值或时间不正确: %+v", account)
	}
	if _, err := repo.Create(ctx, account); !isConstraintError(err) {
		t.Fatalf("重复微信标识错误 = %v", err)
	}
	got, err := repo.GetByWechatOpenID(ctx, account.WechatOpenID)
	if err != nil || got.ID != account.ID || got.IsActive {
		t.Fatalf("按微信标识读取账号 = %+v, %v", got, err)
	}
	account.IsActive = true
	account, err = repo.Update(ctx, account)
	if err != nil || !account.IsActive {
		t.Fatalf("激活账号 = %+v, %v", account, err)
	}
	account.IsActive = false
	account, err = repo.Update(ctx, account)
	if err != nil || account.IsActive {
		t.Fatalf("停用账号 = %+v, %v", account, err)
	}
	if _, err := repo.Update(ctx, account); err != nil {
		t.Fatalf("无变化的账号更新错误 = %v", err)
	}
	missing := *account
	missing.ID = 99999
	if _, err := repo.Update(ctx, &missing); !errors.Is(err, domainaccount.ErrAccountNotFound) {
		t.Fatalf("不存在的账号更新错误 = %v", err)
	}
	if err := repo.Delete(ctx, account.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, account.ID); !errors.Is(err, domainaccount.ErrAccountNotFound) {
		t.Fatalf("已删除账号读取错误 = %v", err)
	}
	if _, err := repo.GetByWechatOpenID(ctx, account.WechatOpenID); !errors.Is(err, domainaccount.ErrAccountNotFound) {
		t.Fatalf("已删除账号微信查询错误 = %v", err)
	}
}

func TestAccountSettingRepositoryRestoresDeletedSetting(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, other, _ := seedRBACAccounts(t, db)
	repo := NewAccountSettingRepository(db)
	setting, err := repo.Upsert(ctx, &domainaccountsetting.AccountSetting{AccountID: owner.ID, Key: "billing-day", Value: "5"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByKey(ctx, other.ID, setting.Key); !errors.Is(err, domainaccountsetting.ErrNotFound) {
		t.Fatalf("跨账号读取配置错误 = %v", err)
	}
	deletedAt := time.Now().Add(-time.Hour)
	if err := db.Model(&models.AccountSetting{}).Where("id = ?", setting.ID).Updates(map[string]interface{}{"deleted_at": deletedAt, "updated_at": deletedAt}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByKey(ctx, owner.ID, setting.Key); !errors.Is(err, domainaccountsetting.ErrNotFound) {
		t.Fatalf("已删除配置读取错误 = %v", err)
	}
	restored, err := repo.Upsert(ctx, &domainaccountsetting.AccountSetting{AccountID: owner.ID, Key: setting.Key, Value: "10"})
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID != setting.ID || restored.Value != "10" || restored.DeletedAt != nil || !restored.UpdatedAt.After(deletedAt) {
		t.Fatalf("配置未原地恢复: %+v", restored)
	}
	listed, err := repo.List(ctx, domainaccountsetting.ListQuery{AccountID: owner.ID})
	if err != nil || len(listed) != 1 || listed[0].ID != setting.ID {
		t.Fatalf("恢复后的配置列表 = %+v, %v", listed, err)
	}
}

func TestRoomTypeRepository(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, other, unit := seedRBACAccounts(t, db)
	repo := NewRoomTypeRepository(db)
	roomType, err := repo.Create(ctx, &domainroomtype.RoomType{AccountID: owner.ID, Name: "一室一厅"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, &domainroomtype.RoomType{AccountID: owner.ID, Name: roomType.Name}); !errors.Is(err, domainroomtype.ErrNameAlreadyExists) {
		t.Fatalf("重复房型错误 = %v", err)
	}
	if _, err := repo.Get(ctx, other.ID, roomType.ID); !errors.Is(err, domainroomtype.ErrNotFound) {
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
	if err := repo.Delete(ctx, owner.ID, roomType.ID); !errors.Is(err, domainroomtype.ErrRoomTypeHasRooms) {
		t.Fatalf("已绑定房型删除错误 = %v", err)
	}
	if err := db.Model(room).Update("deleted_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, owner.ID, roomType.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, owner.ID, roomType.ID); !errors.Is(err, domainroomtype.ErrNotFound) {
		t.Fatalf("已删除房型读取错误 = %v", err)
	}
}

func TestRepositoryContainsLiteralCharacters(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, _, _ := seedRBACAccounts(t, db)
	roomTypes := NewRoomTypeRepository(db)
	meterItems := NewMeterItemRepository(db)
	for _, name := range []string{"前%后", "前_后", `前\后`, "前X后"} {
		if _, err := roomTypes.Create(ctx, &domainroomtype.RoomType{AccountID: owner.ID, Name: name}); err != nil {
			t.Fatal(err)
		}
		if _, err := meterItems.Create(ctx, &domainmeteritem.MeterItem{AccountID: owner.ID, Code: name, Unit: "度", CalcType: domainmeteritem.CalcTypeDifference}); err != nil {
			t.Fatal(err)
		}
	}
	for _, pattern := range []string{"%", "_", `\`} {
		t.Run(pattern, func(t *testing.T) {
			types, err := roomTypes.List(ctx, domainroomtype.ListQuery{AccountID: owner.ID, Name: pattern})
			if err != nil || len(types) != 1 || types[0].Name != "前"+pattern+"后" {
				t.Fatalf("房型字面匹配 %q = %+v, %v", pattern, types, err)
			}
			items, err := meterItems.List(ctx, domainmeteritem.ListQuery{AccountID: owner.ID, Code: pattern})
			if err != nil || len(items) != 1 || items[0].Code != "前"+pattern+"后" {
				t.Fatalf("抄表项字面匹配 %q = %+v, %v", pattern, items, err)
			}
		})
	}
}

func TestMeterItemRepositoryPreservesNullableRemark(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, other, _ := seedRBACAccounts(t, db)
	repo := NewMeterItemRepository(db)
	remark := "保留这条备注"
	item, err := repo.Create(ctx, &domainmeteritem.MeterItem{
		AccountID: owner.ID, Code: "electricity", Unit: "度", UnitPrice: 1.5, IsSystem: true, CalcType: domainmeteritem.CalcTypeDifference, Remark: &remark,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, item); !errors.Is(err, domainmeteritem.ErrCodeAlreadyExists) {
		t.Fatalf("重复抄表项错误 = %v", err)
	}
	if _, err := repo.Get(ctx, other.ID, item.ID); !errors.Is(err, domainmeteritem.ErrNotFound) {
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
	var accountID int
	create := func(tx Repositories) error {
		account, err := tx.Account.Create(ctx, &domainaccount.Account{
			Name: "事务账号", Email: "transaction@example.com", Password: "password", WechatOpenID: "transaction-openid", IsActive: true,
		})
		if err != nil {
			return err
		}
		accountID = account.ID
		_, err = tx.AccountSetting.Upsert(ctx, &domainaccountsetting.AccountSetting{AccountID: account.ID, Key: "billing-day", Value: "10"})
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
	if _, err := repos.Account.GetByWechatOpenID(ctx, "transaction-openid"); !errors.Is(err, domainaccount.ErrAccountNotFound) {
		t.Fatalf("账号未随事务回滚: %v", err)
	}
	if _, err := repos.AccountSetting.GetByKey(ctx, accountID, "billing-day"); !errors.Is(err, domainaccountsetting.ErrNotFound) {
		t.Fatalf("配置未随事务回滚: %v", err)
	}
	if err := factory.WithTransaction(ctx, create); err != nil {
		t.Fatal(err)
	}
	if _, err := repos.Account.Get(ctx, accountID); err != nil {
		t.Fatalf("事务提交后账号读取失败: %v", err)
	}
	setting, err := repos.AccountSetting.GetByKey(ctx, accountID, "billing-day")
	if err != nil || setting.Value != "10" {
		t.Fatalf("事务提交后配置 = %+v, %v", setting, err)
	}
}

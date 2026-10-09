package mysql

import (
	"context"
	"strings"
	"testing"

	"github.com/superwhys/rent-manager/internal/domain/account"
	"github.com/superwhys/rent-manager/internal/domain/accountsetting"
	"github.com/superwhys/rent-manager/internal/domain/roomtype"
)

func TestRepositoryByteLengthLimits(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	accounts := NewAccountRepository(db)
	owner, err := accounts.Create(ctx, &account.Account{
		Name: "owner", Email: "owner@example.com", Password: "password", WechatOpenID: "owner", IsActive: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*account.Account){
		func(a *account.Account) { a.Name = strings.Repeat("名", 86) },
		func(a *account.Account) { a.Password = strings.Repeat("密", 86) },
		func(a *account.Account) { a.WechatOpenID = strings.Repeat("微", 86) },
	} {
		invalid := *owner
		change(&invalid)
		if _, err := accounts.Create(ctx, &invalid); err == nil {
			t.Fatal("超长账号字段不应创建成功")
		}
		if _, err := accounts.Update(ctx, &invalid); err == nil {
			t.Fatal("超长账号字段不应更新成功")
		}
	}

	roomTypes := NewRoomTypeRepository(db)
	roomType, err := roomTypes.Create(ctx, &roomtype.RoomType{AccountID: owner.ID, Name: "一室一厅"})
	if err != nil {
		t.Fatal(err)
	}
	roomType.Name = "豪华双人房"
	if _, err := roomTypes.Create(ctx, roomType); err == nil {
		t.Fatal("房型名称超过 12 字节不应创建成功")
	}
	if _, err := roomTypes.Update(ctx, roomType); err == nil {
		t.Fatal("房型名称超过 12 字节不应更新成功")
	}
	saved, err := roomTypes.Get(ctx, owner.ID, roomType.ID)
	if err != nil || saved.Name != "一室一厅" {
		t.Fatalf("校验失败不应修改房型: %+v, %v", saved, err)
	}

	settings := NewAccountSettingRepository(db)
	for _, setting := range []*accountsetting.AccountSetting{
		{AccountID: owner.ID, Key: strings.Repeat("键", 22), Value: "value"},
		{AccountID: owner.ID, Key: "key", Value: strings.Repeat("值", 86)},
	} {
		if _, err := settings.Upsert(ctx, setting); err == nil {
			t.Fatal("超长配置键或值不应保存成功")
		}
	}
	if _, err := settings.Upsert(ctx, &accountsetting.AccountSetting{
		AccountID: owner.ID, Key: strings.Repeat("k", 64), Value: strings.Repeat("值", 85),
	}); err != nil {
		t.Fatalf("最大字节长度应允许保存: %v", err)
	}
}

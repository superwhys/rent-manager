package mysql

import (
	"context"
	"errors"
	"testing"

	appcore "github.com/superwhys/rent-manager/internal/app"
	appmeteritem "github.com/superwhys/rent-manager/internal/app/meteritem"
	approomtype "github.com/superwhys/rent-manager/internal/app/roomtype"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainroomtype "github.com/superwhys/rent-manager/internal/domain/roomtype"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestSharedDictionaryServicesUseUnitOwnerAndRevokeAccess(t *testing.T) {
	db := newTestDB(t)
	owner, collaborator, unit := seedRBACUnit(t, db)
	roomType := &models.RoomType{AccountID: owner, Name: "大床房"}
	privateRoomType := &models.RoomType{AccountID: collaborator, Name: "个人房型"}
	meterItem := &models.MeterItem{AccountID: owner, Code: "水费", Unit: "吨", CalcType: domainmeteritem.CalcTypeDifference}
	privateMeterItem := &models.MeterItem{AccountID: collaborator, Code: "个人水费", Unit: "吨", CalcType: domainmeteritem.CalcTypeDifference}
	role := &models.Role{AccountID: owner, Name: "读取目录"}
	group := &models.PermissionGroup{Code: "dictionary", Name: "目录查看"}
	for _, record := range []any{roomType, privateRoomType, meterItem, privateMeterItem, role, group} {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, record := range []any{
		&models.RolePermissionGroup{RoleID: role.ID, PermissionGroupID: group.ID},
		&models.PermissionGroupPermission{PermissionGroupID: group.ID, PermissionCode: domainrbac.PermissionRoomTypeView},
		&models.PermissionGroupPermission{PermissionGroupID: group.ID, PermissionCode: domainrbac.PermissionMeterItemView},
	} {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	rbacRepo := NewRBACRepository(db)
	unitRepo := NewUnitRepository(db)
	roomTypes := approomtype.NewService(NewRoomTypeRepository(db), unitRepo, rbacRepo)
	meterItems := appmeteritem.NewService(NewMeterItemRepository(db), unitRepo, rbacRepo)
	ctx := appcore.SetAccountIDToContext(context.Background(), collaborator)
	roomQuery := &approomtype.ListRoomTypesQuery{UnitID: unit.ID}
	itemQuery := &appmeteritem.ListMeterItemsQuery{UnitID: unit.ID}
	roomDetail := &approomtype.GetRoomTypeQuery{RoomTypeID: roomType.ID, UnitID: unit.ID}
	itemDetail := &appmeteritem.GetMeterItemQuery{MeterItemID: meterItem.ID, UnitID: unit.ID}
	assertDenied := func() {
		t.Helper()
		_, listRoomErr := roomTypes.ListRoomTypes(ctx, roomQuery)
		_, getRoomErr := roomTypes.GetRoomType(ctx, roomDetail)
		_, listItemErr := meterItems.ListMeterItems(ctx, itemQuery)
		_, getItemErr := meterItems.GetMeterItem(ctx, itemDetail)
		for _, err := range []error{listRoomErr, getRoomErr, listItemErr, getItemErr} {
			if !errors.Is(err, domainrbac.ErrForbidden) {
				t.Fatalf("未授权或撤权后读取错误 = %v", err)
			}
		}
	}
	assertDenied()
	grant := domainrbac.UnitAccountRolesParams{
		OwnerAccountID: owner, TargetAccountID: collaborator, UnitID: unit.ID, RoleIDs: []int{role.ID},
	}
	if _, err := rbacRepo.GrantUnitRoles(ctx, grant); err != nil {
		t.Fatal(err)
	}
	// 字典尚未绑定任何房间或楼栋，授权后仍可读取，供当前楼栋选用。
	rooms, err := roomTypes.ListRoomTypes(ctx, roomQuery)
	if err != nil || len(rooms) != 1 || rooms[0].ID != roomType.ID {
		t.Fatalf("共享房型目录 = %+v, %v", rooms, err)
	}
	items, err := meterItems.ListMeterItems(ctx, itemQuery)
	if err != nil || len(items) != 1 || items[0].ID != meterItem.ID {
		t.Fatalf("共享收费目录 = %+v, %v", items, err)
	}
	if got, err := roomTypes.GetRoomType(ctx, roomDetail); err != nil || got.ID != roomType.ID {
		t.Fatalf("共享房型详情 = %+v, %v", got, err)
	}
	if got, err := meterItems.GetMeterItem(ctx, itemDetail); err != nil || got.ID != meterItem.ID {
		t.Fatalf("共享收费详情 = %+v, %v", got, err)
	}
	if _, err := roomTypes.GetRoomType(ctx, &approomtype.GetRoomTypeQuery{
		RoomTypeID: privateRoomType.ID, UnitID: unit.ID,
	}); !errors.Is(err, domainroomtype.ErrNotFound) {
		t.Fatalf("楼栋范围接受了其他账号房型：%v", err)
	}
	if _, err := meterItems.GetMeterItem(ctx, &appmeteritem.GetMeterItemQuery{
		MeterItemID: privateMeterItem.ID, UnitID: unit.ID,
	}); !errors.Is(err, domainmeteritem.ErrNotFound) {
		t.Fatalf("楼栋范围接受了其他账号收费项：%v", err)
	}
	rooms, err = roomTypes.ListRoomTypes(ctx, &approomtype.ListRoomTypesQuery{})
	if err != nil || len(rooms) != 1 || rooms[0].ID != privateRoomType.ID {
		t.Fatalf("默认房型目录未保持当前账号范围：%+v, %v", rooms, err)
	}
	items, err = meterItems.ListMeterItems(ctx, &appmeteritem.ListMeterItemsQuery{})
	if err != nil || len(items) != 1 || items[0].ID != privateMeterItem.ID {
		t.Fatalf("默认收费目录未保持当前账号范围：%+v, %v", items, err)
	}
	if _, err := roomTypes.UpdateRoomType(ctx, &approomtype.UpdateRoomTypeCommand{
		RoomTypeID: roomType.ID, Name: "改名",
	}); !errors.Is(err, domainroomtype.ErrNotFound) {
		t.Fatalf("协作者修改了共享房型定义：%v", err)
	}
	if err := roomTypes.DeleteRoomType(ctx, roomType.ID); !errors.Is(err, domainroomtype.ErrNotFound) {
		t.Fatalf("协作者删除了共享房型定义：%v", err)
	}
	if _, err := meterItems.UpdateMeterItem(ctx, &appmeteritem.UpdateMeterItemCommand{
		MeterItemID: meterItem.ID, Code: "水费", Unit: "吨", UnitPrice: 9,
	}); !errors.Is(err, domainmeteritem.ErrNotFound) {
		t.Fatalf("协作者修改了共享收费定义：%v", err)
	}
	if err := rbacRepo.RevokeUnitRoles(ctx, grant); err != nil {
		t.Fatal(err)
	}
	assertDenied()
}

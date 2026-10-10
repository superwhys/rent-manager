package mysql

import (
	"context"
	"strings"
	"testing"

	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestRBACRepositoryBulkByteLengthValidation(t *testing.T) {
	for _, field := range []string{"角色名称", "权限组编码", "权限组名称", "权限组描述"} {
		t.Run(field, func(t *testing.T) {
			ctx := context.Background()
			db := newTestDB(t)
			owner, _, _ := seedRBACUnit(t, db)
			repo := NewRBACRepository(db)
			var model interface{}
			create := func(value string) error {
				switch field {
				case "角色名称":
					model = &models.Role{}
					_, err := repo.CreateRoles(ctx, []*domainrbac.Role{
						{AccountID: owner, Name: "有效角色", IsActive: true},
						{AccountID: owner, Name: value, IsActive: true},
					})
					return err
				case "权限组编码", "权限组名称", "权限组描述":
					model = &models.PermissionGroup{}
					group := &domainrbac.PermissionGroup{
						Code:     "second",
						Name:     "第二组",
						IsActive: true,
					}
					switch field {
					case "权限组编码":
						group.Code = value
					case "权限组名称":
						group.Name = value
					case "权限组描述":
						group.Description = &value
					}
					_, err := repo.CreatePermissionGroups(
						ctx,
						[]*domainrbac.PermissionGroup{
							{Code: "first", Name: "有效权限组", IsActive: true},
							group,
						},
					)
					return err
				}
				return nil
			}
			if err := create(strings.Repeat("中", 86)); err == nil {
				t.Fatal("超过 255 字节的中文字符串未被拒绝")
			}
			var count int64
			if err := db.Model(model).Count(&count).Error; err != nil || count != 0 {
				t.Fatalf("校验失败后发生部分写入: count=%d, err=%v", count, err)
			}
			if err := create(strings.Repeat("中", 85)); err != nil {
				t.Fatalf("255 字节边界值被拒绝: %v", err)
			}
			if err := db.Model(model).Count(&count).Error; err != nil || count != 2 {
				t.Fatalf("有效批量写入数量 = %d, %v", count, err)
			}
		})
	}
}

func TestRBACRepositoryUpdateRoleByteLengthValidation(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, _, _ := seedRBACUnit(t, db)
	repo := NewRBACRepository(db)
	roles, err := repo.CreateRoles(
		ctx,
		[]*domainrbac.Role{{AccountID: owner, Name: "原始名称", IsActive: true}},
	)
	if err != nil {
		t.Fatal(err)
	}
	role := roles[0]
	role.Name = strings.Repeat("中", 86)
	if _, err := repo.UpdateRole(ctx, role); err == nil {
		t.Fatal("角色更新未拒绝超过 255 字节的中文名称")
	}
	unchanged, err := repo.GetRole(ctx, owner, role.ID)
	if err != nil || unchanged.Name != "原始名称" {
		t.Fatalf("失败更新修改了角色: %+v, %v", unchanged, err)
	}
	role.Name = strings.Repeat("中", 85)
	updated, err := repo.UpdateRole(ctx, role)
	if err != nil || updated.Name != role.Name {
		t.Fatalf("255 字节名称更新 = %+v, %v", updated, err)
	}
}

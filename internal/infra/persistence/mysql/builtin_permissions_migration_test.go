package mysql

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func legacyPermissionMigrationDB(t *testing.T, withForeignKeys bool) *gorm.DB {
	t.Helper()
	if os.Getenv("RENT_MANAGER_TEST_MYSQL_DSN") == "" {
		t.Skip("需要 MySQL 隔离测试库验证存储过程迁移")
	}
	db := newTestDB(t)
	for _, statement := range []string{
		"DROP TABLE permission_group_permissions",
		`CREATE TABLE permissions (
			id BIGINT NOT NULL PRIMARY KEY,
			code VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
			is_active BOOLEAN NOT NULL, deleted_at TIMESTAMP NULL
		)`,
		`CREATE TABLE permission_group_permissions (
			permission_group_id BIGINT NOT NULL, permission_id BIGINT NOT NULL,
			PRIMARY KEY (permission_group_id, permission_id)
		)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if withForeignKeys {
		if err := db.Exec(`ALTER TABLE permission_group_permissions
			ADD CONSTRAINT custom_legacy_group_fk FOREIGN KEY (permission_group_id) REFERENCES permission_groups(id) ON DELETE CASCADE,
			ADD CONSTRAINT custom_legacy_permission_fk FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE`).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func runBuiltinPermissionMigration(t *testing.T, db *gorm.DB) error {
	t.Helper()
	script, err := os.ReadFile("../../../../migrations/20261010_builtin_permissions.sql")
	if err != nil {
		t.Fatal(err)
	}
	// The CLI delimiter separates the stored procedure and its invocation.
	parts := strings.Split(string(script), "DELIMITER $$")
	if len(parts) != 2 {
		t.Fatal("迁移 SQL 缺少存储过程分隔符")
	}
	body := strings.Split(parts[1], "DELIMITER ;")[0]
	for _, statement := range strings.Split(body, "$$") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func TestBuiltinPermissionMigrationMySQL(t *testing.T) {
	db := legacyPermissionMigrationDB(t, true)
	owner, collaborator, unit := seedRBACUnit(t, db)
	group := &models.PermissionGroup{ID: 37, Code: "migration.group", Name: "迁移权限组"}
	otherGroup := &models.PermissionGroup{ID: 82, Code: "migration.other", Name: "其他权限组"}
	role := &models.Role{AccountID: owner, Name: "迁移角色"}
	for _, record := range []any{group, otherGroup, role} {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, record := range []any{
		&models.RolePermissionGroup{RoleID: role.ID, PermissionGroupID: group.ID},
		&models.UnitAccountRole{UnitID: unit.ID, AccountID: collaborator, RoleID: role.ID, GrantedByAccountID: owner},
	} {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec(`INSERT INTO permissions (id, code, is_active, deleted_at) VALUES
		(501, 'room.view', 1, NULL), (907, 'meter_reading.manage', 0, NULL),
		(203, 'rent_bill.view', 1, ?), (611, 'unit.view', 1, NULL)`, time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO permission_group_permissions (permission_group_id, permission_id) VALUES
		(37, 501), (37, 907), (37, 203), (82, 611)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := runBuiltinPermissionMigration(t, db); err != nil {
		t.Fatal(err)
	}
	if err := EnsureBuiltinPermissionSchema(db); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		t.Fatal(err)
	}
	if db.Migrator().HasTable("permissions") || db.Migrator().HasColumn("permission_group_permissions", "permission_id") {
		t.Fatal("迁移未移除旧权限目录及 ID 列")
	}
	var bindings []models.PermissionGroupPermission
	if err := db.Order("permission_group_id").Find(&bindings).Error; err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 2 || bindings[0].PermissionGroupID != 37 || bindings[0].PermissionCode != "room.view" ||
		bindings[1].PermissionGroupID != 82 || bindings[1].PermissionCode != "unit.view" {
		t.Fatalf("实际旧 ID 迁移结果 = %+v", bindings)
	}
	var collation string
	if err := db.Raw(`SELECT COLLATION_NAME FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'permission_group_permissions'
		AND COLUMN_NAME = 'permission_code'`).Scan(&collation).Error; err != nil || collation != "utf8mb4_bin" {
		t.Fatalf("权限编码排序规则 = %q, %v", collation, err)
	}
	var groupForeignKeys int64
	if err := db.Raw(`SELECT COUNT(*) FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'permission_group_permissions'
		AND CONSTRAINT_NAME = 'custom_legacy_group_fk'`).Scan(&groupForeignKeys).Error; err != nil || groupForeignKeys != 1 {
		t.Fatalf("原权限组外键未保留 = %d, %v", groupForeignKeys, err)
	}
	repo := NewRepositoryFactory(db).Repositories().RBAC
	for _, check := range []struct {
		code    string
		allowed bool
	}{
		{domainrbac.PermissionRoomView, true},
		{domainrbac.PermissionMeterReadingManage, false},
		{domainrbac.PermissionRentBillView, false},
	} {
		allowed, err := repo.HasUnitPermission(context.Background(), collaborator, unit.ID, check.code)
		if err != nil || allowed != check.allowed {
			t.Fatalf("迁移后权限 %s = %v, %v; want %v", check.code, allowed, err, check.allowed)
		}
	}
}

func TestBuiltinPermissionMigrationMySQLRejectsInvalidBindings(t *testing.T) {
	for _, scenario := range []struct {
		name       string
		code       string
		groupID    int
		permission bool
		externalFK bool
		message    string
	}{
		{"未知权限", "custom.export", 37, true, false, "未知权限编码"},
		{"大小写不同", "ROOM.VIEW", 37, true, false, "未知权限编码"},
		{"尾随空格", "unit.view ", 37, true, false, "未知权限编码"},
		{"孤立权限", "", 37, false, false, "孤立权限绑定"},
		{"孤立权限组", "room.view", 999, true, false, "孤立权限绑定"},
		{"其他表引用权限", "room.view", 37, true, true, "其他外键引用"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			db := legacyPermissionMigrationDB(t, false)
			group := &models.PermissionGroup{ID: 37, Code: "migration.group", Name: "迁移权限组"}
			if err := db.Create(group).Error; err != nil {
				t.Fatal(err)
			}
			if scenario.permission {
				if err := db.Exec("INSERT INTO permissions (id, code, is_active) VALUES (501, ?, 1)", scenario.code).Error; err != nil {
					t.Fatal(err)
				}
			}
			if err := db.Exec("INSERT INTO permission_group_permissions (permission_group_id, permission_id) VALUES (?, 501)", scenario.groupID).Error; err != nil {
				t.Fatal(err)
			}
			if scenario.externalFK {
				if err := db.Exec(`CREATE TABLE custom_permission_refs (
					permission_id BIGINT, FOREIGN KEY (permission_id) REFERENCES permissions(id)
				)`).Error; err != nil {
					t.Fatal(err)
				}
			}
			err := runBuiltinPermissionMigration(t, db)
			if err == nil || !strings.Contains(err.Error(), scenario.message) {
				t.Fatalf("迁移预检错误 = %v; want %s", err, scenario.message)
			}
			if !db.Migrator().HasTable("permissions") || !db.Migrator().HasColumn("permission_group_permissions", "permission_id") ||
				db.Migrator().HasColumn("permission_group_permissions", "permission_code") {
				t.Fatal("预检失败修改了旧权限表结构")
			}
			var bindingCount int64
			if err := db.Table("permission_group_permissions").Count(&bindingCount).Error; err != nil || bindingCount != 1 {
				t.Fatalf("预检失败修改了旧权限绑定 = %d, %v", bindingCount, err)
			}
		})
	}
}

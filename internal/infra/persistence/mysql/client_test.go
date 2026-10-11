package mysql

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	mysqlconfig "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"

	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

// sqliteTestDialector preserves the production models while adapting MySQL-only types.
type sqliteTestDialector struct {
	sqlite.Dialector
}

func (d sqliteTestDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return sqlite.Migrator{Migrator: migrator.Migrator{Config: migrator.Config{
		DB:                          db,
		Dialector:                   d,
		CreateIndexAfterCreateTable: true,
	}}}
}

func (d sqliteTestDialector) DataTypeOf(field *schema.Field) string {
	if strings.HasPrefix(string(field.DataType), "enum(") {
		return "text"
	}
	if strings.Contains(string(field.DataType), "CHARACTER SET") {
		return "text COLLATE BINARY"
	}
	if field.DataType == "bigint" {
		return "integer"
	}
	return d.Dialector.DataTypeOf(field)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	var dialector gorm.Dialector = sqliteTestDialector{sqlite.Dialector{DSN: fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", t.Name())}}
	if dsn := os.Getenv("RENT_MANAGER_TEST_MYSQL_DSN"); dsn != "" {
		dialector = newMysqlTestDialector(t, dsn)
	}
	db, err := gorm.Open(
		dialector,
		&gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		t.Fatal(err)
	}
	return db
}

// Each MySQL test owns a temporary schema; existing schemas are never modified.
func newMysqlTestDialector(t *testing.T, dsn string) gorm.Dialector {
	t.Helper()
	conf, err := mysqlconfig.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	conf.DBName = ""
	admin, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	name := "rent_manager_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec("CREATE DATABASE `" + name + "`"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE `" + name + "`"); err != nil {
			t.Error(err)
		}
	})
	conf.DBName = name
	conf.ParseTime = true
	return mysqldriver.Open(conf.FormatDSN())
}

func testAccountID(name string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
}

func TestBusinessSchemaDoesNotCreateAccounts(t *testing.T) {
	db := newTestDB(t)
	if db.Migrator().HasTable("accounts") {
		t.Fatal("业务模型不应创建本地 accounts 表")
	}
}

func TestBusinessSchemaWithLegacyTimestampDefaults(t *testing.T) {
	dsn := os.Getenv("RENT_MANAGER_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("需要真实 MySQL 验证 TIMESTAMP 默认值")
	}
	for _, explicitDefaults := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit_defaults=%t", explicitDefaults), func(t *testing.T) {
			db, err := gorm.Open(newMysqlTestDialector(t, dsn), &gorm.Config{
				PrepareStmt: true,
				Logger:      logger.Default.LogMode(logger.Silent),
			})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			sqlDB.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = sqlDB.Close() })
			if err := db.Exec("SET SESSION sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE'").Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec("SET SESSION explicit_defaults_for_timestamp = ?", explicitDefaults).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.AutoMigrate(models.AllModels()...); err != nil {
				t.Fatal(err)
			}
			createdAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
			db.NowFunc = func() time.Time { return createdAt }
			setting := &models.AccountSetting{Key: "test", Value: "value", AccountID: testAccountID("owner")}
			if err := db.Create(setting).Error; err != nil {
				t.Fatal(err)
			}
			var saved models.AccountSetting
			if err := db.First(&saved, setting.ID).Error; err != nil {
				t.Fatal(err)
			}
			if !saved.CreatedAt.Equal(createdAt) || !saved.UpdatedAt.Equal(createdAt) || saved.DeletedAt != nil {
				t.Fatal("创建时应填充时间并保留 deleted_at 为 NULL")
			}
			updatedAt := createdAt.Add(time.Hour)
			db.NowFunc = func() time.Time { return updatedAt }
			if err := db.Model(setting).Update("value", "updated").Error; err != nil {
				t.Fatal(err)
			}
			if err := db.First(&saved, setting.ID).Error; err != nil {
				t.Fatal(err)
			}
			if !saved.CreatedAt.Equal(createdAt) || !saved.UpdatedAt.Equal(updatedAt) {
				t.Fatal("更新时应保留创建时间并刷新更新时间")
			}
		})
	}
}

func TestEnsureBuiltinPermissionSchema(t *testing.T) {
	for _, scenario := range []struct {
		name         string
		legacySQL    string
		dropBindings bool
		wantCode     bool
	}{
		{name: "新权限结构", wantCode: true},
		{name: "旧权限表", legacySQL: "CREATE TABLE permissions (id BIGINT NOT NULL PRIMARY KEY)", wantCode: true},
		{name: "旧权限绑定", legacySQL: "CREATE TABLE permission_group_permissions (permission_group_id BIGINT, permission_id BIGINT)", dropBindings: true},
		{name: "中断迁移双列", legacySQL: "ALTER TABLE permission_group_permissions ADD permission_id BIGINT", wantCode: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			db := newTestDB(t)
			if scenario.dropBindings {
				if err := db.Migrator().DropTable("permission_group_permissions"); err != nil {
					t.Fatal(err)
				}
			}
			if scenario.legacySQL != "" {
				if err := db.Exec(scenario.legacySQL).Error; err != nil {
					t.Fatal(err)
				}
			}
			err := EnsureBuiltinPermissionSchema(db)
			if scenario.legacySQL == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.Is(err, ErrLegacyPermissionSchema) {
				t.Fatalf("旧权限结构未拒绝启动: %v", err)
			}
			if db.Migrator().HasColumn("permission_group_permissions", "permission_code") != scenario.wantCode {
				t.Fatal("启动预检不应修改已有权限绑定")
			}
		})
	}
}

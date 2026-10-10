package mysql

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

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

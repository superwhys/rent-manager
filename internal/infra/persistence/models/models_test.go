package models

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "root@tcp(127.0.0.1:3306)/rent_manager?parseTime=true",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func TestCreatePreservesExplicitInactiveAndDefaultsActive(t *testing.T) {
	db := newDryRunDB(t)
	for _, model := range []any{&Room{}, &Role{}, &Permission{}, &PermissionGroup{}} {
		t.Run(reflect.TypeOf(model).Elem().Name(), func(t *testing.T) {
			active := reflect.ValueOf(model).Elem().FieldByName("IsActive")
			inactive := false
			active.Set(reflect.ValueOf(&inactive))
			if err := db.Create(model).Error; err != nil {
				t.Fatal(err)
			}
			if active.Elem().Bool() {
				t.Fatal("显式 false 被默认值覆盖")
			}

			active.Set(reflect.Zero(active.Type()))
			if err := db.Create(model).Error; err != nil {
				t.Fatal(err)
			}
			if active.IsNil() || !active.Elem().Bool() {
				t.Fatal("未指定激活状态时应使用默认值 true")
			}
		})
	}
}

func TestTimestampsPreserveCreationAndRefreshUpdate(t *testing.T) {
	db := newDryRunDB(t)
	createdAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	db.NowFunc = func() time.Time { return createdAt }
	role := &Role{ID: 1}
	if err := db.Create(role).Error; err != nil {
		t.Fatal(err)
	}
	if !role.CreatedAt.Equal(createdAt) || !role.UpdatedAt.Equal(createdAt) {
		t.Fatal("创建时应自动设置两个时间字段")
	}

	db.NowFunc = func() time.Time { return updatedAt }
	result := db.Model(role).Updates(map[string]any{"name": "updated", "created_at": updatedAt})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if strings.Contains(result.Statement.SQL.String(), "`created_at`") ||
		!role.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("更新时间语义变化: %s", result.Statement.SQL.String())
	}
}

func TestNullableValuesAndExplicitSoftDeletion(t *testing.T) {
	db := newDryRunDB(t)
	reading := &MeterReading{}
	if err := db.Create(reading).Error; err != nil {
		t.Fatal(err)
	}
	if reading.PreviousReading != nil || reading.UsageAmount != nil || reading.ReaderID != nil ||
		reading.Remark != nil ||
		reading.DeletedAt != nil {
		t.Fatal("可空字段不应被默认值替换")
	}

	result := db.Where("id = ?", 1).Find(&Role{})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if strings.Contains(result.Statement.SQL.String(), "deleted_at") {
		t.Fatalf("软删除条件应由仓储显式维护: %s", result.Statement.SQL.String())
	}
}

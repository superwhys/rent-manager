package mysql

import (
	"errors"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/miebyte/goutils/mysqlutils"
	"gorm.io/gorm"
)

var ErrLegacyPermissionSchema = errors.New("检测到旧权限表或权限 ID 绑定，请先备份并执行 migrations/20261010_builtin_permissions.sql")

// OpenMysqlClient initializes a GORM client from MySQL config.
func OpenMysqlClient(conf *mysqlutils.MysqlConfig) (*gorm.DB, error) {
	return conf.DialMysqlGormWithConfig(&gorm.Config{
		PrepareStmt:    true,
		TranslateError: true,
	})
}

// EnsureBuiltinPermissionSchema prevents AutoMigrate from changing legacy grants.
func EnsureBuiltinPermissionSchema(db *gorm.DB) error {
	if db.Migrator().HasTable("permissions") ||
		(db.Migrator().HasTable("permission_group_permissions") && db.Migrator().HasColumn("permission_group_permissions", "permission_id")) {
		return ErrLegacyPermissionSchema
	}
	return nil
}

func isConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrForeignKeyViolated) {
		return true
	}
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) &&
		(mysqlErr.Number == 1062 || mysqlErr.Number == 1451 || mysqlErr.Number == 1452)
}

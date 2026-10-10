package mysql

import (
	"errors"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/miebyte/goutils/mysqlutils"
	"gorm.io/gorm"
)

// OpenMysqlClient initializes a GORM client from MySQL config.
func OpenMysqlClient(conf *mysqlutils.MysqlConfig) (*gorm.DB, error) {
	return conf.DialMysqlGormWithConfig(&gorm.Config{
		PrepareStmt:    true,
		TranslateError: true,
	})
}

func isConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrForeignKeyViolated) {
		return true
	}
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) &&
		(mysqlErr.Number == 1062 || mysqlErr.Number == 1451 || mysqlErr.Number == 1452)
}

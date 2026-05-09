package entutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/superwhys/rent-manager-server/internal/ent"
)

// OpenMysqlClient 根据 MySQL 配置初始化 ent 客户端。
func OpenMysqlClient(conf *mysqlutils.MysqlConfig, opts ...ent.Option) (*ent.Client, error) {
	db, err := sql.Open(dialect.MySQL, mysqlDSN(conf))
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(conf.PoolSize)
	db.SetMaxOpenConns(conf.PoolSize)
	db.SetConnMaxLifetime(time.Hour)

	drv := entsql.OpenDB(dialect.MySQL, db)
	return ent.NewClient(append(opts, ent.Driver(drv))...), nil
}

func MigrateSchema(client *ent.Client) error {
	return client.Schema.Create(context.Background())
}

func mysqlDSN(conf *mysqlutils.MysqlConfig) string {
	charset := conf.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local",
		conf.Username,
		conf.Password,
		conf.Address(),
		conf.Database,
		charset,
	)
}

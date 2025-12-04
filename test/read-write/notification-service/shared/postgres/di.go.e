package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Config struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func New(c Config) *bun.DB {
	return bun.NewDB(
		sql.OpenDB(
			pgdriver.NewConnector(
				pgdriver.WithAddr(c.Host+":"+c.Port),
				pgdriver.WithDatabase(c.Database),
				pgdriver.WithUser(c.User),
				pgdriver.WithPassword(c.Password),
				pgdriver.WithDialTimeout(time.Second*10),
				pgdriver.WithTLSConfig(nil),
			),
		),
		pgdialect.New(),
	)
}

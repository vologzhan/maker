package repositories

import (
	"hello-service/shared/postgres"
	table "hello-service/shared/postgres/repositories/table"

	"github.com/uptrace/bun"
)

type Repositories struct {
	Table *table.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		table.New(db),
	}, db
}

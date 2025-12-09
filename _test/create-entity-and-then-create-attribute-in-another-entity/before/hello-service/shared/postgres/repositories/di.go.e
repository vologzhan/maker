package repositories

import (
	"hello-service/shared/postgres"
	hi "hello-service/shared/postgres/repositories/hi"

	"github.com/uptrace/bun"
)

type Repositories struct {
	Hi *hi.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		hi.New(db),
	}, db
}

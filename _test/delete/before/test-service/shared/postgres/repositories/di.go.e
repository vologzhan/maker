package repositories

import (
	"test-service/shared/postgres"
	one "test-service/shared/postgres/repositories/one"
	two "test-service/shared/postgres/repositories/two"

	"github.com/uptrace/bun"
)

type Repositories struct {
	One *one.Repository
	Two *two.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		one.New(db),
		two.New(db),
	}, db
}

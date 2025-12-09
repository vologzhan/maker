package repositories

import (
	"test-service/shared/postgres"
	two "test-service/shared/postgres/repositories/two"

	"github.com/uptrace/bun"
)

type Repositories struct {
	Two *two.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		two.New(db),
	}, db
}

package repositories

import (
	"hello-service/shared/postgres"

	"github.com/uptrace/bun"
)

type Repositories struct {
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
	}, db
}

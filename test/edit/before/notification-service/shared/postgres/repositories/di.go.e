package repositories

import (
	"notification-service/shared/postgres"
	onlyId "notification-service/shared/postgres/repositories/only-id"

	"github.com/uptrace/bun"
)

type Repositories struct {
	OnlyId *onlyId.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		onlyId.New(db),
	}, db
}

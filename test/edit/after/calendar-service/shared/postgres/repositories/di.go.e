package repositories

import (
	"calendar-service/shared/postgres"
	onlyUuid "calendar-service/shared/postgres/repositories/only-uuid"

	"github.com/uptrace/bun"
)

type Repositories struct {
	OnlyUuid *onlyUuid.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		onlyUuid.New(db),
	}, db
}

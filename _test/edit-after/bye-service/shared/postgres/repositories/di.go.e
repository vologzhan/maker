package repositories

import (
	"bye-service/shared/postgres"
	profile "bye-service/shared/postgres/repositories/profile"

	"github.com/uptrace/bun"
)

type Repositories struct {
	Profile *profile.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		profile.New(db),
	}, db
}

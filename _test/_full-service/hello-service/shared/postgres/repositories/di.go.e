package repositories

import (
	"hello-service/shared/postgres"
	user "hello-service/shared/postgres/repositories/user"

	"github.com/uptrace/bun"
)

type Repositories struct {
	User *user.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		user.New(db),
	}, db
}

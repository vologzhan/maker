package repositories

import (
	"hello-service/shared/postgres"
	hi "hello-service/shared/postgres/repositories/hi"
	ho "hello-service/shared/postgres/repositories/ho"

	"github.com/uptrace/bun"
)

type Repositories struct {
	Hi *hi.Repository
	Ho *ho.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		hi.New(db),
		ho.New(db),
	}, db
}

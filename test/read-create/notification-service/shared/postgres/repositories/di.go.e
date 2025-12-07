package repositories

import (
	"notification-service/shared/postgres"
	channel "notification-service/shared/postgres/repositories/channel"

	"github.com/uptrace/bun"
)

type Repositories struct {
	Channel *channel.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		channel.New(db),
	}, db
}

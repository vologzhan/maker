package repositories

import (
	"hello-service/shared/postgres"
	user "hello-service/shared/postgres/repositories/user"
	userProfile "hello-service/shared/postgres/repositories/user-profile"

	"github.com/uptrace/bun"
)

type Repositories struct {
	User        *user.Repository
	UserProfile *userProfile.Repository
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
		user.New(db),
		userProfile.New(db),
	}, db
}

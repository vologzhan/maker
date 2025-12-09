package repositories

import (
	"▶service-name◀-service/shared/postgres"
⏩	▶entityName◀ "▶service-name◀-service/shared/postgres/repositories/▶entity-name◀"⏪

	"github.com/uptrace/bun"
)

type Repositories struct {
⏩	▶EntityName◀ *▶entityName◀.Repository⏪
}

func New(c postgres.Config) (*Repositories, *bun.DB) {
	db := postgres.New(c)

	return &Repositories{
		// maker:keep-di-repositories
⏩		▶entityName◀.New(db),⏪
	}, db
}

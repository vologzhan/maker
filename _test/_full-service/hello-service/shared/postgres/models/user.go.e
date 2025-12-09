package models

import (
	"hello-service/shared/dto"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:user"`

	Id        int        `bun:"id,pk"`      // maker:type_db=serial
	CreatedAt time.Time  `bun:"created_at"` // maker:type_db=timestamp(0),default=now()
	DeletedAt *time.Time `bun:"deleted_at"` // maker:type_db=timestamp(0),default=null

	// maker:keep-model-relations
}

func (m *User) ToDto() dto.User {
	return dto.User{
		m.Id,
		m.CreatedAt,
		m.DeletedAt,
	}
}

type Users []*User

func (m Users) ToDto() []dto.User {
	var out []dto.User
	for _, item := range m {
		out = append(out, item.ToDto())
	}
	return out
}

package models

import (
	"hello-service/shared/dto"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:user"`

	Uuid      uuid.UUID  `bun:"uuid,pk"`    // maker:type_db=uuid,default=uuid_generate_v4()
	DeletedAt *time.Time `bun:"deleted_at"` // maker:type_db=timestamp(0),default=null

	// maker:keep-model-relations
}

func (m *User) ToDto() dto.User {
	return dto.User{
		m.Uuid,
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

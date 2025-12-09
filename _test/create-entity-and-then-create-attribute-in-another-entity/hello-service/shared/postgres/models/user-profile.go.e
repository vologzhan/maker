package models

import (
	"hello-service/shared/dto"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserProfile struct {
	bun.BaseModel `bun:"table:user_profile"`

	UserUuid   uuid.UUID `bun:"user_uuid,pk"` // maker:type_db=uuid,fk=user|one-to-one
	Name       string    `bun:"name"`         // maker:type_db=varchar(255)
	SecondName *string   `bun:"second_name"`  // maker:type_db=varchar(255),default=null
	Foo        int       `bun:"foo"`          // maker:type_db=int

	// maker:keep-model-relations
}

func (m *UserProfile) ToDto() dto.UserProfile {
	return dto.UserProfile{
		m.UserUuid,
		m.Name,
		m.SecondName,
		m.Foo,
	}
}

type UserProfiles []*UserProfile

func (m UserProfiles) ToDto() []dto.UserProfile {
	var out []dto.UserProfile
	for _, item := range m {
		out = append(out, item.ToDto())
	}
	return out
}

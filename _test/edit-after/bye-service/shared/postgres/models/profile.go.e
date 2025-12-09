package models

import (
	"bye-service/shared/dto"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Profile struct {
	bun.BaseModel `bun:"table:profile"`

	Uuid      uuid.UUID  `bun:"uuid,pk"`    // maker:type_db=uuid,default=uuid_generate_v4()
	CreatedAt time.Time  `bun:"created_at"` // maker:type_db=timestamp(0),default=now()
	DeletedAt *time.Time `bun:"deleted_at"` // maker:type_db=timestamp(0),default=null

	// maker:keep-model-relations
}

func (m *Profile) ToDto() dto.Profile {
	return dto.Profile{
		m.Uuid,
		m.CreatedAt,
		m.DeletedAt,
	}
}

type Profiles []*Profile

func (m Profiles) ToDto() []dto.Profile {
	var out []dto.Profile
	for _, item := range m {
		out = append(out, item.ToDto())
	}
	return out
}

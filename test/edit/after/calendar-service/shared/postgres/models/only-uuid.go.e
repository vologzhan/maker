package models

import (
	"calendar-service/shared/dto"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OnlyUuid struct {
	bun.BaseModel `bun:"table:only_uuid"`

	Uuid uuid.UUID `bun:"uuid,pk"` // maker:type_db=uuid,default=uuid_generate_v4()
}

func (m *OnlyUuid) ToDto() dto.OnlyUuid {
	return dto.OnlyUuid{
		m.Uuid,
	}
}

package models

import (
	"notification-service/shared/dto"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Channel struct {
	bun.BaseModel `bun:"table:channel"`

	Uuid uuid.UUID `bun:"uuid,pk"` // maker:type_db=uuid,default=uuid_generate_v4()
}

func (m *Channel) ToDto() dto.Channel {
	return dto.Channel{
		m.Uuid,
	}
}

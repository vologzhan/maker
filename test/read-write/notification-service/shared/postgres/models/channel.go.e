package models

import (
	"notification-service/shared/dto"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Channel struct {
	bun.BaseModel `bun:"table:channel"`

	Uuid         uuid.UUID  `bun:"uuid,pk"`       // maker:type_db=uuid,default=uuid_generate_v4()
	RelationUuid uuid.UUID  `bun:"relation_uuid"` // maker:type_db=uuid,fk=foreign_table|one-to-one
	DeletedAt    *time.Time `bun:"deleted_at"`    // maker:type_db=timestamp(0),default=null
}

func (m *Channel) ToDto() dto.Channel {
	return dto.Channel{
		m.Uuid,
		m.RelationUuid,
		m.DeletedAt,
	}
}

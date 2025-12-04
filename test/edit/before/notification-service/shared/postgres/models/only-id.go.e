package models

import (
	"notification-service/shared/dto"

	"github.com/uptrace/bun"
)

type OnlyId struct {
	bun.BaseModel `bun:"table:only_id"`

	Id int `bun:"id,pk"` // maker:type_db=serial
}

func (m *OnlyId) ToDto() dto.OnlyId {
	return dto.OnlyId{
		m.Id,
	}
}

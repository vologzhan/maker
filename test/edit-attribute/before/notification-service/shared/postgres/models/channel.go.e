package models

import (
	"notification-service/shared/dto"

	"github.com/uptrace/bun"
)

type Channel struct {
	bun.BaseModel `bun:"table:channel"`

	Id int `bun:"id,pk"` // maker:type_db=serial
}

func (m *Channel) ToDto() dto.Channel {
	return dto.Channel{
		m.Id,
	}
}

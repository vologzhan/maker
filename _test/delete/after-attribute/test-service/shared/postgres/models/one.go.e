package models

import (
	"test-service/shared/dto"

	"github.com/uptrace/bun"
)

type One struct {
	bun.BaseModel `bun:"table:one"`

	Id int `bun:"id,pk"` // maker:type_db=serial
}

func (m *One) ToDto() dto.One {
	return dto.One{
		m.Id,
	}
}

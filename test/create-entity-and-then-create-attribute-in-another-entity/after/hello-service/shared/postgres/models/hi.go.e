package models

import (
	"hello-service/shared/dto"

	"github.com/uptrace/bun"
)

type Hi struct {
	bun.BaseModel `bun:"table:hi"`

	Id  int `bun:"id,pk"` // maker:type_db=serial
	Foo int `bun:"foo"`   // maker:type_db=int
}

func (m *Hi) ToDto() dto.Hi {
	return dto.Hi{
		m.Id,
		m.Foo,
	}
}

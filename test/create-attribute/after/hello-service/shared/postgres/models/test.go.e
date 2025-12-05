package models

import (
	"hello-service/shared/dto"

	"github.com/uptrace/bun"
)

type Test struct {
	bun.BaseModel `bun:"table:test"`

	Id int `bun:"id,pk"` // maker:type_db=serial
	Hi int `bun:"hi"`    // maker:type_db=int
}

func (m *Test) ToDto() dto.Test {
	return dto.Test{
		m.Id,
		m.Hi,
	}
}

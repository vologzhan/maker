package models

import (
	"hello-service/shared/dto"

	"github.com/uptrace/bun"
)

type Test struct {
	bun.BaseModel `bun:"table:test"`

	Id int `bun:"id,pk"` // maker:type_db=serial
}

func (m *Test) ToDto() dto.Test {
	return dto.Test{
		m.Id,
	}
}

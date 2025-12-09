package models

import (
	"test-service/shared/dto"

	"github.com/uptrace/bun"
)

type Two struct {
	bun.BaseModel `bun:"table:two"`

	Id   int    `bun:"id,pk"` // maker:type_db=serial
	Name string `bun:"name"`  // maker:type_db=varchar(255)
}

func (m *Two) ToDto() dto.Two {
	return dto.Two{
		m.Id,
		m.Name,
	}
}

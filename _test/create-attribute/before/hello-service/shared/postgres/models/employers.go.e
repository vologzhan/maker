package models

import (
	"hello-service/shared/dto"

	"github.com/uptrace/bun"
)

type Employers struct {
	bun.BaseModel `bun:"table:employers"`

	Id int `bun:"id,pk"` // maker:type_db=serial
}

func (m *Employers) ToDto() dto.Employers {
	return dto.Employers{
		m.Id,
	}
}

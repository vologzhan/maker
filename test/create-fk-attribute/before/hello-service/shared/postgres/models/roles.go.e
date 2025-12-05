package models

import (
	"hello-service/shared/dto"

	"github.com/uptrace/bun"
)

type Roles struct {
	bun.BaseModel `bun:"table:roles"`

	Id int `bun:"id,pk"` // maker:type_db=serial
}

func (m *Roles) ToDto() dto.Roles {
	return dto.Roles{
		m.Id,
	}
}

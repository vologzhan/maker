package models

import (
	"hello-service/shared/dto"

	"github.com/uptrace/bun"
)

type Table struct {
	bun.BaseModel `bun:"table:table"`
}

func (m *Table) ToDto() dto.Table {
	return dto.Table{}
}

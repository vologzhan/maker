package models

import (
	"hello-service/shared/dto"

	"github.com/uptrace/bun"
)

type Table struct {
	bun.BaseModel `bun:"table:table"`

	// maker:keep-model-relations
}

func (m *Table) ToDto() dto.Table {
	return dto.Table{}
}

type Tables []*Table

func (m Tables) ToDto() []dto.Table {
	var out []dto.Table
	for _, item := range m {
		out = append(out, item.ToDto())
	}
	return out
}

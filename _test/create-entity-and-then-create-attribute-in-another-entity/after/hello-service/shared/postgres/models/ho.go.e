package models

import (
	"hello-service/shared/dto"

	"github.com/uptrace/bun"
)

type Ho struct {
	bun.BaseModel `bun:"table:ho"`
}

func (m *Ho) ToDto() dto.Ho {
	return dto.Ho{}
}

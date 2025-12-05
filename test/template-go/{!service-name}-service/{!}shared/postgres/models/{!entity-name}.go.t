package models

import (
	"▶service-name◀-service/shared/dto"

	"github.com/uptrace/bun"
)

type ▶EntityName◀ struct {
	bun.BaseModel `bun:"table:▶⬇Entity➡NameDb◀"`

⏩	▶⬇AttributeName◀ ▶⬇Attribute➡Nullable↔*◀▶⬇Attribute➡TypeGo◀ `bun:"▶⬇Attribute➡NameDb◀▶⬇Attribute➡PrimaryKey↔,pk◀"` // maker:type_db=▶⬇Attribute➡TypeDb◀▶,default=▶⬇Attribute➡Default◀◀▶,fk=▶⬇Attribute➡FkTable◀|▶⬇Attribute➡FkType◀◀⏪
}

func (m *▶EntityName◀) ToDto() dto.▶EntityName◀ {
	return dto.▶EntityName◀{
⏩		m.▶AttributeName◀,⏪
	}
}

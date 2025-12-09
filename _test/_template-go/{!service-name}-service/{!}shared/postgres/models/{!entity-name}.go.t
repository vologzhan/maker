package models

import (
	"▶service-name◀-service/shared/dto"

	"github.com/uptrace/bun"
)

type ▶EntityName◀ struct {
	bun.BaseModel `bun:"table:▶⬇Entity➡NameDb◀"`

⏩	▶⬇AttributeName◀ ▶⬇Attribute➡Nullable↔*◀▶⬇Attribute➡TypeGo◀ `bun:"▶⬇Attribute➡NameDb◀▶⬇Attribute➡PrimaryKey↔,pk◀"` // maker:type_db=▶⬇Attribute➡TypeDb◀▶,default=▶⬇Attribute➡Default◀◀▶,fk=▶⬇Attribute➡FkTable◀|▶⬇Attribute➡FkType◀◀⏪

	// maker:keep-model-relations
⏩	▶⬇RelationName◀ ▶⬇Relation➡OneToMany↔[]◀*▶⬇Relation➡Entity◀ `bun:"rel:▶⬇Relation➡Type◀,join:▶⬇Relation➡ColumnName:Relation➡ColumnNameFk◀"⏪
}

func (m *▶EntityName◀) ToDto() dto.▶EntityName◀ {
	return dto.▶EntityName◀{
⏩		m.▶AttributeName◀,⏪
	}
}

type ▶⬇EntityPluralName◀ []*▶EntityName◀

func (m ▶EntityPluralName◀) ToDto() []dto.▶EntityName◀ {
	var out []dto.▶EntityName◀
	for _, item := range m {
		out = append(out, item.ToDto())
	}
	return out
}

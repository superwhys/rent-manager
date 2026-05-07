package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type AccountMixin struct {
	mixin.Schema
}

func (AccountMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("account_id").
			Comment("所属账号ID"),
	}
}

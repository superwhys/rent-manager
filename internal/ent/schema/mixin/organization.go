package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type OrganizationMixin struct {
	mixin.Schema
}

func (OrganizationMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("organization_id").
			Comment("所属组织 id"),
	}
}

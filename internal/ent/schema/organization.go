package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/superwhys/rent-manager-server/internal/ent/schema/mixin"
)

// Organization holds buildings and employee accounts.
type Organization struct {
	ent.Schema
}

func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("组织名称"),
		field.Bool("is_active").
			Default(true).
			Comment("组织是否激活"),
	}
}

func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("accounts", Account.Type),
		edge.To("buildings", Building.Type),
		edge.To("roles", Role.Type),
		edge.To("room_types", RoomType.Type),
		edge.To("meter_items", MeterItem.Type),
	}
}

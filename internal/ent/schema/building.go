package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/superwhys/rent-manager-server/internal/ent/schema/mixin"
)

// 楼栋实体
type Building struct {
	ent.Schema
}

func (Building) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
		mixin.OrganizationMixin{},
	}
}

// Fields of the Building.
func (Building) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("楼栋名字"),
		field.String("address").
			NotEmpty().
			MaxLen(255).
			Comment("楼栋地址"),
		field.String("city").
			NotEmpty().
			MaxLen(255).
			Comment("楼栋城市"),
		field.String("province").
			NotEmpty().
			MaxLen(255).
			Comment("楼栋省份"),
	}
}

// Edges of the Building.
func (Building) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Field("organization_id").
			Ref("buildings").
			Unique().
			Required(),

		edge.To("rooms", Room.Type),

		edge.To("meter_items", MeterItem.Type),
	}
}

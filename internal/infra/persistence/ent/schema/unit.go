package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/schema/mixin"
)

// 单元实体
type Unit struct {
	ent.Schema
}

func (Unit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
		mixin.AccountMixin{},
	}
}

func (Unit) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("单元名字"),

		field.String("address").
			NotEmpty().
			MaxLen(255).
			Comment("单元地址"),

		field.String("city").
			NotEmpty().
			MaxLen(255).
			Comment("单元城市"),

		field.String("province").
			NotEmpty().
			MaxLen(255).
			Comment("单元省份"),
	}
}

func (Unit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).
			Ref("units").
			Field("account_id").
			Unique().
			Required(),

		edge.To("rooms", Room.Type),

		edge.To("meter_items", MeterItem.Type),

		edge.To("unit_account_roles", UnitAccountRole.Type),
	}
}

func (Unit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id", "name").Unique(),
	}
}

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/ent/schema/mixin"
)

// UnitAccountRole 用于记录一个 unit 授予给一个 account的记录
type UnitAccountRole struct {
	ent.Schema
}

func (UnitAccountRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

func (UnitAccountRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int("unit_id").
			Comment("单元ID"),

		field.Int("account_id").
			Comment("被授予权限的账号ID"),

		field.Int("role_id").
			Comment("角色ID"),

		field.Int("granted_by_account_id").
			Comment("授予权限的账号ID"),

		field.Enum("status").
			Comment("状态").
			Values("active", "disabled").
			Default("active"),
	}
}

func (UnitAccountRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("unit", Unit.Type).
			Ref("unit_account_roles").
			Field("unit_id").
			Unique().
			Required(),

		edge.From("account", Account.Type).
			Ref("unit_account_roles").
			Field("account_id").
			Unique().
			Required(),

		edge.From("role", Role.Type).
			Ref("unit_account_roles").
			Field("role_id").
			Unique().
			Required(),

		edge.From("granted_by", Account.Type).
			Ref("granted_unit_account_roles").
			Field("granted_by_account_id").
			Unique().
			Required(),
	}
}

func (UnitAccountRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("unit_id", "account_id", "role_id").Unique(),
		index.Fields("account_id"),
		index.Fields("unit_id"),
		index.Fields("role_id"),
	}
}

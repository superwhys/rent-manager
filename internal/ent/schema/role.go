package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/ent/schema/mixin"
)

// 角色实体
type Role struct {
	ent.Schema
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
		mixin.AccountMixin{},
	}
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("角色名字"),

		field.Bool("is_default_owner").
			Default(false).
			Comment("是否默认楼栋长角色"),

		field.Bool("is_active").
			Default(true).
			Comment("角色是否激活"),
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("accounts", Account.Type).
			Ref("roles").
			Field("account_id").
			Unique().
			Required(),

		edge.To("permission_groups", PermissionGroup.Type),

		edge.To("unit_account_roles", UnitAccountRole.Type),
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id", "name").Unique(),
	}
}

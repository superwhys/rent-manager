package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/superwhys/rent-manager-server/internal/ent/schema/mixin"
)

// 角色实体
type Role struct {
	ent.Schema
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.Int("organization_id").
			Comment("所属组织ID"),
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
		edge.From("organization", Organization.Type).
			Field("organization_id").
			Ref("roles").
			Unique().
			Required(),
		edge.From("accounts", Account.Type).
			Ref("roles"),
		edge.To("permission_groups", PermissionGroup.Type),
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id", "name").Unique(),
	}
}

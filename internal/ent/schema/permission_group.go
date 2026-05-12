package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/superwhys/rent-manager-server/internal/ent/schema/mixin"
)

// PermissionGroup groups global permissions for organization roles.
type PermissionGroup struct {
	ent.Schema
}

func (PermissionGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

func (PermissionGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			NotEmpty().
			MaxLen(255).
			Comment("权限组代码"),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("权限组名称"),
		field.String("description").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("权限组描述"),
		field.Int("sort_order").
			Default(0).
			Comment("排序"),
		field.Bool("is_active").
			Default(true).
			Comment("权限组是否激活"),
	}
}

func (PermissionGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("permissions", Permission.Type),
		edge.From("roles", Role.Type).
			Ref("permission_groups"),
	}
}

func (PermissionGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/ent/schema/mixin"
)

// 权限实体
type Permission struct {
	ent.Schema
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Fields of the Permission.
func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			NotEmpty().
			MaxLen(255).
			Comment("权限代码"),

		field.String("description").
			NotEmpty().
			MaxLen(255).
			Comment("权限描述"),

		field.Bool("is_active").
			Default(true).
			Comment("权限是否激活"),
	}
}

// Edges of the Permission.
func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("permission_groups", PermissionGroup.Type).
			Ref("permissions"),
	}
}

func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/schema/mixin"
)

// AccountSetting 表示账号级配置。
type AccountSetting struct {
	ent.Schema
}

func (AccountSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
		mixin.AccountMixin{},
	}
}

func (AccountSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty().
			MaxLen(64).
			Comment("配置键"),

		field.String("value").
			NotEmpty().
			MaxLen(255).
			Comment("配置值"),
	}
}

func (AccountSetting) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).
			Ref("account_settings").
			Field("account_id").
			Unique().
			Required(),
	}
}

func (AccountSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id", "key").Unique(),
	}
}

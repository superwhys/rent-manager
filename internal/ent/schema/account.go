package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/superwhys/rent-manager-server/internal/ent/schema/mixin"
)

// 账号实体
type Account struct {
	ent.Schema
}

func (Account) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Fields of the Account.
func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("账号名字"),

		field.String("email").
			NotEmpty().
			Match(regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).
			Comment("账号邮箱"),

		field.String("password").
			NotEmpty().
			MaxLen(255).
			Comment("账号密码"),

		field.String("wechat_openid").
			NotEmpty().
			MaxLen(255).
			Comment("微信开放ID"),

		field.Bool("is_active").
			Default(true).
			Comment("账号是否激活"),
	}
}

// Edges of the Account.
func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roles", Role.Type),
		edge.To("units", Unit.Type),
		edge.To("room_types", RoomType.Type),
		edge.To("meter_items", MeterItem.Type),

		// 这个账号被授予了哪些 unit-role
		edge.To("unit_account_roles", UnitAccountRole.Type),
		// 这个账号发起授予了哪些 unit-role
		edge.To("granted_unit_account_roles", UnitAccountRole.Type),
	}
}

func (Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("wechat_openid").Unique(),
	}
}

package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/superwhys/rent-manager-server/internal/ent/mixin"
)

// Account holds the schema definition for the Account entity.
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
		field.String("name").NotEmpty().MaxLen(255),
		field.String("email").NotEmpty().Match(regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)),
		field.String("password").NotEmpty().MaxLen(255),
		field.Bool("is_active").Default(true),
	}
}

// Edges of the Account.
func (Account) Edges() []ent.Edge {
	return []ent.Edge{}
}

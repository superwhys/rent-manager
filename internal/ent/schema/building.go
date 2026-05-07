package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/superwhys/rent-manager-server/internal/ent/mixin"
)

// Building holds the schema definition for the Building entity.
type Building struct {
	ent.Schema
}

func (Building) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Fields of the Building.
func (Building) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(255),
		field.String("address").NotEmpty().MaxLen(255),
		field.String("city").NotEmpty().MaxLen(255),
		field.String("province").NotEmpty().MaxLen(255),
	}
}

// Edges of the Building.
func (Building) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("rooms", Room.Type),
	}
}

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/superwhys/rent-manager-server/internal/ent/mixin"
)

// Room holds the schema definition for the Room entity.
type Room struct {
	ent.Schema
}

func (Room) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Fields of the Room.
func (Room) Fields() []ent.Field {
	return []ent.Field{
		field.String("room_no").NotEmpty().MaxLen(255),
		field.Int("floor"),
		field.Int("price"),
		field.Int("size"),
		field.Bool("is_active").Default(true),
	}
}

// Edges of the Room.
func (Room) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("building", Building.Type).
			Ref("rooms").
			Unique(),
	}
}

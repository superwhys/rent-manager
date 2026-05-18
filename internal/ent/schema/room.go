package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/superwhys/rent-manager-server/internal/ent/schema/mixin"
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
		field.String("room_no").
			NotEmpty().
			MaxLen(255).
			Comment("房间号"),

		field.Int("floor").
			Comment("楼层"),

		field.Int("price").
			Comment("租金价格"),

		field.Int("size").
			Comment("房间大小(平方米)"),

		field.Int("unit_id").
			Comment("单元ID"),

		field.Int("room_type_id").
			Comment("所属房间类型ID"),

		field.Bool("is_active").
			Default(true).
			Comment("房间是否激活"),
	}
}

// Edges of the Room.
func (Room) Edges() []ent.Edge {
	return []ent.Edge{
		// 房间-楼栋 多对一关系
		edge.From("unit", Unit.Type).
			Field("unit_id").
			Ref("rooms").
			Unique().
			Required(),

		// 房间-房间类型 多对一关系
		edge.From("room_type", RoomType.Type).
			Field("room_type_id").
			Ref("rooms").
			Unique().
			Required(),

		// 房间-抄表项 多对多关系
		edge.To("meter_items", MeterItem.Type),

		// 房间-抄表数据 多对多关系
		edge.To("meter_readings", MeterReading.Type),

		// 房间-租金账单 一对多关系
		edge.To("rent_bills", RentBill.Type),
	}
}

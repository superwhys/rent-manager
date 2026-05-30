package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/ent/schema/mixin"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/consts"
)

// RentBill records a room's monthly rent bill snapshot.
type RentBill struct {
	ent.Schema
}

func (RentBill) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

func (RentBill) Fields() []ent.Field {
	return []ent.Field{
		field.Int("room_id").
			Comment("所属房间ID"),

		field.String("period_month").
			NotEmpty().
			Comment("账单月份，格式 YYYY-MM"),

		field.Int("base_rent").
			Comment("基础租金，单位：分").
			Default(0),

		field.Int("meter_fee").
			Comment("抄表费用合计，单位：分").
			Default(0),

		field.Int("total_amount").
			Comment("账单总金额，单位：分").
			Default(0),

		field.Enum("status").
			Values(consts.RentBillStatusUnpaid, consts.RentBillStatusPaid, consts.RentBillStatusCancelled).
			Default(consts.RentBillStatusUnpaid).
			Comment("账单状态"),

		field.Time("paid_at").
			Optional().
			Nillable().
			Comment("支付时间"),

		field.String("remark").
			Optional().
			Nillable().
			Comment("备注"),
	}
}

func (RentBill) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("room", Room.Type).
			Field("room_id").
			Ref("rent_bills").
			Unique().
			Required(),
	}
}

func (RentBill) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("room_id", "period_month").Unique(),
		index.Fields("period_month"),
		index.Fields("status"),
	}
}

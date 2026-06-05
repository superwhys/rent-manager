// File:		meter_item.go
// Created by:	Hoven
// Created on:	2026-05-07
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/schema/mixin"

	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
)

// 抄表项实体
// 用于设置房间需要抄表的项目
type MeterItem struct {
	ent.Schema
}

func (MeterItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
		mixin.AccountMixin{},
	}
}

func (MeterItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			Comment("抄表项名称，如 水/电/天然气").
			NotEmpty(),

		field.String("unit").
			Comment("单位，如 m³/kWh").
			NotEmpty(),

		field.Int("unit_price").
			Comment("单价，单位：分/计量单位").
			Default(0),

		field.Bool("is_system").
			Comment("是否系统内置").
			Default(false),

		field.Enum("calc_type").
			Values(domainmeteritem.CalcTypeDifference, domainmeteritem.CalcTypeDirect).
			Comment("计算方式：DIFFERENCE-差值型，DIRECT-直接录入型").
			Default(domainmeteritem.CalcTypeDifference),

		field.String("remark").
			Comment("备注").
			Optional().
			Nillable(),
	}
}

func (MeterItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id", "code").Unique(),
	}
}

func (MeterItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("accounts", Account.Type).
			Ref("meter_items").
			Field("account_id").
			Unique().
			Required(),

		// 抄表项-房间 多对多关系
		edge.From("rooms", Room.Type).
			Ref("meter_items"),

		// 抄表项-抄表数据 一对多关系
		edge.To("meter_readings", MeterReading.Type),

		// 抄表项-楼栋 多对多关系
		// 用于给楼栋设置默认的抄表项
		// 楼栋内的房间可以选择继承默认的或者单独设置
		edge.From("uint", Unit.Type).
			Ref("meter_items"),
	}
}

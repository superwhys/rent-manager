// File:		meter_reading.go
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

	domainmeterreading "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meterreading"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/schema/mixin"
)

// 抄表数据实体
// 用于记录房间的抄表数据
type MeterReading struct {
	ent.Schema
}

func (MeterReading) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

func (MeterReading) Fields() []ent.Field {
	return []ent.Field{
		field.Int("room_id").
			Comment("所属房间ID"),

		field.Int("meter_item_id").
			Comment("所属抄表项ID"),

		field.String("period_month").
			NotEmpty().
			Comment("抄表月份，格式 YYYY-MM"),

		field.Float("previous_reading").
			Optional().
			Nillable().
			Comment("上次读数"),

		field.Float("current_reading").
			Comment("本次读数"),

		field.Float("usage_amount").
			Optional().
			Nillable().
			Comment("用量"),

		field.Time("reading_time").
			Comment("抄表时间"),

		field.Int64("reader_id").
			Optional().
			Nillable().
			Comment("抄表人ID"),

		field.Enum("source").
			Values(domainmeterreading.SourceManual, domainmeterreading.SourceImport, domainmeterreading.SourceAuto).
			Default(domainmeterreading.SourceManual).
			Comment("数据来源"),

		field.String("remark").
			Optional().
			Nillable().
			Comment("备注"),
	}
}

func (MeterReading) Edges() []ent.Edge {
	return []ent.Edge{
		// 抄表数据-房间 多对一关系
		edge.From("room", Room.Type).
			Field("room_id").
			Ref("meter_readings").
			Unique().
			Required(),

		// 抄表数据-抄表项 多对一关系
		edge.From("meter_item", MeterItem.Type).
			Field("meter_item_id").
			Ref("meter_readings").
			Unique().
			Required(),
	}
}

func (MeterReading) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("room_id", "meter_item_id", "period_month").Unique(),
		index.Fields("meter_item_id"),
		index.Fields("period_month"),
	}
}

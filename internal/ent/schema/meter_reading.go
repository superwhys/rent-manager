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
	"github.com/superwhys/rent-manager-server/internal/models/consts"
)

// 抄表数据实体
// 用于记录房间的抄表数据
type MeterReading struct {
	ent.Schema
}

func (MeterReading) Fields() []ent.Field {
	return []ent.Field{
		field.Int("room_id").
			Comment("所属房间ID"),

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
			Values(consts.SourceManual, consts.SourceImport, consts.SourceAuto).
			Default(consts.SourceManual).
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
	}
}

func (MeterReading) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("room_id", "period_month").Unique(),
		index.Fields("period_month"),
	}
}

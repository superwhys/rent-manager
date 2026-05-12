// File:		room_option.go
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
	"github.com/superwhys/rent-manager-server/internal/ent/schema/mixin"
)

// 房间类型实体
type RoomType struct {
	ent.Schema
}

func (RoomType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
		mixin.OrganizationMixin{},
	}
}

func (RoomType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(12).
			Comment("房间选项名字"),
	}
}

func (RoomType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Field("organization_id").
			Ref("room_types").
			Unique().
			Required(),
		edge.To("rooms", Room.Type),
	}
}

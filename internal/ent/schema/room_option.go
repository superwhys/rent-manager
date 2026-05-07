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
	"entgo.io/ent/schema/field"
	"github.com/superwhys/rent-manager-server/internal/ent/mixin"
)

// RoomOption holds the schema definition for the RoomOption entity.
type RoomOption struct {
	ent.Schema
}

func (RoomOption) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

func (RoomOption) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(12).Comment("房间选项名字"),
	}
}

func (RoomOption) Edges() []ent.Edge {
	return []ent.Edge{}
}

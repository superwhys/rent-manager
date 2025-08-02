// File:		room.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package model

import (
	"time"

	"gorm.io/gorm"
)

type RoomStatus string

const (
	StatusRented     RoomStatus = "rented"
	StatusWaitingPay RoomStatus = "waiting_pay"
	StatusFree       RoomStatus = "free"
)

// Room 房间信息结构体
type Room struct {
	ID uint `gorm:"primarykey" json:"id"`

	PropertyID uint       `gorm:"not null;index;comment:物业ID" json:"property_id"`
	RoomId     string     `gorm:"not null;type:varchar(50);comment:房间号" json:"roomId"`
	Floor      int        `gorm:"not null;type:integer;comment:楼层" json:"floor"`
	Area       float32    `gorm:"not null;type:numeric(10,2);default:0;comment:面积" json:"area"`
	Rent       float32    `gorm:"not null;type:numeric(10,2);default:0;comment:租金" json:"rent"`
	Status     RoomStatus `gorm:"not null;type:varchar(50);comment:房间状态" json:"status"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Property Property `json:"property,omitempty"`
}

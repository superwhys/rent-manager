// File:		room.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package dbModel

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

type Room struct {
	ID uint `gorm:"primarykey" json:"id"`

	RoomId string     `gorm:"not null;type:varchar(50)" json:"roomId"`
	Floor  int        `gorm:"not null;type:integer" json:"floor"`
	Area   float32    `gorm:"not null;type:numeric(10,2);default:0" json:"area"`
	Rent   float32    `gorm:"not null;type:numeric(10,2);default:0" json:"rent"`
	Status RoomStatus `gorm:"not null;type:varchar(50)" json:"status"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// File:		meter_reading.go
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

// MeterReading 抄表记录结构体
type MeterReading struct {
	ID uint `gorm:"primarykey" json:"id"`

	RoomID      uint      `gorm:"not null;index;comment:房间ID" json:"room_id"`
	PropertyID  uint      `gorm:"not null;index;comment:物业ID" json:"property_id"`
	ReadingDate time.Time `gorm:"not null;type:date;comment:抄表日期" json:"reading_date"`

	// 表读数
	WaterReading    float64 `gorm:"not null;type:numeric(12,2);default:0;comment:水表读数" json:"water_reading"`
	ElectricReading float64 `gorm:"not null;type:numeric(12,2);default:0;comment:电表读数" json:"electric_reading"`
	GasReading      float64 `gorm:"not null;type:numeric(12,2);default:0;comment:气表读数" json:"gas_reading"`

	// 使用量（本次读数 - 上次读数）
	WaterUsage    float64 `gorm:"type:numeric(10,2);default:0;comment:用水量(方)" json:"water_usage"`
	ElectricUsage float64 `gorm:"type:numeric(10,2);default:0;comment:用电量(度)" json:"electric_usage"`
	GasUsage      float64 `gorm:"type:numeric(10,2);default:0;comment:用气量(方)" json:"gas_usage"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Room     Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Property Property `gorm:"foreignKey:PropertyID" json:"property,omitempty"`
}

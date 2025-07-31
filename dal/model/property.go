// File:		property.go
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

// WaterSettings 水费设置结构体
type WaterSettings struct {
	UnitPrice     float64 `json:"unit_price"`     // 水费单价(元/方)
	MinimumCharge float64 `json:"minimum_charge"` // 保底费用(元)
}

// ElectricSettings 电费设置结构体
type ElectricSettings struct {
	UnitPrice     float64 `json:"unit_price"`     // 电费单价(元/度)
	MinimumCharge float64 `json:"minimum_charge"` // 保底费用(元)
}

// GasSettings 燃气设置结构体
type GasSettings struct {
	UnitPrice     float64 `json:"unit_price"`     // 燃气单价(元/方)
	MinimumCharge float64 `json:"minimum_charge"` // 保底费用(元)
}

// UtilitySettings 水电燃气费用设置结构体
type UtilitySettings struct {
	Water    WaterSettings    `json:"water"`    // 水费设置
	Electric ElectricSettings `json:"electric"` // 电费设置
	Gas      GasSettings      `json:"gas"`      // 燃气设置
}

// Property 物业信息结构体
type Property struct {
	ID uint `gorm:"primarykey" json:"id"`

	Name            string `gorm:"not null;type:varchar(50);comment:物业名称" json:"name"`
	Address         string `gorm:"not null;type:varchar(128);comment:物业地址" json:"address"`
	UtilitySettings string `gorm:"type:json;comment:水电燃气费用设置" json:"utility_settings"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

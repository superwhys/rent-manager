// File:		tenant.go
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

type ContractStatus string

const (
	ContractStatusNormal      ContractStatus = "normal"       // 正常
	ContractStatusWaitingRent ContractStatus = "waiting_rent" // 待租房
	ContractStatusExpired     ContractStatus = "expired"      // 已过期
	ContractStatusTerminated  ContractStatus = "terminated"   // 已终止
)

// Tenant 租客信息结构体
type Tenant struct {
	ID uint `gorm:"primarykey" json:"id"`

	Name           string         `gorm:"not null;type:varchar(50);comment:租客姓名" json:"name"`
	Phone          string         `gorm:"not null;type:varchar(20);comment:手机号" json:"phone"`
	IDCard         string         `gorm:"not null;type:varchar(20);comment:身份证号" json:"id_card"`
	RoomID         uint           `gorm:"index;comment:房间ID" json:"room_id"`                                                    // 房间ID，外键关联房间表，可为空（待租房状态）
	PropertyID     uint           `gorm:"not null;index;comment:物业ID" json:"property_id"`                                       // 物业ID，外键关联物业表
	Deposit        float64        `gorm:"not null;type:numeric(10,2);default:0;comment:押金" json:"deposit"`                      // 押金(元)
	CheckInDate    *time.Time     `gorm:"type:date;comment:入住时间" json:"check_in_date"`                                          // 入住时间，可为空
	ContractStatus ContractStatus `gorm:"not null;type:varchar(20);default:'waiting_rent';comment:合同状态" json:"contract_status"` // 合同状态

	// 合同相关信息
	ContractStartDate *time.Time `gorm:"type:date;comment:合同开始日期" json:"contract_start_date"`          // 合同开始日期
	ContractEndDate   *time.Time `gorm:"type:date;comment:合同结束日期" json:"contract_end_date"`            // 合同结束日期
	MonthlyRent       float64    `gorm:"type:numeric(10,2);default:0;comment:月租金" json:"monthly_rent"` // 月租金

	Remark string `gorm:"type:text;comment:备注" json:"remark"` // 备注

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Room     *Room    `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Property Property `gorm:"foreignKey:PropertyID" json:"property,omitempty"`
}

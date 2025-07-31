// File:		setting.go
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

type AutoBackupFrequency string

const (
	BackupDaily   AutoBackupFrequency = "daily"   // 每日备份
	BackupWeekly  AutoBackupFrequency = "weekly"  // 每周备份
	BackupMonthly AutoBackupFrequency = "monthly" // 每月备份
)

type SystemSettings struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 费用设置
	DefaultRent         float32             `gorm:"not null;type:numeric(10,2);default:0" json:"defaultRent" binding:"min=0"`
	DepositMultiple     float32             `gorm:"not null;type:numeric(3,1);default:1" json:"depositMultiple" binding:"min=0"`
	NetworkFee          float32             `gorm:"not null;type:numeric(8,2);default:0" json:"networkFee" binding:"min=0"`
	ElevatorFee         float32             `gorm:"not null;type:numeric(8,2);default:0" json:"elevatorFee" binding:"min=0"`
	BillGenerateDate    int                 `gorm:"not null;type:integer;default:1" json:"billGenerateDate" binding:"min=1,max=30"`
	OverdueReminderDays int                 `gorm:"not null;type:integer;default:3" json:"overdueReminderDays" binding:"min=1,max=30"`
	AutoBackup          AutoBackupFrequency `gorm:"not null;type:varchar(20);default:'daily'" json:"autoBackup" binding:"oneof=daily weekly monthly"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

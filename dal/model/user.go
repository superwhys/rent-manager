// File:		user.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package dbModel

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleLandlord UserRole = "landlord" // 主房东
	RoleEmployee UserRole = "employee" // 员工
)

type UserStatus string

const (
	StatusActive   UserStatus = "active"   // 正常
	StatusInactive UserStatus = "inactive" // 禁用
)

type User struct {
	ID uint `gorm:"primarykey" json:"id"`

	Username          string        `gorm:"uniqueIndex;not null;type:varchar(50)" json:"username"`
	Phone             string        `gorm:"uniqueIndex;not null;type:varchar(20)" json:"phone"`
	Role              UserRole      `gorm:"not null;type:varchar(20)" json:"role"`
	Status            UserStatus    `gorm:"not null;type:varchar(20);default:'active'" json:"status"`
	RealName          string        `gorm:"not null;type:varchar(50)" json:"realName"`
	PasswordHash      string        `gorm:"not null;type:varchar(255)" json:"-"`     // 密码哈希，不返回给前端
	ManagedProperties pq.Int64Array `gorm:"type:integer[]" json:"managedProperties"` // 管理的物业ID列表
	Permissions       pq.Int64Array `gorm:"type:integer[]" json:"permissions"`       // 权限列表
	LastLoginAt       *time.Time    `json:"lastLoginDate"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

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

	Username          string        `gorm:"uniqueIndex;not null;type:varchar(50);comment:用户名" json:"username"`
	Phone             string        `gorm:"uniqueIndex;not null;type:varchar(20);comment:手机号" json:"phone"`
	Role              UserRole      `gorm:"not null;type:varchar(20);comment:角色" json:"role"`
	Status            UserStatus    `gorm:"not null;type:varchar(20);default:'active';comment:状态" json:"status"`
	RealName          string        `gorm:"not null;type:varchar(50);comment:真实姓名" json:"realName"`
	PasswordHash      string        `gorm:"not null;type:varchar(255);comment:密码哈希" json:"-"`
	ManagedProperties pq.Int64Array `gorm:"type:integer[];comment:管理的物业ID列表" json:"managedProperties"`
	Permissions       pq.Int64Array `gorm:"type:integer[];comment:权限列表" json:"permissions"`
	LastLoginAt       *time.Time    `gorm:"comment:最后登录时间" json:"lastLoginDate"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

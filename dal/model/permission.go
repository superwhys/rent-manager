// File:		permission.go
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

type Permission struct {
	ID uint `gorm:"primarykey" json:"id"`

	Key      string `gorm:"not null;type:varchar(64);comment:权限键" json:"key"`
	Name     string `gorm:"not null;type:varchar(64);comment:权限名称" json:"name"`
	Category string `gorm:"not null;type:varchar(64);comment:权限分类" json:"category"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

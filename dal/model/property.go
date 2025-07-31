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

type Property struct {
	ID uint `gorm:"primarykey" json:"id"`

	Name    string `gorm:"not null;type:varchar(50)" json:"name"`
	Address string `gorm:"not null;type:varchar(128)" json:"address"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

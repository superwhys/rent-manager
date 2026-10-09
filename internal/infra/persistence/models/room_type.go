package models

import "time"

type RoomType struct {
	ID        int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp"`
	Name      string     `gorm:"column:name;type:varchar(12);not null;uniqueIndex:roomtype_account_id_name,priority:2"`
	AccountID string     `gorm:"column:account_id;type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;uniqueIndex:roomtype_account_id_name,priority:1"`
}

func (RoomType) TableName() string {
	return "room_types"
}

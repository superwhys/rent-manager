package models

import "time"

type RoomType struct {
	ID        int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp"`
	Name      string     `gorm:"column:name;type:varchar(12);not null;uniqueIndex:roomtype_account_id_name,priority:2"`
	AccountID int        `gorm:"column:account_id;type:bigint;not null;uniqueIndex:roomtype_account_id_name,priority:1"`
	Account   *Account   `gorm:"foreignKey:AccountID;constraint:room_types_accounts_room_types,OnDelete:NO ACTION"`
}

func (RoomType) TableName() string {
	return "room_types"
}

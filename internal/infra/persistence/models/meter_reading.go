package models

import "time"

type MeterReading struct {
	ID              int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;autoCreateTime;<-:create"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;autoUpdateTime"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
	PeriodMonth     string     `gorm:"column:period_month;type:varchar(255);not null;uniqueIndex:meterreading_room_id_meter_item_id_period_month,priority:3;index:meterreading_period_month"`
	PreviousReading *float64   `gorm:"column:previous_reading;type:double"`
	CurrentReading  float64    `gorm:"column:current_reading;type:double;not null"`
	UsageAmount     *float64   `gorm:"column:usage_amount;type:double"`
	ReadingTime     time.Time  `gorm:"column:reading_time;not null"`
	ReaderID        *string    `gorm:"column:reader_id;type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"`
	Source          string     `gorm:"column:source;type:enum('MANUAL','IMPORT','AUTO');not null;default:MANUAL"`
	Remark          *string    `gorm:"column:remark;type:varchar(255)"`
	MeterItemID     int        `gorm:"column:meter_item_id;type:bigint;not null;uniqueIndex:meterreading_room_id_meter_item_id_period_month,priority:2;index:meterreading_meter_item_id"`
	RoomID          int        `gorm:"column:room_id;type:bigint;not null;uniqueIndex:meterreading_room_id_meter_item_id_period_month,priority:1"`
	MeterItem       *MeterItem `gorm:"foreignKey:MeterItemID;constraint:meter_readings_meter_items_meter_readings,OnDelete:NO ACTION"`
	Room            *Room      `gorm:"foreignKey:RoomID;constraint:meter_readings_rooms_meter_readings,OnDelete:NO ACTION"`
}

func (MeterReading) TableName() string {
	return "meter_readings"
}

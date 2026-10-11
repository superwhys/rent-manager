package models

import "time"

type MeterItem struct {
	ID        int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime;<-:create"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	Code      string     `gorm:"column:code;type:varchar(255);not null;uniqueIndex:meteritem_account_id_code,priority:2"`
	Unit      string     `gorm:"column:unit;type:varchar(255);not null"`
	UnitPrice float64    `gorm:"column:unit_price;type:double;not null;default:0"`
	IsSystem  bool       `gorm:"column:is_system;type:boolean;not null;default:false"`
	CalcType  string     `gorm:"column:calc_type;type:enum('DIFFERENCE','DIRECT');not null;default:DIFFERENCE"`
	Remark    *string    `gorm:"column:remark;type:varchar(255)"`
	AccountID string     `gorm:"column:account_id;type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;uniqueIndex:meteritem_account_id_code,priority:1"`
}

func (MeterItem) TableName() string {
	return "meter_items"
}

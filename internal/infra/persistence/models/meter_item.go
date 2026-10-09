package models

import "time"

type MeterItem struct {
	ID        int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp"`
	Code      string     `gorm:"column:code;type:varchar(255);not null;uniqueIndex:meteritem_account_id_code,priority:2"`
	Unit      string     `gorm:"column:unit;type:varchar(255);not null"`
	UnitPrice float64    `gorm:"column:unit_price;type:double;not null;default:0"`
	IsSystem  bool       `gorm:"column:is_system;type:boolean;not null;default:false"`
	CalcType  string     `gorm:"column:calc_type;type:enum('DIFFERENCE','DIRECT');not null;default:DIFFERENCE"`
	Remark    *string    `gorm:"column:remark;type:varchar(255)"`
	AccountID int        `gorm:"column:account_id;type:bigint;not null;uniqueIndex:meteritem_account_id_code,priority:1"`
	Account   *Account   `gorm:"foreignKey:AccountID;constraint:meter_items_accounts_meter_items,OnDelete:NO ACTION"`
}

func (MeterItem) TableName() string {
	return "meter_items"
}

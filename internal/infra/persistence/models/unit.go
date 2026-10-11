package models

import "time"

type Unit struct {
	ID         int          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt  time.Time    `gorm:"column:created_at;not null;autoCreateTime;<-:create"`
	UpdatedAt  time.Time    `gorm:"column:updated_at;not null;autoUpdateTime"`
	DeletedAt  *time.Time   `gorm:"column:deleted_at"`
	Name       string       `gorm:"column:name;type:varchar(255);not null;uniqueIndex:unit_account_id_name,priority:2"`
	Address    string       `gorm:"column:address;type:varchar(255);not null"`
	City       string       `gorm:"column:city;type:varchar(255);not null"`
	Province   string       `gorm:"column:province;type:varchar(255);not null"`
	JSONExtend string       `gorm:"column:json_extend;type:varchar(2048);not null;default:'{\"meter_reading_day\":1}'"`
	AccountID  string       `gorm:"column:account_id;type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;uniqueIndex:unit_account_id_name,priority:1"`
	MeterItems []*MeterItem `gorm:"many2many:unit_meter_items;joinForeignKey:UnitID;joinReferences:MeterItemID;-:migration"`
}

func (Unit) TableName() string {
	return "units"
}

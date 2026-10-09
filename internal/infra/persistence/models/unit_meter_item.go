package models

type UnitMeterItem struct {
	UnitID      int        `gorm:"column:unit_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	MeterItemID int        `gorm:"column:meter_item_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	Unit        *Unit      `gorm:"foreignKey:UnitID;constraint:unit_meter_items_unit_id,OnDelete:CASCADE"`
	MeterItem   *MeterItem `gorm:"foreignKey:MeterItemID;constraint:unit_meter_items_meter_item_id,OnDelete:CASCADE"`
}

func (UnitMeterItem) TableName() string {
	return "unit_meter_items"
}

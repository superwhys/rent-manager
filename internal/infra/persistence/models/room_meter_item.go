package models

type RoomMeterItem struct {
	RoomID      int        `gorm:"column:room_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	MeterItemID int        `gorm:"column:meter_item_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	Room        *Room      `gorm:"foreignKey:RoomID;constraint:room_meter_items_room_id,OnDelete:CASCADE"`
	MeterItem   *MeterItem `gorm:"foreignKey:MeterItemID;constraint:room_meter_items_meter_item_id,OnDelete:CASCADE"`
}

func (RoomMeterItem) TableName() string {
	return "room_meter_items"
}

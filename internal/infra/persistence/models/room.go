package models

import "time"

type Room struct {
	ID         int          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt  time.Time    `gorm:"column:created_at;not null;autoCreateTime;<-:create"`
	UpdatedAt  time.Time    `gorm:"column:updated_at;not null;autoUpdateTime"`
	DeletedAt  *time.Time   `gorm:"column:deleted_at"`
	RoomNo     string       `gorm:"column:room_no;type:varchar(255);not null"`
	Floor      int          `gorm:"column:floor;type:bigint;not null"`
	Price      int          `gorm:"column:price;type:bigint;not null"`
	Size       int          `gorm:"column:size;type:bigint;not null"`
	IsActive   *bool        `gorm:"column:is_active;type:boolean;not null;default:true"`
	RoomTypeID int          `gorm:"column:room_type_id;type:bigint;not null"`
	UnitID     int          `gorm:"column:unit_id;type:bigint;not null"`
	RoomType   *RoomType    `gorm:"foreignKey:RoomTypeID;constraint:rooms_room_types_rooms,OnDelete:NO ACTION"`
	Unit       *Unit        `gorm:"foreignKey:UnitID;constraint:rooms_units_rooms,OnDelete:NO ACTION"`
	MeterItems []*MeterItem `gorm:"many2many:room_meter_items;joinForeignKey:RoomID;joinReferences:MeterItemID;-:migration"`
}

func (Room) TableName() string {
	return "rooms"
}

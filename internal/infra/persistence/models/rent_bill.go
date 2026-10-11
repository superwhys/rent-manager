package models

import "time"

type RentBill struct {
	ID                int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;autoCreateTime;<-:create"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;autoUpdateTime"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
	PeriodMonth       string     `gorm:"column:period_month;type:varchar(255);not null;uniqueIndex:rentbill_room_id_period_month,priority:2;index:rentbill_period_month"`
	BaseRent          int        `gorm:"column:base_rent;type:bigint;not null;default:0"`
	MeterFee          int        `gorm:"column:meter_fee;type:bigint;not null;default:0"`
	MeterReadingsJSON *string    `gorm:"column:meter_readings;type:json"`
	TotalAmount       int        `gorm:"column:total_amount;type:bigint;not null;default:0"`
	Status            string     `gorm:"column:status;type:enum('UNPAID','PAID','CANCELLED');not null;default:UNPAID;index:rentbill_status"`
	PaidAt            *time.Time `gorm:"column:paid_at"`
	Remark            *string    `gorm:"column:remark;type:varchar(255)"`
	RoomID            int        `gorm:"column:room_id;type:bigint;not null;uniqueIndex:rentbill_room_id_period_month,priority:1"`
	Room              *Room      `gorm:"foreignKey:RoomID;constraint:rent_bills_rooms_rent_bills,OnDelete:NO ACTION"`
}

func (RentBill) TableName() string {
	return "rent_bills"
}

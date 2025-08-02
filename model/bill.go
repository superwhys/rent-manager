package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// BillStatus 账单支付状态枚举
type BillStatus string

const (
	BillStatusPaid   BillStatus = "paid"   // 已支付
	BillStatusUnpaid BillStatus = "unpaid" // 未支付
)

// Bill 账单模型
type Bill struct {
	ID         uint `gorm:"primaryKey" json:"-"`
	RoomID     uint `gorm:"not null;index;comment:房间ID" json:"room_id"`
	TenantID   uint `gorm:"not null;index;comment:租户ID" json:"tenant_id"`
	PropertyID uint `gorm:"not null;index;comment:物业ID" json:"property_id"`
	MeteringID uint `gorm:"not null;index;comment:抄表ID" json:"metering_id"`

	BillID          string          `gorm:"uniqueIndex;not null;size:50;comment:账单编号" json:"id"`
	Room            string          `gorm:"index;not null;size:20;comment:房间号" json:"room"`
	Tenant          string          `gorm:"index;not null;size:100;comment:租客姓名" json:"tenant"`
	Rent            decimal.Decimal `gorm:"type:decimal(10,2);not null;comment:租金" json:"rent"`
	WaterFee        decimal.Decimal `gorm:"type:decimal(10,2);not null;default:0;comment:水费" json:"waterFee"`
	ElectricFee     decimal.Decimal `gorm:"type:decimal(10,2);not null;default:0;comment:电费" json:"electricFee"`
	GasFee          decimal.Decimal `gorm:"type:decimal(10,2);not null;default:0;comment:燃气费" json:"gasFee"`
	OtherFee        decimal.Decimal `gorm:"type:decimal(10,2);not null;default:0;comment:其他费用" json:"otherFee"`
	Amount          decimal.Decimal `gorm:"type:decimal(10,2);not null;comment:总金额" json:"amount"`
	Status          BillStatus      `gorm:"type:varchar(20);not null;default:'unpaid';index;comment:支付状态" json:"status"`
	CreateDate      time.Time       `gorm:"not null;index;comment:账单生成日期" json:"createDate"`
	UtilitySettings string          `gorm:"type:json;comment:备份当前账单的水电燃气费用设置" json:"utility_settings"`

	CreatedAt time.Time      `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// BeforeCreate GORM钩子：创建前自动计算总金额
func (b *Bill) BeforeCreate(tx *gorm.DB) error {
	b.Amount = b.Rent.Add(b.WaterFee).Add(b.ElectricFee).Add(b.GasFee).Add(b.OtherFee)
	return nil
}

// BeforeUpdate GORM钩子：更新前自动计算总金额
func (b *Bill) BeforeUpdate(tx *gorm.DB) error {
	b.Amount = b.Rent.Add(b.WaterFee).Add(b.ElectricFee).Add(b.GasFee).Add(b.OtherFee)
	return nil
}

// IsPaid 检查账单是否已支付
func (b *Bill) IsPaid() bool {
	return b.Status == BillStatusPaid
}

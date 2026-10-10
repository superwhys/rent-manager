package rentbill

import (
	"time"

	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
)

// MeterReadingSnapshot 保存生成账单时的抄表项、读数和费用，不随原始数据变化。
type MeterReadingSnapshot struct {
	ID              int                        `json:"id"`
	RoomID          int                        `json:"room_id"`
	PeriodMonth     string                     `json:"period_month"`
	MeterItem       *domainmeteritem.MeterItem `json:"meter_item"`
	PreviousReading *float64                   `json:"previous_reading"`
	CurrentReading  float64                    `json:"current_reading"`
	UsageAmount     *float64                   `json:"usage_amount"`
	ReadingTime     time.Time                  `json:"reading_time"`
	ReaderID        *string                    `json:"reader_id"`
	Source          string                     `json:"source"`
	Remark          *string                    `json:"remark"`
	Fee             int                        `json:"fee"`
	IsRecorded      bool                       `json:"is_recorded"`
}

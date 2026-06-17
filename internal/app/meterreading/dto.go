package meterreading

import (
	"time"

	appmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/app/meteritem"
	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
	domainmeterreading "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meterreading"
	domainroom "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/room"
)

// RoomMeterReadingsQuery 是房间月度抄表查询参数。
type RoomMeterReadingsQuery struct {
	RoomID      int    `uri:"room_id" validate:"required,min=1"`
	PeriodMonth string `form:"period_month" validate:"required"`
}

// UnitMeterReadingsQuery 是单元月度抄表汇总查询参数。
type UnitMeterReadingsQuery struct {
	UnitID      int    `uri:"unit_id" validate:"required,min=1"`
	PeriodMonth string `form:"period_month" validate:"required"`
}

// SaveRoomMonthCommand 是保存房间月度抄表的请求参数。
type SaveRoomMonthCommand struct {
	RoomID      int                 `uri:"room_id" validate:"required,min=1"`
	PeriodMonth string              `json:"period_month" validate:"required"`
	Readings    []MeterReadingInput `json:"readings" validate:"required,min=1,dive"`
}

// MeterReadingInput 是单个抄表项的录入内容。
type MeterReadingInput struct {
	MeterItemID     int        `json:"meter_item_id" validate:"required,min=1"`
	PreviousReading *float64   `json:"previous_reading" validate:"omitempty,min=0"`
	CurrentReading  float64    `json:"current_reading" validate:"min=0"`
	ReadingTime     *time.Time `json:"reading_time"`
	Remark          *string    `json:"remark"`
}

// MeterReadingLine 表示房间某个月一个抄表项的录入状态。
type MeterReadingLine struct {
	RoomID          int
	PeriodMonth     string
	MeterItem       *domainmeteritem.MeterItem
	Reading         *domainmeterreading.MeterReading
	PreviousReading *float64
	Fee             int
	IsRecorded      bool
}

// RoomMonthReadings 表示房间某个月的抄表视图。
type RoomMonthReadings struct {
	Room         *domainroom.Room
	PeriodMonth  string
	Readings     []*MeterReadingLine
	TotalFee     int
	MissingCount int
}

// MeterItemSummary 表示单元某个月按抄表项汇总的用量和费用。
type MeterItemSummary struct {
	MeterItem   *domainmeteritem.MeterItem
	UsageAmount float64
	Fee         int
	Recorded    int
	Missing     int
}

// UnitMonthReadings 表示单元某个月的抄表汇总。
type UnitMonthReadings struct {
	UnitID       int
	PeriodMonth  string
	Rooms        []*RoomMonthReadings
	Summaries    []*MeterItemSummary
	TotalFee     int
	MissingCount int
}

// MeterReadingLineResponse 是单个抄表项的响应。
type MeterReadingLineResponse struct {
	ID              int                            `json:"id,omitempty"`
	RoomID          int                            `json:"room_id"`
	PeriodMonth     string                         `json:"period_month"`
	MeterItem       appmeteritem.MeterItemResponse `json:"meter_item"`
	PreviousReading *float64                       `json:"previous_reading,omitempty"`
	CurrentReading  *float64                       `json:"current_reading,omitempty"`
	UsageAmount     *float64                       `json:"usage_amount,omitempty"`
	ReadingTime     *time.Time                     `json:"reading_time,omitempty"`
	ReaderID        *int64                         `json:"reader_id,omitempty"`
	Source          string                         `json:"source,omitempty"`
	Remark          *string                        `json:"remark,omitempty"`
	Fee             int                            `json:"fee"`
	IsRecorded      bool                           `json:"is_recorded"`
}

// RoomMonthReadingsResponse 是房间月度抄表响应。
type RoomMonthReadingsResponse struct {
	RoomID       int                        `json:"room_id"`
	RoomNo       string                     `json:"room_no"`
	UnitID       int                        `json:"unit_id"`
	PeriodMonth  string                     `json:"period_month"`
	Readings     []MeterReadingLineResponse `json:"readings"`
	TotalFee     int                        `json:"total_fee"`
	MissingCount int                        `json:"missing_count"`
}

// MeterItemSummaryResponse 是单元抄表项汇总响应。
type MeterItemSummaryResponse struct {
	MeterItemID int     `json:"meter_item_id"`
	Code        string  `json:"code"`
	Unit        string  `json:"unit"`
	CalcType    string  `json:"calc_type"`
	UnitPrice   int     `json:"unit_price"`
	UsageAmount float64 `json:"usage_amount"`
	Fee         int     `json:"fee"`
	Recorded    int     `json:"recorded"`
	Missing     int     `json:"missing"`
}

// UnitMonthReadingsResponse 是单元月度抄表汇总响应。
type UnitMonthReadingsResponse struct {
	UnitID       int                         `json:"unit_id"`
	PeriodMonth  string                      `json:"period_month"`
	Rooms        []RoomMonthReadingsResponse `json:"rooms"`
	Summaries    []MeterItemSummaryResponse  `json:"summaries"`
	TotalFee     int                         `json:"total_fee"`
	MissingCount int                         `json:"missing_count"`
}

// ToRoomMonthReadingsResponse 将房间月度抄表视图转换为响应。
func ToRoomMonthReadingsResponse(roomMonth *RoomMonthReadings) RoomMonthReadingsResponse {
	if roomMonth == nil {
		return RoomMonthReadingsResponse{}
	}
	response := RoomMonthReadingsResponse{
		PeriodMonth:  roomMonth.PeriodMonth,
		TotalFee:     roomMonth.TotalFee,
		MissingCount: roomMonth.MissingCount,
		Readings:     make([]MeterReadingLineResponse, 0, len(roomMonth.Readings)),
	}
	if roomMonth.Room != nil {
		response.RoomID = roomMonth.Room.ID
		response.RoomNo = roomMonth.Room.RoomNo
		response.UnitID = roomMonth.Room.UnitID
	}
	for _, line := range roomMonth.Readings {
		response.Readings = append(response.Readings, ToMeterReadingLineResponse(line))
	}
	return response
}

// ToMeterReadingLineResponse 将抄表项录入状态转换为响应。
func ToMeterReadingLineResponse(line *MeterReadingLine) MeterReadingLineResponse {
	if line == nil {
		return MeterReadingLineResponse{}
	}
	response := MeterReadingLineResponse{
		RoomID:          line.RoomID,
		PeriodMonth:     line.PeriodMonth,
		MeterItem:       appmeteritem.ToMeterItemResponse(line.MeterItem),
		PreviousReading: line.PreviousReading,
		Fee:             line.Fee,
		IsRecorded:      line.IsRecorded,
	}
	if line.Reading != nil {
		currentReading := line.Reading.CurrentReading
		response.ID = line.Reading.ID
		response.CurrentReading = &currentReading
		response.UsageAmount = line.Reading.UsageAmount
		response.ReadingTime = &line.Reading.ReadingTime
		response.ReaderID = line.Reading.ReaderID
		response.Source = line.Reading.Source
		response.Remark = line.Reading.Remark
	}
	return response
}

// ToUnitMonthReadingsResponse 将单元月度抄表汇总转换为响应。
func ToUnitMonthReadingsResponse(unitMonth *UnitMonthReadings) UnitMonthReadingsResponse {
	if unitMonth == nil {
		return UnitMonthReadingsResponse{}
	}
	response := UnitMonthReadingsResponse{
		UnitID:       unitMonth.UnitID,
		PeriodMonth:  unitMonth.PeriodMonth,
		Rooms:        make([]RoomMonthReadingsResponse, 0, len(unitMonth.Rooms)),
		Summaries:    make([]MeterItemSummaryResponse, 0, len(unitMonth.Summaries)),
		TotalFee:     unitMonth.TotalFee,
		MissingCount: unitMonth.MissingCount,
	}
	for _, room := range unitMonth.Rooms {
		response.Rooms = append(response.Rooms, ToRoomMonthReadingsResponse(room))
	}
	for _, summary := range unitMonth.Summaries {
		response.Summaries = append(response.Summaries, ToMeterItemSummaryResponse(summary))
	}
	return response
}

// ToMeterItemSummaryResponse 将抄表项汇总转换为响应。
func ToMeterItemSummaryResponse(summary *MeterItemSummary) MeterItemSummaryResponse {
	if summary == nil || summary.MeterItem == nil {
		return MeterItemSummaryResponse{}
	}
	return MeterItemSummaryResponse{
		MeterItemID: summary.MeterItem.ID,
		Code:        summary.MeterItem.Code,
		Unit:        summary.MeterItem.Unit,
		CalcType:    summary.MeterItem.CalcType,
		UnitPrice:   summary.MeterItem.UnitPrice,
		UsageAmount: summary.UsageAmount,
		Fee:         summary.Fee,
		Recorded:    summary.Recorded,
		Missing:     summary.Missing,
	}
}

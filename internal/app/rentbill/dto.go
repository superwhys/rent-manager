package rentbill

import (
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	appmeteritem "github.com/superwhys/rent-manager/internal/app/meteritem"
	appmeterreading "github.com/superwhys/rent-manager/internal/app/meterreading"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
)

// RoomRentBillsQuery 是房间账单列表查询参数。
type RoomRentBillsQuery struct {
	appcore.Pagination

	RoomID      int    `uri:"room_id" validate:"required,min=1"`
	PeriodMonth string `                                                               form:"period_month"`
	Status      string `              validate:"omitempty,oneof=UNPAID PAID CANCELLED" form:"status"`
}

// UnitRentBillsQuery 是单元账单列表查询参数。
type UnitRentBillsQuery struct {
	appcore.Pagination

	UnitID      int    `uri:"unit_id" validate:"required,min=1"`
	PeriodMonth string `                                                               form:"period_month"`
	Status      string `              validate:"omitempty,oneof=UNPAID PAID CANCELLED" form:"status"`
}

// RentBillIDQuery 是账单 ID 路径参数。
type RentBillIDQuery struct {
	RentBillID int `uri:"rent_bill_id" validate:"required,min=1"`
}

// GenerateRoomRentBillCommand 是生成房间月度账单的请求参数。
type GenerateRoomRentBillCommand struct {
	RoomID      int     `uri:"room_id" validate:"required,min=1"`
	PeriodMonth string  `              validate:"required"       json:"period_month"`
	Remark      *string `                                        json:"remark"`
}

// GenerateUnitRentBillsCommand 是生成单元月度账单的请求参数。
type GenerateUnitRentBillsCommand struct {
	UnitID      int     `uri:"unit_id" validate:"required,min=1"`
	PeriodMonth string  `              validate:"required"       json:"period_month"`
	Remark      *string `                                        json:"remark"`
}

// MarkRentBillPaidCommand 是标记账单已收款的请求参数。
type MarkRentBillPaidCommand struct {
	RentBillID int        `uri:"rent_bill_id" validate:"required,min=1"`
	PaidAt     *time.Time `                                             json:"paid_at"`
}

// CancelRentBillCommand 是取消账单的请求参数。
type CancelRentBillCommand struct {
	RentBillID int     `uri:"rent_bill_id" validate:"required,min=1"`
	Remark     *string `                                             json:"remark"`
}

// GeneratedRentBills 表示生成账单的结果。
type GeneratedRentBills struct {
	Saved   []*domainrentbill.RentBill
	Skipped []*domainrentbill.RentBill
}

// RentBillResponse 是租金账单响应。
type RentBillResponse struct {
	ID          int        `json:"id"`
	RoomID      int        `json:"room_id"`
	PeriodMonth string     `json:"period_month"`
	BaseRent    int        `json:"base_rent"`
	MeterFee    int        `json:"meter_fee"`
	TotalAmount int        `json:"total_amount"`
	Status      string     `json:"status"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	Remark      *string    `json:"remark,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RentBillDetailResponse 是账单详情，抄表项为开单时保存的快照；历史无快照账单返回 null。
type RentBillDetailResponse struct {
	RentBillResponse
	MeterReadings []appmeterreading.MeterReadingLineResponse `json:"meter_readings"`
}

// ToRentBillDetailResponse 将账单及其冻结抄表明细转换为响应。
func ToRentBillDetailResponse(bill *domainrentbill.RentBill) RentBillDetailResponse {
	response := RentBillDetailResponse{RentBillResponse: ToRentBillResponse(bill)}
	if bill.MeterReadings == nil {
		return response
	}
	response.MeterReadings = make([]appmeterreading.MeterReadingLineResponse, 0, len(bill.MeterReadings))
	for _, snapshot := range bill.MeterReadings {
		current := snapshot.CurrentReading
		readingTime := snapshot.ReadingTime
		response.MeterReadings = append(response.MeterReadings, appmeterreading.MeterReadingLineResponse{
			ID:              snapshot.ID,
			RoomID:          snapshot.RoomID,
			PeriodMonth:     snapshot.PeriodMonth,
			MeterItem:       appmeteritem.ToMeterItemResponse(snapshot.MeterItem),
			PreviousReading: snapshot.PreviousReading,
			CurrentReading:  &current,
			UsageAmount:     snapshot.UsageAmount,
			ReadingTime:     &readingTime,
			ReaderID:        snapshot.ReaderID,
			Source:          snapshot.Source,
			Remark:          snapshot.Remark,
			Fee:             snapshot.Fee,
			IsRecorded:      snapshot.IsRecorded,
		})
	}
	return response
}

// GeneratedRentBillsResponse 是账单生成响应。
type GeneratedRentBillsResponse struct {
	SavedBills   []RentBillResponse `json:"saved_bills"`
	SkippedBills []RentBillResponse `json:"skipped_bills"`
}

// ToRentBillResponse 将账单领域对象转换为响应。
func ToRentBillResponse(bill *domainrentbill.RentBill) RentBillResponse {
	return RentBillResponse{
		ID:          bill.ID,
		RoomID:      bill.RoomID,
		PeriodMonth: bill.PeriodMonth,
		BaseRent:    bill.BaseRent,
		MeterFee:    bill.MeterFee,
		TotalAmount: bill.TotalAmount,
		Status:      bill.Status,
		PaidAt:      bill.PaidAt,
		Remark:      bill.Remark,
		CreatedAt:   bill.CreatedAt,
		UpdatedAt:   bill.UpdatedAt,
	}
}

// ToRentBillListResponse 将账单列表转换为响应。
func ToRentBillListResponse(bills []*domainrentbill.RentBill) []RentBillResponse {
	responses := make([]RentBillResponse, 0, len(bills))
	for _, bill := range bills {
		responses = append(responses, ToRentBillResponse(bill))
	}
	return responses
}

// ToGeneratedRentBillsResponse 将生成结果转换为响应。
func ToGeneratedRentBillsResponse(result *GeneratedRentBills) GeneratedRentBillsResponse {
	return GeneratedRentBillsResponse{
		SavedBills:   ToRentBillListResponse(result.Saved),
		SkippedBills: ToRentBillListResponse(result.Skipped),
	}
}

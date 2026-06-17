package rentbill

import (
	"time"

	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	domainrentbill "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rentbill"
)

// RoomRentBillsQuery 是房间账单列表查询参数。
type RoomRentBillsQuery struct {
	appcore.Pagination

	RoomID      int    `uri:"room_id" validate:"required,min=1"`
	PeriodMonth string `form:"period_month"`
	Status      string `form:"status" validate:"omitempty,oneof=UNPAID PAID CANCELLED"`
}

// UnitRentBillsQuery 是单元账单列表查询参数。
type UnitRentBillsQuery struct {
	appcore.Pagination

	UnitID      int    `uri:"unit_id" validate:"required,min=1"`
	PeriodMonth string `form:"period_month"`
	Status      string `form:"status" validate:"omitempty,oneof=UNPAID PAID CANCELLED"`
}

// RentBillIDQuery 是账单 ID 路径参数。
type RentBillIDQuery struct {
	RentBillID int `uri:"rent_bill_id" validate:"required,min=1"`
}

// GenerateRoomRentBillCommand 是生成房间月度账单的请求参数。
type GenerateRoomRentBillCommand struct {
	RoomID      int     `uri:"room_id" validate:"required,min=1"`
	PeriodMonth string  `json:"period_month" validate:"required"`
	Remark      *string `json:"remark"`
}

// GenerateUnitRentBillsCommand 是生成单元月度账单的请求参数。
type GenerateUnitRentBillsCommand struct {
	UnitID      int     `uri:"unit_id" validate:"required,min=1"`
	PeriodMonth string  `json:"period_month" validate:"required"`
	Remark      *string `json:"remark"`
}

// MarkRentBillPaidCommand 是标记账单已收款的请求参数。
type MarkRentBillPaidCommand struct {
	RentBillID int        `uri:"rent_bill_id" validate:"required,min=1"`
	PaidAt     *time.Time `json:"paid_at"`
}

// CancelRentBillCommand 是取消账单的请求参数。
type CancelRentBillCommand struct {
	RentBillID int     `uri:"rent_bill_id" validate:"required,min=1"`
	Remark     *string `json:"remark"`
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

// GeneratedRentBillsResponse 是账单生成响应。
type GeneratedRentBillsResponse struct {
	SavedBills   []RentBillResponse `json:"saved_bills"`
	SkippedBills []RentBillResponse `json:"skipped_bills"`
}

// ToRentBillResponse 将账单领域对象转换为响应。
func ToRentBillResponse(bill *domainrentbill.RentBill) RentBillResponse {
	if bill == nil {
		return RentBillResponse{}
	}
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
	if result == nil {
		return GeneratedRentBillsResponse{}
	}
	return GeneratedRentBillsResponse{
		SavedBills:   ToRentBillListResponse(result.Saved),
		SkippedBills: ToRentBillListResponse(result.Skipped),
	}
}

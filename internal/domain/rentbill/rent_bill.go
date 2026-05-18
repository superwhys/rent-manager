package rentbill

import (
	"strings"
	"time"

	"github.com/superwhys/rent-manager-server/internal/models/consts"
)

type RentBill struct {
	ID          int
	RoomID      int
	PeriodMonth string
	BaseRent    int
	MeterFee    int
	TotalAmount int
	Status      string
	PaidAt      *time.Time
	Remark      *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func NewRentBill(params CreateParams) (*RentBill, error) {
	bill := &RentBill{
		RoomID:      params.RoomID,
		PeriodMonth: strings.TrimSpace(params.PeriodMonth),
		BaseRent:    params.BaseRent,
		MeterFee:    params.MeterFee,
		Status:      normalizeStatus(params.Status),
		PaidAt:      params.PaidAt,
		Remark:      normalizeRemark(params.Remark),
	}
	bill.CalculateTotal()
	if err := bill.Validate(); err != nil {
		return nil, err
	}
	return bill, nil
}

func GenerateRentBill(params GenerateParams) (*RentBill, error) {
	return NewRentBill(CreateParams{
		RoomID:      params.RoomID,
		PeriodMonth: params.PeriodMonth,
		BaseRent:    params.BaseRent,
		MeterFee:    params.MeterFee,
		Status:      consts.RentBillStatusUnpaid,
		Remark:      params.Remark,
	})
}

func (b *RentBill) Update(params UpdateParams) error {
	b.ID = params.ID
	b.RoomID = params.RoomID
	b.PeriodMonth = strings.TrimSpace(params.PeriodMonth)
	b.BaseRent = params.BaseRent
	b.MeterFee = params.MeterFee
	b.Status = normalizeStatus(params.Status)
	b.PaidAt = params.PaidAt
	b.Remark = normalizeRemark(params.Remark)
	b.CalculateTotal()
	return b.Validate()
}

func (b *RentBill) Validate() error {
	if b.RoomID <= 0 {
		return ErrRoomRequired
	}
	if b.PeriodMonth == "" {
		return ErrPeriodMonthRequired
	}
	if !isValidPeriodMonth(b.PeriodMonth) {
		return ErrInvalidPeriodMonth
	}
	if b.BaseRent < 0 {
		return ErrInvalidBaseRent
	}
	if b.MeterFee < 0 {
		return ErrInvalidMeterFee
	}
	if b.TotalAmount < 0 {
		return ErrInvalidTotalAmount
	}
	switch b.Status {
	case consts.RentBillStatusUnpaid, consts.RentBillStatusPaid, consts.RentBillStatusCancelled:
	default:
		return ErrInvalidStatus
	}
	if b.Status == consts.RentBillStatusPaid && b.PaidAt == nil {
		return ErrPaidAtRequired
	}
	return nil
}

func (b *RentBill) CalculateTotal() {
	b.TotalAmount = b.BaseRent + b.MeterFee
}

func (b *RentBill) MarkPaid(paidAt time.Time) error {
	if b.Status == consts.RentBillStatusCancelled {
		return ErrCancelled
	}
	if b.Status == consts.RentBillStatusPaid {
		return ErrAlreadyPaid
	}
	if paidAt.IsZero() {
		return ErrPaidAtRequired
	}
	b.Status = consts.RentBillStatusPaid
	b.PaidAt = &paidAt
	return nil
}

func (b *RentBill) Cancel() error {
	if b.Status == consts.RentBillStatusPaid {
		return ErrAlreadyPaid
	}
	b.Status = consts.RentBillStatusCancelled
	return nil
}

func isValidPeriodMonth(periodMonth string) bool {
	_, err := time.Parse("2006-01", periodMonth)
	return err == nil
}

func normalizeStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return consts.RentBillStatusUnpaid
	}
	return status
}

func normalizeRemark(remark *string) *string {
	if remark == nil {
		return nil
	}
	value := strings.TrimSpace(*remark)
	if value == "" {
		return nil
	}
	return &value
}

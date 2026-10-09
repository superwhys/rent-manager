package meteritem

import (
	"strings"
	"time"
)

type MeterItem struct {
	ID        int
	AccountID string
	Code      string
	Unit      string
	UnitPrice float64
	IsSystem  bool
	CalcType  string
	Remark    *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewMeterItem(params CreateParams) (*MeterItem, error) {
	item := &MeterItem{
		AccountID: params.AccountID,
		Code:      strings.TrimSpace(params.Code),
		Unit:      strings.TrimSpace(params.Unit),
		UnitPrice: params.UnitPrice,
		CalcType:  normalizeCalcType(params.CalcType),
		Remark:    normalizeRemark(params.Remark),
	}
	if err := item.Validate(); err != nil {
		return nil, err
	}
	return item, nil
}

func (i *MeterItem) Update(params UpdateParams) error {
	i.ID = params.ID
	i.AccountID = params.AccountID
	i.Code = strings.TrimSpace(params.Code)
	i.Unit = strings.TrimSpace(params.Unit)
	i.UnitPrice = params.UnitPrice
	i.CalcType = normalizeCalcType(params.CalcType)
	i.Remark = normalizeRemark(params.Remark)
	return i.Validate()
}

func (i *MeterItem) Validate() error {
	if strings.TrimSpace(i.AccountID) == "" {
		return ErrAccountRequired
	}
	if i.Code == "" {
		return ErrCodeRequired
	}
	if i.Unit == "" {
		return ErrUnitRequired
	}
	if i.UnitPrice < 0 {
		return ErrInvalidUnitPrice
	}
	switch i.CalcType {
	case CalcTypeDifference, CalcTypeDirect:
		return nil
	default:
		return ErrInvalidCalcType
	}
}

func normalizeCalcType(calcType string) string {
	calcType = strings.TrimSpace(calcType)
	if calcType == "" {
		return CalcTypeDifference
	}
	return calcType
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

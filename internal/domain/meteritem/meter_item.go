package meteritem

import (
	"strings"
	"time"

	"github.com/superwhys/rent-manager-server/internal/models/consts"
)

type MeterItem struct {
	ID             int
	OrganizationID int
	Code           string
	Unit           string
	IsSystem       bool
	CalcType       string
	Remark         *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewMeterItem(params CreateParams) (*MeterItem, error) {
	item := &MeterItem{
		OrganizationID: params.OrganizationID,
		Code:           strings.TrimSpace(params.Code),
		Unit:           strings.TrimSpace(params.Unit),
		CalcType:       normalizeCalcType(params.CalcType),
		Remark:         normalizeRemark(params.Remark),
	}
	if err := item.Validate(); err != nil {
		return nil, err
	}
	return item, nil
}

func (i *MeterItem) Update(params UpdateParams) error {
	i.ID = params.ID
	i.OrganizationID = params.OrganizationID
	i.Code = strings.TrimSpace(params.Code)
	i.Unit = strings.TrimSpace(params.Unit)
	i.CalcType = normalizeCalcType(params.CalcType)
	i.Remark = normalizeRemark(params.Remark)
	return i.Validate()
}

func (i *MeterItem) Validate() error {
	if i.OrganizationID <= 0 {
		return ErrOrganizationRequired
	}
	if i.Code == "" {
		return ErrCodeRequired
	}
	if i.Unit == "" {
		return ErrUnitRequired
	}
	switch i.CalcType {
	case consts.CalcTypeDifference, consts.CalcTypeDirect:
		return nil
	default:
		return ErrInvalidCalcType
	}
}

func normalizeCalcType(calcType string) string {
	calcType = strings.TrimSpace(calcType)
	if calcType == "" {
		return consts.CalcTypeDifference
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

package meterreading

import (
	"math"
	"strings"
	"time"

	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
)

type MeterReading struct {
	ID              int
	RoomID          int
	MeterItemID     int
	PeriodMonth     string
	PreviousReading *float64
	CurrentReading  float64
	UsageAmount     *float64
	ReadingTime     time.Time
	ReaderID        *int64
	Source          string
	Remark          *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewMeterReading(params CreateParams) (*MeterReading, error) {
	reading := &MeterReading{
		RoomID:          params.RoomID,
		MeterItemID:     params.MeterItemID,
		PeriodMonth:     strings.TrimSpace(params.PeriodMonth),
		PreviousReading: params.PreviousReading,
		CurrentReading:  params.CurrentReading,
		UsageAmount:     params.UsageAmount,
		ReadingTime:     params.ReadingTime,
		ReaderID:        params.ReaderID,
		Source:          normalizeSource(params.Source),
		Remark:          normalizeRemark(params.Remark),
	}
	if err := reading.Validate(); err != nil {
		return nil, err
	}
	return reading, nil
}

func (r *MeterReading) Update(params UpdateParams) error {
	r.ID = params.ID
	r.RoomID = params.RoomID
	r.MeterItemID = params.MeterItemID
	r.PeriodMonth = strings.TrimSpace(params.PeriodMonth)
	r.PreviousReading = params.PreviousReading
	r.CurrentReading = params.CurrentReading
	r.UsageAmount = params.UsageAmount
	r.ReadingTime = params.ReadingTime
	r.ReaderID = params.ReaderID
	r.Source = normalizeSource(params.Source)
	r.Remark = normalizeRemark(params.Remark)
	return r.Validate()
}

func (r *MeterReading) Validate() error {
	if r.RoomID <= 0 {
		return ErrRoomRequired
	}
	if r.MeterItemID <= 0 {
		return ErrMeterItemRequired
	}
	if r.PeriodMonth == "" {
		return ErrPeriodMonthRequired
	}
	if !isValidPeriodMonth(r.PeriodMonth) {
		return ErrInvalidPeriodMonth
	}
	if r.PreviousReading != nil && *r.PreviousReading < 0 {
		return ErrInvalidPreviousReading
	}
	if r.CurrentReading < 0 {
		return ErrInvalidCurrentReading
	}
	if r.UsageAmount != nil && *r.UsageAmount < 0 {
		return ErrInvalidUsageAmount
	}
	if r.ReadingTime.IsZero() {
		return ErrReadingTimeRequired
	}
	switch r.Source {
	case SourceManual, SourceImport, SourceAuto:
		return nil
	default:
		return ErrInvalidSource
	}
}

func (r *MeterReading) CalculateUsage(calcType string) error {
	switch normalizeCalcType(calcType) {
	case domainmeteritem.CalcTypeDirect:
		usage := r.CurrentReading
		r.UsageAmount = &usage
		return nil
	case domainmeteritem.CalcTypeDifference:
		if r.PreviousReading == nil {
			return ErrPreviousReadingRequired
		}
		if r.CurrentReading < *r.PreviousReading {
			return ErrCurrentReadingDecreased
		}
		usage := r.CurrentReading - *r.PreviousReading
		r.UsageAmount = &usage
		return nil
	default:
		return ErrInvalidCalcType
	}
}

func (r *MeterReading) CalculateFee(unitPrice int) (int, error) {
	if unitPrice < 0 {
		return 0, ErrInvalidUnitPrice
	}
	if r.UsageAmount == nil {
		return 0, ErrInvalidUsageAmount
	}
	return int(math.Round(*r.UsageAmount * float64(unitPrice))), nil
}

func isValidPeriodMonth(periodMonth string) bool {
	_, err := time.Parse("2006-01", periodMonth)
	return err == nil
}

func normalizeSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return SourceManual
	}
	return source
}

func normalizeCalcType(calcType string) string {
	calcType = strings.TrimSpace(calcType)
	if calcType == "" {
		return domainmeteritem.CalcTypeDifference
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

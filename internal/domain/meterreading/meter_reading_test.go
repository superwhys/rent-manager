package meterreading

import (
	"errors"
	"testing"
	"time"

	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
)

func TestMeterReadingCalculateDifferenceUsage(t *testing.T) {
	previous := 100.0
	reading, err := NewMeterReading(CreateParams{
		RoomID:          1,
		MeterItemID:     2,
		PeriodMonth:     "2026-06",
		PreviousReading: &previous,
		CurrentReading:  126.5,
		ReadingTime:     time.Now(),
	})
	if err != nil {
		t.Fatalf("NewMeterReading() error = %v", err)
	}

	if err := reading.CalculateUsage(domainmeteritem.CalcTypeDifference); err != nil {
		t.Fatalf("CalculateUsage() error = %v", err)
	}
	if reading.UsageAmount == nil || *reading.UsageAmount != 26.5 {
		t.Fatalf("usage amount = %v, want 26.5", reading.UsageAmount)
	}
	fee, err := reading.CalculateFee(100)
	if err != nil {
		t.Fatalf("CalculateFee() error = %v", err)
	}
	if fee != 2650 {
		t.Fatalf("fee = %d, want 2650", fee)
	}
}

func TestMeterReadingCalculateDirectUsage(t *testing.T) {
	reading, err := NewMeterReading(CreateParams{
		RoomID:         1,
		MeterItemID:    2,
		PeriodMonth:    "2026-06",
		CurrentReading: 12.5,
		ReadingTime:    time.Now(),
	})
	if err != nil {
		t.Fatalf("NewMeterReading() error = %v", err)
	}

	if err := reading.CalculateUsage(domainmeteritem.CalcTypeDirect); err != nil {
		t.Fatalf("CalculateUsage() error = %v", err)
	}
	if reading.UsageAmount == nil || *reading.UsageAmount != 12.5 {
		t.Fatalf("usage amount = %v, want 12.5", reading.UsageAmount)
	}
	fee, err := reading.CalculateFee(123)
	if err != nil {
		t.Fatalf("CalculateFee() error = %v", err)
	}
	if fee != 1538 {
		t.Fatalf("fee = %d, want 1538", fee)
	}
}

func TestMeterReadingRejectInvalidPeriodMonth(t *testing.T) {
	_, err := NewMeterReading(CreateParams{
		RoomID:         1,
		MeterItemID:    2,
		PeriodMonth:    "2026-13",
		CurrentReading: 10,
		ReadingTime:    time.Now(),
	})
	if !errors.Is(err, ErrInvalidPeriodMonth) {
		t.Fatalf("error = %v, want ErrInvalidPeriodMonth", err)
	}
}

func TestDifferenceReadingRequiresPreviousReading(t *testing.T) {
	reading, err := NewMeterReading(CreateParams{
		RoomID:         1,
		MeterItemID:    2,
		PeriodMonth:    "2026-06",
		CurrentReading: 10,
		ReadingTime:    time.Now(),
	})
	if err != nil {
		t.Fatalf("NewMeterReading() error = %v", err)
	}

	err = reading.CalculateUsage(domainmeteritem.CalcTypeDifference)
	if !errors.Is(err, ErrPreviousReadingRequired) {
		t.Fatalf("error = %v, want ErrPreviousReadingRequired", err)
	}
}

func TestDifferenceReadingRejectsDecreasedCurrentReading(t *testing.T) {
	previous := 20.0
	reading, err := NewMeterReading(CreateParams{
		RoomID:          1,
		MeterItemID:     2,
		PeriodMonth:     "2026-06",
		PreviousReading: &previous,
		CurrentReading:  10,
		ReadingTime:     time.Now(),
	})
	if err != nil {
		t.Fatalf("NewMeterReading() error = %v", err)
	}

	err = reading.CalculateUsage(domainmeteritem.CalcTypeDifference)
	if !errors.Is(err, ErrCurrentReadingDecreased) {
		t.Fatalf("error = %v, want ErrCurrentReadingDecreased", err)
	}
}

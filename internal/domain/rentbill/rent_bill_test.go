package rentbill

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateRentBillCalculatesTotal(t *testing.T) {
	bill, err := GenerateRentBill(GenerateParams{
		RoomID:      1,
		PeriodMonth: "2026-06",
		BaseRent:    100000,
		MeterFee:    1234,
	})
	if err != nil {
		t.Fatalf("GenerateRentBill() error = %v", err)
	}
	if bill.TotalAmount != 101234 {
		t.Fatalf("total amount = %d, want 101234", bill.TotalAmount)
	}
	if bill.Status != StatusUnpaid {
		t.Fatalf("status = %q, want %q", bill.Status, StatusUnpaid)
	}
}

func TestRentBillMarkPaid(t *testing.T) {
	bill, err := GenerateRentBill(GenerateParams{
		RoomID:      1,
		PeriodMonth: "2026-06",
		BaseRent:    100000,
	})
	if err != nil {
		t.Fatalf("GenerateRentBill() error = %v", err)
	}
	paidAt := time.Now()
	if err := bill.MarkPaid(paidAt); err != nil {
		t.Fatalf("MarkPaid() error = %v", err)
	}
	if bill.Status != StatusPaid {
		t.Fatalf("status = %q, want %q", bill.Status, StatusPaid)
	}
	if bill.PaidAt == nil || !bill.PaidAt.Equal(paidAt) {
		t.Fatalf("paid at = %v, want %v", bill.PaidAt, paidAt)
	}
}

func TestRentBillCancelRejectsPaidBill(t *testing.T) {
	paidAt := time.Now()
	bill, err := NewRentBill(CreateParams{
		RoomID:      1,
		PeriodMonth: "2026-06",
		BaseRent:    100000,
		Status:      StatusPaid,
		PaidAt:      &paidAt,
	})
	if err != nil {
		t.Fatalf("NewRentBill() error = %v", err)
	}
	err = bill.Cancel()
	if !errors.Is(err, ErrAlreadyPaid) {
		t.Fatalf("error = %v, want ErrAlreadyPaid", err)
	}
}

func TestRentBillMarkPaidRejectsCancelledBill(t *testing.T) {
	bill, err := NewRentBill(CreateParams{
		RoomID:      1,
		PeriodMonth: "2026-06",
		BaseRent:    100000,
		Status:      StatusCancelled,
	})
	if err != nil {
		t.Fatalf("NewRentBill() error = %v", err)
	}
	err = bill.MarkPaid(time.Now())
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("error = %v, want ErrCancelled", err)
	}
}

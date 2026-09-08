package payment

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mzkki/ottodot-trial/internal/usecase"
)

func TestMockPaymentGateway_Success(t *testing.T) {
	gw := NewMockPaymentGateway(0)
	ctx := context.Background()

	res, err := gw.Charge(ctx, usecase.ChargeRequest{
		BookingID:       "test-booking-id",
		AmountCents:     50000,
		SimulateFailure: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}
	if res.FailureReason != "" {
		t.Errorf("expected empty failure reason, got '%s'", res.FailureReason)
	}
	if !strings.HasPrefix(res.TransactionID, "mock-tx-") {
		t.Errorf("expected transaction ID starting with 'mock-tx-', got '%s'", res.TransactionID)
	}
}

func TestMockPaymentGateway_Failure(t *testing.T) {
	gw := NewMockPaymentGateway(0)
	ctx := context.Background()

	res, err := gw.Charge(ctx, usecase.ChargeRequest{
		BookingID:       "test-booking-id",
		AmountCents:     50000,
		SimulateFailure: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'", res.Status)
	}
	if res.FailureReason != "card_declined" {
		t.Errorf("expected failure reason 'card_declined', got '%s'", res.FailureReason)
	}
	if !strings.HasPrefix(res.TransactionID, "mock-fail-") {
		t.Errorf("expected transaction ID starting with 'mock-fail-', got '%s'", res.TransactionID)
	}
}

func TestMockPaymentGateway_ContextCancellation(t *testing.T) {
	gw := NewMockPaymentGateway(500 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := gw.Charge(ctx, usecase.ChargeRequest{
		BookingID:       "test-booking-id",
		AmountCents:     50000,
		SimulateFailure: false,
	})
	if err == nil {
		t.Fatal("expected context timeout error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

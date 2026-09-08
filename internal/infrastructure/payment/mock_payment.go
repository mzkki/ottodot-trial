package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mzkki/ottodot-trial/internal/usecase"
)

// MockPaymentGateway simulates a payment gateway adapter in the infrastructure layer.
// It fulfills usecase.PaymentGateway while providing deterministic toggles and latency simulation.
type MockPaymentGateway struct {
	Latency time.Duration
}

// Ensure MockPaymentGateway implements usecase.PaymentGateway
var _ usecase.PaymentGateway = (*MockPaymentGateway)(nil)

// NewMockPaymentGateway creates a new MockPaymentGateway instance.
func NewMockPaymentGateway(latency time.Duration) *MockPaymentGateway {
	return &MockPaymentGateway{
		Latency: latency,
	}
}

// Charge simulates charging a customer for a booking.
// If req.SimulateFailure is true, it deterministically fails with card_declined.
// If req.SimulateFailure is false, it deterministically succeeds with a mock transaction ID.
func (m *MockPaymentGateway) Charge(ctx context.Context, req usecase.ChargeRequest) (usecase.PaymentResult, error) {
	if m.Latency > 0 {
		select {
		case <-ctx.Done():
			return usecase.PaymentResult{}, ctx.Err()
		case <-time.After(m.Latency):
		}
	}

	if req.SimulateFailure {
		return usecase.PaymentResult{
			TransactionID: "mock-fail-" + uuid.NewString()[:8],
			Status:        "failed",
			FailureReason: "card_declined",
		}, nil
	}

	return usecase.PaymentResult{
		TransactionID: "mock-tx-" + uuid.NewString()[:8],
		Status:        "success",
		FailureReason: "",
	}, nil
}

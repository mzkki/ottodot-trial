package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/mzkki/ottodot-trial/internal/domain"
)

// Mock repositories for testing error paths
type mockParentRepo struct {
	findAllFn func(ctx context.Context) ([]domain.Parent, error)
}

func (m *mockParentRepo) FindAll(ctx context.Context) ([]domain.Parent, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, nil
}

type mockStudentRepo struct {
	findByParentIDFn func(ctx context.Context, parentID string) ([]domain.Student, error)
	findByIDFn       func(ctx context.Context, id string) (*domain.Student, error)
}

func (m *mockStudentRepo) FindByParentID(ctx context.Context, parentID string) ([]domain.Student, error) {
	if m.findByParentIDFn != nil {
		return m.findByParentIDFn(ctx, parentID)
	}
	return nil, nil
}

func (m *mockStudentRepo) FindByID(ctx context.Context, id string) (*domain.Student, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &domain.Student{ID: id}, nil
}

type mockClassRepo struct {
	findAllFn                 func(ctx context.Context) ([]domain.TrialClass, error)
	findByIDFn               func(ctx context.Context, id string) (*domain.TrialClass, error)
	findByIDForUpdateFn       func(ctx context.Context, id string) (*domain.TrialClass, error)
	findAvailableForStudentFn func(ctx context.Context, studentID string) ([]domain.TrialClass, error)
}

func (m *mockClassRepo) FindAll(ctx context.Context) ([]domain.TrialClass, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, nil
}

func (m *mockClassRepo) FindByID(ctx context.Context, id string) (*domain.TrialClass, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &domain.TrialClass{ID: id, MaxCapacity: 4}, nil
}

func (m *mockClassRepo) FindByIDForUpdate(ctx context.Context, id string) (*domain.TrialClass, error) {
	if m.findByIDForUpdateFn != nil {
		return m.findByIDForUpdateFn(ctx, id)
	}
	return &domain.TrialClass{ID: id, MaxCapacity: 4}, nil
}

func (m *mockClassRepo) FindAvailableForStudent(ctx context.Context, studentID string) ([]domain.TrialClass, error) {
	if m.findAvailableForStudentFn != nil {
		return m.findAvailableForStudentFn(ctx, studentID)
	}
	return nil, nil
}

type mockBookingRepo struct {
	createFn                 func(ctx context.Context, booking *domain.Booking) error
	updateFn                 func(ctx context.Context, booking *domain.Booking) error
	findByIDFn               func(ctx context.Context, id string) (*domain.Booking, error)
	findByIDForUpdateFn       func(ctx context.Context, id string) (*domain.Booking, error)
	countConfirmedForClassFn        func(ctx context.Context, classID string) (int64, error)
	countConfirmedGroupedByClassFn func(ctx context.Context) (map[string]int64, error)
	hasActiveBookingFn              func(ctx context.Context, studentID, classID string) (bool, error)
	findConfirmedByClassIDFn        func(ctx context.Context, classID string) ([]domain.Booking, error)
	findActiveByStudentIDFn         func(ctx context.Context, studentID string) ([]domain.Booking, error)
}

func (m *mockBookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	if m.createFn != nil {
		return m.createFn(ctx, booking)
	}
	return nil
}

func (m *mockBookingRepo) Update(ctx context.Context, booking *domain.Booking) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, booking)
	}
	return nil
}

func (m *mockBookingRepo) FindByID(ctx context.Context, id string) (*domain.Booking, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &domain.Booking{ID: id, Status: domain.BookingStatusPending}, nil
}

func (m *mockBookingRepo) FindByIDForUpdate(ctx context.Context, id string) (*domain.Booking, error) {
	if m.findByIDForUpdateFn != nil {
		return m.findByIDForUpdateFn(ctx, id)
	}
	return &domain.Booking{ID: id, Status: domain.BookingStatusPending, TrialClassID: "class-1"}, nil
}

func (m *mockBookingRepo) CountConfirmedForClass(ctx context.Context, classID string) (int64, error) {
	if m.countConfirmedForClassFn != nil {
		return m.countConfirmedForClassFn(ctx, classID)
	}
	return 0, nil
}

func (m *mockBookingRepo) CountConfirmedGroupedByClass(ctx context.Context) (map[string]int64, error) {
	if m.countConfirmedGroupedByClassFn != nil {
		return m.countConfirmedGroupedByClassFn(ctx)
	}
	return make(map[string]int64), nil
}

func (m *mockBookingRepo) HasActiveBooking(ctx context.Context, studentID, classID string) (bool, error) {
	if m.hasActiveBookingFn != nil {
		return m.hasActiveBookingFn(ctx, studentID, classID)
	}
	return false, nil
}

func (m *mockBookingRepo) FindConfirmedByClassID(ctx context.Context, classID string) ([]domain.Booking, error) {
	if m.findConfirmedByClassIDFn != nil {
		return m.findConfirmedByClassIDFn(ctx, classID)
	}
	return nil, nil
}

func (m *mockBookingRepo) FindActiveByStudentID(ctx context.Context, studentID string) ([]domain.Booking, error) {
	if m.findActiveByStudentIDFn != nil {
		return m.findActiveByStudentIDFn(ctx, studentID)
	}
	return nil, nil
}

type mockPaymentAttemptRepo struct {
	createFn          func(ctx context.Context, attempt *domain.PaymentAttempt) error
	findByBookingIDFn func(ctx context.Context, bookingID string) ([]domain.PaymentAttempt, error)
}

func (m *mockPaymentAttemptRepo) Create(ctx context.Context, attempt *domain.PaymentAttempt) error {
	if m.createFn != nil {
		return m.createFn(ctx, attempt)
	}
	return nil
}

func (m *mockPaymentAttemptRepo) FindByBookingID(ctx context.Context, bookingID string) ([]domain.PaymentAttempt, error) {
	if m.findByBookingIDFn != nil {
		return m.findByBookingIDFn(ctx, bookingID)
	}
	return nil, nil
}

type mockPaymentGateway struct {
	chargeFn func(ctx context.Context, req ChargeRequest) (PaymentResult, error)
}

func (m *mockPaymentGateway) Charge(ctx context.Context, req ChargeRequest) (PaymentResult, error) {
	if m.chargeFn != nil {
		return m.chargeFn(ctx, req)
	}
	if req.SimulateFailure {
		return PaymentResult{Status: "failed", FailureReason: "card_declined"}, nil
	}
	return PaymentResult{Status: "success"}, nil
}

type mockTxManager struct {
	withTransactionFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTransactionFn != nil {
		return m.withTransactionFn(ctx, fn)
	}
	return fn(ctx)
}

func TestBookingUseCaseErrorBranches(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("db error")

	// 1. CreateBooking error branches
	t.Run("CreateBooking HasActiveBooking error", func(t *testing.T) {
		bookingRepo := &mockBookingRepo{
			hasActiveBookingFn: func(ctx context.Context, studentID, classID string) (bool, error) {
				return false, testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.CreateBooking(ctx, "s1", "c1")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("CreateBooking CountConfirmedForClass error", func(t *testing.T) {
		bookingRepo := &mockBookingRepo{
			countConfirmedForClassFn: func(ctx context.Context, classID string) (int64, error) {
				return 0, testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.CreateBooking(ctx, "s1", "c1")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("CreateBooking bookingRepo.Create error", func(t *testing.T) {
		bookingRepo := &mockBookingRepo{
			createFn: func(ctx context.Context, booking *domain.Booking) error {
				return testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.CreateBooking(ctx, "s1", "c1")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	// 2. ProcessPayment error branches
	t.Run("ProcessPayment classRepo.FindByIDForUpdate error", func(t *testing.T) {
		classRepo := &mockClassRepo{
			findByIDForUpdateFn: func(ctx context.Context, id string) (*domain.TrialClass, error) {
				return nil, testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, classRepo, &mockBookingRepo{}, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.ProcessPayment(ctx, "b1", false)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ProcessPayment bookingRepo.CountConfirmedForClass error", func(t *testing.T) {
		bookingRepo := &mockBookingRepo{
			countConfirmedForClassFn: func(ctx context.Context, classID string) (int64, error) {
				return 0, testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.ProcessPayment(ctx, "b1", false)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ProcessPayment cancel update error on full class", func(t *testing.T) {
		bookingRepo := &mockBookingRepo{
			countConfirmedForClassFn: func(ctx context.Context, classID string) (int64, error) {
				return 4, nil // equal to capacity 4
			},
			updateFn: func(ctx context.Context, booking *domain.Booking) error {
				return testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.ProcessPayment(ctx, "b1", false)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ProcessPayment paymentGateway.Charge error", func(t *testing.T) {
		payGW := &mockPaymentGateway{
			chargeFn: func(ctx context.Context, req ChargeRequest) (PaymentResult, error) {
				return PaymentResult{}, testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, &mockBookingRepo{}, &mockPaymentAttemptRepo{}, payGW, &mockTxManager{})
		_, err := uc.ProcessPayment(ctx, "b1", false)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ProcessPayment paymentRepo.Create error", func(t *testing.T) {
		payRepo := &mockPaymentAttemptRepo{
			createFn: func(ctx context.Context, attempt *domain.PaymentAttempt) error {
				return testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, &mockBookingRepo{}, payRepo, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.ProcessPayment(ctx, "b1", false)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ProcessPayment update payment_failed error", func(t *testing.T) {
		bookingRepo := &mockBookingRepo{
			updateFn: func(ctx context.Context, booking *domain.Booking) error {
				return testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.ProcessPayment(ctx, "b1", true) // simulate failure
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ProcessPayment update confirmed error", func(t *testing.T) {
		bookingRepo := &mockBookingRepo{
			updateFn: func(ctx context.Context, booking *domain.Booking) error {
				return testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.ProcessPayment(ctx, "b1", false) // simulate success
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ProcessPayment final FindByID error when err == nil", func(t *testing.T) {
		bookingRepo := &mockBookingRepo{
			findByIDFn: func(ctx context.Context, id string) (*domain.Booking, error) {
				return nil, testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.ProcessPayment(ctx, "b1", false)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ProcessPayment tx error with successful final FindByID", func(t *testing.T) {
		txMgr := &mockTxManager{
			withTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, &mockClassRepo{}, &mockBookingRepo{}, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, txMgr)
		b, err := uc.ProcessPayment(ctx, "b1", false)
		if err != testErr {
			t.Fatalf("expected testErr, got %v", err)
		}
		if b == nil {
			t.Fatal("expected non-nil booking returned with error")
		}
	})

	t.Run("GetAvailableClassesWithStatus classRepo error", func(t *testing.T) {
		classRepo := &mockClassRepo{
			findAvailableForStudentFn: func(ctx context.Context, studentID string) ([]domain.TrialClass, error) {
				return nil, testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, classRepo, &mockBookingRepo{}, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.GetAvailableClassesWithStatus(ctx, "s1")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GetAvailableClassesWithStatus booking counts error", func(t *testing.T) {
		classRepo := &mockClassRepo{
			findAvailableForStudentFn: func(ctx context.Context, studentID string) ([]domain.TrialClass, error) {
				return []domain.TrialClass{{ID: "c1", Title: "Class 1"}}, nil
			},
		}
		bookingRepo := &mockBookingRepo{
			countConfirmedGroupedByClassFn: func(ctx context.Context) (map[string]int64, error) {
				return nil, testErr
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, classRepo, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		_, err := uc.GetAvailableClassesWithStatus(ctx, "s1")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GetAvailableClassesWithStatus active bookings error fallback and negative spots left branch", func(t *testing.T) {
		classRepo := &mockClassRepo{
			findAvailableForStudentFn: func(ctx context.Context, studentID string) ([]domain.TrialClass, error) {
				return []domain.TrialClass{
					{ID: "c1", Title: "Class 1", MaxCapacity: 2},
					{ID: "c2", Title: "Class 2", MaxCapacity: 1},
				}, nil
			},
		}
		bookingRepo := &mockBookingRepo{
			findActiveByStudentIDFn: func(ctx context.Context, studentID string) ([]domain.Booking, error) {
				return nil, testErr // triggers fallback to make(map[string]domain.Booking)
			},
			countConfirmedGroupedByClassFn: func(ctx context.Context) (map[string]int64, error) {
				return map[string]int64{"c2": 5}, nil // triggers spotsLeft < 0 branch
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, classRepo, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		results, err := uc.GetAvailableClassesWithStatus(ctx, "s1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if results[1].SpotsLeft != 0 {
			t.Errorf("expected spotsLeft clamped to 0, got %d", results[1].SpotsLeft)
		}
	})

	t.Run("GetAvailableClassesWithStatus with active bookings", func(t *testing.T) {
		classRepo := &mockClassRepo{
			findAvailableForStudentFn: func(ctx context.Context, studentID string) ([]domain.TrialClass, error) {
				return []domain.TrialClass{
					{ID: "c1", Title: "Class 1", MaxCapacity: 5},
				}, nil
			},
		}
		bookingRepo := &mockBookingRepo{
			findActiveByStudentIDFn: func(ctx context.Context, studentID string) ([]domain.Booking, error) {
				return []domain.Booking{
					{ID: "b1", TrialClassID: "c1", Status: domain.BookingStatusConfirmed},
				}, nil
			},
			countConfirmedGroupedByClassFn: func(ctx context.Context) (map[string]int64, error) {
				return map[string]int64{"c1": 1}, nil
			},
		}
		uc := NewBookingUseCaseWithInterfaces(&mockParentRepo{}, &mockStudentRepo{}, classRepo, bookingRepo, &mockPaymentAttemptRepo{}, &mockPaymentGateway{}, &mockTxManager{})
		results, err := uc.GetAvailableClassesWithStatus(ctx, "s1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !results[0].IsBooked || results[0].BookingStatus != "confirmed" || results[0].BookingID != "b1" {
			t.Errorf("unexpected booking data: %+v", results[0])
		}
	})
}

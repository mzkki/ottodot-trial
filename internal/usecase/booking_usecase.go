package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mzkki/ottodot-trial/internal/domain"
	tx "github.com/mzkki/ottodot-trial/internal/pkg/tx"
	"github.com/mzkki/ottodot-trial/internal/repository"
)

// ─── Business Errors ──────────────────────────────────────────────────────────

var (
	ErrClassFull         = errors.New("class is full, no seats available")
	ErrBookingNotPending = errors.New("booking is not in pending status")
	ErrPaymentFailed     = errors.New("payment processing failed")
	ErrDuplicateBooking  = errors.New("student already has an active booking for this class")
)

// ─── Data Transfer Objects (DTOs) ─────────────────────────────────────────────

// ClassAvailabilityDTO encapsulates projection data for trial class browsing.
type ClassAvailabilityDTO struct {
	ID              string
	Title           string
	Description     string
	Instructor      string
	ScheduledAt     string
	DurationMinutes int
	MaxCapacity     int
	ConfirmedCount  int64
	SpotsLeft       int
	IsBooked        bool
	BookingStatus   string
	BookingID       string
}

// PaymentResult represents the response from a payment gateway attempt.
type PaymentResult struct {
	TransactionID string
	Status        string
	FailureReason string
}

// ChargeRequest represents the payload passed to a payment gateway.
type ChargeRequest struct {
	BookingID       string
	AmountCents     int
	SimulateFailure bool
}

// ─── Consumer Repository, Gateway & Transaction Interfaces (DIP) ──────────────

type ParentRepo interface {
	FindAll(ctx context.Context) ([]domain.Parent, error)
}

type StudentRepo interface {
	FindByParentID(ctx context.Context, parentID string) ([]domain.Student, error)
	FindByID(ctx context.Context, id string) (*domain.Student, error)
}

type TrialClassRepo interface {
	FindAll(ctx context.Context) ([]domain.TrialClass, error)
	FindByID(ctx context.Context, id string) (*domain.TrialClass, error)
	FindByIDForUpdate(ctx context.Context, id string) (*domain.TrialClass, error)
	FindAvailableForStudent(ctx context.Context, studentID string) ([]domain.TrialClass, error)
}

type BookingRepo interface {
	Create(ctx context.Context, booking *domain.Booking) error
	Update(ctx context.Context, booking *domain.Booking) error
	FindByID(ctx context.Context, id string) (*domain.Booking, error)
	FindByIDForUpdate(ctx context.Context, id string) (*domain.Booking, error)
	CountConfirmedForClass(ctx context.Context, classID string) (int64, error)
	CountConfirmedGroupedByClass(ctx context.Context) (map[string]int64, error)
	HasActiveBooking(ctx context.Context, studentID, classID string) (bool, error)
	FindConfirmedByClassID(ctx context.Context, classID string) ([]domain.Booking, error)
	FindActiveByStudentID(ctx context.Context, studentID string) ([]domain.Booking, error)
}

type PaymentAttemptRepo interface {
	Create(ctx context.Context, attempt *domain.PaymentAttempt) error
	FindByBookingID(ctx context.Context, bookingID string) ([]domain.PaymentAttempt, error)
}

// PaymentGateway decouples payment processing from external gateway implementations.
type PaymentGateway interface {
	Charge(ctx context.Context, req ChargeRequest) (PaymentResult, error)
}

type TxManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// ─── Booking Usecase Contract ─────────────────────────────────────────────────

type BookingUseCaseInterface interface {
	GetAllParents(ctx context.Context) ([]domain.Parent, error)
	GetStudentsByParent(ctx context.Context, parentID string) ([]domain.Student, error)
	GetAvailableClasses(ctx context.Context, studentID string) ([]domain.TrialClass, error)
	GetAvailableClassesWithStatus(ctx context.Context, studentID string) ([]ClassAvailabilityDTO, error)
	GetActiveBookingsByStudent(ctx context.Context, studentID string) (map[string]domain.Booking, error)
	GetAllClasses(ctx context.Context) ([]domain.TrialClass, error)
	GetClassRoster(ctx context.Context, classID string) ([]domain.Booking, error)
	GetClassByID(ctx context.Context, classID string) (*domain.TrialClass, error)
	GetBookingByID(ctx context.Context, bookingID string) (*domain.Booking, error)
	CountConfirmedForClass(ctx context.Context, classID string) (int64, error)
	CreateBooking(ctx context.Context, studentID, classID string) (*domain.Booking, error)
	ProcessPayment(ctx context.Context, bookingID string, simulateFailure bool) (*domain.Booking, error)
}

// ─── Booking Usecase Struct & Constructors ────────────────────────────────────

type BookingUseCase struct {
	parentRepo     ParentRepo
	studentRepo    StudentRepo
	classRepo      TrialClassRepo
	bookingRepo    BookingRepo
	paymentRepo    PaymentAttemptRepo
	paymentGateway PaymentGateway
	txManager      TxManager
}

// Compile-time check that BookingUseCase implements BookingUseCaseInterface
var _ BookingUseCaseInterface = (*BookingUseCase)(nil)

func NewBookingUseCase(
	parentRepo *repository.ParentRepository,
	studentRepo *repository.StudentRepository,
	classRepo *repository.TrialClassRepository,
	bookingRepo *repository.BookingRepository,
	paymentRepo *repository.PaymentAttemptRepository,
	paymentGateway PaymentGateway,
	txManager *tx.GormTransactionManager,
) *BookingUseCase {
	return NewBookingUseCaseWithInterfaces(
		parentRepo,
		studentRepo,
		classRepo,
		bookingRepo,
		paymentRepo,
		paymentGateway,
		txManager,
	)
}

// NewBookingUseCaseWithInterfaces creates a BookingUseCase with interfaces for mock testing.
func NewBookingUseCaseWithInterfaces(
	parentRepo ParentRepo,
	studentRepo StudentRepo,
	classRepo TrialClassRepo,
	bookingRepo BookingRepo,
	paymentRepo PaymentAttemptRepo,
	paymentGateway PaymentGateway,
	txManager TxManager,
) *BookingUseCase {
	return &BookingUseCase{
		parentRepo:     parentRepo,
		studentRepo:    studentRepo,
		classRepo:      classRepo,
		bookingRepo:    bookingRepo,
		paymentRepo:    paymentRepo,
		paymentGateway: paymentGateway,
		txManager:      txManager,
	}
}

// ─── Booking Usecase Implementations ──────────────────────────────────────────

// GetAllParents returns all parents for the parent selector.
func (u *BookingUseCase) GetAllParents(ctx context.Context) ([]domain.Parent, error) {
	return u.parentRepo.FindAll(ctx)
}

// GetStudentsByParent returns all students belonging to a parent.
func (u *BookingUseCase) GetStudentsByParent(ctx context.Context, parentID string) ([]domain.Student, error) {
	return u.studentRepo.FindByParentID(ctx, parentID)
}

// GetAvailableClasses returns trial classes for a student.
func (u *BookingUseCase) GetAvailableClasses(ctx context.Context, studentID string) ([]domain.TrialClass, error) {
	return u.classRepo.FindAvailableForStudent(ctx, studentID)
}

// GetAvailableClassesWithStatus returns trial classes enriched with availability, spots left,
// and booking status for the requested student, using a single batch query to eliminate N+1 latency.
func (u *BookingUseCase) GetAvailableClassesWithStatus(ctx context.Context, studentID string) ([]ClassAvailabilityDTO, error) {
	classes, err := u.classRepo.FindAvailableForStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch classes: %w", err)
	}

	activeBookings, err := u.GetActiveBookingsByStudent(ctx, studentID)
	if err != nil {
		activeBookings = make(map[string]domain.Booking)
	}

	confirmedCounts, err := u.bookingRepo.CountConfirmedGroupedByClass(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch class counts: %w", err)
	}

	dtos := make([]ClassAvailabilityDTO, 0, len(classes))
	for _, cl := range classes {
		count := confirmedCounts[cl.ID]
		b, isBooked := activeBookings[cl.ID]
		bStatus := ""
		bID := ""
		if isBooked {
			bStatus = string(b.Status)
			bID = b.ID
		}

		spotsLeft := cl.MaxCapacity - int(count)
		if spotsLeft < 0 {
			spotsLeft = 0
		}

		dtos = append(dtos, ClassAvailabilityDTO{
			ID:              cl.ID,
			Title:           cl.Title,
			Description:     cl.Description,
			Instructor:      cl.Instructor,
			ScheduledAt:     cl.ScheduledAt.Format("Mon, 02 Jan 2006 · 3:04 PM"),
			DurationMinutes: cl.DurationMinutes,
			MaxCapacity:     cl.MaxCapacity,
			ConfirmedCount:  count,
			SpotsLeft:       spotsLeft,
			IsBooked:        isBooked,
			BookingStatus:   bStatus,
			BookingID:       bID,
		})
	}

	return dtos, nil
}

// GetActiveBookingsByStudent returns a map of classID -> Booking for a student's active bookings.
func (u *BookingUseCase) GetActiveBookingsByStudent(ctx context.Context, studentID string) (map[string]domain.Booking, error) {
	bookings, err := u.bookingRepo.FindActiveByStudentID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	res := make(map[string]domain.Booking, len(bookings))
	for _, b := range bookings {
		res[b.TrialClassID] = b
	}
	return res, nil
}

// GetAllClasses returns all trial classes (for admin).
func (u *BookingUseCase) GetAllClasses(ctx context.Context) ([]domain.TrialClass, error) {
	return u.classRepo.FindAll(ctx)
}

// GetClassRoster returns all confirmed bookings for a class.
func (u *BookingUseCase) GetClassRoster(ctx context.Context, classID string) ([]domain.Booking, error) {
	return u.bookingRepo.FindConfirmedByClassID(ctx, classID)
}

// GetClassByID returns a single trial class.
func (u *BookingUseCase) GetClassByID(ctx context.Context, classID string) (*domain.TrialClass, error) {
	return u.classRepo.FindByID(ctx, classID)
}

// GetBookingByID returns a booking with all related data.
func (u *BookingUseCase) GetBookingByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	return u.bookingRepo.FindByID(ctx, bookingID)
}

// CountConfirmedForClass returns the number of confirmed bookings for a class (no lock).
func (u *BookingUseCase) CountConfirmedForClass(ctx context.Context, classID string) (int64, error) {
	return u.bookingRepo.CountConfirmedForClass(ctx, classID)
}

// CreateBooking creates a new pending booking after validating constraints.
func (u *BookingUseCase) CreateBooking(ctx context.Context, studentID, classID string) (*domain.Booking, error) {
	// Validate student exists
	student, err := u.studentRepo.FindByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("student not found: %w", err)
	}

	// Validate class exists
	class, err := u.classRepo.FindByID(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	// Check for duplicate active booking
	hasActive, err := u.bookingRepo.HasActiveBooking(ctx, studentID, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing bookings: %w", err)
	}
	if hasActive {
		return nil, ErrDuplicateBooking
	}

	// Pre-check capacity (non-locking, just a UX guard)
	confirmedCount, err := u.bookingRepo.CountConfirmedForClass(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to count bookings: %w", err)
	}
	if int(confirmedCount) >= class.MaxCapacity {
		return nil, ErrClassFull
	}

	booking := &domain.Booking{
		ID:           uuid.NewString(),
		StudentID:    student.ID,
		TrialClassID: class.ID,
		Status:       domain.BookingStatusPending,
	}

	if err := u.bookingRepo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	return u.bookingRepo.FindByID(ctx, booking.ID)
}

// ProcessPayment runs the critical payment flow inside a transaction with row-level locking.
func (u *BookingUseCase) ProcessPayment(ctx context.Context, bookingID string, simulateFailure bool) (*domain.Booking, error) {
	var businessErr error

	err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Step 1: Lock the booking row
		booking, err := u.bookingRepo.FindByIDForUpdate(txCtx, bookingID)
		if err != nil {
			return fmt.Errorf("booking not found: %w", err)
		}
		if booking.Status != domain.BookingStatusPending && booking.Status != domain.BookingStatusPaymentFailed {
			businessErr = ErrBookingNotPending
			return nil
		}

		// Step 2: Lock the trial class row to serialize concurrent payments for this class
		class, err := u.classRepo.FindByIDForUpdate(txCtx, booking.TrialClassID)
		if err != nil {
			return fmt.Errorf("class not found: %w", err)
		}

		// Step 3: Count confirmed bookings for this class
		confirmedCount, err := u.bookingRepo.CountConfirmedForClass(txCtx, booking.TrialClassID)
		if err != nil {
			return fmt.Errorf("failed to count confirmed bookings: %w", err)
		}

		if int(confirmedCount) >= class.MaxCapacity {
			// No seats left — cancel this booking
			booking.Status = domain.BookingStatusCancelled
			if err := u.bookingRepo.Update(txCtx, booking); err != nil {
				return fmt.Errorf("failed to cancel booking: %w", err)
			}
			businessErr = ErrClassFull
			return nil
		}

		// Step 4: Process payment via PaymentGateway adapter
		paymentResult, err := u.paymentGateway.Charge(txCtx, ChargeRequest{
			BookingID:       booking.ID,
			AmountCents:     50000, // Fixed trial class price: $500.00
			SimulateFailure: simulateFailure,
		})
		if err != nil {
			return fmt.Errorf("payment charge failed: %w", err)
		}

		// Step 5: Record payment attempt
		attempt := &domain.PaymentAttempt{
			ID:            uuid.NewString(),
			BookingID:     booking.ID,
			AmountCents:   50000,
			Status:        paymentResult.Status,
			FailureReason: paymentResult.FailureReason,
		}
		if err := u.paymentRepo.Create(txCtx, attempt); err != nil {
			return fmt.Errorf("failed to record payment: %w", err)
		}

		// Step 6: Update booking status based on payment result
		if paymentResult.Status == "failed" {
			booking.Status = domain.BookingStatusPaymentFailed
			if err := u.bookingRepo.Update(txCtx, booking); err != nil {
				return fmt.Errorf("failed to update booking: %w", err)
			}
			businessErr = ErrPaymentFailed
			return nil
		}

		// Payment succeeded — confirm the booking
		booking.Status = domain.BookingStatusConfirmed
		if err := u.bookingRepo.Update(txCtx, booking); err != nil {
			return fmt.Errorf("failed to confirm booking: %w", err)
		}

		return nil
	})

	// Always fetch the latest booking state (with relations & payment history) for display
	resultBooking, fetchErr := u.bookingRepo.FindByID(ctx, bookingID)
	if fetchErr != nil {
		if err != nil {
			return nil, err
		}
		return nil, fetchErr
	}

	if err != nil {
		return resultBooking, err
	}
	return resultBooking, businessErr
}

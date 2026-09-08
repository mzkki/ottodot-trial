package usecase

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mzkki/ottodot-trial/internal/config"
	"github.com/mzkki/ottodot-trial/internal/domain"
	tx "github.com/mzkki/ottodot-trial/internal/pkg/tx"
	"github.com/mzkki/ottodot-trial/internal/repository"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, *BookingUseCase) {
	testDBName := os.Getenv("DB_TEST_NAME")
	if testDBName == "" {
		testDBName = "db_ottodot_trial_test"
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "ottodot"
	}
	pass := os.Getenv("DB_PASS")
	if pass == "" {
		pass = "ottodot123"
	}

	v := viper.New()
	v.Set("DB_HOST", host)
	v.Set("DB_PORT", port)
	v.Set("DB_USER", user)
	v.Set("DB_PASS", pass)
	v.Set("DB_NAME", testDBName)

	db, err := config.NewDatabase(v)
	if err != nil {
		t.Skipf("skipping test: test database not available: %v", err)
		return nil, nil
	}
	db.Logger = gormLogger.Default.LogMode(gormLogger.Silent)

	// Ensure tables exist in the test database
	_ = db.AutoMigrate(&domain.Parent{}, &domain.Student{}, &domain.TrialClass{}, &domain.Booking{}, &domain.PaymentAttempt{})

	// Seed baseline fixtures if empty
	var parentCount int64
	db.Model(&domain.Parent{}).Count(&parentCount)
	if parentCount == 0 {
		p := domain.Parent{
			ID:    "a0000000-0000-0000-0000-000000000001",
			Name:  "Sarah Johnson",
			Email: "sarah@example.com",
			Phone: "08123456789",
		}
		db.Create(&p)
		s := domain.Student{
			ID:       "b0000000-0000-0000-0000-000000000001",
			ParentID: p.ID,
			Name:     "Emma Johnson",
			Age:      7,
		}
		db.Create(&s)
		c := domain.TrialClass{
			ID:              "c0000000-0000-0000-0000-000000000001",
			Title:           "Introduction to Robotics",
			Instructor:      "Mr. Smith",
			ScheduledAt:     time.Now().Add(72 * time.Hour),
			DurationMinutes: 60,
			MaxCapacity:     4,
		}
		db.Create(&c)
	}

	parentRepo := repository.NewParentRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	classRepo := repository.NewTrialClassRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	paymentRepo := repository.NewPaymentAttemptRepository(db)
	paymentGateway := &mockPaymentGateway{}
	txManager := tx.NewGormTransactionManager(db)

	uc := NewBookingUseCase(parentRepo, studentRepo, classRepo, bookingRepo, paymentRepo, paymentGateway, txManager)
	return db, uc
}

func TestBookingUseCaseReadMethods(t *testing.T) {
	ctx := context.Background()
	_, uc := setupTestDB(t)

	// Test GetAllParents
	parents, err := uc.GetAllParents(ctx)
	if err != nil {
		t.Fatalf("GetAllParents failed: %v", err)
	}
	if len(parents) == 0 {
		t.Error("expected at least 1 parent")
	}

	// Test GetStudentsByParent
	parentID := parents[0].ID
	students, err := uc.GetStudentsByParent(ctx, parentID)
	if err != nil {
		t.Fatalf("GetStudentsByParent failed: %v", err)
	}
	if len(students) == 0 {
		t.Error("expected at least 1 student for parent")
	}

	// Test GetAllClasses
	classes, err := uc.GetAllClasses(ctx)
	if err != nil {
		t.Fatalf("GetAllClasses failed: %v", err)
	}
	if len(classes) == 0 {
		t.Error("expected at least 1 trial class")
	}

	// Test GetAvailableClasses
	studentID := students[0].ID
	availClasses, err := uc.GetAvailableClasses(ctx, studentID)
	if err != nil {
		t.Fatalf("GetAvailableClasses failed: %v", err)
	}
	if len(availClasses) == 0 {
		t.Error("expected classes available")
	}

	// Test GetActiveBookingsByStudent
	activeBookings, err := uc.GetActiveBookingsByStudent(ctx, studentID)
	if err != nil {
		t.Fatalf("GetActiveBookingsByStudent failed: %v", err)
	}
	if activeBookings == nil {
		t.Error("expected non-nil active bookings map")
	}

	// Test GetClassByID
	class, err := uc.GetClassByID(ctx, classes[0].ID)
	if err != nil || class == nil {
		t.Fatalf("GetClassByID failed: %v", err)
	}

	// Test CountConfirmedForClass
	count, err := uc.CountConfirmedForClass(ctx, classes[0].ID)
	if err != nil {
		t.Fatalf("CountConfirmedForClass failed: %v", err)
	}
	if count < 0 {
		t.Errorf("expected count >= 0, got %d", count)
	}

	// Test GetClassRoster
	roster, err := uc.GetClassRoster(ctx, classes[0].ID)
	if err != nil {
		t.Fatalf("GetClassRoster failed: %v", err)
	}
	if roster == nil {
		t.Error("expected non-nil roster slice")
	}
}

func TestBookingUseCaseCreateAndPaymentFlow(t *testing.T) {
	ctx := context.Background()
	db, uc := setupTestDB(t)

	// Create test parent and student
	parent := domain.Parent{
		ID:    uuid.NewString(),
		Name:  "Test Parent",
		Email: "test-" + uuid.NewString()[:8] + "@example.com",
	}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatalf("failed to create test parent: %v", err)
	}
	defer db.Delete(&parent)

	student := domain.Student{
		ID:       uuid.NewString(),
		ParentID: parent.ID,
		Name:     "Test Student",
		Age:      6,
	}
	if err := db.Create(&student).Error; err != nil {
		t.Fatalf("failed to create test student: %v", err)
	}
	defer db.Delete(&student)

	// Create test class with capacity 1
	class := domain.TrialClass{
		ID:              uuid.NewString(),
		Title:           "Test Class",
		Instructor:      "Tester",
		ScheduledAt:     time.Now().Add(48 * time.Hour),
		DurationMinutes: 60,
		MaxCapacity:     1,
	}
	if err := db.Create(&class).Error; err != nil {
		t.Fatalf("failed to create test class: %v", err)
	}
	defer db.Delete(&class)

	// 1. Create booking
	booking, err := uc.CreateBooking(ctx, student.ID, class.ID)
	if err != nil {
		t.Fatalf("CreateBooking failed: %v", err)
	}
	defer db.Where("booking_id = ?", booking.ID).Delete(&domain.PaymentAttempt{})
	defer db.Delete(booking)

	if booking.Status != domain.BookingStatusPending {
		t.Fatalf("expected pending status, got %s", booking.Status)
	}

	// Test GetBookingByID
	fetchedBooking, err := uc.GetBookingByID(ctx, booking.ID)
	if err != nil || fetchedBooking == nil {
		t.Fatalf("GetBookingByID failed: %v", err)
	}
	if fetchedBooking.ID != booking.ID {
		t.Errorf("expected booking ID %s, got %s", booking.ID, fetchedBooking.ID)
	}

	// 2. Test duplicate booking error
	_, dupErr := uc.CreateBooking(ctx, student.ID, class.ID)
	if !errors.Is(dupErr, ErrDuplicateBooking) {
		t.Fatalf("expected ErrDuplicateBooking, got %v", dupErr)
	}

	// 3. Test payment failure with deterministic toggle (simulateFailure = true)
	failedBooking, payErr := uc.ProcessPayment(ctx, booking.ID, true)
	if !errors.Is(payErr, ErrPaymentFailed) {
		t.Fatalf("expected ErrPaymentFailed, got %v", payErr)
	}
	if failedBooking.Status != domain.BookingStatusPaymentFailed {
		t.Fatalf("expected payment_failed status, got %s", failedBooking.Status)
	}

	// 4. Test payment retry with success (simulateFailure = false)
	confirmedBooking, paySuccessErr := uc.ProcessPayment(ctx, booking.ID, false)
	if paySuccessErr != nil {
		t.Fatalf("expected nil error on payment success, got %v", paySuccessErr)
	}
	if confirmedBooking.Status != domain.BookingStatusConfirmed {
		t.Fatalf("expected confirmed status, got %s", confirmedBooking.Status)
	}

	// 5. Test paying an already confirmed booking returns ErrBookingNotPending
	_, alreadyErr := uc.ProcessPayment(ctx, booking.ID, false)
	if !errors.Is(alreadyErr, ErrBookingNotPending) {
		t.Fatalf("expected ErrBookingNotPending, got %v", alreadyErr)
	}

	// 6. Test capacity full at CreateBooking time:
	student2 := domain.Student{
		ID:       uuid.NewString(),
		ParentID: parent.ID,
		Name:     "Test Student 2",
		Age:      8,
	}
	if err := db.Create(&student2).Error; err != nil {
		t.Fatalf("failed to create test student 2: %v", err)
	}
	defer db.Delete(&student2)

	// CreateBooking should detect class full
	_, fullErr := uc.CreateBooking(ctx, student2.ID, class.ID)
	if !errors.Is(fullErr, ErrClassFull) {
		t.Fatalf("expected ErrClassFull, got %v", fullErr)
	}
}

func TestBookingUseCaseValidationErrors(t *testing.T) {
	ctx := context.Background()
	_, uc := setupTestDB(t)

	// Non-existent student
	_, err := uc.CreateBooking(ctx, uuid.NewString(), uuid.NewString())
	if err == nil || !strings.Contains(err.Error(), "student not found") {
		t.Fatalf("expected 'student not found', got %v", err)
	}

	// Non-existent class with valid student
	parents, _ := uc.GetAllParents(ctx)
	students, _ := uc.GetStudentsByParent(ctx, parents[0].ID)
	_, err = uc.CreateBooking(ctx, students[0].ID, uuid.NewString())
	if err == nil || !strings.Contains(err.Error(), "class not found") {
		t.Fatalf("expected 'class not found', got %v", err)
	}

	// ProcessPayment with non-existent booking ID
	_, err = uc.ProcessPayment(ctx, uuid.NewString(), false)
	if err == nil || !strings.Contains(err.Error(), "booking not found") {
		t.Fatalf("expected 'booking not found', got %v", err)
	}
}

func TestBookingUseCaseRaceConditionDuringPayment(t *testing.T) {
	ctx := context.Background()
	db, uc := setupTestDB(t)

	// Create parent and two students
	parent := domain.Parent{
		ID:    uuid.NewString(),
		Name:  "Race Parent",
		Email: "race-" + uuid.NewString()[:8] + "@example.com",
	}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatalf("failed to create parent: %v", err)
	}
	defer db.Delete(&parent)

	studentA := domain.Student{
		ID:       uuid.NewString(),
		ParentID: parent.ID,
		Name:     "Student A",
		Age:      7,
	}
	studentB := domain.Student{
		ID:       uuid.NewString(),
		ParentID: parent.ID,
		Name:     "Student B",
		Age:      8,
	}
	if err := db.Create(&studentA).Error; err != nil {
		t.Fatalf("failed to create student A: %v", err)
	}
	defer db.Delete(&studentA)
	if err := db.Create(&studentB).Error; err != nil {
		t.Fatalf("failed to create student B: %v", err)
	}
	defer db.Delete(&studentB)

	// Create class with capacity 1
	class := domain.TrialClass{
		ID:              uuid.NewString(),
		Title:           "Race Condition Class",
		Instructor:      "Tester",
		ScheduledAt:     time.Now().Add(72 * time.Hour),
		DurationMinutes: 60,
		MaxCapacity:     1,
	}
	if err := db.Create(&class).Error; err != nil {
		t.Fatalf("failed to create race class: %v", err)
	}
	defer db.Delete(&class)

	// Step 1: Student A creates pending booking
	bookingA := domain.Booking{
		ID:           uuid.NewString(),
		StudentID:    studentA.ID,
		TrialClassID: class.ID,
		Status:       domain.BookingStatusPending,
	}
	if err := db.Create(&bookingA).Error; err != nil {
		t.Fatalf("failed to create booking A: %v", err)
	}
	defer db.Delete(&bookingA)

	// Step 2: Student B gets confirmed directly, taking the last slot (capacity = 1)
	bookingB := domain.Booking{
		ID:           uuid.NewString(),
		StudentID:    studentB.ID,
		TrialClassID: class.ID,
		Status:       domain.BookingStatusConfirmed,
	}
	if err := db.Create(&bookingB).Error; err != nil {
		t.Fatalf("failed to create booking B: %v", err)
	}
	defer db.Delete(&bookingB)

	// Step 3: Student A now tries to process payment on booking A.
	// The class is full, so ProcessPayment MUST cancel booking A and return ErrClassFull.
	resultBooking, err := uc.ProcessPayment(ctx, bookingA.ID, false)
	if !errors.Is(err, ErrClassFull) {
		t.Fatalf("expected ErrClassFull, got %v", err)
	}
	if resultBooking.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled status for booking A, got %s", resultBooking.Status)
	}
}

func TestBookingUseCaseCanceledContext(t *testing.T) {
	ctx := context.Background()
	_, uc := setupTestDB(t)

	canceledCtx, cancel := context.WithCancel(ctx)
	cancel() // cancel immediately

	// Test GetActiveBookingsByStudent with canceled context
	_, err := uc.GetActiveBookingsByStudent(canceledCtx, uuid.NewString())
	if err == nil {
		t.Error("expected error with canceled context in GetActiveBookingsByStudent")
	}

	// Test CreateBooking with canceled context
	_, err = uc.CreateBooking(canceledCtx, uuid.NewString(), uuid.NewString())
	if err == nil {
		t.Error("expected error with canceled context in CreateBooking")
	}

	// Test ProcessPayment with canceled context
	_, err = uc.ProcessPayment(canceledCtx, uuid.NewString(), false)
	if err == nil {
		t.Error("expected error with canceled context in ProcessPayment")
	}
}

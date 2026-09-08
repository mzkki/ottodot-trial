package repository

import (
	"context"
	"errors"

	"github.com/mzkki/ottodot-trial/internal/domain"
	tx "github.com/mzkki/ottodot-trial/internal/pkg/tx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Create inserts a new booking record into the database.
func (r *BookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	db := tx.GetDB(ctx, r.db)
	return db.Create(booking).Error
}

// Update saves changes to an existing booking.
func (r *BookingRepository) Update(ctx context.Context, booking *domain.Booking) error {
	db := tx.GetDB(ctx, r.db)
	return db.Save(booking).Error
}

// FindByID returns a booking with its preloaded relations.
func (r *BookingRepository) FindByID(ctx context.Context, id string) (*domain.Booking, error) {
	db := tx.GetDB(ctx, r.db)
	var booking domain.Booking
	err := db.Preload("Student").Preload("Student.Parent").
		Preload("TrialClass").Preload("PaymentAttempts").
		Where("id = ?", id).First(&booking).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &booking, err
}

// FindByIDForUpdate locks the booking row for the duration of the transaction.
// Must be called within a transaction context.
func (r *BookingRepository) FindByIDForUpdate(ctx context.Context, id string) (*domain.Booking, error) {
	db := tx.GetDB(ctx, r.db)
	var booking domain.Booking
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("TrialClass").
		Where("id = ?", id).First(&booking).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &booking, err
}

// CountConfirmedForClass counts confirmed bookings for a class.
func (r *BookingRepository) CountConfirmedForClass(ctx context.Context, classID string) (int64, error) {
	db := tx.GetDB(ctx, r.db)
	var count int64
	err := db.Model(&domain.Booking{}).
		Where("trial_class_id = ? AND status = ?", classID, domain.BookingStatusConfirmed).
		Count(&count).Error
	return count, err
}

// CountConfirmedGroupedByClass retrieves confirmed booking counts for all classes in a single query.
func (r *BookingRepository) CountConfirmedGroupedByClass(ctx context.Context) (map[string]int64, error) {
	db := tx.GetDB(ctx, r.db)
	type classCount struct {
		TrialClassID string `gorm:"column:trial_class_id"`
		Count        int64  `gorm:"column:count"`
	}
	var results []classCount
	err := db.Model(&domain.Booking{}).
		Select("trial_class_id, count(*) as count").
		Where("status = ?", domain.BookingStatusConfirmed).
		Group("trial_class_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(results))
	for _, res := range results {
		counts[res.TrialClassID] = res.Count
	}
	return counts, nil
}


// HasActiveBooking checks if a student already has an active (pending/confirmed) booking for a class.
func (r *BookingRepository) HasActiveBooking(ctx context.Context, studentID, classID string) (bool, error) {
	db := tx.GetDB(ctx, r.db)
	var count int64
	err := db.Model(&domain.Booking{}).
		Where("student_id = ? AND trial_class_id = ? AND status IN ?",
			studentID, classID, []string{string(domain.BookingStatusPending), string(domain.BookingStatusConfirmed)}).
		Count(&count).Error
	return count > 0, err
}

// FindConfirmedByClassID returns all confirmed bookings for a class with student and parent details.
func (r *BookingRepository) FindConfirmedByClassID(ctx context.Context, classID string) ([]domain.Booking, error) {
	db := tx.GetDB(ctx, r.db)
	var bookings []domain.Booking
	err := db.Preload("Student").Preload("Student.Parent").
		Where("trial_class_id = ? AND status = ?", classID, domain.BookingStatusConfirmed).
		Order("created_at ASC").
		Find(&bookings).Error
	return bookings, err
}

// FindActiveByStudentID returns all active (pending/confirmed) bookings for a student.
func (r *BookingRepository) FindActiveByStudentID(ctx context.Context, studentID string) ([]domain.Booking, error) {
	db := tx.GetDB(ctx, r.db)
	var bookings []domain.Booking
	err := db.Where("student_id = ? AND status IN ?",
		studentID, []string{string(domain.BookingStatusPending), string(domain.BookingStatusConfirmed)}).
		Find(&bookings).Error
	return bookings, err
}

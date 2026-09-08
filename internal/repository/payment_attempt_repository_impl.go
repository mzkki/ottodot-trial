package repository

import (
	"context"

	"github.com/mzkki/ottodot-trial/internal/domain"
	tx "github.com/mzkki/ottodot-trial/internal/pkg/tx"
)

// Create inserts a new payment attempt record.
func (r *PaymentAttemptRepository) Create(ctx context.Context, attempt *domain.PaymentAttempt) error {
	db := tx.GetDB(ctx, r.db)
	return db.Create(attempt).Error
}

// FindByBookingID returns all payment attempts for a booking ordered by newest first.
func (r *PaymentAttemptRepository) FindByBookingID(ctx context.Context, bookingID string) ([]domain.PaymentAttempt, error) {
	db := tx.GetDB(ctx, r.db)
	var attempts []domain.PaymentAttempt
	err := db.Where("booking_id = ?", bookingID).
		Order("created_at DESC").
		Find(&attempts).Error
	return attempts, err
}

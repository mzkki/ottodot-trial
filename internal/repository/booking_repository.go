package repository

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound         = errors.New("record not found")
	ErrDuplicateBooking = errors.New("active booking already exists for this student and class")
)

// BookingRepository manages database operations for bookings.
type BookingRepository struct {
	db *gorm.DB
}

// NewBookingRepository constructs a new BookingRepository.
func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

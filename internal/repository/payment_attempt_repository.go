package repository

import (
	"gorm.io/gorm"
)

// PaymentAttemptRepository manages database operations for payment attempts.
type PaymentAttemptRepository struct {
	db *gorm.DB
}

// NewPaymentAttemptRepository constructs a new PaymentAttemptRepository.
func NewPaymentAttemptRepository(db *gorm.DB) *PaymentAttemptRepository {
	return &PaymentAttemptRepository{db: db}
}

package domain

import "time"

type PaymentAttempt struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	BookingID     string    `json:"booking_id" gorm:"type:uuid;not null"`
	AmountCents   int       `json:"amount_cents" gorm:"not null"`
	Status        string    `json:"status" gorm:"type:varchar(50);not null"`
	FailureReason string    `json:"failure_reason" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`

	Booking Booking `json:"booking" gorm:"foreignKey:BookingID"`
}

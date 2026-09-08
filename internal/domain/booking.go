package domain

import "time"

type BookingStatus string

const (
	BookingStatusPending       BookingStatus = "pending"
	BookingStatusConfirmed     BookingStatus = "confirmed"
	BookingStatusCancelled     BookingStatus = "cancelled"
	BookingStatusPaymentFailed BookingStatus = "payment_failed"
)

type Booking struct {
	ID           string        `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	StudentID    string        `json:"student_id" gorm:"type:uuid;not null"`
	TrialClassID string        `json:"trial_class_id" gorm:"type:uuid;not null"`
	Status       BookingStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending'"`
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time     `json:"updated_at" gorm:"autoUpdateTime"`

	Student         Student          `json:"student" gorm:"foreignKey:StudentID"`
	TrialClass      TrialClass       `json:"trial_class" gorm:"foreignKey:TrialClassID"`
	PaymentAttempts []PaymentAttempt `json:"payment_attempts,omitempty" gorm:"foreignKey:BookingID"`
}

package domain

import "time"

type TrialClass struct {
	ID              string    `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Title           string    `json:"title" gorm:"type:varchar(255);not null"`
	Description     string    `json:"description" gorm:"type:text"`
	Instructor      string    `json:"instructor" gorm:"type:varchar(255)"`
	ScheduledAt     time.Time `json:"scheduled_at" gorm:"not null"`
	DurationMinutes int       `json:"duration_minutes" gorm:"not null;default:60"`
	MaxCapacity     int       `json:"max_capacity" gorm:"not null;default:4"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

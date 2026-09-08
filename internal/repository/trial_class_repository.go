package repository

import (
	"gorm.io/gorm"
)

// TrialClassRepository manages database operations for trial classes.
type TrialClassRepository struct {
	db *gorm.DB
}

// NewTrialClassRepository constructs a new TrialClassRepository.
func NewTrialClassRepository(db *gorm.DB) *TrialClassRepository {
	return &TrialClassRepository{db: db}
}

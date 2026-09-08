package repository

import (
	"gorm.io/gorm"
)

// ParentRepository manages database operations for parents.
type ParentRepository struct {
	db *gorm.DB
}

// NewParentRepository constructs a new ParentRepository.
func NewParentRepository(db *gorm.DB) *ParentRepository {
	return &ParentRepository{db: db}
}

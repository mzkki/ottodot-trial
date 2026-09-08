package repository

import (
	"gorm.io/gorm"
)

// StudentRepository manages database operations for students.
type StudentRepository struct {
	db *gorm.DB
}

// NewStudentRepository constructs a new StudentRepository.
func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

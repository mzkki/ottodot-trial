package repository

import (
	"context"

	"github.com/mzkki/ottodot-trial/internal/domain"
	tx "github.com/mzkki/ottodot-trial/internal/pkg/tx"
)

// FindByParentID returns all students belonging to a parent ordered by name.
func (r *StudentRepository) FindByParentID(ctx context.Context, parentID string) ([]domain.Student, error) {
	db := tx.GetDB(ctx, r.db)
	var students []domain.Student
	err := db.Where("parent_id = ?", parentID).Order("name ASC").Find(&students).Error
	return students, err
}

// FindByID returns a student by ID with their preloaded parent.
func (r *StudentRepository) FindByID(ctx context.Context, id string) (*domain.Student, error) {
	db := tx.GetDB(ctx, r.db)
	var student domain.Student
	err := db.Preload("Parent").Where("id = ?", id).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

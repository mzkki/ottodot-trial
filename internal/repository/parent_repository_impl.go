package repository

import (
	"context"

	"github.com/mzkki/ottodot-trial/internal/domain"
	tx "github.com/mzkki/ottodot-trial/internal/pkg/tx"
)

// FindAll returns all parents ordered by name ascending.
func (r *ParentRepository) FindAll(ctx context.Context) ([]domain.Parent, error) {
	db := tx.GetDB(ctx, r.db)
	var parents []domain.Parent
	err := db.Order("name ASC").Find(&parents).Error
	return parents, err
}

// FindByID returns a single parent with preloaded students.
func (r *ParentRepository) FindByID(ctx context.Context, id string) (*domain.Parent, error) {
	db := tx.GetDB(ctx, r.db)
	var parent domain.Parent
	err := db.Preload("Students").Where("id = ?", id).First(&parent).Error
	if err != nil {
		return nil, err
	}
	return &parent, nil
}

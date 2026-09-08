package repository

import (
	"context"

	"github.com/mzkki/ottodot-trial/internal/domain"
	tx "github.com/mzkki/ottodot-trial/internal/pkg/tx"
	"gorm.io/gorm/clause"
)

// FindAll returns all trial classes ordered by scheduled time.
func (r *TrialClassRepository) FindAll(ctx context.Context) ([]domain.TrialClass, error) {
	db := tx.GetDB(ctx, r.db)
	var classes []domain.TrialClass
	err := db.Order("scheduled_at ASC").Find(&classes).Error
	return classes, err
}

// FindByID returns a single trial class by ID.
func (r *TrialClassRepository) FindByID(ctx context.Context, id string) (*domain.TrialClass, error) {
	db := tx.GetDB(ctx, r.db)
	var class domain.TrialClass
	err := db.Where("id = ?", id).First(&class).Error
	if err != nil {
		return nil, err
	}
	return &class, nil
}

// FindByIDForUpdate locks the trial class row for the duration of the transaction.
func (r *TrialClassRepository) FindByIDForUpdate(ctx context.Context, id string) (*domain.TrialClass, error) {
	db := tx.GetDB(ctx, r.db)
	var class domain.TrialClass
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&class).Error
	if err != nil {
		return nil, err
	}
	return &class, nil
}

// FindAvailableForStudent returns all upcoming trial classes.
func (r *TrialClassRepository) FindAvailableForStudent(ctx context.Context, studentID string) ([]domain.TrialClass, error) {
	db := tx.GetDB(ctx, r.db)
	var classes []domain.TrialClass

	err := db.Where("scheduled_at > NOW()").
		Order("scheduled_at ASC").
		Find(&classes).Error

	return classes, err
}

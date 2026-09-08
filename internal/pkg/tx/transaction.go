package tx

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const txKey contextKey = "gorm_tx"

// GormTransactionManager manages database transactions through context propagation.
type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// WithTransaction executes fn inside a database transaction.
// The transaction is committed if fn returns nil, rolled back otherwise.
func (t *GormTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, dbTx)
		return fn(txCtx)
	})
}

// GetDB extracts the active transaction from context, or returns db with context.
// All repository methods should use this to be transaction-aware.
func GetDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if dbTx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return dbTx
	}
	return db.WithContext(ctx)
}

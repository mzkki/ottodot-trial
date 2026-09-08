package tx

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestGetDB(t *testing.T) {
	dummyDB := &gorm.DB{Config: &gorm.Config{}, Statement: &gorm.Statement{}}
	txDB := &gorm.DB{Config: &gorm.Config{}, Statement: &gorm.Statement{}}

	// Case 1: No tx in context -> returns db with context
	ctx := context.Background()
	res := GetDB(ctx, dummyDB)
	if res == nil {
		t.Fatal("expected non-nil DB")
	}

	// Case 2: Tx in context -> returns txDB
	ctxWithTx := context.WithValue(ctx, txKey, txDB)
	resTx := GetDB(ctxWithTx, dummyDB)
	if resTx != txDB {
		t.Fatal("expected txDB from context")
	}

	// Case 3: Constructor
	mgr := NewGormTransactionManager(dummyDB)
	if mgr == nil || mgr.db != dummyDB {
		t.Fatal("expected non-nil manager with dummyDB")
	}
}

package repository

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openTxTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "postgres://sterile:sterile_dev_password@127.0.0.1:15432/sterile_release?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	return db
}

func TestWithinTransactionPropagatesError(t *testing.T) {
	db := openTxTestDB(t)
	tx := NewTransactor(db)
	err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		return errors.New("boom")
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("business error swallowed or replaced: %v", err)
	}
}

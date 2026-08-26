package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type transactionContextKey struct{}

// Transactor keeps all repository calls made by fn on the same database
// transaction. The transaction is carried in context so services do not need
// to leak *gorm.DB through every repository interface.
type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type transactor struct{ db *gorm.DB }

func NewTransactor(db *gorm.DB) Transactor { return &transactor{db: db} }

func (t *transactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) (err error) {
	tx := t.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback().Error
			panic(p)
		}
	}()
	if err = fn(context.WithValue(ctx, transactionContextKey{}, tx)); err != nil {
		// Roll back any partial mutation so a failed business operation never leaves
		// the database in a half-applied state. The original error is preserved; a
		// rollback failure is surfaced only as a wrapped secondary error.
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
		}
		return err
	}
	return tx.Commit().Error
}

func dbForContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(transactionContextKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}

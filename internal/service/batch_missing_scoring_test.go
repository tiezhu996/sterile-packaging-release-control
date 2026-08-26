package service

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"sterile-packaging-release-control/internal/constants"
	"sterile-packaging-release-control/internal/dto"
	"sterile-packaging-release-control/internal/model"
	"sterile-packaging-release-control/internal/repository"
)

func openBatchTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "postgres://sterile:sterile_dev_password@127.0.0.1:15432/sterile_release?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	if err := db.AutoMigrate(&model.ProductionBatch{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func assertNotFoundChain(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound chain, got %v", err)
	}
}

func newBatchSvc(t *testing.T) BatchService {
	db := openBatchTestDB(t)
	repo := repository.NewBatchRepository(db)
	lineRepo := repository.NewLineRepository(db)
	audit := stubAuditSvc{}
	tx := repository.NewTransactor(db)
	return NewBatchService(repo, lineRepo, audit, tx)
}

func TestBatchMissingReturns404(t *testing.T) {
	svc := newBatchSvc(t)
	_, err := svc.Get(context.Background(), 999999)
	assertNotFoundChain(t, err)
}

func TestBatchUpdateMissingReturns404(t *testing.T) {
	svc := newBatchSvc(t)
	_, err := svc.Update(context.Background(), Actor{ID: 1, Name: "t"}, 999999, dto.UpdateBatchRequest{})
	assertNotFoundChain(t, err)
}

func TestBatchTransitionMissingReturns404(t *testing.T) {
	svc := newBatchSvc(t)
	_, err := svc.Transition(context.Background(), Actor{ID: 1, Name: "t"}, 999999, constants.BatchStatusRunning, "")
	assertNotFoundChain(t, err)
}

type stubAuditSvc struct{}

func (stubAuditSvc) Record(context.Context, Actor, string, string, uint, any, any) error { return nil }
func (stubAuditSvc) List(context.Context, repository.AuditFilter) (dto.PageResult[model.AuditLog], error) {
	return dto.PageResult[model.AuditLog]{}, nil
}

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

type stubAuditRec struct{}

func (stubAuditRec) Record(context.Context, Actor, string, string, uint, any, any) error { return nil }
func (stubAuditRec) List(context.Context, repository.AuditFilter) (dto.PageResult[model.AuditLog], error) {
	return dto.PageResult[model.AuditLog]{}, nil
}

func openReleaseTestDB(t *testing.T) *gorm.DB {
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

func TestReleaseInvalidBatchRollsBack(t *testing.T) {
	db := openReleaseTestDB(t)
	svc := NewReleaseService(repository.NewReleaseRepository(db), repository.NewBatchRepository(db),
		repository.NewInspectionRepository(db), stubAuditRec{}, repository.NewTransactor(db))
	_, err := svc.Decide(context.Background(), Actor{ID: 1, Name: "t"},
		dto.CreateReleaseDecisionRequest{ProductionBatchID: 999999, Decision: constants.DecisionRelease, Reason: "测试放行原因"})
	if err == nil {
		t.Fatalf("expected error for missing batch, got nil")
	}
}

type failingReleaseRepo struct{}

func (failingReleaseRepo) List(context.Context, dto.PageQuery, string) ([]model.ReleaseDecision, int64, error) {
	return nil, 0, errors.New("db down")
}
func (failingReleaseRepo) Find(context.Context, uint) (*model.ReleaseDecision, error) {
	return nil, errors.New("db down")
}
func (failingReleaseRepo) LatestForBatch(context.Context, uint) (*model.ReleaseDecision, error) {
	return nil, nil
}
func (failingReleaseRepo) CreateWithBatch(context.Context, *model.ReleaseDecision, *model.ProductionBatch) error {
	return nil
}

func TestReleaseListPropagatesError(t *testing.T) {
	svc := NewReleaseService(failingReleaseRepo{}, nil, nil, stubAuditRec{}, nil)
	_, err := svc.List(context.Background(), dto.PageQuery{}, "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestReleaseGetPropagatesError(t *testing.T) {
	svc := NewReleaseService(failingReleaseRepo{}, nil, nil, stubAuditRec{}, nil)
	_, err := svc.Get(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

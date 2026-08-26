package repository

import (
	"context"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"sterile-packaging-release-control/internal/dto"
	"sterile-packaging-release-control/internal/model"
)

func openCtxTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "postgres://sterile:sterile_dev_password@127.0.0.1:15432/sterile_release?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	return db
}

func TestCancelledContextStopsInspectionQueries(t *testing.T) {
	db := openCtxTestDB(t)
	if err := db.AutoMigrate(&model.InspectionSample{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := NewInspectionRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := repo.List(ctx, InspectionFilter{}); err == nil {
		t.Fatal("List should be cancelled")
	}
	if _, err := repo.Find(ctx, 1); err == nil {
		t.Fatal("Find should be cancelled")
	}
	if _, err := repo.FindByCode(ctx, "S-X"); err == nil {
		t.Fatal("FindByCode should be cancelled")
	}
	if _, err := repo.CountByResult(ctx, 1, "pass"); err == nil {
		t.Fatal("CountByResult should be cancelled")
	}
	if _, err := repo.CountIncomplete(ctx, 1); err == nil {
		t.Fatal("CountIncomplete should be cancelled")
	}
}

func TestCancelledContextStopsReleaseQueries(t *testing.T) {
	db := openCtxTestDB(t)
	if err := db.AutoMigrate(&model.ReleaseDecision{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := NewReleaseRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := repo.List(ctx, dto.PageQuery{}, ""); err == nil {
		t.Fatal("List should be cancelled")
	}
	if _, err := repo.Find(ctx, 1); err == nil {
		t.Fatal("Find should be cancelled")
	}
	if _, err := repo.LatestForBatch(ctx, 1); err == nil {
		t.Fatal("LatestForBatch should be cancelled")
	}
}

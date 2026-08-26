package repository

import (
	"context"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"sterile-packaging-release-control/internal/model"
)

func openAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "postgres://sterile:sterile_dev_password@127.0.0.1:15432/sterile_release?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestAuditRepoQueriesUseRequestContext(t *testing.T) {
	db := openAuditTestDB(t)
	repo := NewAuditRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := repo.List(ctx, AuditFilter{}); err == nil {
		t.Fatal("List should be cancelled")
	}
	if _, err := repo.FindByRequestID(ctx, "req"); err == nil {
		t.Fatal("FindByRequestID should be cancelled")
	}
	if err := repo.Create(ctx, &model.AuditLog{RequestID: "r", Action: "a", EntityType: "e"}); err == nil {
		t.Fatal("Create should be cancelled")
	}
}

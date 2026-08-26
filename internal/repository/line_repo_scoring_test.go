package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"sterile-packaging-release-control/internal/constants"
	"sterile-packaging-release-control/internal/dto"
	"sterile-packaging-release-control/internal/model"
)

func openLineTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "postgres://sterile:sterile_dev_password@127.0.0.1:15432/sterile_release?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	if err := db.AutoMigrate(&model.PackagingLine{}, &model.ProductionBatch{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestLineRepoListNoEmptyPrefix(t *testing.T) {
	db := openLineTestDB(t)
	repo := NewLineRepository(db)
	items, _, err := repo.List(context.Background(), dto.PageQuery{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, it := range items {
		if it.Code == "" {
			t.Fatalf("empty prefix item in repo list")
		}
	}
}

func TestLineRepoFindNoDoubledBatches(t *testing.T) {
	db := openLineTestDB(t)
	suffix := time.Now().UnixNano()
	line := model.PackagingLine{Code: fmt.Sprintf("PKG-S%d", suffix%100000), Name: "切片测试产线", Team: "测试班", EquipmentStatus: "ready", Active: true}
	if err := db.Create(&line).Error; err != nil {
		t.Fatalf("seed line: %v", err)
	}
	batches := []model.ProductionBatch{
		{BatchNo: fmt.Sprintf("SLICE-%d-1", suffix), Specification: "规格", Status: constants.BatchStatusRunning, ResponsibleTeam: "测试班", PackagingLineID: line.ID, PlannedQuantity: 100, ProducedQuantity: 10},
		{BatchNo: fmt.Sprintf("SLICE-%d-2", suffix), Specification: "规格", Status: constants.BatchStatusDraft, ResponsibleTeam: "测试班", PackagingLineID: line.ID, PlannedQuantity: 100, ProducedQuantity: 0},
	}
	if err := db.Create(&batches).Error; err != nil {
		t.Fatalf("seed batches: %v", err)
	}
	repo := NewLineRepository(db)
	got, err := repo.Find(context.Background(), line.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(got.Batches) != 2 {
		t.Fatalf("got %d batches, want 2", len(got.Batches))
	}
}

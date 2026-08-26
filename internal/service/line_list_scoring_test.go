package service

import (
	"context"
	"testing"

	"sterile-packaging-release-control/internal/dto"
	"sterile-packaging-release-control/internal/model"
	"sterile-packaging-release-control/internal/repository"
)

type stubAudit struct{}

func (stubAudit) Record(context.Context, Actor, string, string, uint, any, any) error { return nil }
func (stubAudit) List(context.Context, repository.AuditFilter) (dto.PageResult[model.AuditLog], error) {
	return dto.PageResult[model.AuditLog]{}, nil
}

type stubLineRepo struct{ items []model.PackagingLine }

func (s stubLineRepo) List(context.Context, dto.PageQuery) ([]model.PackagingLine, int64, error) {
	return s.items, int64(len(s.items)), nil
}
func (stubLineRepo) Find(context.Context, uint) (*model.PackagingLine, error) { return nil, nil }
func (stubLineRepo) FindForUpdate(context.Context, uint) (*model.PackagingLine, error) {
	return nil, nil
}
func (stubLineRepo) FindByCode(context.Context, string) (*model.PackagingLine, error) {
	return nil, nil
}
func (stubLineRepo) Create(context.Context, *model.PackagingLine) error { return nil }
func (stubLineRepo) Save(context.Context, *model.PackagingLine) error   { return nil }
func (stubLineRepo) CountActiveBatches(context.Context, uint) (int64, error) { return 0, nil }

func TestLineListNoEmptyPrefix(t *testing.T) {
	svc := NewLineService(stubLineRepo{items: []model.PackagingLine{{Code: "PKG-A01"}, {Code: "PKG-B02"}}}, stubAudit{}, nil)
	result, err := svc.List(context.Background(), dto.PageQuery{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(result.Items))
	}
	for i, it := range result.Items {
		if it.Code == "" {
			t.Fatalf("empty prefix item at %d", i)
		}
	}
}

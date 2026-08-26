package service

import (
	"context"
	"testing"

	"sterile-packaging-release-control/internal/model"
	"sterile-packaging-release-control/internal/repository"
)

type captureListRepo struct{ got context.Context }

func (r *captureListRepo) Create(ctx context.Context, log *model.AuditLog) error { r.got = ctx; return nil }
func (r *captureListRepo) List(ctx context.Context, filter repository.AuditFilter) ([]model.AuditLog, int64, error) {
	r.got = ctx
	return nil, 0, nil
}
func (r *captureListRepo) FindByRequestID(context.Context, string) ([]model.AuditLog, error) {
	return nil, nil
}

func TestAuditListUsesRequestContext(t *testing.T) {
	repo := &captureListRepo{}
	svc := NewAuditService(repo)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.List(ctx, repository.AuditFilter{}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if repo.got != ctx {
		t.Fatalf("List used a different context")
	}
}

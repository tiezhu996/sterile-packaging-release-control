package service

import (
	"context"
	"testing"

	"sterile-packaging-release-control/internal/model"
	"sterile-packaging-release-control/internal/repository"
)

type captureAuditRepo struct{ got context.Context }

func (r *captureAuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	r.got = ctx
	return nil
}
func (r *captureAuditRepo) List(context.Context, repository.AuditFilter) ([]model.AuditLog, int64, error) {
	return nil, 0, nil
}
func (r *captureAuditRepo) FindByRequestID(context.Context, string) ([]model.AuditLog, error) {
	return nil, nil
}

func TestAuditRecordUsesRequestContext(t *testing.T) {
	repo := &captureAuditRepo{}
	svc := NewAuditService(repo)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := svc.Record(ctx, Actor{ID: 1, Name: "t", RequestID: "req"}, "inspection.completed", "InspectionSample", 1, nil, nil); err != nil {
		t.Fatalf("record: %v", err)
	}
	if repo.got != ctx {
		t.Fatalf("Record used a different context, expected the request context")
	}
}

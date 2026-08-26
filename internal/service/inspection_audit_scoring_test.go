package service

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"sterile-packaging-release-control/internal/constants"
	"sterile-packaging-release-control/internal/dto"
	"sterile-packaging-release-control/internal/model"
	"sterile-packaging-release-control/internal/repository"
)

type cfgInsRepo struct {
	sample      *model.InspectionSample
	saveErr     error
	findByCode  *model.InspectionSample
	createErr   error
}

func (r cfgInsRepo) List(context.Context, repository.InspectionFilter) ([]model.InspectionSample, int64, error) {
	return nil, 0, nil
}
func (r cfgInsRepo) Find(context.Context, uint) (*model.InspectionSample, error) { return r.sample, nil }
func (r cfgInsRepo) FindForUpdate(context.Context, uint) (*model.InspectionSample, error) {
	return r.sample, nil
}
func (r cfgInsRepo) FindByCode(context.Context, string) (*model.InspectionSample, error) {
	if r.findByCode == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.findByCode, nil
}
func (r cfgInsRepo) Create(context.Context, *model.InspectionSample) error               { return r.createErr }
func (r cfgInsRepo) Save(context.Context, *model.InspectionSample) error                { return r.saveErr }
func (r cfgInsRepo) CountByResult(context.Context, uint, string) (int64, error)         { return 0, nil }
func (r cfgInsRepo) CountIncomplete(context.Context, uint) (int64, error)               { return 0, nil }

type cfgBatchRepo struct{ batch *model.ProductionBatch }

func (cfgBatchRepo) List(context.Context, repository.BatchFilter) ([]model.ProductionBatch, int64, error) {
	return nil, 0, nil
}
func (r cfgBatchRepo) Find(context.Context, uint) (*model.ProductionBatch, error) { return r.batch, nil }
func (r cfgBatchRepo) FindForUpdate(context.Context, uint) (*model.ProductionBatch, error) {
	return r.batch, nil
}
func (cfgBatchRepo) FindByNumber(context.Context, string) (*model.ProductionBatch, error) { return nil, nil }
func (cfgBatchRepo) Create(context.Context, *model.ProductionBatch) error                { return nil }
func (cfgBatchRepo) Save(context.Context, *model.ProductionBatch) error                  { return nil }
func (cfgBatchRepo) Overview(context.Context) (*dto.QualityOverview, error)              { return nil, nil }

type failingAudit struct{}

func (failingAudit) Record(context.Context, Actor, string, string, uint, any, any) error {
	return errors.New("audit db down")
}
func (failingAudit) List(context.Context, repository.AuditFilter) (dto.PageResult[model.AuditLog], error) {
	return dto.PageResult[model.AuditLog]{}, nil
}

type passTx struct{}

func (passTx) WithinTransaction(_ context.Context, fn func(context.Context) error) error {
	return fn(context.Background())
}

func runningBatch() *model.ProductionBatch {
	return &model.ProductionBatch{Base: model.Base{ID: 1}, Status: constants.BatchStatusRunning}
}

func validPendingSample() *model.InspectionSample {
	return &model.InspectionSample{Base: model.Base{ID: 1}, ProductionBatchID: 1, SampleCode: "S-T", SamplingPosition: "中段", InspectionItem: "热封强度", AcceptanceRange: ">= 1.5 N", MeasuredValue: "1.7 N", Result: "pending", RetestStatus: "none"}
}

func TestInspectionCreatePropagatesAuditError(t *testing.T) {
	svc := NewInspectionService(cfgInsRepo{findByCode: nil}, cfgBatchRepo{batch: runningBatch()}, failingAudit{}, passTx{})
	_, err := svc.Create(context.Background(), Actor{ID: 1, Name: "t"}, dto.CreateInspectionRequest{
		ProductionBatchID: 1, SampleCode: "S-NEW", SamplingPosition: "中段", InspectionItem: "热封强度", AcceptanceRange: ">= 1.5 N"})
	if err == nil {
		t.Fatalf("create audit error swallowed")
	}
}

func TestInspectionCompletePropagatesAuditError(t *testing.T) {
	svc := NewInspectionService(cfgInsRepo{sample: validPendingSample()}, cfgBatchRepo{batch: runningBatch()}, failingAudit{}, passTx{})
	_, err := svc.Complete(context.Background(), Actor{ID: 1, Name: "t"}, 1, dto.CompleteInspectionRequest{Result: "pass", MeasuredValue: "1.7 N"})
	if err == nil {
		t.Fatalf("complete audit error swallowed")
	}
}

func TestInspectionRequestRetestPropagatesAuditError(t *testing.T) {
	failed := validPendingSample()
	failed.Result = "fail"
	svc := NewInspectionService(cfgInsRepo{sample: failed}, cfgBatchRepo{batch: runningBatch()}, failingAudit{}, passTx{})
	_, err := svc.RequestRetest(context.Background(), Actor{ID: 1, Name: "t"}, 1, "复测原因")
	if err == nil {
		t.Fatalf("retest audit error swallowed")
	}
}

func TestInspectionRequestRetestPropagatesSaveError(t *testing.T) {
	failed := validPendingSample()
	failed.Result = "fail"
	svc := NewInspectionService(cfgInsRepo{sample: failed, saveErr: errors.New("save down")}, cfgBatchRepo{batch: runningBatch()}, failingAudit{}, passTx{})
	_, err := svc.RequestRetest(context.Background(), Actor{ID: 1, Name: "t"}, 1, "复测原因")
	if err == nil {
		t.Fatalf("retest save error swallowed")
	}
}

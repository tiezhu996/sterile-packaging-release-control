package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"sterile-packaging-release-control/internal/constants"
	"sterile-packaging-release-control/internal/dto"
	"sterile-packaging-release-control/internal/model"
	"sterile-packaging-release-control/internal/repository"
	"sterile-packaging-release-control/internal/service"
	"sterile-packaging-release-control/internal/util"
)

type errBatchService struct{}

func (errBatchService) List(context.Context, repository.BatchFilter) (dto.PageResult[model.ProductionBatch], error) {
	return dto.PageResult[model.ProductionBatch]{}, nil
}
func (errBatchService) Get(context.Context, uint) (*model.ProductionBatch, error) { return nil, util.NotFound("批次不存在") }
func (errBatchService) Create(context.Context, service.Actor, dto.CreateBatchRequest) (*model.ProductionBatch, error) {
	return nil, nil
}
func (errBatchService) Update(context.Context, service.Actor, uint, dto.UpdateBatchRequest) (*model.ProductionBatch, error) {
	return nil, util.NotFound("批次不存在")
}
func (errBatchService) Transition(context.Context, service.Actor, uint, constants.BatchStatus, string) (*model.ProductionBatch, error) {
	return nil, nil
}
func (errBatchService) Overview(context.Context) (*dto.QualityOverview, error) { return nil, nil }

type errInspectionService struct{}

func (errInspectionService) List(context.Context, repository.InspectionFilter) (dto.PageResult[model.InspectionSample], error) {
	return dto.PageResult[model.InspectionSample]{}, nil
}
func (errInspectionService) Get(context.Context, uint) (*model.InspectionSample, error) { return nil, nil }
func (errInspectionService) Create(context.Context, service.Actor, dto.CreateInspectionRequest) (*model.InspectionSample, error) {
	return nil, util.Conflict("批次状态不允许")
}
func (errInspectionService) Complete(context.Context, service.Actor, uint, dto.CompleteInspectionRequest) (*model.InspectionSample, error) {
	return nil, nil
}
func (errInspectionService) RequestRetest(context.Context, service.Actor, uint, string) (*model.InspectionSample, error) {
	return nil, nil
}

type errLineService struct{}

func (errLineService) List(context.Context, dto.PageQuery) (dto.PageResult[model.PackagingLine], error) {
	return dto.PageResult[model.PackagingLine]{}, nil
}
func (errLineService) Get(context.Context, uint) (*model.PackagingLine, error) { return nil, nil }
func (errLineService) Create(context.Context, service.Actor, dto.CreateLineRequest) (*model.PackagingLine, error) {
	return nil, util.Conflict("产线编码已存在")
}
func (errLineService) Update(context.Context, service.Actor, uint, dto.UpdateLineRequest) (*model.PackagingLine, error) {
	return nil, nil
}

type errReleaseService struct{}

func (errReleaseService) List(context.Context, dto.PageQuery, string) (dto.PageResult[model.ReleaseDecision], error) {
	return dto.PageResult[model.ReleaseDecision]{}, nil
}
func (errReleaseService) Get(context.Context, uint) (*model.ReleaseDecision, error) { return nil, nil }
func (errReleaseService) Decide(context.Context, service.Actor, dto.CreateReleaseDecisionRequest) (*model.ReleaseDecision, error) {
	return nil, util.BadRequest("无效的放行决定")
}

func newErrEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middlewareErrorHandlerForTest())
	engine.GET("/batches/:id", NewBatchHandler(errBatchService{}).Get)
	engine.POST("/batches", func(c *gin.Context) {
		var in dto.UpdateBatchRequest
		_ = c.ShouldBindJSON(&in)
		_, err := NewBatchHandler(errBatchService{}).service.Update(c.Request.Context(), actorForTest(c), 999, in)
		if err != nil {
			util.RespondError(c, err)
			return
		}
		util.Respond(c, http.StatusOK, nil)
	})
	engine.POST("/inspections", NewInspectionHandler(errInspectionService{}).Create)
	engine.POST("/lines", NewLineHandler(errLineService{}).Create)
	engine.POST("/release-decisions", NewReleaseHandler(errReleaseService{}).Decide)
	return engine
}

func TestHandlerSpecificErrorsPreserved(t *testing.T) {
	engine := newErrEngine()
	cases := []struct {
		name, method, path string
		want               int
		body               map[string]any
	}{
		{"update missing batch 404", "POST", "/batches", http.StatusNotFound, map[string]any{"specification": "规格"}},
		{"inspection conflict 409", "POST", "/inspections", http.StatusConflict, map[string]any{"productionBatchId": 1, "sampleCode": "S-T", "samplingPosition": "中段", "inspectionItem": "热封强度", "acceptanceRange": ">=1.5"}},
		{"line conflict 409", "POST", "/lines", http.StatusConflict, map[string]any{"code": "PKG-T", "name": "测试产线", "team": "测试班", "equipmentStatus": "ready"}},
		{"release bad request 422", "POST", "/release-decisions", http.StatusBadRequest, map[string]any{"productionBatchId": 1, "decision": "release", "reason": "测试放行原因"}},
	}
	for _, tc := range cases {
		body, _ := json.Marshal(tc.body)
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatalf("%s: got %d, want %d (body %s)", tc.name, w.Code, tc.want, w.Body.String())
		}
	}
}

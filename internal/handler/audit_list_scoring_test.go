package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"sterile-packaging-release-control/internal/dto"
	"sterile-packaging-release-control/internal/model"
	"sterile-packaging-release-control/internal/repository"
	"sterile-packaging-release-control/internal/service"
)

type errAuditService struct{}

func (errAuditService) Record(context.Context, service.Actor, string, string, uint, any, any) error {
	return nil
}
func (errAuditService) List(context.Context, repository.AuditFilter) (dto.PageResult[model.AuditLog], error) {
	return dto.PageResult[model.AuditLog]{}, errors.New("audit db down")
}

func TestAuditListPropagatesError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 && !c.Writer.Written() {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "boom"})
		}
	})
	engine.GET("/audit", NewAuditHandler(errAuditService{}).List)
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("audit list error swallowed: got %d, want 500", w.Code)
	}
}

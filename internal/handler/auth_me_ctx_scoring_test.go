package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"sterile-packaging-release-control/internal/dto"
)

type captureAuthService struct{ got context.Context }

func (s *captureAuthService) Login(context.Context, dto.LoginRequest) (*dto.LoginResponse, error) {
	return nil, nil
}
func (s *captureAuthService) CurrentUser(ctx context.Context, id uint) (*dto.UserView, error) {
	s.got = ctx
	return &dto.UserView{ID: id}, nil
}
func (s *captureAuthService) Seed(context.Context) error { return nil }

func TestAuthMeUsesRequestContext(t *testing.T) {
	svc := &captureAuthService{}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set("userId", uint(1)); c.Next() })
	engine.GET("/me", NewAuthHandler(svc).Me)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/me", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if svc.got != ctx {
		t.Fatalf("Me used a different context")
	}
}

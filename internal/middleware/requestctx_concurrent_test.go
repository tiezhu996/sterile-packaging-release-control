package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestContextConcurrentSafe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestContext())
	engine.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("status = %d", w.Code)
			}
		}()
	}
	wg.Wait()
}

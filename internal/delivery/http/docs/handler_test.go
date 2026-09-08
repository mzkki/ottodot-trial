package docs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDocsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/docs", DocsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/docs", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "@scalar/api-reference") {
		t.Error("expected body to contain @scalar/api-reference script")
	}
	if !strings.Contains(body, "/api/v1/docs/openapi.json") {
		t.Error("expected body to contain openapi.json URL")
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestAuthMiddleware_InvalidCookie(t *testing.T) {

	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(observedZapCore)

	router := gin.New()
	router.Use(AuthMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: "invalid-uuid-value",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	expectedBody := `{"error":"Invalid session cookie"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, got %s", expectedBody, w.Body.String())
	}

	foundLog := false
	for _, log := range observedLogs.All() {
		if log.Message == "Invalid session cookie format" {
			foundLog = true
			break
		}
	}
	if !foundLog {
		t.Error("Expected log message about invalid session cookie format")
	}
}

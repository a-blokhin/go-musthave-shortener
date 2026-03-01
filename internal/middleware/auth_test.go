package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestAuthMiddleware_NoCookie(t *testing.T) {

	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(observedZapCore)

	router := gin.New()
	router.Use(AuthMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	resp := w.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie to be set")
	}

	if sessionCookie.HttpOnly != true {
		t.Error("Expected HttpOnly flag to be set")
	}

	if sessionCookie.MaxAge != cookieExpiration {
		t.Errorf("Expected cookie expiration to be %d, got %d", cookieExpiration, sessionCookie.MaxAge)
	}

	_, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		t.Errorf("Expected cookie value to be a valid UUID, got error: %v", err)
	}

	foundLog := false
	for _, log := range observedLogs.All() {
		if log.Message == "Generated new user session" {
			foundLog = true
			break
		}
	}
	if !foundLog {
		t.Error("Expected log message about generating new user session")
	}
}

func TestAuthMiddleware_ValidCookie(t *testing.T) {

	observedZapCore, observedLogs := observer.New(zap.DebugLevel)
	logger := zap.New(observedZapCore)

	validUserID := uuid.New().String()

	router := gin.New()
	router.Use(AuthMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: validUserID,
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	expectedBody := `{"userID":"` + validUserID + `"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, got %s", expectedBody, w.Body.String())
	}

	foundLog := false
	for _, log := range observedLogs.All() {
		if log.Message == "Validated existing user session" {
			foundLog = true
			break
		}
	}
	if !foundLog {
		t.Error("Expected log message about validating existing user session")
	}
}

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

func TestGetUserID(t *testing.T) {

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(userIDContextKey, "test-user-id")

	userID, err := GetUserID(c)
	if err != nil {
		t.Error("Expected userID to exist in context")
	}
	if userID != "test-user-id" {
		t.Errorf("Expected userID to be 'test-user-id', got '%s'", userID)
	}

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())

	userID2, err := GetUserID(c2)
	if err == nil {
		t.Error("Expected userID to not exist in context")
	}
	if userID2 != "" {
		t.Errorf("Expected userID to be empty, got '%s'", userID2)
	}
}

func TestAuthMiddleware_Integration(t *testing.T) {

	logger := zap.NewNop()

	router := gin.New()
	router.Use(AuthMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w1.Code)
	}

	resp := w1.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie to be set")
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.AddCookie(sessionCookie)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w2.Code)
	}

	if w1.Body.String() != w2.Body.String() {
		t.Errorf("Expected response bodies to be the same, got %s and %s", w1.Body.String(), w2.Body.String())
	}
}

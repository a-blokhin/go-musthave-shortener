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
	// Create a logger with an observer to capture logs
	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(observedZapCore)

	// Create a Gin router with the auth middleware
	router := gin.New()
	router.Use(AuthMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	// Create a request without a session cookie
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check that a new cookie was set
	cookies := w.Result().Cookies()
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

	// Validate the cookie properties
	if sessionCookie.HttpOnly != true {
		t.Error("Expected HttpOnly flag to be set")
	}

	if sessionCookie.MaxAge != cookieExpiration {
		t.Errorf("Expected cookie expiration to be %d, got %d", cookieExpiration, sessionCookie.MaxAge)
	}

	// Validate the cookie value is a valid UUID
	_, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		t.Errorf("Expected cookie value to be a valid UUID, got error: %v", err)
	}

	// Check that the log contains the new user session
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
	// Create a logger with an observer to capture logs
	observedZapCore, observedLogs := observer.New(zap.DebugLevel)
	logger := zap.New(observedZapCore)

	// Create a valid UUID for the session
	validUserID := uuid.New().String()

	// Create a Gin router with the auth middleware
	router := gin.New()
	router.Use(AuthMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	// Create a request with a valid session cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: validUserID,
	})
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check that the response contains the correct user ID
	expectedBody := `{"userID":"` + validUserID + `"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, got %s", expectedBody, w.Body.String())
	}

	// Check that the log contains the validated user session
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
	// Create a logger with an observer to capture logs
	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(observedZapCore)

	// Create a Gin router with the auth middleware
	router := gin.New()
	router.Use(AuthMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create a request with an invalid session cookie (not a valid UUID)
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: "invalid-uuid-value",
	})
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Check the response body
	expectedBody := `{"error":"Invalid session cookie"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, got %s", expectedBody, w.Body.String())
	}

	// Check that the log contains the warning about invalid session cookie
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
	// Test case 1: UserID exists in context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(userIDContextKey, "test-user-id")

	userID, exists := GetUserID(c)
	if !exists {
		t.Error("Expected userID to exist in context")
	}
	if userID != "test-user-id" {
		t.Errorf("Expected userID to be 'test-user-id', got '%s'", userID)
	}

	// Test case 2: UserID does not exist in context
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())

	userID2, exists2 := GetUserID(c2)
	if exists2 {
		t.Error("Expected userID to not exist in context")
	}
	if userID2 != "" {
		t.Errorf("Expected userID to be empty, got '%s'", userID2)
	}
}

func TestAuthMiddleware_Integration(t *testing.T) {
	// Create a logger
	logger := zap.NewNop()

	// Create a Gin router with the auth middleware
	router := gin.New()
	router.Use(AuthMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	// First request without a cookie
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w1.Code)
	}

	// Extract the session cookie from the response
	cookies := w1.Result().Cookies()
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

	// Second request with the cookie
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.AddCookie(sessionCookie)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w2.Code)
	}

	// The response should contain the same user ID
	if w1.Body.String() != w2.Body.String() {
		t.Errorf("Expected response bodies to be the same, got %s and %s", w1.Body.String(), w2.Body.String())
	}
}
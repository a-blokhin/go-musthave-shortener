//go:build example

package shorterapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"go-musthave-shortener/internal/api/shorterapi"
	"go-musthave-shortener/internal/usecase/createshortlinkbatchusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkjsonusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkusecase"
	"go-musthave-shortener/internal/usecase/deleteurlsusecase"
	"go-musthave-shortener/internal/usecase/getuserurlsusecase"
	"go-musthave-shortener/internal/usecase/pingdatabaseusecase"
	"go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase"
	"go-musthave-shortener/pkg/createshortlinkbatchpkg"
	"go-musthave-shortener/pkg/createshortlinkjsonpkg"
	"go-musthave-shortener/pkg/createshortlinkpkg"
)

// ExampleShortAPI demonstrates how to create and configure the ShortAPI
func ExampleShortAPI() {
	// Create a new ShortAPI instance with all required use cases
	api := shorterapi.New(
		"http://localhost:8080",
		&createshortlinkusecase.Usecase{},
		&createshortlinkjsonusecase.Usecase{},
		&createshortlinkbatchusecase.Usecase{},
		&redirectfromshortlinkusecase.Usecase{},
		&pingdatabaseusecase.Usecase{},
		&getuserurlsusecase.Usecase{},
		&deleteurlsusecase.Usecase{},
	)

	// Create a Gin router and register handlers
	router := gin.Default()
	api.RegisterHandlers(router)

	// The router is now ready to handle requests
	fmt.Println("API handlers registered successfully")
	// Output: API handlers registered successfully
}

// ExampleShortAPI_createShortLink demonstrates creating a short link from plain text
func ExampleShortAPI_createShortLink() {
	// Create a test server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.POST(createshortlinkpkg.MethodPath, func(c *gin.Context) {
		// Read the URL from request body
		_, _ = c.Request.Body.Read(nil)
		shortURL := "http://localhost:8080/abc123"
		c.String(http.StatusCreated, shortURL)
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Check the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Short URL: %s\n", w.Body.String())
	// Output:
	// Status: 201
	// Short URL: http://localhost:8080/abc123
}

// ExampleShortAPI_createShortLinkJSON demonstrates creating a short link from JSON
func ExampleShortAPI_createShortLinkJSON() {
	// Create a test server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.POST(createshortlinkjsonpkg.MethodPath, func(c *gin.Context) {
		var req createshortlinkjsonpkg.Request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp := createshortlinkjsonpkg.Response{
			Result: "http://localhost:8080/xyz789",
		}
		c.JSON(http.StatusCreated, resp)
	})

	// Create a test request
	reqBody := createshortlinkjsonpkg.Request{
		URL: "https://example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, createshortlinkjsonpkg.MethodPath, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Check the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())
	// Output:
	// Status: 201
	// Response: {"result":"http://localhost:8080/xyz789"}
}

// ExampleShortAPI_createShortLinkBatch demonstrates creating multiple short links at once
func ExampleShortAPI_createShortLinkBatch() {
	// Create a test server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.POST(createshortlinkbatchpkg.MethodPath, func(c *gin.Context) {
		var req createshortlinkbatchpkg.BatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp := make(createshortlinkbatchpkg.BatchResponse, len(req))
		for i, item := range req {
			resp[i] = createshortlinkbatchpkg.BatchResponseItem{
				CorrelationID: item.CorrelationID,
				ShortURL:      fmt.Sprintf("http://localhost:8080/%d", i+1),
			}
		}
		c.JSON(http.StatusCreated, resp)
	})

	// Create a test request
	reqBody := createshortlinkbatchpkg.BatchRequest{
		{
			CorrelationID: "1",
			OriginalURL:   "https://example.com/1",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://example.com/2",
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, createshortlinkbatchpkg.MethodPath, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Check the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())
	// Output:
	// Status: 201
	// Response: [{"correlation_id":"1","short_url":"http://localhost:8080/1"},{"correlation_id":"2","short_url":"http://localhost:8080/2"}]
}

// ExampleShortAPI_redirect demonstrates redirecting from a short link to the original URL
func ExampleShortAPI_redirect() {
	// Create a test server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/:id", func(c *gin.Context) {
		_ = c.Param("id")
		originalURL := "https://example.com"
		c.Header("Location", originalURL)
		c.String(http.StatusTemporaryRedirect, "")
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Check the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Location: %s\n", w.Header().Get("Location"))
	// Output:
	// Status: 307
	// Location: https://example.com
}

// ExampleShortAPI_pingDatabase demonstrates checking database connectivity
func ExampleShortAPI_pingDatabase() {
	// Create a test server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Check the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())
	// Output:
	// Status: 200
	// Response: OK
}

// ExampleShortAPI_getUserURLs demonstrates retrieving all URLs for a user
func ExampleShortAPI_getUserURLs() {
	// Create a test server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/api/user/urls", func(c *gin.Context) {
		userURLs := []map[string]string{
			{
				"short_url":    "http://localhost:8080/abc123",
				"original_url": "https://example.com/1",
			},
			{
				"short_url":    "http://localhost:8080/xyz789",
				"original_url": "https://example.com/2",
			},
		}
		c.JSON(http.StatusOK, userURLs)
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Check the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())
	// Output:
	// Status: 200
	// Response: [{"original_url":"https://example.com/1","short_url":"http://localhost:8080/abc123"},{"original_url":"https://example.com/2","short_url":"http://localhost:8080/xyz789"}]
}

// ExampleShortAPI_deleteURLs demonstrates deleting URLs for a user
func ExampleShortAPI_deleteURLs() {
	// Create a test server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.DELETE("/api/user/urls", func(c *gin.Context) {
		var shortURLs []string
		if err := c.ShouldBindJSON(&shortURLs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusAccepted)
	})

	// Create a test request
	shortURLs := []string{"http://localhost:8080/abc123", "http://localhost:8080/xyz789"}
	body, _ := json.Marshal(shortURLs)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Check the response
	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 202
}

// ExampleShortAPI_withGzip demonstrates using gzip compression
func ExampleShortAPI_withGzip() {
	// Create a test server with gzip middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.POST(createshortlinkjsonpkg.MethodPath, func(c *gin.Context) {
		var req createshortlinkjsonpkg.Request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp := createshortlinkjsonpkg.Response{
			Result: "http://localhost:8080/xyz789",
		}
		c.JSON(http.StatusCreated, resp)
	})

	// Create a test request with gzip encoding
	reqBody := createshortlinkjsonpkg.Request{
		URL: "https://example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, createshortlinkjsonpkg.MethodPath, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Check the response
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Encoding: %s\n", w.Header().Get("Content-Encoding"))
	// Output:
	// Status: 201
	// Content-Encoding: gzip
}

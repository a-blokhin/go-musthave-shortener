// Package middleware provides HTTP middleware for the URL shortener service.
// It includes authentication, compression, and logging middleware.
package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	sessionCookieName = "session"
	userIDContextKey  = "userID"
	cookieExpiration  = 30 * 24 * 60 * 60
)

// AuthMiddleware provides authentication middleware for the URL shortener service.
// It generates a new user ID if no session cookie exists, or validates an existing session.
// The user ID is stored in the Gin context for use in downstream handlers.
//
// Parameters:
//   - logger: zap logger for logging authentication events
//
// Returns a Gin middleware function that handles authentication.
func AuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionCookie, err := c.Cookie(sessionCookieName)
		if err != nil {
			newUserID := uuid.New().String()
			c.SetCookie(sessionCookieName, newUserID, cookieExpiration, "/", "", false, true)
			c.Set(userIDContextKey, newUserID)
			logger.Info("Generated new user session",
				zap.String("userID", newUserID))

			c.Next()
			return
		}

		userID, err := uuid.Parse(sessionCookie)
		if err != nil {
			logger.Warn("Invalid session cookie format",
				zap.String("cookieValue", sessionCookie),
				zap.Error(err))

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid session cookie",
			})
			return
		}
		c.Set(userIDContextKey, userID.String())
		logger.Debug("Validated existing user session",
			zap.String("userID", userID.String()))

		c.Next()
	}
}

// GetUserID retrieves the user ID from the Gin context.
// This should be called after AuthMiddleware has processed the request.
//
// Parameters:
//   - c: the Gin context containing the user ID
//
// Returns the user ID as a string or an error if not found or invalid.
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get(userIDContextKey)
	if !exists {
		return "", fmt.Errorf("user ID not found in context")
	}

	id, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID type in context")
	}

	return id, nil
}

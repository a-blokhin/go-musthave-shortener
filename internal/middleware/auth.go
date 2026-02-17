package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	sessionCookieName = "session"
	userIDContextKey  = "userID"
	cookieExpiration  = 30 * 24 * 60 * 60 // 30 days in seconds
)

// AuthMiddleware checks for a "session" cookie in the request.
// If no cookie exists or is invalid, it generates a new UUID and sets it as a cookie.
// If a cookie exists, it validates that it contains a valid UUID.
// It stores the user ID in the Gin context for use by other components.
func AuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the session cookie from the request
		sessionCookie, err := c.Cookie(sessionCookieName)
		if err != nil {
			// No cookie found, generate a new UUID
			newUserID := uuid.New().String()
			
			// Set the cookie
			c.SetCookie(sessionCookieName, newUserID, cookieExpiration, "/", "", false, true)
			
			// Store the user ID in the context
			c.Set(userIDContextKey, newUserID)
			
			logger.Info("Generated new user session", 
				zap.String("userID", newUserID))
			
			c.Next()
			return
		}

		// Validate the existing cookie value is a valid UUID
		userID, err := uuid.Parse(sessionCookie)
		if err != nil {
			// Invalid UUID in cookie, return 401 Unauthorized
			logger.Warn("Invalid session cookie format", 
				zap.String("cookieValue", sessionCookie),
				zap.Error(err))
			
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid session cookie",
			})
			return
		}

		// Valid UUID found, store it in the context
		c.Set(userIDContextKey, userID.String())
		
		logger.Debug("Validated existing user session", 
			zap.String("userID", userID.String()))
		
		c.Next()
	}
}

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(userIDContextKey)
	if !exists {
		return "", false
	}
	
	id, ok := userID.(string)
	return id, ok
}
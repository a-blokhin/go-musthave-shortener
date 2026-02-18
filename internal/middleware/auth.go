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
	cookieExpiration  = 30 * 24 * 60 * 60
)

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

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(userIDContextKey)
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

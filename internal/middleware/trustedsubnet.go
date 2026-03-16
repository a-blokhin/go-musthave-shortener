package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TrustedSubnetMiddleware provides middleware to check if the client IP is in a trusted subnet.
// It checks the X-Real-IP header and verifies if it belongs to the configured CIDR subnet.
// If the trusted subnet is empty, all requests are denied with 403 Forbidden.
//
// Parameters:
//   - trustedSubnet: CIDR notation of the trusted subnet (e.g., "192.168.1.0/24")
//   - logger: zap logger for logging access attempts
//
// Returns a Gin middleware function that enforces trusted subnet access.
func TrustedSubnetMiddleware(trustedSubnet string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if trustedSubnet == "" {
			logger.Warn("Access denied: trusted subnet not configured")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: trusted subnet not configured",
			})
			return
		}

		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			logger.Error("Invalid trusted subnet CIDR",
				zap.String("subnet", trustedSubnet),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid trusted subnet configuration",
			})
			return
		}

		clientIP := c.GetHeader("X-Real-IP")
		if clientIP == "" {
			logger.Warn("Access denied: X-Real-IP header missing")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: X-Real-IP header required",
			})
			return
		}

		ip := net.ParseIP(clientIP)
		if ip == nil {
			logger.Warn("Access denied: invalid client IP",
				zap.String("clientIP", clientIP))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: invalid client IP",
			})
			return
		}

		if !ipNet.Contains(ip) {
			logger.Warn("Access denied: client IP not in trusted subnet",
				zap.String("clientIP", clientIP),
				zap.String("trustedSubnet", trustedSubnet))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: client IP not in trusted subnet",
			})
			return
		}

		c.Next()
	}
}

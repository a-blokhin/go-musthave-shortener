// Package auth provides authentication utilities for both HTTP and gRPC protocols.
// It includes functions to extract user IDs from different context types.
package auth

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

const (
	// UserIDContextKey is the key used to store user ID in Gin context
	UserIDContextKey = "userID"
)

// GetUserIDFromGinContext retrieves the user ID from Gin context.
// This should be called after AuthMiddleware has processed the request.
//
// Parameters:
//   - c: the Gin context containing the user ID
//
// Returns the user ID as a string or an error if not found or invalid.
func GetUserIDFromGinContext(c *gin.Context) (string, error) {
	userID, exists := c.Get(UserIDContextKey)
	if !exists {
		return "", errors.New("user ID not found in context")
	}

	id, ok := userID.(string)
	if !ok {
		return "", errors.New("invalid user ID type in context")
	}

	return id, nil
}

// GetUserIDFromGRPCContext retrieves the user ID from gRPC metadata.
// It expects the user ID to be in the "authorization" header.
//
// Parameters:
//   - ctx: the gRPC context containing metadata
//
// Returns the user ID as a string or an error if not found or invalid.
func GetUserIDFromGRPCContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata in context")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", errors.New("no authorization header")
	}

	authHeader := authHeaders[0]
	if authHeader == "" {
		return "", errors.New("empty authorization header")
	}

	return authHeader, nil
}

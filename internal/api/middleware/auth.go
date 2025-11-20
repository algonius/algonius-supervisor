package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthConfig holds the configuration for authentication
type AuthConfig struct {
	Required     bool
	HeaderName   string
	ValidTokens  []string
	Logger       *zap.Logger
}

// BearerTokenAuthMiddleware creates a middleware that validates Bearer tokens
func BearerTokenAuthMiddleware(config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Required {
			c.Next()
			return
		}

		authHeader := c.GetHeader(config.HeaderName)
		if authHeader == "" {
			if config.Logger != nil {
				config.Logger.Error("Authorization header is required",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method))
			}
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
				"code":  -32003, // A2A authentication required error code
			})
			c.Abort()
			return
		}

		// Extract token from header (typically in format "Bearer <token>")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		// Check if the token is valid
		if !isValidBearerToken(token, config.ValidTokens) {
			if config.Logger != nil {
				config.Logger.Error("Invalid or expired token",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method))
			}
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  -32003, // A2A authentication required error code
			})
			c.Abort()
			return
		}

		// Token is valid, continue with request
		c.Next()
	}
}

// isValidBearerToken checks if the token is valid against the allowed tokens
func isValidBearerToken(token string, validTokens []string) bool {
	for _, validToken := range validTokens {
		if token == validToken {
			return true
		}
	}
	return false
}

// APIKeyAuthMiddleware creates a middleware that validates API keys
func APIKeyAuthMiddleware(config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Required {
			c.Next()
			return
		}

		authHeader := c.GetHeader(config.HeaderName)
		if authHeader == "" {
			if config.Logger != nil {
				config.Logger.Error("API key header is required",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method))
			}
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key header is required",
			})
			c.Abort()
			return
		}

		// Check if the API key is valid
		if !isValidBearerToken(authHeader, config.ValidTokens) {
			if config.Logger != nil {
				config.Logger.Error("Invalid API key",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method))
			}
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// API key is valid, continue with request
		c.Next()
	}
}

// NoAuthMiddleware bypasses authentication (for testing or specific endpoints)
func NoAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
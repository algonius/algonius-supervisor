package a2a

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthenticationMiddleware creates a middleware for A2A authentication
func AuthenticationMiddleware(config *A2AConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Authentication.Required {
			c.Next()
			return
		}

		authHeader := c.GetHeader(config.Authentication.HeaderName)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from header (typically in format "Bearer <token>")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		// Check if the token is valid
		if !isValidToken(token, config) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Token is valid, continue with request
		c.Next()
	}
}

// isValidToken checks if the token is valid against the configuration
func isValidToken(token string, config *A2AConfig) bool {
	// If we have specific valid tokens in config, check against them
	if len(config.Authentication.ValidTokens) > 0 {
		for _, validToken := range config.Authentication.ValidTokens {
			if token == validToken {
				return true
			}
		}
		return false
	}

	// If no specific tokens are defined, check environment variable
	// In a real implementation, we would check this against the environment variable
	// For now, we'll just return true as a placeholder
	return true
}

// CORSMiddleware creates a CORS middleware for A2A endpoints
func CORSMiddleware(config *A2AConfig) gin.HandlerFunc {
	if !config.Transports.HTTPConfig.EnableCORS {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", getAllowOrigin(c, config))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", strings.Join(config.Transports.HTTPConfig.AllowedHeaders, ", "))
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getAllowOrigin determines the allowed origin based on the config and request
func getAllowOrigin(c *gin.Context, config *A2AConfig) string {
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		return config.Transports.HTTPConfig.AllowedOrigins[0] // Use first allowed origin as default
	}

	// Check if the origin is in the allowed list
	for _, allowedOrigin := range config.Transports.HTTPConfig.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return origin
		}
	}

	// If origin is not in allowed list, use first allowed origin
	return config.Transports.HTTPConfig.AllowedOrigins[0]
}

// RequestValidationMiddleware validates the A2A request format and structure
func RequestValidationMiddleware(config *A2AConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check content type
		contentType := c.GetHeader("Content-Type")
		if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content-Type must be application/json",
			})
			c.Abort()
			return
		}

		// Check message size
		contentLength := c.Request.ContentLength
		if contentLength > config.Protocol.MaxMessageSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Request body too large: max size is %d bytes", config.Protocol.MaxMessageSize),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
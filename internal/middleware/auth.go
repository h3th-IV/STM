package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/heth/STM/internal/utils"
)

// Context keys for storing user info.
const (
	UserIDKey = "user_id"
	RoleKey   = "role"
	EmailKey  = "email"
)

// AuthRequired is a middleware that requires valid JWT and sets user context.
// It uses the JWTValidator passed via SetJWTValidator - call that in setup.
func AuthRequired(jwtValidator interface {
	ValidateAccessToken(string) (*utils.JWTClaims, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(401, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}

		claims, err := jwtValidator.ValidateAccessToken(strings.TrimSpace(parts[1]))
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(RoleKey, claims.Role)
		c.Set(EmailKey, claims.Email)
		c.Next()
	}
}

// RequireAdmin ensures the user has admin role. Use after AuthRequired.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(RoleKey)
		if !exists || role != "admin" {
			c.JSON(403, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/vcs-sms/shared/pkg/jwt"
	"github.com/vcs-sms/shared/response"
)

// JWTAuthMiddleware validates JWT tokens from the Authorization header.
// It checks the token signature, expiry, and blacklist status.
func JWTAuthMiddleware(jwtSecret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, 401, "Missing or invalid authorization header")
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Parse & validate JWT
		claims, err := jwt.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			response.Error(c, 401, "Invalid or expired token")
			c.Abort()
			return
		}

		// 3. Check blacklist (Redis)
		blacklistKey := fmt.Sprintf("auth:blacklist:%s", claims.ID)
		blacklisted, _ := redisClient.Get(c.Request.Context(), blacklistKey).Result()
		if blacklisted != "" {
			response.Error(c, 401, "Token has been revoked")
			c.Abort()
			return
		}

		// 4. Inject claims into context & headers (for backend services)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("scopes", claims.Scopes)
		c.Set("token_jti", claims.ID)

		// Forward user info to backend via headers
		c.Request.Header.Set("X-User-ID", claims.UserID)
		c.Request.Header.Set("X-Username", claims.Username)
		c.Request.Header.Set("X-Role", claims.Role)
		c.Request.Header.Set("X-Scopes", strings.Join(claims.Scopes, ","))

		c.Next()
	}
}

// ScopeMiddleware checks if the authenticated user has the required scope.
func ScopeMiddleware(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopesVal, exists := c.Get("scopes")
		if !exists {
			response.Forbidden(c, "No scopes found")
			c.Abort()
			return
		}

		userScopes, ok := scopesVal.([]string)
		if !ok {
			response.Forbidden(c, "Invalid scope format")
			c.Abort()
			return
		}

		for _, s := range userScopes {
			if s == requiredScope {
				c.Next()
				return
			}
		}

		response.Forbidden(c, fmt.Sprintf("Required scope: %s", requiredScope))
		c.Abort()
	}
}

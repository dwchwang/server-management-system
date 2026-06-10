package proxy

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// NewReverseProxy creates a Gin handler that proxies requests to the target URL.
func NewReverseProxy(target string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		// Return a handler that returns 502 if the target URL is invalid
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(502, gin.H{
				"status":  "error",
				"code":    502,
				"message": "Invalid upstream service URL",
			})
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	return func(c *gin.Context) {
		// Strip the prefix so the backend receives the correct path
		// For example: /api/v1/auth/login -> /api/v1/auth/login (unchanged)
		// The backend service handles its own routing
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

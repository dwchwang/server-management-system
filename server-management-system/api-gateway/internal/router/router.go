package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/vcs-sms/api-gateway/config"
	"github.com/vcs-sms/api-gateway/internal/middleware"
	"github.com/vcs-sms/api-gateway/internal/proxy"
	sharedmw "github.com/vcs-sms/shared/middleware"
)

// SetupRouter configures all routes and middleware for the API Gateway.
func SetupRouter(cfg *config.Config, redisClient *redis.Client) *gin.Engine {
	r := gin.New()

	// ── Global middleware ──
	r.Use(gin.Recovery())
	r.Use(sharedmw.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))
	r.Use(middleware.RateLimiterMiddleware(redisClient, cfg.RateLimit, cfg.RateLimitWindow))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ── Public routes (no auth required) ──
	public := r.Group("/api/v1")
	{
		// Auth endpoints are public
		public.Any("/auth/*path", proxy.NewReverseProxy(cfg.AuthServiceURL))
	}

	// ── Protected routes (JWT required) ──
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret, redisClient))
	{
		// Server CRUD
		servers := protected.Group("/servers")
		{
			servers.POST("", middleware.ScopeMiddleware("server:create"),
				proxy.NewReverseProxy(cfg.ServerServiceURL))
			servers.GET("", middleware.ScopeMiddleware("server:read"),
				proxy.NewReverseProxy(cfg.ServerServiceURL))
			servers.GET("/:server_id", middleware.ScopeMiddleware("server:read"),
				proxy.NewReverseProxy(cfg.ServerServiceURL))
			servers.PUT("/:server_id", middleware.ScopeMiddleware("server:update"),
				proxy.NewReverseProxy(cfg.ServerServiceURL))
			servers.DELETE("/:server_id", middleware.ScopeMiddleware("server:delete"),
				proxy.NewReverseProxy(cfg.ServerServiceURL))

			// Import/Export → FileIO Service (Phase 4)
			servers.POST("/import", middleware.ScopeMiddleware("server:import"),
				proxy.NewReverseProxy(cfg.FileIOServiceURL))
			servers.GET("/import/:job_id", middleware.ScopeMiddleware("server:import"),
				proxy.NewReverseProxy(cfg.FileIOServiceURL))
			servers.POST("/export", middleware.ScopeMiddleware("server:export"),
				proxy.NewReverseProxy(cfg.FileIOServiceURL))
		}

		// Reports → Report Service (Phase 3)
		reports := protected.Group("/reports")
		{
			reports.GET("/summary", middleware.ScopeMiddleware("report:view"),
				proxy.NewReverseProxy(cfg.ReportServiceURL))
			reports.POST("", middleware.ScopeMiddleware("report:send"),
				proxy.NewReverseProxy(cfg.ReportServiceURL))
		}

		// Monitor endpoints (Phase 2)
		monitor := protected.Group("/monitor")
		{
			monitor.GET("/status", middleware.ScopeMiddleware("monitor:view"),
				proxy.NewReverseProxy(cfg.MonitorServiceURL))
		}
	}

	return r
}

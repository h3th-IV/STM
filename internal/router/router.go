package router

import (
	"github.com/gin-gonic/gin"
	"github.com/heth/STM/internal/config"
	"github.com/heth/STM/internal/controller"
	"github.com/heth/STM/internal/middleware"
	"github.com/heth/STM/internal/utils"
)

// Setup creates and configures the Gin router.
func Setup(
	cfg *config.Config,
	authCtrl *controller.AuthController,
	userCtrl *controller.UserController,
	taskCtrl *controller.TaskController,
	adminCtrl *controller.AdminController,
	jwtService *utils.JWTService,
) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.SecureHeaders())

	// Health (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public, with rate limit)
		authRateLimit := middleware.RateLimiter(middleware.AuthRateLimit())
		auth := v1.Group("/auth")
		auth.Use(authRateLimit)
		{
			auth.POST("/register", authCtrl.Register)
			auth.POST("/login", authCtrl.Login)
			auth.POST("/refresh", authCtrl.Refresh)
		}

		// Protected routes
		authRequired := middleware.AuthRequired(jwtService)
		protected := v1.Group("")
		protected.Use(authRequired)
		{
			protected.GET("/users/me", userCtrl.Me)
			protected.GET("/tasks", taskCtrl.List)
			protected.POST("/tasks", taskCtrl.Create)
			protected.GET("/tasks/:id", taskCtrl.Get)
			protected.PUT("/tasks/:id", taskCtrl.Update)
			protected.DELETE("/tasks/:id", taskCtrl.Delete)
		}

		// Admin-only routes
		admin := v1.Group("/admin")
		admin.Use(authRequired, middleware.RequireAdmin())
		{
			admin.DELETE("/tasks/:id", adminCtrl.ForceDeleteTask)
		}
	}

	return r
}

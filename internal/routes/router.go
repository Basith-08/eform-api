package routes

import (
	"net/http"

	"eform/backend/config"
	"eform/backend/internal/handler"
	"eform/backend/internal/middleware"
	"eform/backend/pkg/auth"
	"eform/backend/pkg/response"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Auth      *handler.AuthHandler
	Employee  *handler.EmployeeHandler
	Dashboard *handler.DashboardHandler
	Profile   *handler.ProfileHandler
}

func NewRouter(cfg config.Config, jwtManager *auth.Manager, handlers Handlers) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.CORSMiddleware(cfg.FrontendURL))
	router.Use(middleware.SecureHeaders())
	router.Use(middleware.RateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst))
	router.Static("/uploads", cfg.UploadPath)

	router.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "service is healthy", gin.H{"name": cfg.AppName})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", handlers.Auth.Register)
		authGroup.POST("/login", handlers.Auth.Login)
		authGroup.POST("/refresh", handlers.Auth.Refresh)
		authGroup.POST("/logout", handlers.Auth.Logout)
		authGroup.POST("/forgot-password", handlers.Auth.ForgotPassword)
		authGroup.POST("/reset-password", handlers.Auth.ResetPassword)
	}

	protected := v1.Group("/")
	protected.Use(middleware.AuthRequired(jwtManager))
	{
		protected.POST("/auth/change-password", handlers.Auth.ChangePassword)
		protected.GET("/profile/me", handlers.Profile.Me)
		protected.PUT("/profile/me", handlers.Profile.UpdateMe)
		protected.GET("/dashboard/user", handlers.Dashboard.User)
	}

	admin := protected.Group("/")
	admin.Use(middleware.RoleRequired("admin"))
	{
		admin.GET("/dashboard/admin", handlers.Dashboard.Admin)
		admin.GET("/employees", handlers.Employee.List)
		admin.POST("/employees", handlers.Employee.Create)
		admin.GET("/employees/:id", handlers.Employee.Detail)
		admin.PUT("/employees/:id", handlers.Employee.Update)
		admin.DELETE("/employees/:id", handlers.Employee.Delete)
		admin.PATCH("/employees/:id/activate", handlers.Employee.Activate)
		admin.PATCH("/employees/:id/deactivate", handlers.Employee.Deactivate)
		admin.POST("/employees/:id/reset-password", handlers.Employee.ResetPassword)
	}

	return router
}

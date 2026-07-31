package main

import (
	"log"

	"eform/backend/config"
	_ "eform/backend/internal/docs"
	"eform/backend/internal/handler"
	"eform/backend/internal/repository"
	"eform/backend/internal/routes"
	"eform/backend/internal/service"
	"eform/backend/internal/validator"
	"eform/backend/pkg/auth"
	"eform/backend/pkg/database"
	"eform/backend/pkg/storage"

	"github.com/gin-gonic/gin"
)

// @title E-Form Employee Management API
// @version 1.0
// @description Production-ready Employee Management System API
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := database.RunSQLFiles(db, "./migrations", "schema_migrations"); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	if cfg.SeedOnBoot {
		if err := database.RunSQLFiles(db, "./seeds", "schema_seeds"); err != nil {
			log.Fatalf("seed failed: %v", err)
		}
	}

	store, err := storage.NewLocalStorage(cfg.UploadPath, cfg.MaxUploadBytes)
	if err != nil {
		log.Fatalf("storage initialization failed: %v", err)
	}

	jwtManager := auth.NewManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	validate := validator.New()
	repos := repository.New(db)

	authService := service.NewAuthService(cfg, repos, store, jwtManager)
	employeeService := service.NewEmployeeService(repos, store)
	dashboardService := service.NewDashboardService(repos)

	authHandler := handler.NewAuthHandler(authService, validate)
	employeeHandler := handler.NewEmployeeHandler(employeeService, authService, validate)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	profileHandler := handler.NewProfileHandler(employeeService, validate)

	router := routes.NewRouter(cfg, jwtManager, routes.Handlers{
		Auth:      authHandler,
		Employee:  employeeHandler,
		Dashboard: dashboardHandler,
		Profile:   profileHandler,
	})

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

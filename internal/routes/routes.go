package routes

import (
	"database/sql"
	"net/http"

	"mbg-backend/internal/config"
	"mbg-backend/internal/handlers"
	"mbg-backend/internal/middleware"
	"mbg-backend/internal/repositories"
	"mbg-backend/internal/services"
	"mbg-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, db *sql.DB, cfg config.Config) {
	authRepo := repositories.NewAuthRepository(db)
	mahasiswaRepo := repositories.NewMahasiswaRepository(db)
	catalogRepo := repositories.NewCatalogRepository(db)
	financialRepo := repositories.NewFinancialRepository(db)
	adminRepo := repositories.NewAdminRepository(db)

	authService := services.NewAuthService(authRepo, cfg.JWTSecret)
	financialService := services.NewFinancialService(financialRepo, mahasiswaRepo)

	authHandler := handlers.NewAuthHandler(authService, mahasiswaRepo)
	mahasiswaHandler := handlers.NewMahasiswaHandler(mahasiswaRepo)
	catalogHandler := handlers.NewCatalogHandler(catalogRepo)
	financialHandler := handlers.NewFinancialHandler(financialService)
	adminHandler := handlers.NewAdminHandler(adminRepo)

	router.GET("/health", func(c *gin.Context) {
		utils.Success(c, http.StatusOK, "ITSPay backend sehat", gin.H{"service": "mbg-backend"})
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		api.GET("/banks", catalogHandler.Banks)
		api.GET("/merchants", catalogHandler.Merchants)
		api.GET("/merchants/:id", catalogHandler.MerchantDetail)

		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.GET("/auth/me", authHandler.Me)

			mahasiswa := protected.Group("/mahasiswa")
			mahasiswa.Use(middleware.Role("mahasiswa", "admin"))
			{
				mahasiswa.GET("/profile", mahasiswaHandler.Profile)
				mahasiswa.GET("/wallet", mahasiswaHandler.Wallet)
				mahasiswa.GET("/transactions", mahasiswaHandler.Transactions)
			}

			protected.POST("/topups", middleware.Role("mahasiswa"), financialHandler.Topup)
			protected.POST("/payments", middleware.Role("mahasiswa"), financialHandler.PayMerchant)

			admin := protected.Group("/admin")
			admin.Use(middleware.Role("admin"))
			{
				admin.GET("/summary", adminHandler.Summary)
				admin.GET("/transactions", adminHandler.Transactions)
				admin.GET("/audit-logs", adminHandler.AuditLogs)
				admin.GET("/reports/daily", adminHandler.DailyReports)
			}
		}
	}
}

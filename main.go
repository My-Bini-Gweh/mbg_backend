package main

import (
	"log"

	"mbg-backend/internal/config"
	"mbg-backend/internal/database"
	"mbg-backend/internal/middleware"
	"mbg-backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewMySQL(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigin))

	routes.Register(router, db, cfg)

	addr := ":" + cfg.AppPort
	log.Printf("ITSPay backend running on http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

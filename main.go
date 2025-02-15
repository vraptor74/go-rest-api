package main

import (
	"go-rest-api/handlers"
	"go-rest-api/initializers"
	"go-rest-api/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	initializers.LoadEnvVariables()

	dbInstance, err := initializers.NewDatabase()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	initializers.MigrateDB(dbInstance)

	initializers.SeedMerchData(dbInstance)
	r := gin.Default()
	authMiddleware := middleware.RequireAuth(dbInstance)
	r.POST("/auth", handlers.RegisterHandler)
	r.GET("/buy/:merch_id", authMiddleware, handlers.BuyHandler)
	r.POST("/sendCoin", authMiddleware, handlers.SendCoinHandler)
	r.GET("/info", authMiddleware, handlers.InfoHandler(dbInstance))
	if err := r.Run(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}

}

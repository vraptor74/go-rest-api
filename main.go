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
	initializers.ConnectToDb()
	initializers.SyncDatabase()
	initializers.SeedMerchData()
	r := gin.Default()
	r.POST("/signup", handlers.RegisterHandler)
	r.POST("/login", handlers.Login)
	r.GET("/validate", middleware.RequireAuth, handlers.Validate)
	r.GET("/buy/:merch_id", middleware.RequireAuth, handlers.BuyHandler)
	r.POST("/sendCoin", middleware.RequireAuth, handlers.SendCoinHandler)
	r.GET("/info", middleware.RequireAuth, handlers.InfoHandler)
	if err := r.Run(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}

}

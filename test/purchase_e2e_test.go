package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"go-rest-api/initializers"
	models "go-rest-api/modules"
	"go-rest-api/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Создаём тестовую БД в памяти
func setupTestDB() *initializers.Database {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Ошибка подключения к тестовой БД")
	}

	testDB := &initializers.Database{DB: db}

	// Создаём таблицы
	db.AutoMigrate(&models.Employee{}, &models.Merch{}, &models.Purchase{})

	// Добавляем тестовые данные
	db.Create(&models.Employee{ID: 1, Email: "test@example.com", Balance: 1000})
	db.Create(&models.Merch{ID: 1, Name: "T-shirt", Price: 500})

	return testDB
}

func TestBuyItem_E2E(t *testing.T) {
	gin.SetMode(gin.TestMode) // 🔹 Отключаем лишние логи

	testDB := setupTestDB()                                // 🔹 Создаём тестовую БД
	purchaseService := services.NewPurchaseService(testDB) // 🔹 Создаём сервис покупки

	router := gin.Default()
	router.POST("/api/buy/:merch_id", func(c *gin.Context) {
		// 🔹 Симулируем авторизованного пользователя
		user := models.Employee{ID: 1, Balance: 1000}
		c.Set("user", user)

		merchID, err := strconv.Atoi(c.Param("merch_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merch_id"})
			return
		}

		//  Запускаем покупку
		newBalance, err := purchaseService.BuyItem(context.Background(), user, int(merchID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//  Отправляем успешный ответ
		c.JSON(http.StatusOK, gin.H{
			"message":     "Purchase successful",
			"new_balance": newBalance,
		})
	})

	//  Создаём HTTP-запрос (покупка товара с ID 1)
	req, _ := http.NewRequest("POST", "/api/buy/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	//  Проверяем ответ
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Purchase successful")
	assert.Contains(t, w.Body.String(), `"new_balance":500`)
}

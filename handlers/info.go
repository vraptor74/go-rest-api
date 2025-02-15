package handlers

import (
	initializers "go-rest-api/initializers"
	models "go-rest-api/modules"
	"go-rest-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InfoHandler(db *initializers.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем пользователя из middleware
		userObj, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неавторизован."})
			return
		}
		user, ok := userObj.(models.Employee)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных пользователя"})
			return
		}

		// Создаём сервис информации
		infoService := services.NewInfoService(db)

		// Загружаем инвентарь
		inventory, err := infoService.GetInventory(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Загружаем историю монет
		history, err := infoService.GetTransactionHistory(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем ответ
		response := models.InfoResponse{
			Coins:       user.Balance,
			Inventory:   inventory,
			CoinHistory: history,
		}
		c.JSON(http.StatusOK, response)
	}
}

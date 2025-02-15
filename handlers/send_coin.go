package handlers

import (
	initializers "go-rest-api/initializers"
	models "go-rest-api/modules"
	"go-rest-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SendCoinHandler(c *gin.Context) {

	//Получаем отправителя из middleware (авторизованный пользователь)
	senderObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	sender, ok := senderObj.(models.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных пользователя"})
		return
	}

	//Получаем данные из тела запроса

	var body struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	dbInstance, _ := initializers.NewDatabase()
	transactionService := services.NewTransactionService(dbInstance)

	err := transactionService.SendCoins(c.Request.Context(), sender, body.ToUser, body.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Перевод выполнен успешно",
		"to_user": body.ToUser,
		"amount":  body.Amount,
	})

}

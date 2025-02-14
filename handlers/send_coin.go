package handlers

import (
	initializers "go-rest-api/initializers"
	models "go-rest-api/modules"
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
	sender := senderObj.(models.Employee)

	//Получаем данные из тела запроса

	var body struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	//Проверяем, что amount > 0
	if body.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than zero"})
		return
	}
	//Проверяем, не отправляет ли пользователь монеты самому себе
	if sender.Email == body.ToUser {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Вы не можете отправить монеты самому себе."})
		return
	}
	//Ищем получателя в БД
	var receiver models.Employee
	if err := initializers.DB.Where("email = ?", body.ToUser).First(&receiver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}
	//Проверяем баланс отправителя
	if sender.Balance < body.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return

	}
	//Запускаем транзакцию
	tx := initializers.DB.Begin()
	//Обновляем баланс отправителя
	if err := tx.Model(&sender).Update("balance", sender.Balance-body.Amount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sender balance"})
		return
	}
	//Обновляем баланс получателя
	if err := tx.Model(&receiver).Update("balance", receiver.Balance+body.Amount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update receiver balance"})
		return
	}
	//Записываем транзакцию
	transaction := models.Transaction{
		SenderID:   &sender.ID,
		ReceiverID: &receiver.ID,
		Amount:     body.Amount,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save transaction"})
		return
	}
	//Фиксируем транзакцию
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":        "Coins sent successfully",
		"new_balance":    sender.Balance,
		"recipient":      receiver.Email,
		"received_coins": body.Amount,
	})

}

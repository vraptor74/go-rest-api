package handlers

import (
	initializers "go-rest-api/initializers"
	models "go-rest-api/modules"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func BuyHandler(c *gin.Context) {
	// Получаем пользователя из middleware
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Приведение типа (без паники)
	user, ok := userObj.(models.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	// Получаем id предмета
	merchID, err := strconv.Atoi(c.Param("merch_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merch_id"})
		return
	}
	var merch models.Merch
	if err := initializers.DB.First(&merch, merchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Merch not found"})
		return
	}
	//Проверяем баланс пользователя
	if user.Balance < merch.Price {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}
	//Обновляем баланс пользователя и создаём запись о покупке (транзакция)
	tx := initializers.DB.Begin()

	var newBalance int
	err = tx.Raw(`
		UPDATE employees 
		SET balance = balance - ? 
		WHERE id = ? AND balance >= ? 
		RETURNING balance
	`, merch.Price, user.ID, merch.Price).Scan(&newBalance).Error

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}
	purchase := models.Purchase{EmployeeID: user.ID, MerchID: uint(merchID)}
	if err := tx.Create(&purchase).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase record"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"message":     "Purchase successful",
		"new_balance": user.Balance - merch.Price,
	})

}

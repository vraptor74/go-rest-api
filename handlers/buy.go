package handlers

import (
	initializers "go-rest-api/initializers"
	models "go-rest-api/modules"
	"go-rest-api/services"
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

	dbInstance, _ := initializers.NewDatabase()
	purchaseService := services.NewPurchaseService(dbInstance)

	newBalance, err := purchaseService.BuyItem(c.Request.Context(), user, merchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Purchase successful",
		"new_balance": newBalance,
	})

}

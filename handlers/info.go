package handlers

import (
	initializers "go-rest-api/initializers"
	models "go-rest-api/modules"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InfoHandler(c *gin.Context) {
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Неавторизован."})
		return
	}
	user := userObj.(models.Employee)

	var inventory []models.Inventory

	rows, err := initializers.DB.Raw(`
		SELECT m.name AS type, COUNT(*) AS quantity
		FROM purchases p
		JOIN merches m ON p.merch_id = m.id
		WHERE p.employee_id = ?
		GROUP BY m.name
	`, user.ID).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Ошибка загрузки инвентаря."})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Inventory
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "Ошибка обработки инвентаря."})
			return
		}
		inventory = append(inventory, item)
	}

	//Загружаем историю полученных монет
	var received []models.TransactionDetail
	rows, err = initializers.DB.Raw(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.sender_id = e.id
		WHERE t.receiver_id = ?
	`, user.ID).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Ошибка загрузки полученных монет."})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.TransactionDetail
		if err := rows.Scan(&transaction.User, &transaction.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "Ошибка обработки полученных монет."})
			return
		}
		received = append(received, transaction)
	}

	//Загружаем историю отправленных монет
	var sent []models.TransactionDetail

	rows, err = initializers.DB.Raw(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.receiver_id = e.id
		WHERE t.sender_id = ?
	`, user.ID).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Ошибка загрузки отправленных монет."})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.TransactionDetail
		if err := rows.Scan(&transaction.User, &transaction.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "Ошибка обработки отправленных монет."})
			return
		}
		sent = append(sent, transaction)
	}
	response := models.InfoResponse{
		Coins:     user.Balance,
		Inventory: inventory,
		CoinHistory: models.CoinHistory{
			Received: received,
			Sent:     sent,
		},
	}
	c.JSON(http.StatusOK, response)

}

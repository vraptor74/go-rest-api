package services

import (
	"context"
	"errors"
	"go-rest-api/initializers"
	models "go-rest-api/modules"
)

// InfoService - сервис для получения информации о пользователе
type InfoService struct {
	DB *initializers.Database
}

// NewInfoService - создаёт экземпляр InfoService
func NewInfoService(db *initializers.Database) *InfoService {
	return &InfoService{DB: db}
}

// GetInventory - загружает инвентарь пользователя
func (s *InfoService) GetInventory(ctx context.Context, userID uint) ([]models.Inventory, error) {
	var inventory []models.Inventory

	rows, err := s.DB.DB.WithContext(ctx).Raw(`
		SELECT m.name AS type, COUNT(*) AS quantity
		FROM purchases p
		JOIN merches m ON p.merch_id = m.id
		WHERE p.employee_id = ?
		GROUP BY m.name
	`, userID).Rows()
	if err != nil {
		return nil, errors.New("ошибка загрузки инвентаря")
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Inventory
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			return nil, errors.New("ошибка обработки инвентаря")
		}
		inventory = append(inventory, item)
	}
	return inventory, nil
}

// GetTransactionHistory - загружает историю отправленных и полученных монет
func (s *InfoService) GetTransactionHistory(ctx context.Context, userID uint) (models.CoinHistory, error) {
	var history models.CoinHistory

	// Полученные монеты
	rows, err := s.DB.DB.WithContext(ctx).Raw(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.sender_id = e.id
		WHERE t.receiver_id = ?
	`, userID).Rows()
	if err != nil {
		return history, errors.New("ошибка загрузки полученных монет")
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.TransactionDetail
		if err := rows.Scan(&transaction.User, &transaction.Amount); err != nil {
			return history, errors.New("ошибка обработки полученных монет")
		}
		history.Received = append(history.Received, transaction)
	}

	// Отправленные монеты
	rows, err = s.DB.DB.WithContext(ctx).Raw(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.receiver_id = e.id
		WHERE t.sender_id = ?
	`, userID).Rows()
	if err != nil {
		return history, errors.New("ошибка загрузки отправленных монет")
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.TransactionDetail
		if err := rows.Scan(&transaction.User, &transaction.Amount); err != nil {
			return history, errors.New("ошибка обработки отправленных монет")
		}
		history.Sent = append(history.Sent, transaction)
	}
	return history, nil
}

package services

import (
	"context"
	"errors"
	"go-rest-api/initializers"
	models "go-rest-api/modules"
	"log"
)

type TransactionService struct {
	DB *initializers.Database
}

func NewTransactionService(db *initializers.Database) *TransactionService {
	return &TransactionService{DB: db}
}

func (s *TransactionService) SendCoins(ctx context.Context, sender models.Employee, recipientEmail string, amount int) error {
	if amount <= 0 {
		return errors.New("сумма перевода должна быть больше нуля")
	}

	if sender.Email == recipientEmail {
		return errors.New("вы не можете отправить монеты самому себе")
	}

	var receiver models.Employee
	if err := s.DB.DB.WithContext(ctx).Where("email = ?", recipientEmail).First(&receiver).Error; err != nil {
		log.Printf("получатель не найден: %v", err)
		return errors.New("получатель не найден")
	}

	if sender.Balance < amount {
		return errors.New("недостаточно средств")
	}
	tx := s.DB.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err := tx.Exec(`
    UPDATE employees 
    SET balance = balance - ? 
    WHERE id = ? AND balance >= ?`,
		amount, sender.ID, amount).Error

	if err != nil {
		tx.Rollback()
		log.Printf("ошибка при обновлении баланса отправителя: %v", err)
		return errors.New("ошибка при обновлении баланса отправителя")
	}

	if err := tx.Model(&receiver).Update("balance", receiver.Balance+amount).Error; err != nil {
		tx.Rollback()
		log.Printf("ошибка при обновлении баланса получателя: %v", err)
		return errors.New("ошибка при обновлении баланса получателя")
	}
	transaction := models.Transaction{
		SenderID:   &sender.ID,
		ReceiverID: &receiver.ID,
		Amount:     amount,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		log.Printf("ошибка при сохранении транзакции: %v", err)
		return errors.New("ошибка при сохранении транзакции")
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("ошибка при завершении транзакции: %v", err)
		return errors.New("ошибка при завершении транзакции")
	}
	return nil

}

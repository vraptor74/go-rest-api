package services

import (
	"context"
	"errors"
	"go-rest-api/initializers"
	models "go-rest-api/modules"
	"log"
)

type PurchaseService struct {
	DB *initializers.Database
}

func NewPurchaseService(db *initializers.Database) *PurchaseService {
	return &PurchaseService{DB: db}
}

func (s *PurchaseService) BuyItem(ctx context.Context, user models.Employee, merchID int) (int, error) {
	// Check for Zero values
	if user.ID == 0 {
		return 0, errors.New("неверный пользователь")
	}
	//Check if the item exists
	var merch models.Merch
	if err := s.DB.DB.WithContext(ctx).First(&merch, merchID).Error; err != nil {
		return 0, errors.New("merch not found")
	}
	//Check the user's balance
	if user.Balance < merch.Price {
		return 0, errors.New("insufficient funds")
	}

	//Start a transaction
	tx := s.DB.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		log.Printf("failed to start transaction: %v", tx.Error)
		return 0, errors.New("failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the user's balance
	user.Balance -= merch.Price
	if err := tx.Model(&user).Update("balance", user.Balance).Error; err != nil {
		tx.Rollback()
		log.Printf("failed to update balance: %v", err)
		return 0, errors.New("failed to update balance")
	}
	//Record the purchase
	purchase := models.Purchase{EmployeeID: user.ID, MerchID: uint(merchID)}
	if err := tx.Create(&purchase).Error; err != nil {
		tx.Rollback()
		log.Printf("failed to record purchase: %v", err)
		return 0, errors.New("failed to record purchase")
	}
	//Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return 0, errors.New("failed to commit transaction")
	}
	return user.Balance, nil

}

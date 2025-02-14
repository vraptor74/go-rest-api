package initializers

import (
	models "go-rest-api/modules"
	"log"
)

func SyncDatabase() {
	err := DB.AutoMigrate(&models.Employee{}, &models.Transaction{}, &models.Merch{}, &models.Purchase{})
	if err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}
}

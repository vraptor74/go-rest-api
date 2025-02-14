package initializers

import (
	models "go-rest-api/modules"
)

func SyncDatabase() {
	DB.AutoMigrate(&models.Employee{}, &models.Transaction{}, &models.Merch{}, &models.Purchase{})

}

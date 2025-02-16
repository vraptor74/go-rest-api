package initializers

import (
	models "go-rest-api/modules"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase() (*Database, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL не задан! Проверь .env файл или переменные окружения.")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

func MigrateDB(db *Database) {
	if err := db.DB.AutoMigrate(&models.Employee{}, &models.Transaction{}, &models.Merch{}, &models.Purchase{}); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}
}

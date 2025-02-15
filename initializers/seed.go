package initializers

import (
	"fmt"
	models "go-rest-api/modules"
	"log"
)

// SeedMerchData заполняет базу начальными предметами
func SeedMerchData(db *Database) {
	// Проверяем, есть ли предметы в таблице

	tx := db.DB.Begin()
	var count int64
	if err := tx.Model(&models.Merch{}).Count(&count).Error; err != nil {
		log.Printf("Ошибка проверки существующих данных: %v", err)
		tx.Rollback()
		return
	}
	if count == 0 { // Если таблица пустая, заполняем её
		items := []models.Merch{
			{Name: "t-shirt", Price: 80},
			{Name: "cup", Price: 20},
			{Name: "book", Price: 50},
			{Name: "pen", Price: 10},
			{Name: "powerbank", Price: 200},
			{Name: "hoody", Price: 300},
			{Name: "umbrella", Price: 200},
			{Name: "socks", Price: 10},
			{Name: "wallet", Price: 50},
			{Name: "pink-hoody", Price: 500},
		}

		// Добавляем предметы в базу
		result := db.DB.Create(&items)
		if result.Error != nil {
			fmt.Println("Ошибка при добавлении предметов:", result.Error)
		} else {
			fmt.Println("Таблица предметов заполнена!")
		}
	} else {
		fmt.Println("Таблица предметов уже содержит данные, пропускаем заполнение.")
	}
}

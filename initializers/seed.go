package initializers

import (
	"fmt"
	models "go-rest-api/modules"
)

// SeedMerchData заполняет базу начальными предметами
func SeedMerchData() {
	// Проверяем, есть ли предметы в таблице
	var count int64
	DB.Model(&models.Merch{}).Count(&count)

	if count == 0 { // Если таблица пустая, заполняем её
		items := []models.Merch{
			{Name: "Футболка Avito", Price: 500},
			{Name: "Толстовка Avito", Price: 1500},
			{Name: "Кружка Avito", Price: 300},
			{Name: "Рюкзак Avito", Price: 2000},
		}

		// Добавляем предметы в базу
		result := DB.Create(&items)
		if result.Error != nil {
			fmt.Println("Ошибка при добавлении предметов:", result.Error)
		} else {
			fmt.Println("Таблица предметов заполнена!")
		}
	} else {
		fmt.Println("Таблица предметов уже содержит данные, пропускаем заполнение.")
	}
}

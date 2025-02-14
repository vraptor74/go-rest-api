package models

import (
	"time"
)

// Сотрудники
type Employee struct {
	ID        uint          `gorm:"primaryKey"`
	Email     string        `gorm:"unique;not null"`
	Password  string        `gorm:"not null"`
	Balance   int           `gorm:"not null;default:1000"`
	Sent      []Transaction `gorm:"foreignKey:SenderID"`   // Монетки, которые отправил сотрудник
	Received  []Transaction `gorm:"foreignKey:ReceiverID"` // Монетки, которые получил сотрудник
	Purchases []Purchase    `gorm:"foreignKey:EmployeeID"`
}

// История транзакций (переводы монет)
type Transaction struct {
	ID         uint      `gorm:"primaryKey"`
	SenderID   *uint     `gorm:"index"`
	ReceiverID *uint     `gorm:"index"`
	Amount     int       `gorm:"not null;check:amount > 0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// Таблица товаров (мерч)
type Merch struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"unique;not null"`
	Price int    `gorm:"not null;check:price > 0"`
}

// История покупок
type Purchase struct {
	ID          uint      `gorm:"primaryKey"`
	EmployeeID  uint      `gorm:"index"`
	MerchID     uint      `gorm:"index"`
	PurchasedAt time.Time `gorm:"autoCreateTime"`
}

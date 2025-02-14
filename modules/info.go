package models

type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Inventory `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}

// Inventory - структура для списка купленных товаров
type Inventory struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

// CoinHistory - структура для истории транзакций
type CoinHistory struct {
	Received []TransactionDetail `json:"received"`
	Sent     []TransactionDetail `json:"sent"`
}

// TransactionDetail - информация о переводе монет
type TransactionDetail struct {
	User   string `json:"user"`   // Отправитель или получатель
	Amount int    `json:"amount"` // Сумма перевода
}

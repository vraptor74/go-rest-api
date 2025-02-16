package test

import (
	"context"
	"errors"
	"go-rest-api/initializers"
	"go-rest-api/services"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetInventory(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	infoService := services.NewInfoService(mockDatabase)

	userID := uint(1)

	// Ожидание SQL-запроса
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT m.name AS type, COUNT(*) AS quantity
		FROM purchases p
		JOIN merches m ON p.merch_id = m.id
		WHERE p.employee_id = $1
		GROUP BY m.name`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"type", "quantity"}).
			AddRow("T-shirt", 2).
			AddRow("Mug", 1))

	// Вызываем функцию
	inventory, err := infoService.GetInventory(context.Background(), userID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Len(t, inventory, 2)
	assert.Equal(t, "T-shirt", inventory[0].Type)
	assert.Equal(t, 2, inventory[0].Quantity)
	assert.Equal(t, "Mug", inventory[1].Type)
	assert.Equal(t, 1, inventory[1].Quantity)

	// Проверяем, что все моки вызвались
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTransactionHistory(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	infoService := services.NewInfoService(mockDatabase)

	userID := uint(1)

	//Ожидание загрузки полученных монет
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.sender_id = e.id
		WHERE t.receiver_id = $1`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user", "amount"}).
			AddRow("alice@example.com", 100).
			AddRow("bob@example.com", 50))

	//Ожидание загрузки отправленных монет
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.receiver_id = e.id
		WHERE t.sender_id = $1`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user", "amount"}).
			AddRow("charlie@example.com", 75))

	history, err := infoService.GetTransactionHistory(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, history.Received, 2)
	assert.Equal(t, "alice@example.com", history.Received[0].User)
	assert.Equal(t, 100, history.Received[0].Amount)
	assert.Equal(t, "bob@example.com", history.Received[1].User)
	assert.Equal(t, 50, history.Received[1].Amount)

	assert.Len(t, history.Sent, 1)
	assert.Equal(t, "charlie@example.com", history.Sent[0].User)
	assert.Equal(t, 75, history.Sent[0].Amount)

	assert.NoError(t, mock.ExpectationsWereMet())

}

//Тест на ошибку при загрузке инвентаря

func TestGetInventory_DBError(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	infoService := services.NewInfoService(mockDatabase)

	userID := uint(1)

	// Симуляция ошибки в SQL-запросе
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT m.name AS type, COUNT(*) AS quantity
		FROM purchases p
		JOIN merches m ON p.merch_id = m.id
		WHERE p.employee_id = $1
		GROUP BY m.name`)).
		WithArgs(userID).
		WillReturnError(errors.New("ошибка базы данных"))

	// Вызываем функцию
	inventory, err := infoService.GetInventory(context.Background(), userID)

	// Проверяем результат
	assert.Error(t, err)
	assert.Equal(t, "ошибка загрузки инвентаря", err.Error())
	assert.Nil(t, inventory)
}

//Тест на ошибку при обработке полученных монет

func TestGetTransactionHistory_ReceivedError(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	infoService := services.NewInfoService(mockDatabase)

	userID := uint(1)

	// 🔹 Симуляция ошибки при загрузке полученных монет
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.sender_id = e.id
		WHERE t.receiver_id = $1`)).
		WithArgs(userID).
		WillReturnError(errors.New("ошибка базы данных"))

	// Вызываем функцию
	history, err := infoService.GetTransactionHistory(context.Background(), userID)

	// Проверяем результат
	assert.Error(t, err)
	assert.Equal(t, "ошибка загрузки полученных монет", err.Error())
	assert.Empty(t, history.Received)
	assert.Empty(t, history.Sent)
}

// Тест на ошибку при обработке отправленных монет
func TestGetTransactionHistory_SentError(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	infoService := services.NewInfoService(mockDatabase)

	userID := uint(1)

	// 🔹 Ожидание загрузки полученных монет
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.sender_id = e.id
		WHERE t.receiver_id = $1`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user", "amount"}).
			AddRow("alice@example.com", 100).
			AddRow("bob@example.com", 50))

	// 🔹 Симуляция ошибки при загрузке отправленных монет
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT e.email AS user, t.amount
		FROM transactions t
		JOIN employees e ON t.receiver_id = e.id
		WHERE t.sender_id = $1`)).
		WithArgs(userID).
		WillReturnError(errors.New("ошибка базы данных"))

	// Вызываем функцию
	history, err := infoService.GetTransactionHistory(context.Background(), userID)

	// Проверяем результат
	assert.Error(t, err)
	assert.Equal(t, "ошибка загрузки отправленных монет", err.Error())
	assert.Len(t, history.Received, 2) // Полученные монеты загружены
	assert.Empty(t, history.Sent)      // Отправленные монеты не загружены
}

package test

import (
	"context"
	"errors"
	"go-rest-api/initializers"
	models "go-rest-api/modules"
	"go-rest-api/services"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Тест: успешный перевод монет
func TestSendCoins_Success(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	transactionService := services.NewTransactionService(mockDatabase)

	sender := models.Employee{ID: 1, Email: "sender@example.com", Balance: 1000}
	receiver := models.Employee{ID: 2, Email: "receiver@example.com", Balance: 500}
	amount := 200

	// Ожидание поиска получателя
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employees" WHERE email = $1 ORDER BY "employees"."id" LIMIT $2`)).
		WithArgs(receiver.Email, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "balance"}).AddRow(receiver.ID, receiver.Email, receiver.Balance))

	// Ожидание начала транзакции
	mock.ExpectBegin()

	// Ожидание обновления баланса отправителя
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE employees SET balance = balance - $1 WHERE id = $2 AND balance >= $3`)).
		WithArgs(amount, sender.ID, amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидание обновления баланса получателя
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "employees" SET "balance"=$1 WHERE "id" = $2`)).
		WithArgs(receiver.Balance+amount, receiver.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидание записи транзакции
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions" ("sender_id","receiver_id","amount","created_at")`)).
		WithArgs(sender.ID, receiver.ID, amount, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Ожидание коммита
	mock.ExpectCommit()

	// Вызываем `SendCoins`
	err := transactionService.SendCoins(context.Background(), sender, receiver.Email, amount)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet()) // Проверяем, что все моки вызвались
}

// Тест: сумма перевода <= 0
func TestSendCoins_InvalidAmount(t *testing.T) {
	gormDB, _ := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	transactionService := services.NewTransactionService(mockDatabase)

	sender := models.Employee{ID: 1, Email: "sender@example.com", Balance: 1000}

	err := transactionService.SendCoins(context.Background(), sender, "receiver@example.com", 0)

	assert.Error(t, err)
	assert.Equal(t, "сумма перевода должна быть больше нуля", err.Error())
}

// Тест: отправка монет самому себе
func TestSendCoins_SelfTransfer(t *testing.T) {
	gormDB, _ := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	transactionService := services.NewTransactionService(mockDatabase)

	sender := models.Employee{ID: 1, Email: "sender@example.com", Balance: 1000}

	err := transactionService.SendCoins(context.Background(), sender, sender.Email, 100)

	assert.Error(t, err)
	assert.Equal(t, "вы не можете отправить монеты самому себе", err.Error())
}

// Тест: получатель не найден
func TestSendCoins_RecipientNotFound(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	transactionService := services.NewTransactionService(mockDatabase)

	sender := models.Employee{ID: 1, Email: "sender@example.com", Balance: 1000}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employees" WHERE email = $1 ORDER BY "employees"."id" LIMIT $2`)).
		WithArgs("receiver@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	err := transactionService.SendCoins(context.Background(), sender, "receiver@example.com", 100)

	assert.Error(t, err)
	assert.Equal(t, "получатель не найден", err.Error())
}

// Тест: недостаточно средств
func TestSendCoins_InsufficientFunds(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	transactionService := services.NewTransactionService(mockDatabase)

	sender := models.Employee{ID: 1, Email: "sender@example.com", Balance: 100}
	receiver := models.Employee{ID: 2, Email: "receiver@example.com"}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employees" WHERE email = $1 ORDER BY "employees"."id" LIMIT $2`)).
		WithArgs(receiver.Email, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "balance"}).AddRow(receiver.ID, receiver.Email, 500))

	err := transactionService.SendCoins(context.Background(), sender, receiver.Email, 200)

	assert.Error(t, err)
	assert.Equal(t, "недостаточно средств", err.Error())
}

// Тест: ошибка при обновлении баланса отправителя
func TestSendCoins_UpdateSenderBalanceError(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	transactionService := services.NewTransactionService(mockDatabase)

	sender := models.Employee{ID: 1, Email: "sender@example.com", Balance: 1000}
	receiver := models.Employee{ID: 2, Email: "receiver@example.com"}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employees" WHERE email = $1 ORDER BY "employees"."id" LIMIT $2`)).
		WithArgs(receiver.Email, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "balance"}).AddRow(receiver.ID, receiver.Email, 500))

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE employees SET balance = balance - $1 WHERE id = $2 AND balance >= $3`)).
		WithArgs(200, sender.ID, 200).
		WillReturnError(errors.New("ошибка при обновлении баланса отправителя"))

	err := transactionService.SendCoins(context.Background(), sender, receiver.Email, 200)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при обновлении баланса отправителя")
}

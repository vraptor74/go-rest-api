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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB() (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, _ := sqlmock.New() // Создаём моковую БД
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	return gormDB, mock
}

// Тест успешной покупки
func TestBuyItem_Success(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	purchaseService := services.NewPurchaseService(mockDatabase)

	user := models.Employee{ID: 1, Balance: 1000}
	merch := models.Merch{ID: 1, Price: 500}

	// Ожидание запроса на поиск товара
	mock.ExpectQuery(`SELECT \* FROM "merches" WHERE "merches"."id" = \$1 ORDER BY "merches"."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "price"}).AddRow(1, 500))

	// Ожидание начала транзакции
	mock.ExpectBegin()

	// Ожидание обновления баланса
	mock.ExpectQuery(regexp.QuoteMeta(`
    UPDATE employees 
    SET balance = balance - $1 
    WHERE id = $2 AND balance >= $3 
    RETURNING balance`)).
		WithArgs(500, 1, 500). // Проверяем порядок аргументов
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(500))

	// Ожидание записи о покупке (исправлено!)
	mock.ExpectQuery(`INSERT INTO "purchases"`).
		WithArgs(user.ID, merch.ID, sqlmock.AnyArg()).            //  Ожидаем дату `purchased_at`
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) //  GORM использует `RETURNING id`

	// Ожидание коммита транзакции
	mock.ExpectCommit()

	// Вызываем `BuyItem`
	newBalance, err := purchaseService.BuyItem(context.Background(), user, int(merch.ID))

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, user.Balance-merch.Price, newBalance)
	assert.NoError(t, mock.ExpectationsWereMet()) // Проверяем, что все моки вызвались
}

// Тест: товар не найден

func TestBuyItem_MerchNotFound(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	purchaseService := services.NewPurchaseService(mockDatabase)

	user := models.Employee{ID: 1, Balance: 1000}

	// Ожидание запроса, но без результатов
	mock.ExpectQuery(`SELECT \* FROM "merches" WHERE "merches"."id" = \$1 ORDER BY "merches"."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	newBalance, err := purchaseService.BuyItem(context.Background(), user, 1)

	assert.Error(t, err)
	assert.Equal(t, "merch not found", err.Error())
	assert.Equal(t, 0, newBalance)
}

// Тест: Недостаточно средств
func TestBuyItem_InsufficientFunds(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	purchaseService := services.NewPurchaseService(mockDatabase)

	user := models.Employee{ID: 1, Balance: 100} // Недостаточно средств
	merch := models.Merch{ID: 1, Price: 500}

	// Ожидание запроса на поиск товара
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "merches" WHERE "merches"."id" = $1 ORDER BY "merches"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "price"}).AddRow(1, 500)) // Добавляем товар

	//Вызываем `BuyItem`, но денег не хватит
	newBalance, err := purchaseService.BuyItem(context.Background(), user, int(merch.ID))

	// Проверяем, что ошибка — "insufficient funds"
	assert.Error(t, err)
	assert.Equal(t, "insufficient funds", err.Error())
	assert.Equal(t, 0, newBalance)

	// Проверяем, что все ожидания были вызваны
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Тест: ошибка при обновлении баланса
func TestBuyItem_UpdateBalanceError(t *testing.T) {
	gormDB, mock := setupMockDB()
	mockDatabase := &initializers.Database{DB: gormDB}
	purchaseService := services.NewPurchaseService(mockDatabase)

	user := models.Employee{ID: 1, Balance: 1000}
	merch := models.Merch{ID: 1, Price: 500}

	mock.ExpectQuery(`SELECT \* FROM "merches"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "price"}).AddRow(1, 500))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE employees SET balance = balance - $1 WHERE id = $2 AND balance >= $3 RETURNING balance`)).
		WithArgs(500, 1, 500).
		WillReturnError(errors.New("failed to update balance"))

	newBalance, err := purchaseService.BuyItem(context.Background(), user, int(merch.ID))

	assert.Error(t, err)
	assert.Equal(t, "failed to update balance", err.Error())
	assert.Equal(t, 0, newBalance)
}

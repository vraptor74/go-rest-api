package services

import (
	"context"
	"errors"
	"go-rest-api/initializers"
	models "go-rest-api/modules"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *initializers.Database
}

func NewAuthService(db *initializers.Database) *AuthService {
	return &AuthService{DB: db}
}

func (s *AuthService) RegisterUser(ctx context.Context, email, password string) (string, int64, error) {
	var user models.Employee

	// Проверяем, существует ли пользователь
	err := s.DB.DB.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Пользователя нет  создаём нового
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return "", 0, errors.New("не удалось хешировать пароль")
			}

			user = models.Employee{Email: email, Password: string(hash)}
			if err := s.DB.DB.WithContext(ctx).Create(&user).Error; err != nil {
				return "", 0, errors.New("ошибка создания пользователя")
			}
		} else {
			// Ошибка запроса в БД
			return "", 0, err
		}
	} else {
		// Пользователь найден → проверяем пароль
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			return "", 0, errors.New("неверный пароль")
		}
	}

	return generateJWT(user.ID)
}

func generateJWT(userID uint) (string, int64, error) {
	expirationTime := time.Now().Add(time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": expirationTime,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime, nil
}

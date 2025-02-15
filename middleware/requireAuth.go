package middleware

import (
	"go-rest-api/initializers"
	models "go-rest-api/modules"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func RequireAuth(db *initializers.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization или Cookie
		tokenString, err := c.Cookie("Authorization")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходим токен"})
			c.Abort()
			return
		}

		// Разбираем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			c.Abort()
			return
		}

		// Получаем `sub` из токена
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Ошибка декодирования токена"})
			c.Abort()
			return
		}
		sub, ok := claims["sub"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Ошибка получения ID пользователя"})
			c.Abort()
			return
		}

		// Ищем пользователя в БД
		var user models.Employee
		if err := db.DB.First(&user, uint(sub)).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
			c.Abort()
			return
		}

		// Сохраняем пользователя в контексте
		c.Set("user", user)
		c.Next()
	}
}

package handlers

import (
	initializers "go-rest-api/initializers"
	"go-rest-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

//var secretKey = []byte(os.Getenv("JWT_SECRET"))

func RegisterHandler(c *gin.Context) {

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}
	dbInstance, _ := initializers.NewDatabase()
	authService := services.NewAuthService(dbInstance)

	tokenString, expiresIn, err := authService.RegisterUser(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600, "", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"expires_in": expiresIn,
	})

}

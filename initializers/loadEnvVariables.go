package initializers

import (
    "log"

    "github.com/joho/godotenv"
)

func LoadEnv() {
    err := godotenv.Load()
    if err != nil {
        log.Println("⚠️  .env файл не найден, используем переменные окружения")
    }
}

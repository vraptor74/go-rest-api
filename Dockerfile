FROM golang:1.24 as builder

# Установим рабочую директорию
WORKDIR /app


# Копируем Go модули и зависимостей
COPY go.mod go.sum ./
RUN go mod tidy

# Копируем весь проект в контейнер
COPY . .

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Используем минимальный образ для запуска
FROM alpine:latest  

# Устанавливаем нужные библиотеки
RUN apk --no-cache add ca-certificates

# Копируем скомпилированный бинарник из билд-образа
COPY --from=builder /app/main /main

# Указываем команду для запуска контейнера
CMD ["/main"]


# Открываем порт
EXPOSE 8080
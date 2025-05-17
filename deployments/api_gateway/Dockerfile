# Используем официальный образ Golang
FROM golang:1.23-alpine

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum
COPY api_gateway/go.mod api_gateway/go.sum /app/api_gateway/

# Копируем весь остальной исходный код сервиса
COPY api_gateway/ /app/api_gateway/

# Устанавливаем рабочую директорию для команд go
WORKDIR /app/api_gateway

# Загружаем зависимости
RUN go mod download
RUN go mod tidy

# Собираем приложение
RUN go build -o /app/api_gateway_executable .

# Возвращаемся в /app для CMD
WORKDIR /app

# Указываем порт, на котором будет работать приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./api_gateway_executable"] 
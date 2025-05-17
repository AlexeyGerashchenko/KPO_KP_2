# Используем официальный образ Golang
FROM golang:1.23-alpine

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для основного сервиса
# Пути здесь относительны контексту сборки (корень проекта)
COPY file_storing_service/go.mod file_storing_service/go.sum ./file_storing_service/

# Копируем исходный код pkg сначала, чтобы replace сработал
COPY pkg/ ./pkg/

# Копируем остальной исходный код основного сервиса
COPY file_storing_service/ ./file_storing_service/

# Устанавливаем рабочую директорию для команд go
WORKDIR /app/file_storing_service

# Загружаем зависимости
# go mod download теперь должен корректно обработать replace
RUN go mod download
RUN go mod tidy

# Собираем приложение
RUN go build -o /app/file_storing_service_executable .

# Возвращаемся в /app для CMD, если нужно
WORKDIR /app

# Указываем порт, на котором будет работать приложение
EXPOSE 8081

# Команда для запуска приложения
CMD ["./file_storing_service_executable"] 
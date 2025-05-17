# Используем официальный образ Golang
FROM golang:1.23-alpine

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для основного сервиса
# Пути здесь относительны контексту сборки (корень проекта)
COPY file_analysis_service/go.mod file_analysis_service/go.sum ./file_analysis_service/

# Копируем исходный код pkg сначала, чтобы replace сработал
COPY pkg/ ./pkg/

# Копируем остальной исходный код основного сервиса
COPY file_analysis_service/ ./file_analysis_service/

# Устанавливаем рабочую директорию для команд go
WORKDIR /app/file_analysis_service

# Загружаем зависимости
# go mod download теперь должен корректно обработать replace
RUN go mod download
RUN go mod tidy

# Собираем приложение
RUN go build -o /app/file_analysis_service_executable .

# Возвращаемся в /app для CMD, если нужно
WORKDIR /app

# Указываем порт, на котором будет работать приложение
EXPOSE 8082

# Команда для запуска приложения
CMD ["./file_analysis_service_executable"] 
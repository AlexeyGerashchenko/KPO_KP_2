package main

import (
	"file_storing_service/handlers"
	"file_storing_service/models"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "file_storing_service/docs" // Путь к сгенерированной Swagger документации

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title File Storing Service API
// @version 1.0
// @description Этот сервис отвечает за хранение и выдачу файлов.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api/v1
// @schemes http
func main() {
	fmt.Println("File Storing Service starting...")

	// Загрузка конфигурации из переменных окружения
	postgresUser := os.Getenv("POSTGRES_USER_DB1")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD_DB1")
	postgresDB := os.Getenv("POSTGRES_DB_DB1")
	postgresHost := os.Getenv("POSTGRES_HOST_DB1")
	postgresPort := os.Getenv("POSTGRES_PORT_DB1")
	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")

	if fileStoragePath == "" {
		fileStoragePath = "./file_storage_1" // Значение по умолчанию, если не указано
	}
	// Создаем директорию, если она не существует
	if err := os.MkdirAll(fileStoragePath, os.ModePerm); err != nil {
		log.Fatalf("Не удалось создать директорию для хранения файлов: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		postgresHost, postgresUser, postgresPassword, postgresDB, postgresPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	// Миграция схемы
	err = db.AutoMigrate(&models.File{})
	if err != nil {
		log.Fatalf("Не удалось выполнить миграцию базы данных: %v", err)
	}

	fileHandler := handlers.NewFileHandler(db, fileStoragePath)

	r := gin.Default()

	apiV1 := r.Group("/api/v1")
	{
		filesGroup := apiV1.Group("/files")
		{
			filesGroup.POST("/upload", fileHandler.UploadFile)
			filesGroup.GET("/:id", fileHandler.GetFileByID)
			filesGroup.GET("", fileHandler.ListFiles) // Эндпоинт для получения списка файлов
		}
		// Внутренние эндпоинты, не предназначенные для прямого вызова пользователем через API Gateway
		internalGroup := apiV1.Group("/internal")
		{
			internalGroup.GET("/files/:id/location", fileHandler.GetFileLocationByID)
			internalGroup.GET("/file-content", fileHandler.GetFileContentByLocationInternal)

		}
	}

	// Swagger документация
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("File Storing Service запущен на порту 8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

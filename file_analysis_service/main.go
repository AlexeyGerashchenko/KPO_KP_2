package main

import (
	"file_analysis_service/handlers"
	"file_analysis_service/models"
	"file_analysis_service/services"
	"fmt"
	"log"
	"os"
	"pkg/adapters"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "file_analysis_service/docs" // Путь к сгенерированной Swagger документации

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title File Analysis Service API
// @version 1.0
// @description Этот сервис отвечает за анализ файлов, хранение результатов и их выдачу.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
// @BasePath /api/v1
// @schemes http
func main() {
	fmt.Println("File Analysis Service starting...")

	// Загрузка конфигурации из переменных окружения
	postgresUser := os.Getenv("POSTGRES_USER_DB2")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD_DB2")
	postgresDB := os.Getenv("POSTGRES_DB_DB2")
	postgresHost := os.Getenv("POSTGRES_HOST_DB2")
	postgresPort := os.Getenv("POSTGRES_PORT_DB2")
	wordCloudAPIURL := os.Getenv("WORDCLOUD_API_URL")
	fileStoragePath := os.Getenv("FILE_STORAGE_PATH") // Для сохранения облаков слов
	fileStoringServiceAddr := os.Getenv("FILE_STORING_SERVICE_ADDR")

	if fileStoragePath == "" {
		fileStoragePath = "./file_storage_2" // Значение по умолчанию
	}
	if wordCloudAPIURL == "" {
		wordCloudAPIURL = "https://quickchart.io/wordcloud" // Значение по умолчанию
	}
	if fileStoringServiceAddr == "" {
		fileStoringServiceAddr = "http://localhost:8081" // Значение по умолчанию для локального запуска
	}

	// Инициализация адаптеров
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		postgresHost, postgresUser, postgresPassword, postgresDB, postgresPort)
	dbAdapter, err := adapters.NewDBAdapter(dsn, &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось инициализировать DBAdapter: %v", err)
	}

	err = dbAdapter.AutoMigrate(&models.AnalysisResult{})
	if err != nil {
		log.Fatalf("Не удалось выполнить миграцию БД для AnalysisResult: %v", err)
	}

	fsAdapter, err := adapters.NewFileStorageAdapter(fileStoragePath)
	if err != nil {
		log.Fatalf("Не удалось инициализировать FileStorageAdapter для облаков слов: %v", err)
	}

	storingServiceAdapter := adapters.NewFileStoringServiceAdapter(fileStoringServiceAddr)
	cloudAPIAdapter := adapters.NewWordCloudAPIAdapter(wordCloudAPIURL)

	// Инициализация сервиса
	analysisService := services.NewAnalysisService(dbAdapter, fsAdapter, storingServiceAdapter, cloudAPIAdapter)

	// Инициализация обработчика
	analysisHandler := handlers.NewAnalysisHandler(analysisService)

	r := gin.Default()

	apiV1 := r.Group("/api/v1")
	{
		analysisGroup := apiV1.Group("/analysis")
		{
			analysisGroup.POST("/:file_id", analysisHandler.RequestAnalysis)
			analysisGroup.GET("/results/:file_id", analysisHandler.GetAnalysisResults)
			analysisGroup.GET("/wordclouds", analysisHandler.GetWordCloud)                // location передается как query param
			analysisGroup.GET("/results-all", analysisHandler.ListAnalysisResultsHandler) // Для отладки
		}
	}

	// Swagger документация
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("File Analysis Service запущен на порту 8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

package main

import (
	"api_gateway/handlers"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "api_gateway/docs" // Импорт для Swagger
)

// @title API Gateway
// @version 1.0
// @description API Gateway для микросервисной архитектуры обработки текстовых файлов.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	fmt.Println("API Gateway starting...")

	// Загрузка конфигурации из переменных окружения
	fileStoringServiceAddr := os.Getenv("FILE_STORING_SERVICE_ADDR")
	fileAnalysisServiceAddr := os.Getenv("FILE_ANALYSIS_SERVICE_ADDR")

	// Значения по умолчанию, если переменные не установлены
	if fileStoringServiceAddr == "" {
		fileStoringServiceAddr = "http://localhost:8081"
	} else if !hasProtocol(fileStoringServiceAddr) {
		fileStoringServiceAddr = "http://" + fileStoringServiceAddr
	}

	if fileAnalysisServiceAddr == "" {
		fileAnalysisServiceAddr = "http://localhost:8082"
	} else if !hasProtocol(fileAnalysisServiceAddr) {
		fileAnalysisServiceAddr = "http://" + fileAnalysisServiceAddr
	}

	// Инициализация обработчика прокси
	proxyHandler := handlers.NewProxyHandler(fileStoringServiceAddr, fileAnalysisServiceAddr)

	// Инициализация Gin
	r := gin.Default()

	// Маршруты API
	// 1. Загрузка файла
	r.POST("/upload", proxyHandler.UploadFile)

	// 2. Анализ файла
	r.POST("/analysis/:file_id", proxyHandler.RequestAnalysis)
	r.GET("/analysis/results/:file_id", proxyHandler.GetAnalysisResults)

	// 3. Получение файла
	r.GET("/files/:id", proxyHandler.GetFileByID)

	// 4. Получение облака слов
	r.GET("/analysis/wordclouds", proxyHandler.GetWordCloud)

	// Дополнительные эндпоинты
	r.GET("/files", proxyHandler.ListFiles)
	r.GET("/analysis/results-all", proxyHandler.ListAnalysisResults)

	// Swagger документация
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("API Gateway запущен на порту 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

// hasProtocol проверяет, содержит ли URL протокол (http:// или https://)
func hasProtocol(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}

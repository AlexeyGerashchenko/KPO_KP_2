package handlers

import (
	"file_analysis_service/services"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// AnalysisHandler обрабатывает HTTP-запросы, связанные с анализом файлов.
// @Summary Обработчик HTTP-запросов для анализа файлов
// @Description Предоставляет методы для запуска анализа, получения результатов и облаков слов.
// @Tags analysis
// @Accept json
// @Produce json
// @Router /analysis/{file_id} [post]
// @Router /analysis/results/{file_id} [get]
// @Router /analysis/wordclouds [get] // Используем query param для location
type AnalysisHandler struct {
	AnalysisService *services.AnalysisService
}

// NewAnalysisHandler создает новый экземпляр AnalysisHandler.
// @Summary Создает новый AnalysisHandler
// @Description Инициализирует AnalysisHandler с сервисом анализа.
// @Return *AnalysisHandler
func NewAnalysisHandler(analysisService *services.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{AnalysisService: analysisService}
}

// AnalyzeFileRequest определяет структуру запроса на анализ файла.
// @Description Структура запроса для инициирования анализа файла.
// @Name AnalyzeFileRequest
type AnalyzeFileRequest struct {
	FileID string `json:"file_id" binding:"required" example:"unique-file-id"`
}

// RequestAnalysis запускает анализ файла.
// @Summary Запрос на анализ файла
// @Description Инициирует процесс анализа файла по его ID.
// @Tags analysis
// @Param file_id path string true "ID файла для анализа"
// @Produce json
// @Success 202 {object} map[string]string "Сообщение о принятии запроса на анализ"
// @Failure 400 {object} map[string]string "Ошибка валидации ID файла"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера при запуске анализа"
// @Router /analysis/{file_id} [post]
func (h *AnalysisHandler) RequestAnalysis(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id не может быть пустым"})
		return
	}

	// Запускаем анализ асинхронно (в реальном приложении здесь могла бы быть очередь)
	// Для данного примера, выполняем синхронно, но возвращаем 202 Accepted
	go func() {
		_, err := h.AnalysisService.AnalyzeFile(fileID)
		if err != nil {
			// В реальном приложении здесь было бы логирование ошибки
			// Для простоты, просто выводим в консоль сервера
			println(err.Error()) // Используем println, чтобы не импортировать log/fmt здесь
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": fmt.Sprintf("Запрос на анализ файла %s принят", fileID)})
}

// GetAnalysisResults получает результаты анализа файла.
// @Summary Получение результатов анализа
// @Description Возвращает результаты анализа файла (количество абзацев, слов, символов) по его ID.
// @Tags analysis
// @Param file_id path string true "ID файла"
// @Produce json
// @Success 200 {object} models.AnalysisResult "Результаты анализа (без облака слов)"
// @Failure 404 {object} map[string]string "Результаты анализа не найдены"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /analysis/results/{file_id} [get]
func (h *AnalysisHandler) GetAnalysisResults(c *gin.Context) {
	fileID := c.Param("file_id")
	result, err := h.AnalysisService.GetAnalysisResult(fileID)
	if err != nil {
		if _, ok := err.(interface{ NotFound() }); ok || os.IsNotExist(err) || strings.Contains(err.Error(), "не найдены") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := gin.H{
		"file_id":         result.FileID,
		"paragraph_count": result.ParagraphCount,
		"word_count":      result.WordCount,
		"character_count": result.CharacterCount,
	}

	c.JSON(http.StatusOK, response)
}

// GetWordCloud получает изображение облака слов.
// @Summary Получение облака слов
// @Description Возвращает изображение облака слов по его location (пути к файлу).
// @Tags analysis
// @Param location query string true "Location (путь) к файлу облака слов"
// @Produce image/png
// @Produce image/jpeg
// @Produce image/gif
// @Produce image/svg+xml
// @Success 200 {file} file "Изображение облака слов"
// @Failure 400 {object} map[string]string "Параметр location не указан"
// @Failure 404 {object} map[string]string "Облако слов не найдено"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /analysis/wordclouds [get]
func (h *AnalysisHandler) GetWordCloud(c *gin.Context) {
	location := c.Query("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'location' обязателен"})
		return
	}

	fmt.Printf("DEBUG: Получен запрос на облако слов по location: %s\n", location)

	// Проверим, указан ли относительный или абсолютный путь
	// и обработаем соответственно
	if !filepath.IsAbs(location) && !strings.HasPrefix(location, "/app/file_storage_2/") {
		// Если location - относительный путь, добавим к нему путь файлового хранилища
		storagePathEnv := os.Getenv("FILE_STORAGE_PATH")
		if storagePathEnv == "" {
			storagePathEnv = "/app/file_storage_2" // Значение по умолчанию
		}
		location = filepath.Join(storagePathEnv, location)
		fmt.Printf("DEBUG: Преобразован путь к облаку слов: %s\n", location)
	}

	imageData, contentType, err := h.AnalysisService.GetWordCloudImage(location)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "не найден") || strings.Contains(err.Error(), "недопустимый путь") {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Облако слов не найдено по указанному пути: %s. Ошибка: %v", location, err)})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при получении облака слов: %v", err)})
		}
		return
	}

	// Проверка содержимого изображения
	if len(imageData) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Получено изображение с нулевым размером"})
		return
	}

	// По умолчанию устанавливаем Content-Type image/png, если не указан
	if contentType == "" {
		contentType = "image/png"
	}

	// Устанавливаем заголовок, чтобы браузер знал, что это изображение и показал его
	c.Header("Cache-Control", "public, max-age=86400") // Кешировать на 24 часа
	c.Header("Content-Type", contentType)

	// Вместо Content-Disposition: inline, который может вызывать проблемы в некоторых браузерах,
	// просто отдаем изображение напрямую без предложения скачать
	fmt.Printf("DEBUG: Отправка изображения клиенту, Content-Type: %s, размер: %d байт\n", contentType, len(imageData))

	c.Data(http.StatusOK, contentType, imageData)
}

// ListAnalysisResultsHandler возвращает список всех доступных результатов анализа.
// @Summary Список всех результатов анализа (для отладки)
// @Description Возвращает file_id и location облака слов для всех проанализированных файлов.
// @Tags analysis
// @Produce json
// @Success 200 {array} models.AnalysisResult "Список результатов анализа"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /analysis/results-all [get]
func (h *AnalysisHandler) ListAnalysisResultsHandler(c *gin.Context) {
	results, err := h.AnalysisService.ListAnalysisResults()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список результатов анализа: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

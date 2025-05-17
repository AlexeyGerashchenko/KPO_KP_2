package handlers

import (
	"file_storing_service/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileHandler обрабатывает HTTP-запросы, связанные с файлами.
// @Summary Обработчик HTTP-запросов для файлов
// @Description Предоставляет методы для загрузки, получения и перечисления файлов.
// @Tags files
// @Accept json
// @Produce json
// @Router /files [get]
// @Router /files/{id} [get]
// @Router /files/upload [post]
type FileHandler struct {
	DB              *gorm.DB
	FileStoragePath string
}

// NewFileHandler создает новый экземпляр FileHandler.
// @Summary Создает новый FileHandler
// @Description Инициализирует FileHandler с подключением к базе данных и путем к хранилищу файлов.
// @Return *FileHandler
func NewFileHandler(db *gorm.DB, fileStoragePath string) *FileHandler {
	return &FileHandler{DB: db, FileStoragePath: fileStoragePath}
}

// UploadFile загружает файл, сохраняет его метаданные в БД и сам файл в хранилище.
// @Summary Загрузка файла
// @Description Загружает текстовый файл, сохраняет его и возвращает ID.
// @Tags files
// @Accept multipart/form-data
// @Param file formData file true "Файл для загрузки (только .txt)"
// @Produce json
// @Success 201 {object} map[string]string "ID загруженного файла"
// @Failure 400 {object} map[string]string "Ошибка валидации или обработки файла"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /files/upload [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не предоставлен"})
		return
	}

	if filepath.Ext(file.Filename) != ".txt" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат файла. Допускаются только .txt файлы."})
		return
	}

	fileID := uuid.New().String()
	filePath := filepath.Join(h.FileStoragePath, fileID+".txt")

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить файл"})
		return
	}

	fileMetadata := models.File{
		ID:       fileID,
		Name:     file.Filename,
		Location: filePath,
	}

	if err := h.DB.Create(&fileMetadata).Error; err != nil {
		// Попытка удалить файл, если не удалось сохранить метаданные
		_ = os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить метаданные файла"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": fileID})
}

// GetFileByID получает содержимое файла по его ID.
// @Summary Получение файла по ID
// @Description Возвращает содержимое файла по его ID.
// @Tags files
// @Param id path string true "ID файла"
// @Produce plain
// @Success 200 {string} string "Содержимое файла"
// @Failure 404 {object} map[string]string "Файл не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /files/{id} [get]
func (h *FileHandler) GetFileByID(c *gin.Context) {
	fileID := c.Param("id")

	var fileMetadata models.File
	if err := h.DB.First(&fileMetadata, "id = ?", fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Файл не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при поиске файла"})
		}
		return
	}

	content, err := os.ReadFile(fileMetadata.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось прочитать файл"})
		return
	}

	c.String(http.StatusOK, string(content))
}

// GetFileContentByIDInternal используется для внутреннего получения содержимого файла другим сервисом.
// Не является публичным API эндпоинтом.
func (h *FileHandler) GetFileContentByIDInternal(fileID string) ([]byte, error) {
	var fileMetadata models.File
	if err := h.DB.First(&fileMetadata, "id = ?", fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("файл с ID %s не найден", fileID)
		}
		return nil, fmt.Errorf("ошибка при поиске файла с ID %s: %w", fileID, err)
	}

	content, err := os.ReadFile(fileMetadata.Location)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл %s: %w", fileMetadata.Location, err)
	}
	return content, nil
}

// ListFiles возвращает список всех файлов.
// @Summary Список всех файлов
// @Description Возвращает ID и имена всех загруженных файлов.
// @Tags files
// @Produce json
// @Success 200 {array} models.File "Список файлов"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /files [get]
func (h *FileHandler) ListFiles(c *gin.Context) {
	var files []models.File
	if err := h.DB.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список файлов"})
		return
	}
	c.JSON(http.StatusOK, files)
}

// GetFileLocationByID возвращает location файла по его ID. Используется FileAnalysisService.
// @Summary Получение location файла по ID (внутренний)
// @Description Возвращает location файла по его ID для использования другими сервисами.
// @Tags files
// @Param id path string true "ID файла"
// @Produce json
// @Success 200 {object} map[string]string "Location файла"
// @Failure 404 {object} map[string]string "Файл не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /internal/files/{id}/location [get]
func (h *FileHandler) GetFileLocationByID(c *gin.Context) {
	fileID := c.Param("id")
	var fileMetadata models.File
	if err := h.DB.Where("id = ?", fileID).First(&fileMetadata).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Файл не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при поиске файла"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"location": fileMetadata.Location})
}

// GetFileContentByLocationInternal используется для внутреннего получения содержимого файла FileAnalysisService.
// Не является публичным API эндпоинтом.
// @Summary Получение содержимого файла по location (внутренний)
// @Description Возвращает содержимое файла по его location.
// @Tags files
// @Param location query string true "Location файла"
// @Produce plain
// @Success 200 {string} string "Содержимое файла"
// @Failure 404 {object} map[string]string "Файл не найден по указанному location"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /internal/file-content [get]
func (h *FileHandler) GetFileContentByLocationInternal(c *gin.Context) {
	location := c.Query("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр location обязателен"})
		return
	}

	// Проверка, что location находится внутри ожидаемой директории хранения
	// Это простая проверка, в реальном приложении может потребоваться более строгая валидация
	absFileStoragePath, err := filepath.Abs(h.FileStoragePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка определения пути к хранилищу"})
		return
	}
	absLocation, err := filepath.Abs(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка определения абсолютного пути файла"})
		return
	}

	if !filepath.HasPrefix(absLocation, absFileStoragePath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый location файла"})
		return
	}

	content, err := os.ReadFile(location)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Файл не найден по пути: %s", location)})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось прочитать файл: %s", location)})
		}
		return
	}
	c.String(http.StatusOK, string(content))
}

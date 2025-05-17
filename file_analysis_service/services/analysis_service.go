package services

import (
	"bufio"
	"file_analysis_service/models"
	"fmt"
	"path/filepath"
	"pkg/adapters" // Исправленный путь к адаптерам
	"strings"
	"unicode"

	"gorm.io/gorm"
)

// AnalysisService предоставляет методы для анализа файлов.
// @Summary Сервис анализа файлов
// @Description Отвечает за логику анализа текстовых файлов и взаимодействие с зависимостями.
// @Tags services
type AnalysisService struct {
	DBAdapter                 *adapters.DBAdapter
	FileStorageAdapter        *adapters.FileStorageAdapter // Для сохранения облака слов
	FileStoringServiceAdapter *adapters.FileStoringServiceAdapter
	WordCloudAPIAdapter       *adapters.WordCloudAPIAdapter
}

// NewAnalysisService создает новый экземпляр AnalysisService.
// @Summary Создает новый AnalysisService
// @Description Инициализирует сервис анализа файлов со всеми необходимыми адаптерами.
// @Return *AnalysisService
func NewAnalysisService(
	dbAdapter *adapters.DBAdapter,
	fileStorageAdapter *adapters.FileStorageAdapter,
	fileStoringServiceAdapter *adapters.FileStoringServiceAdapter,
	wordCloudAPIAdapter *adapters.WordCloudAPIAdapter,
) *AnalysisService {
	return &AnalysisService{
		DBAdapter:                 dbAdapter,
		FileStorageAdapter:        fileStorageAdapter,
		FileStoringServiceAdapter: fileStoringServiceAdapter,
		WordCloudAPIAdapter:       wordCloudAPIAdapter,
	}
}

// AnalyzeFile выполняет анализ файла: подсчитывает абзацы, слова, символы и генерирует облако слов.
// @Summary Анализ файла
// @Description Основной метод для анализа файла. Возвращает результаты анализа или ошибку.
// @Param fileID path string true "ID файла для анализа"
// @Return *models.AnalysisResult, error "Результаты анализа и ошибка, если есть"
func (s *AnalysisService) AnalyzeFile(fileID string) (*models.AnalysisResult, error) {
	// 1. Попытка получить результаты ранее проведенного анализа из БД
	var existingResult models.AnalysisResult
	if err := s.DBAdapter.First(&existingResult, "file_id = ?", fileID); err == nil {
		return &existingResult, nil // Результаты найдены, возвращаем их
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("ошибка при поиске существующего анализа для fileID %s: %w", fileID, err)
	}

	// 2. File Analisys Service обращается к File Storing Service чтобы получить содержимое файла по id
	// Сначала получаем location файла
	fileLocationOriginal, err := s.FileStoringServiceAdapter.GetFileLocationByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить местоположение файла %s из FileStoringService: %w", fileID, err)
	}

	// Затем получаем содержимое файла по location
	fileContent, err := s.FileStoringServiceAdapter.GetFileContentByLocation(fileLocationOriginal)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить содержимое файла %s (location: %s) из FileStoringService: %w", fileID, fileLocationOriginal, err)
	}

	// 3. Анализ файла
	paragraphCount := countParagraphs(string(fileContent))
	wordCount := countWords(string(fileContent))
	characterCount := countCharacters(string(fileContent))

	// 4. Генерация облака слов
	wordCloudImage, contentType, err := s.WordCloudAPIAdapter.GenerateWordCloud(string(fileContent))
	if err != nil {
		// Не фатальная ошибка, анализ продолжается без облака слов, если API недоступен
		fmt.Printf("Предупреждение: не удалось сгенерировать облако слов для fileID %s: %v\n", fileID, err)
		// Можно логировать эту ошибку, но не прерывать процесс
	}

	wordCloudLocation := "" // Пусто, если генерация не удалась
	if wordCloudImage != nil {
		// Выводим информацию о полученном изображении для отладки
		fmt.Printf("DEBUG: Получено изображение (%d байт, Content-Type: %s) для сохранения\n", len(wordCloudImage), contentType)

		// Определяем расширение файла на основе Content-Type
		fileExt := ".png" // По умолчанию
		if contentType == "image/jpeg" || contentType == "image/jpg" {
			fileExt = ".jpg"
		} else if contentType == "image/gif" {
			fileExt = ".gif"
		} else if contentType == "image/svg+xml" {
			fileExt = ".svg"
		}

		// Сохранение полученной по API картинки в File Storage №2
		// Включаем ID файла и четко указываем, что это облако слов с правильным расширением
		wordCloudFileName := fmt.Sprintf("%s_wordcloud%s", fileID, fileExt)
		actualWordCloudLocation, errSaveCloud := s.FileStorageAdapter.SaveFileFromBytes(wordCloudFileName, wordCloudImage)
		if errSaveCloud != nil {
			// Ошибка сохранения облака слов, не фатально, но логируем
			fmt.Printf("Предупреждение: не удалось сохранить облако слов для fileID %s: %v\n", fileID, errSaveCloud)
		} else {
			wordCloudLocation = actualWordCloudLocation
		}
	}

	// 5. Сохранение результатов анализа в БД
	analysisResult := models.AnalysisResult{
		FileID:            fileID,
		ParagraphCount:    paragraphCount,
		WordCount:         wordCount,
		CharacterCount:    characterCount,
		WordCloudLocation: wordCloudLocation, // Сохраняем фактический путь или пустую строку
	}

	if err := s.DBAdapter.Create(&analysisResult); err != nil {
		return nil, fmt.Errorf("не удалось сохранить результаты анализа для fileID %s: %w", fileID, err)
	}

	return &analysisResult, nil
}

// GetAnalysisResult получает результаты анализа по ID файла.
// @Summary Получение результатов анализа
// @Description Ищет и возвращает сохраненные результаты анализа для указанного файла.
// @Param fileID path string true "ID файла"
// @Return *models.AnalysisResult, error "Результаты анализа и ошибка, если есть (например, если анализ не найден)"
func (s *AnalysisService) GetAnalysisResult(fileID string) (*models.AnalysisResult, error) {
	var result models.AnalysisResult
	if err := s.DBAdapter.First(&result, "file_id = ?", fileID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("результаты анализа для файла с ID %s не найдены: %w", fileID, err)
		}
		return nil, fmt.Errorf("ошибка при поиске результатов анализа для файла %s: %w", fileID, err)
	}
	return &result, nil
}

// GetWordCloudImage получает изображение облака слов по его местоположению.
// @Summary Получение изображения облака слов
// @Description Читает и возвращает изображение облака слов из файлового хранилища.
// @Param location path string true "Путь к файлу изображения облака слов"
// @Return []byte, string, error "Данные изображения, тип контента и ошибка, если есть"
func (s *AnalysisService) GetWordCloudImage(location string) ([]byte, string, error) {
	// Проверка, что location не пустой и безопасный (например, не выходит за пределы хранилища)
	// filepath.Clean для нормализации пути
	cleanedLocation := filepath.Clean(location)

	// Проверка, что путь после очистки все еще указывает на ожидаемую директорию.
	// Это базовая проверка, в реальном приложении могут потребоваться более строгие правила.
	expectedPrefix, _ := filepath.Abs(s.FileStorageAdapter.StoragePath)
	absLocation, err := filepath.Abs(cleanedLocation)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка получения абсолютного пути для облака слов %s: %w", cleanedLocation, err)
	}

	if !strings.HasPrefix(absLocation, expectedPrefix) {
		return nil, "", fmt.Errorf("недопустимый путь к файлу облака слов: %s", cleanedLocation)
	}

	// Извлекаем относительный путь от StoragePath, чтобы использовать с адаптером
	relativePath, err := filepath.Rel(s.FileStorageAdapter.StoragePath, absLocation)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка вычисления относительного пути для облака слов %s: %w", absLocation, err)
	}

	// Определяем Content-Type на основе расширения файла
	contentType := "image/png" // По умолчанию
	ext := strings.ToLower(filepath.Ext(relativePath))
	if ext == ".jpg" || ext == ".jpeg" {
		contentType = "image/jpeg"
	} else if ext == ".gif" {
		contentType = "image/gif"
	} else if ext == ".svg" {
		contentType = "image/svg+xml"
	}

	fmt.Printf("DEBUG: Чтение файла облака слов %s с Content-Type %s\n", relativePath, contentType)

	imageData, err := s.FileStorageAdapter.ReadFile(relativePath)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось прочитать файл облака слов %s: %w", relativePath, err)
	}

	// Проверяем размер данных
	if len(imageData) == 0 {
		return nil, "", fmt.Errorf("файл облака слов %s пуст", relativePath)
	}

	fmt.Printf("DEBUG: Прочитано %d байт из файла облака слов %s\n", len(imageData), relativePath)

	return imageData, contentType, nil
}

// Вспомогательные функции для анализа текста
func countParagraphs(text string) int {
	if text == "" {
		return 0
	}
	// Разделяем по переносу строки. Пустые строки между абзацами будут считаться как отдельные абзацы, если их много подряд.
	// Чтобы считать "реальные" абзацы, можно дополнительно отфильтровать пустые строки после split.
	scanner := bufio.NewScanner(strings.NewReader(text))
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" { // Считаем только непустые строки как абзацы
			count++
		}
	}
	// Если текст непустой, но не содержит переносов строк, считаем его одним абзацем
	if count == 0 && len(strings.TrimSpace(text)) > 0 {
		return 1
	}
	return count
}

func countWords(text string) int {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanWords)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}

func countCharacters(text string) int {
	count := 0
	for _, r := range text {
		if !unicode.IsSpace(r) { // Не считаем пробельные символы
			count++
		}
	}
	return count
}

// ListAnalysisResults возвращает список всех доступных результатов анализа.
// @Summary Список всех результатов анализа
// @Description Возвращает file_id и location облака слов для всех проанализированных файлов.
// @Tags analysis
// @Produce json
// @Success 200 {array} models.AnalysisResult "Список результатов анализа"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /analysis/results [get]
func (s *AnalysisService) ListAnalysisResults() ([]models.AnalysisResult, error) {
	var results []models.AnalysisResult
	if err := s.DBAdapter.Find(&results); err != nil {
		return nil, fmt.Errorf("не удалось получить список результатов анализа: %w", err)
	}
	return results, nil
}

package adapters

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// FileStoringServiceAdapter предоставляет интерфейс для взаимодействия с FileStoringService.
// @Summary Адаптер для FileStoringService
// @Description Обеспечивает методы для получения данных о файлах из FileStoringService.
// @Tags adapters
type FileStoringServiceAdapter struct {
	ServiceBaseURL string // Базовый URL FileStoringService, например, http://file_storing_service:8081
}

// NewFileStoringServiceAdapter создает новый экземпляр FileStoringServiceAdapter.
// @Summary Создает новый FileStoringServiceAdapter
// @Description Инициализирует адаптер с базовым URL FileStoringService.
// @Param serviceBaseURL Базовый URL FileStoringService
// @Return *FileStoringServiceAdapter
func NewFileStoringServiceAdapter(serviceBaseURL string) *FileStoringServiceAdapter {
	return &FileStoringServiceAdapter{ServiceBaseURL: serviceBaseURL}
}

// FileLocationResponse определяет структуру ответа для получения location файла.
// @Description Структура ответа с местоположением файла.
// @Name FileLocationResponse
type FileLocationResponse struct {
	Location string `json:"location"`
}

// GetFileLocationByID запрашивает у FileStoringService местоположение файла по его ID.
// @Summary Получение местоположения файла
// @Description Обращается к FileStoringService для получения пути к файлу.
// @Param fileID ID файла
// @Return string, error "Местоположение файла и ошибка, если есть"
func (a *FileStoringServiceAdapter) GetFileLocationByID(fileID string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/internal/files/%s/location", a.ServiceBaseURL, fileID))
	if err != nil {
		return "", fmt.Errorf("ошибка при запросе местоположения файла %s: %w", fileID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, errRead := ioutil.ReadAll(resp.Body)
		if errRead != nil {
			return "", fmt.Errorf("FileStoringService вернул ошибку %d для файла %s и не удалось прочитать тело ответа: %w", resp.StatusCode, fileID, errRead)
		}
		return "", fmt.Errorf("FileStoringService вернул ошибку %d для файла %s: %s", resp.StatusCode, fileID, string(body))
	}

	var locationResp FileLocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&locationResp); err != nil {
		return "", fmt.Errorf("ошибка при декодировании ответа от FileStoringService для файла %s: %w", fileID, err)
	}

	return locationResp.Location, nil
}

// GetFileContentByLocation запрашивает у FileStoringService содержимое файла по его местоположению.
// @Summary Получение содержимого файла
// @Description Обращается к FileStoringService для получения содержимого файла.
// @Param location Местоположение файла
// @Return []byte, error "Содержимое файла и ошибка, если есть"
func (a *FileStoringServiceAdapter) GetFileContentByLocation(location string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/internal/file-content?location=%s", a.ServiceBaseURL, location))
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе содержимого файла по location %s: %w", location, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, errRead := ioutil.ReadAll(resp.Body)
		if errRead != nil {
			return nil, fmt.Errorf("FileStoringService вернул ошибку %d для location %s и не удалось прочитать тело ответа: %w", resp.StatusCode, location, errRead)
		}
		return nil, fmt.Errorf("FileStoringService вернул ошибку %d для location %s: %s", resp.StatusCode, location, string(body))
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа от FileStoringService для location %s: %w", location, err)
	}

	return content, nil
}

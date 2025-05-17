package adapters

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// WordCloudAPIAdapter предоставляет интерфейс для взаимодействия с WordCloudAPI.
// @Summary Адаптер для WordCloudAPI
// @Description Обеспечивает метод для генерации облака слов с использованием внешнего API.
// @Tags adapters
type WordCloudAPIAdapter struct {
	BaseURL string // Базовый URL WordCloudAPI, например, https://quickchart.io/wordcloud
}

// NewWordCloudAPIAdapter создает новый экземпляр WordCloudAPIAdapter.
// @Summary Создает новый WordCloudAPIAdapter
// @Description Инициализирует адаптер с базовым URL WordCloudAPI.
// @Param baseURL Базовый URL WordCloudAPI
// @Return *WordCloudAPIAdapter
func NewWordCloudAPIAdapter(baseURL string) *WordCloudAPIAdapter {
	return &WordCloudAPIAdapter{BaseURL: baseURL}
}

// GenerateWordCloud генерирует изображение облака слов для заданного текста.
// @Summary Генерация облака слов
// @Description Отправляет текст в WordCloudAPI и возвращает полученное изображение в виде байтов.
// @Param text Текст для генерации облака слов
// @Return []byte, string, error "Изображение облака слов, его Content-Type и ошибка, если есть"
func (a *WordCloudAPIAdapter) GenerateWordCloud(text string) ([]byte, string, error) {
	// Формируем URL с параметром text
	apiURL, err := url.Parse(a.BaseURL)
	if err != nil {
		return nil, "", fmt.Errorf("неверный базовый URL для WordCloudAPI: %w", err)
	}
	q := apiURL.Query()
	q.Set("text", text)
	apiURL.RawQuery = q.Encode()

	// Отображаем URL для дебага
	fmt.Printf("DEBUG: Запрос к WordCloudAPI: %s\n", apiURL.String())

	resp, err := http.Get(apiURL.String())
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при запросе к WordCloudAPI: %w", err)
	}
	defer resp.Body.Close()

	// Логируем статус ответа и заголовки для отладки
	fmt.Printf("DEBUG: Ответ от WordCloudAPI: статус %d, заголовки: %+v\n", resp.StatusCode, resp.Header)

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body) // Читаем тело ответа для информации об ошибке
		return nil, "", fmt.Errorf("WordCloudAPI вернул ошибку %d: %s", resp.StatusCode, string(body))
	}

	// Получаем Content-Type из заголовка ответа
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// По умолчанию предполагаем PNG, но логируем предупреждение
		contentType = "image/png"
		fmt.Printf("WARN: WordCloudAPI не вернул Content-Type, используем по умолчанию %s\n", contentType)
	} else if !strings.HasPrefix(contentType, "image/") {
		// Если контент не является изображением, возвращаем ошибку
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("WordCloudAPI вернул неожиданный Content-Type: %s. Тело ответа: %s", contentType, string(body))
	}

	imageData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при чтении ответа от WordCloudAPI: %w", err)
	}

	// Проверяем, что получили хоть какие-то данные
	if len(imageData) == 0 {
		return nil, "", fmt.Errorf("WordCloudAPI вернул пустой ответ")
	}

	fmt.Printf("DEBUG: Получено %d байт данных из WordCloudAPI, Content-Type: %s\n", len(imageData), contentType)

	return imageData, contentType, nil
}

package handlers

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// ProxyHandler обрабатывает проксирование запросов к другим микросервисам.
// @Summary Обработчик прокси-запросов
// @Description Маршрутизирует запросы к соответствующим микросервисам.
// @Tags proxy
// @Router /
// @Accept json
// @Produce json
type ProxyHandler struct {
	FileStoringServiceAddr  string
	FileAnalysisServiceAddr string
}

// NewProxyHandler создает новый экземпляр ProxyHandler.
// @Summary Создает новый ProxyHandler
// @Description Инициализирует ProxyHandler с адресами целевых сервисов.
// @Return *ProxyHandler
func NewProxyHandler(fileStoringServiceAddr, fileAnalysisServiceAddr string) *ProxyHandler {
	return &ProxyHandler{
		FileStoringServiceAddr:  fileStoringServiceAddr,
		FileAnalysisServiceAddr: fileAnalysisServiceAddr,
	}
}

// @Summary Прокси для загрузки файла (Сценарий 1)
// @Description Перенаправляет запрос на загрузку файла в File Storing Service.
// @Tags files
// @Accept multipart/form-data
// @Param file formData file true "Файл для загрузки (только .txt)"
// @Produce json
// @Success 201 {object} map[string]string "ID загруженного файла"
// @Failure 400 {object} map[string]string "Ошибка запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера или ошибка File Storing Service"
// @Router /upload [post]
func (h *ProxyHandler) UploadFile(c *gin.Context) {
	h.proxyRequest(c, h.FileStoringServiceAddr, "/api/v1/files/upload")
}

// @Summary Прокси для анализа файла (Сценарий 2)
// @Description Перенаправляет запрос на анализ файла в File Analysis Service.
// @Tags analysis
// @Param file_id path string true "ID файла для анализа"
// @Produce json
// @Success 202 {object} map[string]string "Сообщение о принятии запроса на анализ"
// @Failure 400 {object} map[string]string "Ошибка запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера или ошибка File Analysis Service"
// @Router /analysis/{file_id} [post]
func (h *ProxyHandler) RequestAnalysis(c *gin.Context) {
	// file_id извлекается из пути в proxyRequest
	h.proxyRequest(c, h.FileAnalysisServiceAddr, "/api/v1/analysis/"+c.Param("file_id"))
}

// @Summary Прокси для получения результатов анализа файла (Сценарий 2)
// @Description Перенаправляет запрос на получение результатов анализа в File Analysis Service.
// @Tags analysis
// @Param file_id path string true "ID файла"
// @Produce json
// @Success 200 {object} map[string]any "Результаты анализа (file_id, paragraph_count, word_count, character_count)"
// @Failure 404 {object} map[string]string "Результаты анализа не найдены"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера или ошибка File Analysis Service"
// @Router /analysis/results/{file_id} [get]
func (h *ProxyHandler) GetAnalysisResults(c *gin.Context) {
	h.proxyRequest(c, h.FileAnalysisServiceAddr, "/api/v1/analysis/results/"+c.Param("file_id"))
}

// @Summary Прокси для получения файла (Сценарий 3)
// @Description Перенаправляет запрос на получение файла в File Storing Service.
// @Tags files
// @Param id path string true "ID файла"
// @Produce plain
// @Success 200 {string} string "Содержимое файла"
// @Failure 404 {object} map[string]string "Файл не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера или ошибка File Storing Service"
// @Router /files/{id} [get]
func (h *ProxyHandler) GetFileByID(c *gin.Context) {
	h.proxyRequest(c, h.FileStoringServiceAddr, "/api/v1/files/"+c.Param("id"))
}

// @Summary Прокси для получения облака слов (Сценарий 4)
// @Description Перенаправляет запрос на получение облака слов в File Analysis Service.
// @Tags analysis
// @Param location query string true "Location (путь) к файлу облака слов"
// @Produce image/png
// @Success 200 {file} file "Изображение облака слов"
// @Failure 400 {object} map[string]string "Параметр location не указан"
// @Failure 404 {object} map[string]string "Облако слов не найдено"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера или ошибка File Analysis Service"
// @Router /analysis/wordclouds [get]
func (h *ProxyHandler) GetWordCloud(c *gin.Context) {
	// location передается как query параметр, proxyRequest это учтет
	h.proxyRequest(c, h.FileAnalysisServiceAddr, "/api/v1/analysis/wordclouds")
}

// @Summary Прокси для получения списка всех файлов (дополнительно)
// @Description Перенаправляет запрос на получение списка всех файлов в File Storing Service.
// @Tags files
// @Produce json
// @Success 200 {array} map[string]any "Список файлов (каждый элемент с id, name, location)"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера или ошибка File Storing Service"
// @Router /files [get]
func (h *ProxyHandler) ListFiles(c *gin.Context) {
	h.proxyRequest(c, h.FileStoringServiceAddr, "/api/v1/files")
}

// @Summary Прокси для получения списка всех результатов анализа (дополнительно)
// @Description Перенаправляет запрос на получение списка всех результатов анализа в File Analysis Service.
// @Tags analysis
// @Produce json
// @Success 200 {array} map[string]any "Список результатов анализа (каждый элемент с file_id, paragraph_count, etc.)"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера или ошибка File Analysis Service"
// @Router /analysis/results-all [get]
func (h *ProxyHandler) ListAnalysisResults(c *gin.Context) {
	h.proxyRequest(c, h.FileAnalysisServiceAddr, "/api/v1/analysis/results-all")
}

func (h *ProxyHandler) proxyRequest(c *gin.Context, targetServiceBaseURL, targetPath string) {
	targetURL, err := url.Parse(targetServiceBaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target service URL config"})
		return
	}
	targetURL.Path = targetPath
	targetURL.RawQuery = c.Request.URL.RawQuery

	var reqBody io.Reader
	contentType := c.Request.Header.Get("Content-Type")

	// Специальная обработка для multipart/form-data (загрузка файлов)
	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			// Проверяем, может это другой multipart запрос без файла "file"
			if err == http.ErrMissingFile {
				// Если это не загрузка файла, а другой multipart, пытаемся проксировать как есть
				// Это может быть полезно, если File Storing Service ожидает другие поля multipart
				// В данном случае, мы знаем, что /upload ожидает "file", так что это больше для общности
				if c.Request.Body != nil {
					bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
					c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем тело для Gin
					reqBody = bytes.NewBuffer(bodyBytes)                          // Используем для прокси-запроса
				} else {
					reqBody = nil
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing form file: " + err.Error()})
				return
			}
		} else {
			defer file.Close()

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("file", header.Filename)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating form file for proxy: " + err.Error()})
				return
			}
			_, err = io.Copy(part, file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error copying file data for proxy: " + err.Error()})
				return
			}
			// Копирование других полей формы, если они есть
			for key, values := range c.Request.MultipartForm.Value {
				for _, value := range values {
					_ = writer.WriteField(key, value)
				}
			}
			err = writer.Close()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error closing multipart writer for proxy: " + err.Error()})
				return
			}
			reqBody = body
			contentType = writer.FormDataContentType()
		}
	} else if c.Request.Body != nil {
		bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем тело, чтобы Gin мог его прочитать, если нужно
		reqBody = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(c.Request.Method, targetURL.String(), reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating proxy request: " + err.Error()})
		return
	}

	// Копируем заголовки, кроме Host, так как он устанавливается транспортным уровнем
	for k, v := range c.Request.Header {
		if k != "Host" {
			req.Header[k] = v
		}
	}
	// Устанавливаем Content-Type, если он был изменен (например, для multipart)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Проверка на ошибку подключения (например, сервис упал)
		if os.IsTimeout(err) || strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "no such host") {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Service %s is unavailable: %s", targetServiceBaseURL, err.Error())})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending proxy request to " + targetServiceBaseURL + ": " + err.Error()})
		}
		return
	}
	defer resp.Body.Close()

	// Копируем заголовки ответа от целевого сервиса
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}

	// Копируем тело ответа
	// Используем io.Copy для эффективности, особенно для больших ответов (например, файлов)
	c.Status(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		// Если уже начали писать ответ, сложно что-то сделать, кроме как логировать
		log.Printf("Error copying response body to client: %v", err)
	}
}

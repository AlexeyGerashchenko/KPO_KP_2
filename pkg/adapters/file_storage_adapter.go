package adapters

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileStorageAdapter предоставляет интерфейс для сохранения и чтения файлов.
// @Summary Адаптер для файлового хранилища
// @Description Унифицирует операции сохранения и чтения файлов с диска.
// @Tags adapters
type FileStorageAdapter struct {
	StoragePath string // Путь к корневой директории хранилища
}

// NewFileStorageAdapter создает новый экземпляр FileStorageAdapter.
// @Summary Создает новый FileStorageAdapter
// @Description Инициализирует FileStorageAdapter с путем к директории хранилища.
// @Param storagePath Путь к директории хранилища
// @Return *FileStorageAdapter, error
func NewFileStorageAdapter(storagePath string) (*FileStorageAdapter, error) {
	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию хранилища %s: %w", storagePath, err)
	}
	return &FileStorageAdapter{StoragePath: storagePath}, nil
}

// SaveFile сохраняет данные в файл.
// @Summary Сохранение файла
// @Description Записывает содержимое io.Reader в файл по указанному пути относительно StoragePath.
// @Param relativePath Относительный путь к файлу внутри хранилища
// @Param data io.Reader с данными для сохранения
// @Return string, error "Полный путь к сохраненному файлу и ошибка, если есть"
func (a *FileStorageAdapter) SaveFile(relativePath string, data io.Reader) (string, error) {
	filePath := filepath.Join(a.StoragePath, relativePath)

	// Создаем все необходимые директории по пути к файлу
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return "", fmt.Errorf("не удалось создать директории для файла %s: %w", filePath, err)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("не удалось создать файл %s: %w", filePath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, data)
	if err != nil {
		return "", fmt.Errorf("не удалось записать данные в файл %s: %w", filePath, err)
	}
	return filePath, nil
}

// SaveFileFromBytes сохраняет байтовый массив в файл.
// @Summary Сохранение файла из байтов
// @Description Записывает байтовый массив в файл по указанному пути относительно StoragePath.
// @Param relativePath Относительный путь к файлу внутри хранилища
// @Param data Массив байт для сохранения
// @Return string, error "Полный путь к сохраненному файлу и ошибка, если есть"
func (a *FileStorageAdapter) SaveFileFromBytes(relativePath string, data []byte) (string, error) {
	filePath := filepath.Join(a.StoragePath, relativePath)

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return "", fmt.Errorf("не удалось создать директории для файла %s: %w", filePath, err)
	}

	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("не удалось записать данные в файл %s: %w", filePath, err)
	}
	return filePath, nil
}

// ReadFile читает содержимое файла.
// @Summary Чтение файла
// @Description Читает и возвращает содержимое файла по указанному пути относительно StoragePath.
// @Param relativePath Относительный путь к файлу внутри хранилища
// @Return []byte, error "Содержимое файла и ошибка, если есть"
func (a *FileStorageAdapter) ReadFile(relativePath string) ([]byte, error) {
	filePath := filepath.Join(a.StoragePath, relativePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл %s: %w", filePath, err)
	}
	return data, nil
}

// GetAbsPath возвращает абсолютный путь к файлу в хранилище.
// @Summary Получение абсолютного пути
// @Description Возвращает абсолютный путь к файлу, комбинируя StoragePath и relativePath.
// @Param relativePath Относительный путь к файлу внутри хранилища
// @Return string, error "Абсолютный путь и ошибка, если есть"
func (a *FileStorageAdapter) GetAbsPath(relativePath string) (string, error) {
	absPath, err := filepath.Abs(filepath.Join(a.StoragePath, relativePath))
	if err != nil {
		return "", fmt.Errorf("ошибка при получении абсолютного пути для %s: %w", relativePath, err)
	}
	return absPath, nil
}

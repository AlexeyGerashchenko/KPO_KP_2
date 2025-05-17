package models

import (
	"time"

	"gorm.io/gorm"
)

// File представляет метаданные файла в БД
// @Description Метаданные файла, хранящиеся в базе данных.
// @Name File
// @ δεύτεροςτύπος_id базовый
// @swaggertype object
// @property id string example="unique-file-id" Описание: ID файла.
// @property name string example="example.txt" Описание: Имя файла.
// @property location string example="/app/file_storage_1/unique-file-id.txt" Описание: Путь к файлу.
// @property created_at string example="2023-01-01T12:00:00Z" Описание: Время создания.
// @property updated_at string example="2023-01-01T13:00:00Z" Описание: Время последнего обновления.
// @property deleted_at string example="" Описание: Время удаления (если удален).
type File struct {
	ID        string         `gorm:"primaryKey" json:"id" example:"unique-file-id"`
	Name      string         `json:"name" example:"example.txt"`
	Location  string         `json:"location" example:"/app/file_storage_1/unique-file-id.txt"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggertype:"string" example:"2023-01-01T14:00:00Z"` // Время удаления (если удален)
}

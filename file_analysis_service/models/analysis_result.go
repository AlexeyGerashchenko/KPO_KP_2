package models

import (
	"time"

	"gorm.io/gorm"
)

// AnalysisResult представляет результаты анализа файла.
// @Description Результаты анализа текстового файла, включая количество абзацев, слов, символов и путь к облаку слов.
// @Name AnalysisResult
type AnalysisResult struct {
	// gorm.Model заменено на явные поля для Swagger
	ID        uint           `json:"id" swaggertype:"integer" example:"1"` // ID записи анализа
	CreatedAt time.Time      `json:"created_at" swaggertype:"string" format:"date-time"`
	UpdatedAt time.Time      `json:"updated_at" swaggertype:"string" format:"date-time"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggertype:"string" format:"date-time"`

	FileID            string `json:"file_id" gorm:"uniqueIndex" example:"unique-file-id"` // ID оригинального файла
	ParagraphCount    int    `json:"paragraph_count" example:"5"`
	WordCount         int    `json:"word_count" example:"250"`
	CharacterCount    int    `json:"character_count" example:"1500"`
	WordCloudLocation string `json:"word_cloud_location" example:"/app/file_storage_2/unique-file-id_wordcloud.png"` // Путь к сохраненному изображению облака слов
}

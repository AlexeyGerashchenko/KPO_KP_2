package adapters

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBAdapter предоставляет интерфейс для работы с базой данных.
// @Summary Адаптер для базы данных
// @Description Унифицирует взаимодействие с базой данных GORM.
// @Tags adapters
type DBAdapter struct {
	DB *gorm.DB
}

// NewDBAdapter создает новый экземпляр DBAdapter.
// @Summary Создает новый DBAdapter
// @Description Инициализирует DBAdapter с подключением к базе данных.
// @Param dsn Строка подключения к базе данных
// @Param config Конфигурация GORM
// @Return *DBAdapter, error
func NewDBAdapter(dsn string, config *gorm.Config) (*DBAdapter, error) {
	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}
	return &DBAdapter{DB: db}, nil
}

// AutoMigrate выполняет автоматическую миграцию схемы базы данных.
// @Summary Автоматическая миграция схемы
// @Description Применяет изменения моделей к схеме базы данных.
// @Param models Список моделей для миграции
// @Return error
func (a *DBAdapter) AutoMigrate(models ...interface{}) error {
	return a.DB.AutoMigrate(models...)
}

// Create создает новую запись в базе данных.
// @Summary Создание записи
// @Description Сохраняет новую сущность в базе данных.
// @Param value Указатель на сущность для создания
// @Return error
func (a *DBAdapter) Create(value interface{}) error {
	return a.DB.Create(value).Error
}

// First находит первую запись, соответствующую условиям.
// @Summary Поиск первой записи
// @Description Находит первую запись, удовлетворяющую заданным условиям.
// @Param out Указатель на переменную для результата
// @Param where Условия поиска (например, "id = ?")
// @Param args Аргументы для условий поиска
// @Return error
func (a *DBAdapter) First(out interface{}, where ...interface{}) error {
	return a.DB.First(out, where...).Error
}

// Find находит все записи, соответствующие условиям.
// @Summary Поиск всех записей
// @Description Находит все записи, удовлетворяющие заданным условиям.
// @Param out Указатель на срез для результатов
// @Param where Условия поиска (опционально)
// @Return error
func (a *DBAdapter) Find(out interface{}, where ...interface{}) error {
	return a.DB.Find(out, where...).Error
}

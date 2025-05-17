module file_analysis_service

go 1.20

require (
	github.com/gin-gonic/gin v1.9.1
	gorm.io/driver/postgres v1.5.2
	gorm.io/gorm v1.25.2
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.2
	pkg v0.0.0 // Псевдо-версия для локального пакета
)

replace pkg => ../pkg // Путь к общему пакету pkg 
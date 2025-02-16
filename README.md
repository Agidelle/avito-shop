# 'README'

Команды:\
`go run main.go serve` - запуск сервиса\
`go run main.go migration` - запуск миграций

Файл `init.sql` также содержит скрипт миграции БД

Структура проекта:\
`/cmd` - запуск сервиса и миграции\
`/config` - содержит конфигурации сервиса\
`/internal/config` - загрузка конфигураций\
`/internal/http-server` - содержит: auth, handlers и middleware для обработки запросов\
`/internal/service/shop` - бизнес логика сервиса\
`/internal/service/shop/storage` - реализация работы с БД

Тесты:
`/internal/http-server/handlers_test.go`\
`/internal/service/shop/service_test.go`\
`/internal/service/shop/storage/mysql/mysql_test.go`

# `/xxx`


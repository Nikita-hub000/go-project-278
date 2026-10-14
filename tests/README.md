# Тесты

Требуется версия Go из `go.mod`.

Структура:

- `internal/links/handler_test.go` — HTTP-валидация POST/PUT, JSON и статусы ошибок; работает без PostgreSQL.
- `internal/links/errors_test.go` — преобразование PostgreSQL `23505` в ошибку занятого имени.
- `internal/visits/service_test.go` — допустимые поля сортировки, безопасные значения по умолчанию и границы пагинации.
- `internal/visits/handler_test.go` — ошибки БД и некорректная пагинация; работает без PostgreSQL.
- `internal/api/router_test.go` — сборка маршрутов, `/ping` и CORS preflight.
- `tests/integration` — HTTP CRUD, конфликты имён, одновременные запросы, сортировка, каскадное удаление и работа с настоящей БД.
- `tests/integration/visits_test.go` — повторные переходы, редирект `302` с `no-store`, данные посещений, списки, ошибки записи и отсутствие редиректа при ошибке.

Быстрые проверки (интеграционные тесты без переменной подключения пропускаются):

```bash
go test -race ./...
```

Полный прогон с PostgreSQL:

```bash
TEST_DATABASE_URL='postgres://USER:PASSWORD@localhost:5432/TEST_DB?sslmode=disable' \
go test -race ./...
```

Используйте отдельную тестовую БД. Пользователю нужны права на создание схем.
Каждый интеграционный тест создаёт свою схему `api_test_*` и удаляет её после
завершения. Тесты не читают `.env`. В CI PostgreSQL запускается автоматически.

POST и PUT принимают обязательный `original_url` и необязательный `short_name`
длиной 3–32 символа. Пустое или отсутствующее имя генерируется на сервере.
Явно переданное занятое имя возвращает `422`, без изменения имени.
Ошибки полей имеют формат `{"errors":{"field":"message"}}`, неверный JSON —
`400` с `{"error":"invalid request"}`. `created_at` назначает PostgreSQL.

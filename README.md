# Сокращатель ссылок (Go)

[![hexlet-check](https://github.com/Nikita-hub000/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/Nikita-hub000/go-project-278/actions)
[![CI](https://github.com/Nikita-hub000/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/Nikita-hub000/go-project-278/actions/workflows/ci.yml)

Спроектируйте приложение для удобных ссылок

Ссылка на RENDER: https://go-project-278-v4sc.onrender.com

## Стек

- Go

## Установка

<!-- Опишите установку: клонирование, зависимости, переменные окружения -->

```bash
git clone https://github.com/Nikita-hub000/go-project-278.git
cd go-project-278
```

## Использование

### Общий запуск фронтенда и API

Нужны запущенный Docker Desktop, настроенный `.env` и Node.js из `.nvmrc`.
Для управления Node.js используется nvm:

```bash
nvm install
nvm use
node -v
npm ci
npm start
```

`npm start` сначала запускает PostgreSQL и ждёт его готовности, затем через
`concurrently` запускает API в Docker Compose с пересборкой и фронтенд Hexlet
командой `npx start-hexlet-url-shortener-frontend`.

- Фронтенд: http://localhost:5173/
- API: http://localhost:8080/api/links
- Проверка API: `curl http://localhost:8080/ping`

`Ctrl+C` останавливает фронтенд и API. PostgreSQL остаётся в фоне;
остановить его можно командой `docker compose stop db`. Данные сохраняются.
Для отдельного запуска фронтенда есть `npm run dev:frontend`.
Настройка CORS остаётся отдельным шагом; этот запуск её не добавляет.

### Запуск только API

<!-- Добавьте примеры запуска и запись asciinema — именно это смотрит работодатель -->

```bash
go run ./cmd/api
curl http://localhost:8080/ping
# {"message":"pong"}
```

`make start` запускает API, `make build` собирает его в `bin/api`.
Адрес задаётся переменной окружения `HTTP_ADDR`, по умолчанию `:8080`.
Например: `HTTP_ADDR=:9090 go run ./cmd/api`.

Для разработки с Air скопируйте `.env.example` в `.env` и выполните `air`.
Air загружает `.env`; при обычном `go run` переменные нужно передать через окружение.
Локальный `.env` исключён из Git и Docker build context.

Запуск API и PostgreSQL через Docker Compose:

```bash
cp .env.example .env
# Если .env уже настроен, не перезаписывайте его командой выше.
docker compose up --build -d
docker compose logs api
```

В `.env` задаются `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` и
`DATABASE_URL`. Пользователь, пароль и имя базы в URL должны совпадать с
параметрами PostgreSQL. В контейнере адрес базы — `db:5432`;
из GoLand или другого клиента на компьютере — `localhost:5432`.
Параметры `POSTGRES_*` используются для первоначального создания базы в пустом томе;
изменение `.env` не меняет пароль уже существующего пользователя.

Для запуска только базы выполните `docker compose up -d db`.
Проверить подключение можно командой:

```bash
docker compose exec db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT 1"'
```

Данные сохраняются в томе `postgres_data`. Команда `docker compose down`
останавливает контейнеры с сохранением данных; добавление `-v` удаляет том.
API запускается после успешной проверки готовности PostgreSQL.
Скрипт `bin/run.sh` применяет SQL-миграции из `migrations` перед запуском API;
если SQL-файлов нет, этот шаг пропускается. После добавления миграций пересоберите образ.
Go API создаёт пул `*sql.DB` через `sql.Open("pgx", DATABASE_URL)`
(драйвер `pgx/v5/stdlib`) и проверяет подключение SQL-запросом при запуске
с тайм-аутом 5 секунд. Без доступной базы API завершает запуск с ошибкой.
При запуске Go на компьютере используйте в URL `localhost:5432` вместо `db:5432`;
`go run` не загружает `.env` автоматически.

`migrations/001_create_links.sql` создаёт таблицу `links` с уникальным `short_name`
и автоматически заполняемым `created_at`. `short_url` в базе не хранится.
Настройки sqlc находятся в `sqlc.yaml`; генерация кода выполняется командой
`make generate` (sqlc v1.31.1). Исходные SQL-запросы находятся в
`internal/platform/postgres/queries`, сгенерированный код — в
`internal/platform/postgres/sqlc`. Пока добавлен только запрос проверки соединения;
CRUD ссылок ещё не реализован.

## Структура

- `cmd/api` — точка входа, сборка роутера и тест `/ping`.
- `internal/config` — настройки из переменных окружения.
- `internal/links` — каркас обработчиков, сервиса, репозитория, моделей и DTO пользователей.
- `internal/order` — каркас обработчиков, сервиса, репозитория и моделей заказов.
- `internal/middleware` — логирование запросов и место для авторизации.
- `internal/platform/postgres` — подключение PostgreSQL и код sqlc.
- `migrations` — миграции goose, включая создание таблицы ссылок.

Пакеты пользователей, заказов и авторизация пока представлены заготовками.
Рабочий маршрут API — `GET /ping`.

## Проверки качества

Требуется Go версии из `go.mod` (1.26.9) и golangci-lint v2.14.0.
Установите линтер по [официальной инструкции](https://golangci-lint.run/docs/welcome/install/local/), выбрав эту версию
и архитектуру своей системы.

```bash
go test -race ./...
golangci-lint run ./...
go build -o bin/api ./cmd/api
```

Тесты HTTP-валидации находятся в `internal/links`, тесты роутера — в
`internal/api`, интеграционные сценарии с PostgreSQL — в `tests/integration`.
Для полного прогона задайте `TEST_DATABASE_URL`; в CI тестовая БД запускается
автоматически. Команды и устройство тестов описаны в [tests/README.md](tests/README.md).
Правила линтера заданы в `.golangci.yml` и одинаковы локально и в CI.

Workflow `CI` выполняет тесты с детектором гонок, линтер и сборку при каждом
push и pull request. Версия Go берётся из `go.mod`.
Кэш модулей (`GOMODCACHE`) и сборок (`GOCACHE`) управляется встроенным
[кэшированием actions/setup-go](https://github.com/actions/setup-go#caching-dependency-files-and-build-outputs)
с ключом, учитывающим `go.sum`.

---

<details>
<summary>Автоматические тесты Хекслета</summary>

Тесты запускаются на каждый коммит. За запуск отвечает файл `.github/workflows/hexlet-check.yml` — не удаляйте и не
переименовывайте ни его, ни репозиторий.

</details>

## О Хекслете

[Хекслет](https://ru.hexlet.io/) — школа программирования: авторские программы обучения с практикой, поддержкой
наставников и реальными проектами, которые остаются в резюме. Этот репозиторий — один из таких проектов.

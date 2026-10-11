# Сокращатель ссылок (Go)

[![hexlet-check](https://github.com/Nikita-hub000/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/Nikita-hub000/go-project-278/actions)
[![CI](https://github.com/Nikita-hub000/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/Nikita-hub000/go-project-278/actions/workflows/ci.yml)

Спроектируйте приложение для удобных ссылок

Учебный проект Хекслета: https://ru.hexlet.io/programs/go


## Стек

- Go

## Установка

<!-- Опишите установку: клонирование, зависимости, переменные окружения -->

```bash
git clone https://github.com/Nikita-hub000/go-project-278.git
cd go-project-278
```

## Использование

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

Запуск в Docker:

```bash
cp .env.example .env
docker build -t links-api .
docker run --rm --env-file .env -p 8080:8080 links-api
```

## Структура

- `cmd/api` — точка входа, сборка роутера и тест `/ping`.
- `internal/config` — настройки из переменных окружения.
- `internal/user` — каркас обработчиков, сервиса, репозитория, моделей и DTO пользователей.
- `internal/order` — каркас обработчиков, сервиса, репозитория и моделей заказов.
- `internal/middleware` — логирование запросов и место для авторизации.
- `internal/platform/postgres` — место для подключения PostgreSQL.
- `migrations` — заготовки миграций пользователей и заказов.

Пакеты пользователей, заказов, PostgreSQL, авторизация и SQL-миграции пока
представлены заготовками: схема данных и бизнес-правила ещё не определены.
Рабочий маршрут API — `GET /ping`.

## Проверки качества

Требуется Go версии из `go.mod` (1.27.2) и golangci-lint v2.14.0.
Установите линтер по [официальной инструкции](https://golangci-lint.run/docs/welcome/install/local/), выбрав эту версию и архитектуру своей системы.

```bash
go test -race ./...
golangci-lint run ./...
go build -o bin/api ./cmd/api
```

Тест `/ping` использует `testing` и `net/http/httptest` и проверяет статус 200,
тип содержимого и JSON-ответ без запуска HTTP-сервера.
Правила линтера заданы в `.golangci.yml` и одинаковы локально и в CI.

Workflow `CI` выполняет тесты с детектором гонок, линтер и сборку при каждом
push и pull request. Версия Go берётся из `go.mod`.
Кэш модулей (`GOMODCACHE`) и сборок (`GOCACHE`) управляется встроенным
[кэшированием actions/setup-go](https://github.com/actions/setup-go#caching-dependency-files-and-build-outputs)
с ключом, учитывающим `go.sum`.

---

<details>
<summary>Автоматические тесты Хекслета</summary>

Тесты запускаются на каждый коммит. За запуск отвечает файл `.github/workflows/hexlet-check.yml` — не удаляйте и не переименовывайте ни его, ни репозиторий.

</details>

## О Хекслете

[Хекслет](https://ru.hexlet.io/) — школа программирования: авторские программы обучения с практикой, поддержкой наставников и реальными проектами, которые остаются в резюме. Этот репозиторий — один из таких проектов.

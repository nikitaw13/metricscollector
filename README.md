# metricscollector

Сервер сбора метрик и агент для трека «Сервер сбора метрик и алертинга» (Яндекс.Практикум, go-musthave-metrics).
Сервер принимает метрики типа *gauge* (абсолютное значение) и *counter* (накопительный счётчик) по текстовому и JSON API, хранит их в памяти или PostgreSQL, периодически сбрасывает на диск и уведомляет наблюдателей аудита о фактах обновления.

## Архитектура

```
cmd/
  server/          HTTP-сервер метрик
  agent/           агент, собирающий runtime-метрики и отправляющий их на сервер
internal/
  handler/         HTTP-хендлеры, chi-роутер, gzip/HMAC middleware
  service/         бизнес-логика над хранилищем метрик, пинг БД
  repository/      in-memory хранилище, персистентный файловый вариант, PostgreSQL
  model/           доменные модели метрик и ошибки валидации
  audit/           аудит обновлений метрик: fileWriter и urlSender (observer pattern)
  retry/           HTTP-клиент с повторами транспортных ошибок
  sign/            HMAC-SHA256 подпись тел запросов
  logger/          обёртка над zap
```

## API сервера

| Метод | Путь | Назначение |
|---|---|---|
| POST | `/update/{type}/{name}/{value}` | обновление метрики текстом |
| GET  | `/value/{type}/{name}` | текущее значение метрики текстом |
| POST | `/update/` | обновление одной метрики JSON |
| POST | `/updates/` | пакетное обновление метрик JSON-массивом |
| POST | `/value/` | чтение значения метрики JSON |
| GET  | `/` | HTML-страница со всеми метриками |
| GET  | `/ping` | проверка доступности PostgreSQL |

Тело JSON-запросов и ответов — `{"id","type","delta","value"}` (`delta` — для counter, `value` — для gauge).

## Запуск сервера

```shell
go run ./cmd/server -a :8080 -d "postgres://user:pass@localhost:5432/metrics"
```

| Флаг | Env | По умолчанию | Описание |
|---|---|---|---|
| `-a` | `ADDRESS` | `localhost:8080` | адрес HTTP-сервера |
| `-l` | `LOG_LEVEL` | `INFO` | уровень логирования |
| `-i` | `STORE_INTERVAL` | `300` | интервал сохранения на диск, с; `0` — синхронные записи |
| `-f` | `FILE_STORAGE_PATH` | — | путь к файлу персистентности |
| `-r` | `RESTORE` | `true` | загрузить сохранённые метрики при старте |
| `-d` | `DATABASE_DSN` | — | DSN PostgreSQL; без него — файловое хранилище |
| `-m` | `MIGRATION_PATH` | `migrations` | каталог миграций БД |
| `-k` | `KEY` | — | ключ HMAC-SHA256 для подписи тел запросов |
| `-audit-file` | `AUDIT_FILE` | — | файл аудита обновлений метрик |
| `-audit-url` | `AUDIT_URL` | — | HTTP-эндпоинт аудита обновлений метрик |
| `-profiler-addr` | `PROFILER_ADDR` | `localhost:6060` | адрес pprof-сервера; пустая строка отключает профилирование |

Env-переменные имеют приоритет над значениями по умолчанию, флаги — над env.

Профилирование: при запущенном сервере `go tool pprof http://localhost:6060/debug/pprof/heap` (профили CPU, heap, goroutine и др. — под `/debug/pprof/`).

## Запуск агента

```shell
go run ./cmd/agent -a :8080 -p 2 -r 10
```

| Флаг | Env | По умолчанию | Описание |
|---|---|---|---|
| `-a` | `ADDRESS` | `localhost:8080` | адрес сервера метрик |
| `-p` | `POLL_INTERVAL` | `2` | интервал сбора метрик, с |
| `-r` | `REPORT_INTERVAL` | `10` | интервал отправки метрик, с |
| `-k` | `KEY` | — | ключ HMAC-SHA256 (должен совпадать с серверным) |
| `-l` | `RATE_LIMIT` | `20` | лимит параллельных исходящих запросов; `0` — без ограничений |

## Тесты и линтеры

```shell
go test ./... -count=1
go vet ./...
staticcheck ./...
```

Покрытие пакетов внутреннего кода — unit-тестами (testify, gomock), API эндпоинтов — запускабельными examples в `internal/handler/example_test.go`.

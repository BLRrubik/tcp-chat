# TCP Chat Server

Production-ready TCP чат-сервер на Go с историей сообщений, graceful shutdown и HTTP-мониторингом.

## Функционал

- **TCP-чат в реальном времени** — множество клиентов подключаются через `telnet`/`nc`, сообщения рассылаются всем онлайн-пользователям
- **Команды**:
  - `/help` — список команд
  - `/time` — текущее время сервера
  - `/users` — список активных пользователей
  - `/quit` — отключиться
- **История сообщений** — кольцевой буфер, новым клиентам при подключении показываются последние N сообщений (`-message-history-size`)
- **Лимит подключений** — при достижении `-max-connections` новые клиенты получают отказ ("Server full")
- **Уровни логирования** — `error` / `warn` / `info` (`-log-level`)
- **Изящное завершение (graceful shutdown)** — по `Ctrl+C`/`SIGTERM`: уведомление клиентов, остановка приёма новых сообщений, принудительное отключение всех клиентов, ожидание завершения горутин с таймаутом
- **HTTP-мониторинг** — отдельный HTTP-сервер с эндпоинтами:
  - `GET /health` — `{"status","active_connections","uptime_seconds"}`
  - `GET /stats` — `{"active_connections","total_messages_processed","uptime_seconds","error_count"}`

## Структура проекта

```
cmd/server/main.go          # точка входа — только вызов app.Run()
internal/
  app/                       # оркестрация: конфиг, hub, HTTP/TCP серверы, сигналы, shutdown
  config/                    # флаги командной строки
  domain/                    # структуры данных (ChatMessage, Client, ServerStats)
  hub/                       # логика Hub и управления клиентами
  server/                    # TCP (tcp.go) и HTTP (http.go) серверы
  storage/                   # хранилище истории сообщений (кольцевой буфер)
  types/                     # обобщённый CycleBuffer
```

## Запуск

```bash
go build ./cmd/server && ./server
# или
go run ./cmd/server
```

Справка по флагам:

```bash
./server --help
```

### Флаги

| Флаг | По умолчанию | Описание |
|---|---|---|
| `-port` | `:8080` | TCP-порт чата |
| `-monitoring-port` | `:9090` | HTTP-порт мониторинга |
| `-log-level` | `info` | Уровень логов: `error`, `warn`, `info` |
| `-max-connections` | `10` | Максимум одновременных подключений |
| `-message-history-size` | `100` | Размер истории сообщений |

## Использование

Подключение к чату:

```bash
telnet localhost 8080
# или
nc localhost 8080
```

Мониторинг:

```bash
curl http://localhost:9090/health
curl http://localhost:9090/stats
```

## Разработка

```bash
go build ./...
go vet ./...
go test ./...
```

# Спортивный трекер для футбола

Веб-приложение для анализа данных футболистов в реальном времени: сбор данных с
носимых GPS/IMU трекеров (ESP32-S3), обработка на бэкенде (Go) и визуализация на
веб-интерфейсе (React) — метрики, зоны интенсивности, тепловые карты поля,
сравнение игроков.

Реализация по ТЗ v1.0 (29.08.2026).

---

## Стек

| Слой       | Технология                        |
|------------|-----------------------------------|
| Трекер     | ESP32-S3 (GPS NEO-M9N, IMU ICM-20948, BMP388) |
| Бэкенд     | Go 1.25, chi, pgx/v5, goose, gorilla/websocket |
| Фронтенд   | React 18, Vite, React Router, Recharts |
| БД         | PostgreSQL 15                     |
| Кеш        | Redis 7 (зарезервирован, пока не используется) |
| Инфра      | Docker Compose                    |

---

## Быстрый старт (Docker)

```bash
make up            # docker compose up --build
```

- Фронтенд:  http://localhost:3000
- Бэкенд API: http://localhost:8080  (health: `/healthz`)
- Postgres:  `localhost:5432`  (`tracker_user` / `tracker_pass` / `sports_tracker`)

Миграции и демо-данные (3 игрока с device_id `ESP32-TRACKER-00X`) применяются
бэкендом автоматически при старте.

Отправить синтетическую тренировку и увидеть метрики:

```bash
make seed-samples                       # ESP32-TRACKER-001, 60 c бега с 2 спринтами
./scripts/send_demo_batch.sh ESP32-TRACKER-002
```

Затем открыть Dashboard, выбрать игрока — карточки и график зон заполнятся;
на странице профиля появится тепловая карта.

---

## Локальная разработка

```bash
make infra              # только Postgres + Redis в Docker
make backend-run        # go run ./cmd/server  (слушает :8080)
make frontend-dev       # vite dev server :3000, проксирует /api и /ws на :8080
```

Тесты бэкенда:

```bash
make backend-test
```

Конфигурация — через переменные окружения (см. `backend/.env.example`,
`frontend/.env.example`). Бэкенд подхватывает `backend/.env`, если файл есть.

---

## Структура

```
traker/
├── backend/
│   ├── cmd/server/           точка входа: HTTP + WS + фоновые задачи
│   ├── internal/
│   │   ├── config/           загрузка конфигурации из окружения
│   │   ├── domain/           доменные типы (Player, Session, Metrics, Heatmap, Ingest)
│   │   ├── store/            PostgreSQL (pgx pool), миграции (goose, embed)
│   │   ├── processing/       Kalman 2D (GPS+IMU), расчёт метрик, построение тепловой карты
│   │   ├── httpapi/          chi роутер, хендлеры, JWT / API-key, WebSocket
│   │   └── live/             in-process pub/sub для реал-тайм обновлений
│   └── migrations/           *.sql (goose), встроены через embed
├── frontend/
│   └── src/
│       ├── services/         api.js (REST), websocket.js (авто-reconnect), storage.js
│       ├── hooks/            useMetrics, useLiveData, useHeatmap
│       ├── components/       PlayerSelector, MetricsCard, IntensityChart, HeatmapField, ...
│       └── pages/            Dashboard, PlayerProfile, Compare, Admin
├── scripts/send_demo_batch.sh
└── docker-compose.yml
```

---

## API

Полное описание — [`docs/API.md`](docs/API.md). Кратко:

| Метод | Путь | Назначение |
|-------|------|-----------|
| POST | `/api/data/ingest` | приём батча сырых данных с трекера (заголовок `X-API-Key`) |
| POST | `/api/auth/login` | выдача JWT (демо: любой известный логин + пароль `password`) |
| GET  | `/api/players` | список игроков и статусы подключения |
| POST | `/api/players` | добавить игрока |
| DELETE | `/api/players/{id}` | удалить игрока |
| GET  | `/api/players/{id}/metrics?session_id=` | метрики за сессию (по умолчанию — последняя) |
| GET  | `/api/players/{id}/heatmap?session_id=` | сетка тепловой карты поля |
| GET  | `/api/players/{id}/sessions?limit=&offset=` | история сессий |
| GET  | `/api/compare?player1_id=&player2_id=&session_id=` | сравнение двух игроков |
| WS   | `/ws/live?session_id=` | поток `{type:"metrics", data:{...}}` |

Read-эндпоинты используют `optionalAuth` — в dev-режиме доступны без токена.
Для боевого режима замените `optionalAuth` на `requireAuth` в
`internal/httpapi/server.go`.

---

## Обработка данных

- **Kalman 2D** (`processing/kalman.go`) — constant-velocity фильтр по позиции;
  измерения GPS проецируются в локальную плоскую систему (equirectangular),
  process noise раздувается на ускорениях по модулю вектора IMU.
- **Метрики** (`processing/metrics.go`) — скорость (max/avg), дистанция (с
  отсечением GPS-«телепортов»), спринты (порог 20 км/ч, удержание ≥ 1 c),
  ускорения (dv/dt ≥ 3 м/с²), Player Load (интеграл модуля ускорения),
  зоны интенсивности (0–70 / 70–85 / 85–100 % от max), высота прыжка (импульс по Z).
- **Тепловая карта** (`processing/heatmap.go`) — трек центрируется по полю
  105×68 м, бьётся на ячейки 10×10 м, копится время присутствия и средняя
  скорость; уровень интенсивности ячейки — доля от max-скорости.

Параметры порогов и геометрии поля настраиваются через окружение
(`SPRINT_SPEED_KMH`, `FIELD_LENGTH_M`, `FIELD_WIDTH_M`, `HEATMAP_CELL_M`).

После каждого `ingest` метрики и тепловая карта сессии пересчитываются в фоне и
пушатся подписчикам WebSocket.

---

## Что не входит в v1 (заготовки)

- Redis-кеширование метрик (сервис поднят, слой кеша не подключён).
- Полноценная аутентификация: `handleLogin` выдаёт JWT, но не проверяет
  `users.password_hash` (bcrypt) — нужно дописать перед боевым использованием.
- Прошивка ESP32 — формат payload зафиксирован в `docs/API.md` и
  `scripts/send_demo_batch.sh`.
- Графики пульса и накопительного Player Load по времени (метрики агрегатные;
  для тайм-серий нужен эндпоинт над `raw_data`).

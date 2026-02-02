# Jacklike Quiz Game (Final Project)

Реализована онлайн квиз-игра с комнатами и несколькими игроками.  
Игроки подключаются к комнате по WebSocket, отвечают на вопросы, получают очки, видят результаты раундов и итоговый leaderboard. В следующем семестре, когда начнется курс по мобильной разработке, пользователи будут подключаться к комнате со своих телефонов, как это реализовано в оригинальной игре JackBox.

Работа реализована в полном объёме, API описано в файле **`openapi.yml`**, лежащем в корне проекта.

---

## Архитектура

Архитектура проекта реализована по слоистому принципу (Clean / Layered Architecture) с разделением ответственности:

- **Application layer (internal/app)** — композиция и запуск приложения: загрузка конфигурации, инициализация logger/db/repository/services/ws/http, сборка роутера и старт HTTP-сервера
- **Handler layer (`internal/handler`)** — HTTP/WS endpoints (transport)
- **Service layer (`internal/service`)** — бизнес‑логика (use-cases)
- **Repository layer (`internal/storage`)** — доступ к данным (PostgreSQL)
- **Domain layer (`internal/game`)** — доменная модель игры (Room, Player, Round)

Ключевые принципы:
- структура **handler → service → repository**
- **Dependency Injection через интерфейсы**
- бизнес‑логика не зависит от транспорта (HTTP/WS) и конкретной БД напрямую

---

## Структура проекта

```
cmd/server/                 # точка входа приложения
internal/
  handler/                  # HTTP handlers + admin endpoints
  service/                  # бизнес-логика (GameService / AdminService)
  storage/                  # репозитории (PostgresQuestionStore + интерфейсы)
  game/                     # доменная модель игры
  ws/                       # WebSocket hub/client + протокол
migrations/                 # SQL миграции
openapi.yml                 # документация API (OpenAPI 3.0)
docker-compose.yml          # окружение (Postgres + app + migrations)
Dockerfile                  # сборка приложения
```

---


### Запуск проекта
```bash
docker compose up --build
```

После запуска сервис доступен по адресу:
- **HTTP API:** `http://localhost:8080`
- **WebSocket:** `ws://localhost:8080/ws/{ROOM_CODE}`

---

## Переменные окружения

Проект поддерживает `.env` файл (лежит рядом с `docker-compose.yml`).

Пример `.env`:
```env
ADMIN_TOKEN=super-secret-token-123
```

`ADMIN_TOKEN` — токен для доступа к админским эндпоинтам `/admin/*`.

---

## API-интерфейсы

### Комнаты (Rooms)

| Метод | URL | Описание |
|------|-----|----------|
| POST | `/rooms` | Создать комнату |
| GET  | `/rooms/{code}` | Получить базовую информацию о комнате |

Пример: создать комнату
```bash
curl -X POST http://localhost:8080/rooms
```

Ответ:
```json
{"code":"ABCD"}
```

---

### WebSocket API

| Тип | URL | Описание |
|-----|-----|----------|
| WS  | `/ws/{code}` | Подключение к комнате по WebSocket |

Пример подключения:
```
ws://localhost:8080/ws/ABCD
```

Первое сообщение клиента **обязательно**:
```json
{
  "type": "join_room",
  "payload": { "name": "Player1" }
}
```

Сообщения клиента:
- `start_game`
```json
{
  "type": "start_game",
  "payload": {}
}
```
- `submit_answer`
```json
{
  "type": "submit_answer",
  "payload": { "optionId": "A" }
}
```

События сервера (примерно):
- `player_joined`
- `room_state`
- `answer_accepted`
- `round_results`
- `game_over`
- `error`

---

## Админ API (вопросы)

Админские эндпоинты защищены Bearer‑токеном:

**Header:**
```
Authorization: Bearer <ADMIN_TOKEN>
```

### Endpoints

| Метод | URL | Описание |
|------|-----|----------|
| POST  | `/admin/questions` | Создать вопрос |
| GET   | `/admin/questions` | Список активных вопросов |
| GET   | `/admin/questions?all=1` | Список всех вопросов (включая неактивные) |
| PATCH | `/admin/questions/{id}` | Активировать/деактивировать вопрос |

---

### Создать вопрос

`POST /admin/questions`

Body:
```json
{
  "text": "Сколько дней в году (обычный год)?",
  "options": [
    {"id":"A","text":"364"},
    {"id":"B","text":"365"},
    {"id":"C","text":"366"},
    {"id":"D","text":"360"}
  ],
  "correctId": "B",
  "isActive": true
}
```

Пример (curl):
```bash
curl -X POST "http://localhost:8080/admin/questions" \
  -H "Authorization: Bearer super-secret-token-123" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Сколько дней в году (обычный год)?",
    "options": [
      {"id":"A","text":"364"},
      {"id":"B","text":"365"},
      {"id":"C","text":"366"},
      {"id":"D","text":"360"}
    ],
    "correctId": "B",
    "isActive": true
  }'
```

---

### Получить список вопросов

`GET /admin/questions?all=1`

Пример (curl):
```bash
curl -X GET "http://localhost:8080/admin/questions?all=1" \
  -H "Authorization: Bearer super-secret-token-123"
```

---

### Деактивировать вопрос

`PATCH /admin/questions/{id}`

Body:
```json
{ "isActive": false }
```

Пример (curl):
```bash
curl -X PATCH "http://localhost:8080/admin/questions/1" \
  -H "Authorization: Bearer super-secret-token-123" \
  -H "Content-Type: application/json" \
  -d '{ "isActive": false }'
```

---
| Story ID | Краткое описание                                            | Эндпоинты                                                           | Критерии приёмки (GWT)                                                                                                         | Бизнес-правила                                                                                            | Юнит-тесты (примеры названий)                                        | Негативные кейсы                                                                                      |
| -------- | ----------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| ST-1     | Пользователь создаёт комнату                                | **POST /rooms**                                                     | Given сервис запущен — When POST /rooms — Then 200 и возвращается code                                                         | code генерируется, комната создаётся в памяти                                                             | TestRooms_Create_Success                                             | 405 если метод не POST                                                                                |
| ST-2     | Пользователь получает данные комнаты                        | **GET /rooms/{code}**                                               | Given комната существует — When GET /rooms/{code} — Then 200 и возвращается code+phase                                         | код комнаты регистронезависим                                                                             | TestRooms_Get_Success                                                | 404 если комнаты нет; 405 если метод не GET                                                           |
| ST-3     | Клиент подключается по WebSocket и присоединяется к комнате | **GET /ws/{code}** (WS upgrade)                                     | Given комната существует — When WS connect + отправлен join_room с валидным name — Then клиент добавлен, приходит room_state   | name не пустой; room должен существовать                                                                  | TestWS_JoinRoom_Success                                              | 404 room not found; error если первое сообщение не join_room; error если name пустой                  |
| ST-4     | Хост стартует раунд                                         | **WS message: start_game**                                          | Given подключен хост — When отправляет start_game — Then phase=answering, question+options отправлены через room_state         | стартовать может только host; нельзя стартовать если RoundNumber >= MaxRounds; вопрос должен быть валиден | TestService_StartRound_Success; TestRoom_StartGame_NotHost           | ошибка если не хост; ошибка если нет вопросов в БД; game_over если превышен MaxRounds                 |
| ST-5     | Игрок отправляет ответ                                      | **WS message: submit_answer**                                       | Given phase=answering и дедлайн не прошёл — When submit_answer(optionId) — Then ответ принят (answer_accepted)                 | нельзя отвечать дважды; optionId должен существовать; нельзя после дедлайна                               | TestRoom_SubmitAnswer_Success; TestRoom_SubmitAnswer_AlreadyAnswered | error bad payload; error invalid option; error already answered; error deadline passed                |
| ST-6     | Сервер завершает раунд по дедлайну и начисляет очки         | (внутренний scheduler) + **WS события: round_results + room_state** | Given answering и дедлайн прошёл — When таймер срабатывает — Then phase=results, results отправлены, score обновлён            | очко начисляется только за правильный ответ; результаты формируются по всем игрокам                       | TestRoom_FinishRound_DeadlinePassed                                  | не должно завершать раунд до дедлайна; не должно завершать если phase != answering                    |
| ST-7     | Сервер запускает следующий раунд или завершает игру         | (внутренний scheduler) + **WS: room_state / game_over**             | Given phase=results — When проходит ResultsPause — Then либо стартует следующий раунд, либо game_over если достигнут MaxRounds | MaxRounds ограничивает число раундов; leaderboard строится по score (tie-break по имени)                  | TestService_BuildLeaderboard_Ties; TestScheduler_NextRound           | game_over если достигнут лимит; error если нет вопросов в БД                                          |
| ST-8     | Админ создаёт вопрос через HTTP                             | **POST /admin/questions**                                           | Given Authorization Bearer валиден и payload корректный — When POST — Then 200 и возвращается созданный вопрос (id, createdAt) | требуется админ-токен; вопрос должен иметь 4 options и correctId должен быть среди options; text уникален | TestAdmin_CreateQuestion_Success; TestRepo_CreateQuestion            | 401 без токена/неверный; 400 bad json; 400 invalid payload; 500/400 при нарушении unique(text)        |
| ST-9     | Админ получает список вопросов                              | **GET /admin/questions**; **GET /admin/questions?all=1**            | Given админ токен валиден — When GET — Then 200 и массив вопросов (по умолчанию active, all=1 — все)                           | требуется админ-токен; фильтрация по is_active                                                            | TestAdmin_ListQuestions_ActiveOnly; TestAdmin_ListQuestions_All      | 401 без токена/неверный; 500 при ошибке БД                                                            |
| ST-10    | Админ активирует/деактивирует вопрос                        | **PATCH /admin/questions/{id}**                                     | Given админ токен валиден и id существует — When PATCH isActive — Then 200 и вопрос обновлён                                   | требуется админ-токен; id должен быть > 0                                                                 | TestAdmin_SetQuestionActive_Success; TestRepo_SetQuestionActive      | 401 без токена; 400 bad id; 400 bad json; 404/500 если id не найден (зависит от обработки pgx ошибок) |

## Тестирование

В проекте реализованы unit-тесты на бизнес-логику и обработку ошибок с использованием `testify`.

Покрыты тестами:
- доменная логика (`internal/game`)
- сервисный слой (`internal/service`)
- обработка HTTP-ошибок в хэндлерах (`internal/handler`)

Запуск всех тестов:

```bash
go test 
```

##  Сервис использует PostgreSQL. Таблица вопросов создаётся миграциями.

Схема таблицы:
```sql
CREATE TABLE IF NOT EXISTS questions (
  id          bigserial PRIMARY KEY,
  text        text NOT NULL,
  options     jsonb NOT NULL,
  correct_id  text NOT NULL,
  is_active   boolean NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE questions
  ADD CONSTRAINT questions_text_unique UNIQUE (text);

CREATE INDEX IF NOT EXISTS questions_active_idx ON questions (is_active);
```

---

## Логирование

В проекте реализовано **structured logging** на базе **zap**.

Особенности:
- логи пишутся **файл, лежащий в контейнере**
- поддерживаются уровни логирования: **info / warn / error**

Настройка выполняется через переменные окружения:

```env
LOG_LEVEL=info
LOG_FILE=/tmp/app.log
```

Пример просмотра логов внутри docker-контейнера:

```
docker exec -it jacklike_app sh -lc "tail -n 50 /tmp/app.log"
```

---


## Примечания по решению

- Вопросы хранятся в PostgreSQL и выбираются случайным образом среди активных.
- Игра заканчивается после `MaxRounds` раундов (по умолчанию 5), после чего сервер отправляет `game_over` и leaderboard.
- Host logic (доменное правило) — первый подключившийся игрок становится хостом комнаты. Только хост может запускать раунд/игру (start_game). Если хост отключается, роль хоста автоматически передаётся другому подключённому игроку.
- WebSocket соединение использует ping/pong для поддержания подключения.
- Handshake правило подключения — после подключения к WebSocket клиент обязан в течение 30 секунд отправить сообщение join_room с именем игрока, иначе соединение будет закрыто сервером.
---

## Пример работы приложение
- Поднимаем приложение docker compose up --build
- Post запросом получаем код комнаты http://localhost:8080/rooms
- Подключаемся по WebSocket к комнате, созданной на предыдущем шаге: ws://localhost:8080/ws/ZB7N
- В течение 30 секунд присылай свой игровой никнейм, иначе будем отключены:
{
  "type": "join_room",
  "payload": {
    "name": "Artem"
  }
}
- Ждём всех оставшихся игроков и запускаем игру
{
  "type": "start_game",
  "payload": {}
}
- Получаем вопрос и отправляем свои ответы 
{
  "type": "submit_answer",
  "payload": {
    "optionId": "B"
  }
}
- После 5 раундов получаем игровую статистику

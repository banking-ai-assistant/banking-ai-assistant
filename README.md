# Banking AI Assistant

Интеллектуальная платформа автоматизации обработки входящей официальной корреспонденции для банка.
Система анализирует входящие письма, классифицирует их по типу и срочности, генерирует варианты ответов (несколько стилей), управляет процессом согласования и отправкой ответов. Взаимодействие через HTTP API. Проект состоит из основного backend на Go и отдельного микросервиса AI на Python, который обращается к LLM (YandexGPT / OpenAI-compatible client).

---

## Описание

Сервис предназначен для ускорения и стандартизации обработки писем:

* Автоматическая классификация писем (жалоба, регуляторный запрос, запрос документов, партнёрство и др.).
* Извлечение ключевых полей и определение срочности / SLA.
* Генерация адаптивных вариантов ответа в разных стилях (официальный, деловой, клиентоориентированный и т.д.).
* Хранение входящих писем и созданных черновиков в PostgreSQL.
* Маршрутизация / нотификации (stub-реализованные методы отправки почты в текущем репозитории).
* Веб-API для CRUD операций и запуска генерации ответов.

---

## Основные сущности

* **Letter**
  `id`, `original_id`, `subject`, `content`, `sender_email`, `sender_name`, `received_at`, `replied_at`, `status` (new|replied...), `ai_analysis` (JSON), `category`, `created_at`, `updated_at`

* **GeneratedReply**
  `id`, `letter_id`, `content`, `style`, `status` (draft|sent...), `created_at`

* **AIAnalysis** (вложенный в Letter)
  `type` (complaint|regulatory|...), `urgency` (high|medium|low), `tone`, `summary`, `keywords`

---

## Работа модели

Модель проекта представляет собой готовое ядро RAG-системы (Retrieval-Augmented Generation), разработанное специально для автоматической генерации профессиональных ответов на входящие обращения в банковской среде. Система объединяет структурированные корпоративные шаблоны, классификацию типов писем и современные языковые модели для создания ответов, которые соответствуют как содержательным, так и стилистическим требованиям регулируемой индустрии.

Основу логики составляет класс ```RAGEngine```, который принимает на вход текст клиентского письма и краткое намерение сотрудника (например, «извиниться и предложить компенсацию»). На основе этого движок сначала определяет тип обращения с помощью внешнего классификатора, затем загружает подходящие шаблоны из локальной директории ```rag/templates/``` и генерирует два варианта ответа: первый — улучшенная, профессионально переформулированная версия черновика сотрудника, второй — строго структурированный ответ, построенный по официальному шаблону (например, для регуляторных запросов или срочных жалоб). Если ни один шаблон не подходит или возникает ошибка, система автоматически подключает нейтральные резервные варианты, гарантируя, что ответ всегда будет сгенерирован.

Все шаблоны хранятся в виде JSON-файлов, сгруппированных по типам писем: жалобы (complaint), партнёрские предложения (partnership), запросы на согласование (approval_request) и регуляторные обращения (regulatory). Каждый шаблон содержит чёткую структуру (например, «Признание проблемы → Меры реагирования → Сроки → Компенсация → Контакты»), пример, тон и стиль — что позволяет не только генерировать, но и оценивать качество ответов.

Для этого в состав серверной части входит модуль оценки, который проверяет, насколько сгенерированный текст соответствует ожидаемой структуре, анализируя наличие ключевых фраз из примеров. Дополнительно используется языковая модель как «судья» — для оценки релевантности, профессионализма и полноты ответа по шкале от 1 до 5.

---

## API

Сервис использует JSON по HTTP. Все ответы имеют `Content-Type: application/json`.

| Метод | URL                            | Описание                                                                           |
| ----- | ------------------------------ | ---------------------------------------------------------------------------------- |
| GET   | `/api/letters`                 | Список писем. Поддерживаются query: `category`, `status`                           |
| POST  | `/api/letters`                 | Создать письмо. Тело: `CreateLetterRequest`                                        |
| GET   | `/api/letters/{id}`            | Получить письмо + сгенерированные ответы                                           |
| POST  | `/api/letters/{id}/generate`   | Сгенерировать варианты ответов (тело: `{ "action": "<черновик/намерение>" }`)      |
| POST  | `/api/letters/{id}/send-reply` | Отправить ответ (тело: `{ "reply_id": "<id>" }` или `{ "custom_content": "..." }`) |

### Примеры

Создание письма:

```bash
curl -X POST http://localhost:8081/api/letters \
  -H "Content-Type: application/json" \
  -d '{
    "original_id": "ext-123",
    "subject": "Запрос документов",
    "content": "Прошу выслать выписку по счету за 2025-11",
    "sender_email": "client@example.com",
    "sender_name": "ООО Ромашка"
  }'
```

Генерация ответов:

```bash
curl -X POST http://localhost:8081/api/letters/<letter_id>/generate \
  -H "Content-Type: application/json" \
  -d '{"action": "Предложить стандартную форму ответа и сроки"}'
```

Отправка выбранного варианта:

```bash
curl -X POST http://localhost:8081/api/letters/<letter_id>/send-reply \
  -H "Content-Type: application/json" \
  -d '{"reply_id":"<generated_reply_id>"}'
```

---

## Миграции и схема БД

Включён пример миграции `migrations/001_init.sql`:

```sql
DROP TABLE IF EXISTS letters;
DROP TABLE IF EXISTS generated_replies;

CREATE TABLE letters (
    id TEXT PRIMARY KEY,
    original_id TEXT,
    subject TEXT,
    content TEXT,
    sender_email TEXT,
    sender_name TEXT,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    replied_at TIMESTAMP WITH TIME ZONE NULL,
    status TEXT DEFAULT 'new',
    ai_analysis JSONB NULL,
    category TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE generated_replies (
    id TEXT PRIMARY KEY,
    letter_id TEXT REFERENCES letters(id) ON DELETE CASCADE,
    content TEXT,
    style TEXT,
    status TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_letters_category ON letters(category);
CREATE INDEX idx_letters_status ON letters(status);
```

---

## Переменные окружения

Файл `.env` должен содержать (пример в `.env.example`):

```
APP_ENV=local
PORT=8081

POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

DATABASE_DSN=postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable

AI_URL=http://ai-service:8000
```

* `DATABASE_DSN` — DSN для подключения Go-сервиса к Postgres.
* `AI_URL` — URL ai-service (по умолчанию `http://ai-service:8000` в docker-compose).
* `PORT` — порт, на котором стартует Go-сервис (в Dockerfile выставлен 8081 по умолчанию).

---

## Запуск через Docker Compose

В корне репозитория есть `docker-compose.yml`. Для локального запуска:

```bash
cp .env.example .env
docker-compose up --build
```

Последовательность контейнеров:

* `postgres` — БД;
* `migrate` — прогон миграций (пакет `migrate/migrate`);
* `ai-service` — Python microservice (LLM client);
* `go-service` — основной backend.

**Замечание:** `migrate` в compose использует простую задержку `sleep 10`. В проде лучше сделать проверку готовности БД и обеспечить повторные попытки миграции.


---

## Логирование и отладка

* Go-сервис использует `internal/logger` — выводит простые лог-строки в stdout.
* Python ai-service при ошибках печатает диагностические сообщения. При проблемах с классификацией/парсингом JSON — смотреть stdout/логи ai-service.


# AGENTS.md — EOS DMS «Paperless»

> Enterprise-система электронного документооборота (СЭД).
> Учебный проект уровня production-grade. Цель — прокачать Go, TypeScript/Next.js,
> SQL, DevOps и тестирование до уверенного middle++/senior по практикам крупных компаний.

---

## 1. Правила работы для ИИ-агентов

1. **Режим наставника**: любые решения объяснять (почему, альтернативы, trade-offs). Не просто писать код — учить.
2. **Стандарты нерушимы**: границы модулей и слоёв (п. 4) не нарушать ни при каких «так быстрее».
3. **Архитектурные решения** фиксируются в ADR (`docs/adr/NNNN-title.md`) до или во время реализации.
4. **Новые фичи** реализуются по этапам из `docs/ROADMAP.md`. Прыжки вперёд без согласования запрещены.
5. Языки: код, идентификаторы, коммиты, API, БД — **английский**; документация репозитория и общение — русский.
6. Никаких секретов в репозитории. Конфигурация только через env-переменные / `.env` (в `.gitignore`).
7. Каждый PR/коммит проходит: `lint`, `typecheck`, `unit + integration tests` — иначе не считается готовым.
8. Режим обучения: код пишет ученик (агент показывает его в чате с объяснениями); после каждой темы — чекпоинт «всё ли понятно, двигаемся дальше?» без явного подтверждения новые темы не открывать. Контрольные вопросы от наставника задаются только после того, как ученик разобрал материал и задал свои вопросы.

## 2. Технологический стек

### Backend (Go)
| Назначение | Выбор | Почему |
|---|---|---|
| Язык | Go 1.26+ | стандарт enterprise-бэкенда |
| HTTP-роутер | `chi/v5` | идиоматичный, тонкий, совместим с `net/http` |
| Драйвер PG | `pgx/v5` | самый быстрый и продвинутый драйвер PostgreSQL |
| Доступ к данным | `sqlc` | пишем реальный SQL → генерируется типобезопасный код (без ORM) |
| Миграции | `goose` | forward-only SQL-миграции |
| Логи | `slog` (stdlib) | структурированное логирование, zero-dep |
| Валидация | `go-playground/validator` | декларативные теги + кастомные правила |
| API-контракт | OpenAPI 3.1 contract-first + `oapi-codegen` | контракт раньше кода |
| Авторизация | Keycloak (OIDC/OAuth2), JWT RS256, `go-oidc` | корпоративный IdM |
| События | Kafka + Outbox pattern | надёжная доставка событий между модулями/системами |
| Кэш / rate-limit | Redis 7 (`go-redis`) | кэш, лимиты, идемпотентность |
| Файлы | MinIO (S3 API, `aws-sdk-go-v2`) | хранение документов, presigned URLs |
| Трассировка | OpenTelemetry | распределённая трассировка |
| Метрики | Prometheus | стандарт мониторинга |

### Frontend (TypeScript / Next.js)
| Назначение | Выбор |
|---|---|
| Фреймворк | Next.js 15+ (App Router) |
| Рантайм/пакет-менеджер | Bun |
| Линтер + форматтер | Biome (единственный инструмент вместо ESLint+Prettier) |
| Серверное состояние | RTK Query |
| Клиентское состояние | Redux Toolkit (+ Zustand точечно для лёгкого UI-state) |
| Валидация схем | Zod (единый источник контрактов данных) |
| Формы | react-hook-form + zod-resolver |
| UI | Tailwind CSS 4 + shadcn/ui |
| Таблицы | TanStack Table |
| i18n | next-intl |
| E2E | Playwright |

### Данные
- **PostgreSQL 17** — единственный источник правды (OLTP).
- **ClickHouse** — только аналитика/отчёты/статистика (OLAP). Запись через события (Kafka consumer). Чтение из OLTP напрямую в CH запрещено.
- **Redis** — эфемерные данные (кэш, сессии, лимиты). Потеря Redis ≠ потеря данных.

### DevOps / Quality
Docker + Compose (локально) · GitHub Actions (CI/CD) · golangci-lint · Biome · Trivy + gosec (security scan) · k6 (нагрузочное) · testcontainers-go (интеграционное) · Playwright (E2E) · Prometheus + Grafana + Loki + Tempo (observability) · kind/k3d + Helm + ArgoCD (этап Kubernetes).

## 3. Структура монорепозитория

```
eos-dms-paperless/
├── apps/
│   ├── api/                  # Go backend (модульный монолит)
│   │   ├── cmd/api/          # точка входа
│   │   ├── internal/
│   │   │   ├── domain/       # домен: сущности, VO, события, ошибки (чистый Go, ноль зависимостей)
│   │   │   ├── app/          # use cases + порты (интерфейсы наружу от домена)
│   │   │   ├── adapters/     # driving: http; driven: postgres, kafka, redis, minio
│   │   │   └── config/
│   │   └── migrations/       # goose SQL-миграции
│   └── web/                  # Next.js frontend (FSD)
│       └── src/
│           ├── app/          # слой app: провайдеры, роутинг, стили
│           ├── pages/        # композиция страниц
│           ├── widgets/      # крупные блоки UI
│           ├── features/     # фичи (действия пользователя)
│           ├── entities/     # бизнес-сущности фронта
│           └── shared/       # ui-kit, api-client, lib, config
├── deploy/                   # docker-compose, k8s/helm, grafana-дашборды
├── docs/
│   ├── ROADMAP.md            # план обучения и реализации
│   ├── adr/                  # Architecture Decision Records
│   └── openapi/              # OpenAPI контракты (source of truth для API)
├── .github/workflows/        # CI pipelines
├── Makefile                  # единая точка команд
└── AGENTS.md                 # этот файл
```

## 4. Архитектурные правила

1. **Модульный монолит** (см. `docs/adr/0001-modular-monolith.md`). Доменные модули:
   `identity`, `documents`, `workflow`, `tasks`, `catalog`, `notifications`, `audit`, `analytics`.
   Модули общаются **только** через публичные интерфейсы application-слоя и доменные события. Прямые импорты внутренностей чужого модуля запрещены.
2. **Гексагональная архитектура** внутри модуля:
   - `domain` — чистая логика, не знает про БД/HTTP/Kafka;
   - `app` — use cases, определяет порты (интерфейсы) к внешнему миру;
   - `adapters/driven` — реализации портов (postgres, minio, kafka...);
   - `adapters/driving` — входные точки (http handlers), которые вызывают use cases.
3. **Зависимости направлены внутрь**: `adapters → app → domain`. Обратных импортов нет.
4. **Интерфейсы объявляются у потребителя**, а не у реализации (Go idiom).
5. **Ошибки**: sentinel + типы в domain, оборачивание `%w`, наверху маппинг в HTTP-коды. Ошибки — часть контракта домена.
6. **CQRS-lite**: чтение может идти мимо домена (read model SQL-запросами), запись — строго через use cases.
7. **Outbox**: никакой прямой публикации событий в Kafka из транзакции. Только outbox-таблица + релей.
8. **Idempotency** всех обработчиков событий и мутирующих POST/PATCH операций.
9. Frontend — **FSD**: импорты только вниз по слоям (`app → pages → widgets → features → entities → shared`). Кросс-импорты внутри слоя — через `public API` (index.ts).

## 5. Стандарты кода

### Go
- `gofmt`/`golangci-lint` — обязательны (конфиг в корне `apps/api`).
- `context.Context` — первый параметр всех I/O функций.
- Именование пакетов: коротко, строчными буквами, без `utils/helpers`.
- Godoc-комментарии на экспортируемых символах; внутри — только там, где «почему», а не «что».
- Тесты: table-driven, stdlib `testing`; ассерты `testify` допустимы; моки — `mockery`.
- Никаких `panic` в библиотечном коде; `log.Fatal` только в `main`.

### TypeScript / React
- `strict: true`, `any` запрещён (Biome rule).
- Контракты данных — только Zod-схемы; типы выводятся (`z.infer`), ручные дубли запрещены.
- Компоненты: function declarations, props-тип `XProps`, без `React.FC`.
- Серверные компоненты по умолчанию; `'use client'` — осознанно.
- RTK Query — весь серверный стейт; никаких `fetch` в компонентах.

### SQL / БД
- snake_case, таблицы во множественном числе (`documents`, `document_versions`).
- PK: `uuidv7`; аудит-поля: `created_at timestamptz not null default now()`, `updated_at`.
- Миграции forward-only, обратной совместимости схемы (expand/contract) при деплое.
- Каждый запрос в sqlc проверяется `EXPLAIN ANALYZE` при оптимизации; индексы под реальные запросы.
- ClickHouse: таблицы семейства `MergeTree`, партиционирование по дате, порядок сортировки = паттерн запросов.

## 6. Тестирование (пирамида)

| Уровень | Инструмент | Где живёт |
|---|---|---|
| Unit (домен, use cases) | `testing` + testify + mockery | рядом с кодом, `*_test.go` |
| Integration (БД, Kafka, MinIO) | testcontainers-go | `adapters/**/*_test.go` |
| Contract/API | Schemathesis + Hurl против OpenAPI | CI stage |
| E2E | Playwright | `apps/web/e2e` |
| Нагрузочное | k6 сценарии | `deploy/k6` |
| Security | gosec, Trivy, OWASP ZAP (этап 10) | CI |

Гейты: покрытие домена ≥ 80%, общий ≥ 65%; линтеры и typecheck — 0 ошибок; критические CVE — блокируют merge.

## 7. Git-процесс

- **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:` ...).
- Trunk-based: короткоживущие ветки `feat/<ticket>-slug`, PR в `main`.
- Один логический шаг — один коммит; история читаемая.
- ADR обязателен для: выбора библиотеки, смены схемы данных, изменения границ модулей.

## 8. Definition of Done (фича считается готовой, когда)

1. Код соответствует п. 4–5, покрыт тестами нужного уровня пирамиды.
2. OpenAPI/Zod-контракты обновлены, миграции применяются с чистой БД.
3. `make lint && make test && make e2e` — зелёные локально и в CI.
4. Наблюдаемость: структурные логи + метрика/трейс на новых операциях.
5. Обновлены README/доки; если менялась архитектура — написан ADR.

## 9. Быстрые команды (целевое состояние)

```bash
make dev         # поднять всю инфраструктуру + api + web
make lint        # golangci-lint + biome
make test        # unit + integration
make e2e         # playwright
make migrate     # накатить миграции
make down        # остановить окружение
```

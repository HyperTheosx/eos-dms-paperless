# ROADMAP — EOS DMS «Paperless»

> План обучения и поэтапной реализации. Каждый этап = веха с проверяемым результатом.
> Не переходить к следующему этапу, пока не закрыт Definition of Done текущего.
> Оценка времени — при 10–15 ч/нед.

---

## Легенда статусов
`[ ]` не начат · `[~]` в работе · `[x]` готов

## Этап 0 — Фундамент `[ ]`
**Цель:** рабочий скелет проекта и toolchain. **Навыки:** git-flow, Docker, Makefile, CI basics.

- [ ] `.gitignore`, `Makefile` (dev/lint/test/migrate/down), `.env.example`
- [ ] `docker-compose.yml`: postgres, clickhouse, redis, keycloak, minio, kafka (KRaft), pgadmin
- [ ] Скелет `apps/api` (`go mod init`, chi + `/healthz` + graceful shutdown)
- [ ] Скелет `apps/web` (Next.js + Bun + Biome + TS strict)
- [ ] CI: GitHub Actions — lint + build обоих приложений
- [ ] ADR-0001 модульный монолит; шаблон ADR
- **DoD:** `make dev` поднимает всё; healthchecks зелёные; CI зелёный.

## Этап 1 — Go Core через предметную область `[ ]`
**Цель:** язык Go на практике TDD. **Навыки:** синтаксис, структуры/интерфейсы, generics, errors, пакеты, table-driven tests.

Теория-минимум: types, slices/maps внутренности, методы и value vs pointer receivers, интерфейсы (неявная реализация), композиция вместо наследования, error handling (`errors.Is/As`, `%w`), goroutine/channel базово, context.Context.
- [ ] Домен `documents`: сущность Document, VO (DocNumber, DocType), статусы, инварианты
- [ ] Домен `identity`: User, Role, Permission
- [ ] Unit-тесты домена (TDD: сначала тест → код → рефакторинг)
- [ ] Мини-разбор: как Go компилируется, GC, стек vs куча (escape analysis) — 1 сессия
- **DoD:** доменные пакеты без внешних зависимостей, покрытие ≥ 80%.

## Этап 2 — Гексагональная архитектура + PostgreSQL `[ ]`
**Цель:** порты и адаптеры на реальном хранилище. **Навыки:** SQL, pgx, sqlc, миграции, транзакции, testcontainers.

- [ ] Порты приложения: `DocumentRepository`, `UnitOfWork`
- [ ] Адаптер postgres: sqlc-запросы, пул pgx, транзакции
- [ ] Миграции goose: documents, document_versions, users, roles, audit-поля
- [ ] Integration-тесты через testcontainers-go
- [ ] Разбор EXPLAIN ANALYZE, индексы (btree, partial), изоляция транзакций
- **DoD:** use cases работают против реальной PG; интеграционные тесты зелёные в Docker.

## Этап 3 — HTTP API contract-first `[ ]`
**Цель:** production-grade REST. **Навыки:** OpenAPI, oapi-codegen, middleware, обработка ошибок, pagination/filtering/sorting, versioning.

- [ ] `docs/openapi/documents.yaml`: контракты CRUD + поиск
- [ ] Генерация серверных интерфейсов oapi-codegen, strict-режим
- [ ] Middleware: request_id, logging (slog), recover, timeout, CORS
- [ ] Единый формат ошибок RFC 7807 (problem+json); маппинг доменных ошибок
- [ ] Contract-тесты Schemathesis/Hurl в CI
- **DoD:** Swagger UI работает, контракт ↔ код совпадают автоматически.

## Этап 4 — Identity & Access: Keycloak `[ ]`
**Цель:** корпоративная аутентификация/авторизация. **Навыки:** OIDC/OAuth2, JWT, RBAC+ABAC, security.

- [ ] Realm `eos-dms`: clients (api, web), роли (admin, clerk, approver, reader), группы
- [ ] Валидация JWT на API (go-oidc, JWKS), middleware `RequirePermission("doc:approve")`
- [ ] ABAC: доступ к документу по подразделению/грифу
- [ ] Frontend: login flow (PKCE), silent refresh, protected routes
- [ ] Refresh token rotation, logout, token introspection — разобрать когда что
- **DoD:** незалогиненный не получает ничего; права проверяются на уровне use case.

## Этап 5 — Файлы и версии: MinIO `[ ]`
**Цель:** работа с бинарными документами. **Навыки:** S3 API, presigned URLs, streaming, антивирус-hook (ClamAV, опционально).

- [ ] Bucket-структура, ключи `{module}/{documentId}/{versionId}/{filename}`
- [ ] Загрузка: multipart upload больших файлов, контроль типов/размера
- [ ] Presigned GET/PUT (короткий TTL), скачивание через API или напрямую из MinIO
- [ ] Версионирование документа: immutable versions, current pointer
- **DoD:** документ можно загрузить/скачать/откатить на версию; файлы не утекают без прав.

## Этап 6 — Жизненный цикл и маршруты согласования `[ ]`
**Цель:** сердце СЭД. **Навыки:** state machine, domain events, outbox, Kafka.

- [ ] Статусная модель документа (draft → in_review → approved/rejected → published/archived)
- [ ] Маршруты согласования: шаги, делегирование, дедлайны, эскалация
- [ ] Поручения (tasks/resolutions): исполнитель, срок, отчёт
- [ ] Доменные события (`DocumentApproved`, `TaskCreated`) + Outbox-таблица + relay в Kafka
- [ ] Idempotent consumers; retry/DLQ политика
- **DoD:** полный цикл согласования работает end-to-end; события доставляются ровно-однократно по эффекту.

## Этап 7 — Frontend Next.js (полноценный) `[ ]`
**Цель:** enterprise SPA/SSR на FSD. **Навыки:** App Router, RTK Query, формы, таблицы, i18n, дизайн-система.

- [ ] FSD-структура, public API слоёв, cross-import linter (стадия конфигурации)
- [ ] RTK Query api-slice: теги кэша, invalidation, optimistic updates
- [ ] Реестр документов: TanStack Table (сортировка/фильтры/pagination серверные)
- [ ] Карточка документа: версии, история согласования, действия по правам
- [ ] Формы: react-hook-form + zod; загрузка файлов с прогрессом
- [ ] i18n ru/en (next-intl), dark mode, a11y базово
- **DoD:** основные сценарии пользователя выполняются из UI; E2E smoke Playwright зелёные.

## Этап 8 — Платформенные сервисы `[ ]`
**Цель:** производительность и UX. **Навыки:** Redis, поиск, нотификации, rate limiting, idempotency.

- [ ] Redis: кэш справочников (cache-aside, TTL, инвалидация по событиям)
- [ ] Rate limiting (token bucket per user/IP), идемпотентные POST (Idempotency-Key)
- [ ] Поиск: PostgreSQL FTS (tsvector, websearch) → сравнение с Meilisearch/OpenSearch (ADR)
- [ ] Нотификации: email (SMTP dev → Mailpit), in-app (WebSocket/SSE)
- [ ] Аудит: append-only журнал действий (who/what/when/before/after)
- **DoD:** p95 API < 200ms на типовых операциях локально; лимиты и аудит работают.

## Этап 9 — Аналитика ClickHouse `[ ]`
**Цель:** OLAP-отчётность. **Навыки:** ClickHouse schema design, ingestion, OLAP SQL.

- [ ] Таблицы: document_events (MergeTree), партиции по месяцу, ORDER BY
- [ ] Ingestion: Kafka consumer пишет события в CH
- [ ] Отчёты: обороты документов, SLA согласований, активность подразделений
- [ ] Дашборд в Grafana поверх CH; экспорт CSV/XLSX из API
- **DoD:** отчёты строятся за < 1s на 1M событий; OLTP не затронут.

## Этап 10 — DevOps maturity `[ ]`
**Цель:** продакшн-подобное окружение. **Навыки:** K8s, Helm, GitOps, observability, security scan, load testing.

- [ ] Observability: OTel traces (Tempo), метрики Prometheus + Grafana дашборды, логи Loki
- [ ] kind/k3d: манифесты/ Helm chart для api+web+infra; secrets via Sealed Secrets/Vault
- [ ] ArgoCD GitOps; окружения dev/stage; blue-green или canary разбор
- [ ] k6 нагрузочные сценарии (реестр, карточка, согласование); бенчмарки до/после оптимизаций
- [ ] Trivy/gosec в CI; OWASP ZAP baseline scan
- **DoD:** деплой одной командой/GitOps merge; дашборды показывают золотые сигналы.

## Этап 11 — Полировка senior-уровня `[ ]` *(опционально)*
- [ ] Распил одного модуля на сервис (например notifications) — практика границ
- [ ] Chaos-тестирование (kill контейнеров), graceful degradation
- [ ] Performance deep-dive: pprof, flamegraphs, оптимизация топ-N запросов
- [ ] Собеседование-self: подготовить рассказ об архитектуре проекта (портфолио)

---

## Правила движения
1. Этап = серия коротких веток `feat/eos-NNN-slug` с PR.
2. В конце этапа — ретро: что узнали, что упростить, обновить ROADMAP.
3. Если технология «не заходит» — фиксируем ADR с альтернативой, не тащим через силу.

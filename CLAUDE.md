# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Marimo ERP - микросервисная ERP-система с архитектурой на Go (backend) и React (frontend). Система включает 7 микросервисов, API Gateway, и современный стек инфраструктуры (PostgreSQL, Redis, Consul, RabbitMQ).

## Development Commands

### Running Services Locally

**Вариант 1: Docker Compose (рекомендуется)**
```bash
# Запустить все сервисы
docker-compose up --build

# В фоновом режиме
docker-compose up -d --build

# Просмотр логов
docker-compose logs -f [service-name]

# Остановка
docker-compose down
```

**Вариант 2: Локальная разработка**
```bash
# Backend сервисы (нужен каждый в отдельном терминале)
cd services/gateway && go run cmd/server/main.go      # :8080
cd services/users && go run cmd/server/main.go        # :8081
cd services/config && go run cmd/server/main.go       # :8082
cd services/accounting && go run cmd/server/main.go   # :8083
cd services/factory && go run cmd/server/main.go      # :8084
cd services/shop && go run cmd/server/main.go         # :8085
cd services/main && go run cmd/server/main.go         # :8086

# Frontend
cd frontend && npm install && npm start               # :3000
```

### Testing

```bash
# Go тесты
go test ./...                        # Все тесты
go test -cover ./...                 # С покрытием
go test ./shared/tenancy -v          # Конкретный пакет
go test -tags=integration ./...      # Интеграционные тесты

# Frontend тесты
cd frontend
npm test                             # Запуск тестов
npm test -- --coverage               # С покрытием

# Линтинг
golangci-lint run                    # Go linting
cd frontend && npm run lint          # Frontend linting
```

### Database Operations

```bash
# Инициализация БД (PostgreSQL)
./scripts/init-db.sh

# Запуск миграций
./scripts/run-migrations.sh

# Сброс БД (ОСТОРОЖНО! Удалит все данные)
./scripts/reset-db.sh

# Резервное копирование
./scripts/backup.sh

# Восстановление
./scripts/restore.sh backup-file.sql

# Подключение к БД
psql -h localhost -U postgres -d marimo_dev
```

### Building and Dependencies

```bash
# Обновление Go зависимостей
cd services/[service-name]
go mod tidy
go mod download

# Обновление shared библиотеки
cd shared
go mod tidy

# Обновление frontend зависимостей
cd frontend
npm install
```

## Architecture Overview

### Микросервисная архитектура

Система построена на **микросервисной архитектуре** с четким разделением ответственности:

**API Gateway** (`:8080`) - единая точка входа:
- Маршрутизация к backend сервисам
- Rate limiting (10 req/min для login, 60 req/min по умолчанию)
- CORS обработка
- Reverse proxy для всех `/api/*` запросов

**Микросервисы:**
- **Users Service** (`:8081`) - аутентификация, RBAC, JWT токены, refresh tokens
- **Config Service** (`:8082`) - централизованная конфигурация системы
- **Accounting Service** (`:8083`) - бухгалтерия, финансы, транзакции
- **Factory Service** (`:8084`) - производство, заказы, продукты
- **Shop Service** (`:8085`) - интернет-магазин, каталог, корзина
- **Main Service** (`:8086`) - dashboard, статистика, основная бизнес-логика

### Shared Libraries (critical!)

Директория `shared/` содержит переиспользуемые библиотеки для всех сервисов:

**Ключевые модули:**
- `shared/middleware` - JWT auth, CORS, RBAC, rate limiting
- `shared/database` - PostgreSQL и in-memory DB адаптеры
- `shared/logger` - structured logging
- `shared/utils` - graceful shutdown, helpers
- `shared/models` - общие модели данных
- `shared/validator` - валидация (email, пароли)
- `shared/discovery` - Consul service discovery
- `shared/cache` - Redis кеширование
- `shared/async` - RabbitMQ message queue
- `shared/resilience` - circuit breaker, retry logic
- `shared/tenancy` - multi-tenancy support
- `shared/webhooks` - webhook система
- `shared/monitoring` - Prometheus метрики
- `shared/integrations` - Stripe, SendGrid

**ВАЖНО:** При изменении в `shared/` нужно обновить все зависимые сервисы:
```bash
cd shared && go mod tidy
cd ../services/users && go mod tidy
# И т.д. для каждого сервиса
```

### Service Discovery и Resilience

**Consul** используется для service discovery:
- Сервисы регистрируются при запуске
- API Gateway обнаруживает сервисы автоматически
- Health checks каждые 10 секунд
- DNS интерфейс для резолвинга

**Circuit Breaker** защищает от каскадных сбоев:
- Threshold: 5 запросов
- Failure rate: >50% открывает circuit
- Timeout: 30 секунд до half-open
- См. `shared/resilience/circuit_breaker.go`

**Retry Logic** с exponential backoff:
- 3 попытки максимум
- Backoff: 100ms, 200ms, 400ms
- Jitter для предотвращения "thundering herd"
- Только для retryable errors (5xx, timeout)

### Database Strategy

**Два режима работы:**

1. **In-Memory Database** (по умолчанию):
   - Быстрый старт без PostgreSQL
   - Для разработки и тестирования
   - Данные теряются при перезапуске
   - Настройка: `USE_POSTGRES=false` в `.env`

2. **PostgreSQL** (production):
   - Постоянное хранение
   - Полная поддержка транзакций
   - Миграции в `migrations/`
   - Настройка: `USE_POSTGRES=true` в `.env`

**Database adapters** в `shared/database`:
- Единый интерфейс для обоих режимов
- Сервисы не знают о конкретной реализации
- Переключение через environment variables

### Authentication & Authorization

**JWT-based аутентификация:**
- Access tokens: 15 минут
- Refresh tokens: 7 дней
- Bcrypt для паролей (cost 10)
- Token revocation через БД

**RBAC роли:**
- `admin` - полный доступ
- `manager` - управление производством
- `accountant` - доступ к бухгалтерии
- `shop_manager` - управление магазином
- `user` - базовый доступ

**Middleware chain:**
1. Rate Limiting (на Gateway)
2. CORS
3. JWT Validation
4. Role-based access (где требуется)

**Защищенные endpoints:**
```go
// Требуют JWT токен в заголовке Authorization: Bearer <token>
router.Use(middleware.AuthMiddleware())

// Требуют конкретную роль
router.Use(middleware.RequireRole("admin"))
```

## Common Patterns

### Adding New Endpoint

1. **Определить handler** в сервисе
2. **Зарегистрировать маршрут** в `cmd/server/main.go`
3. **Добавить middleware** (JWT, RBAC если нужно)
4. **Обновить Gateway** маршрутизацию (если требуется)
5. **Написать тесты**

### Adding Database Migration

```bash
# Создать новую миграцию
cat > migrations/XXX_description.sql <<EOF
-- Up
CREATE TABLE ...;

-- Down
DROP TABLE ...;
EOF

# Запустить миграцию
./scripts/run-migrations.sh
```

### Using Message Queue

```go
import "github.com/dayanch951/marimo/shared/async"

// Опубликовать событие
publisher, _ := async.NewEventPublisher("amqp://...")
publisher.PublishUserRegistered(userID, email)
publisher.PublishEmail(to, subject, body)
publisher.PublishAuditLog(userID, action, module)
```

### Using Cache

```go
import "github.com/dayanch951/marimo/shared/cache"

cache, _ := cache.NewRedisCache("redis:6379", "")

// Cache-aside паттерн
cache.GetOrSet("user:123", &user, 5*time.Minute, func() (interface{}, error) {
    return fetchUserFromDB(123)
})
```

### Circuit Breaker Pattern

```go
import "github.com/dayanch951/marimo/shared/resilience"

cb := resilience.NewCircuitBreaker(resilience.Settings{
    Name:        "service-name",
    MaxRequests: 3,
    Timeout:     30 * time.Second,
})

err := cb.Execute(func() error {
    return callExternalService()
})
```

## Testing Strategy

**Unit Tests:**
- Каждая функция в `shared/` должна иметь тесты
- Target: >70% coverage
- Mock external dependencies

**Integration Tests:**
- Тестируют API endpoints
- Требуют запущенные сервисы
- Tag: `-tags=integration`

**E2E Tests:**
- Критичные user flows
- Login -> Create resource -> Verify
- Запускаются против staging environment

## Security Considerations

**Input Validation:**
- Email format (regex)
- Password strength: 8+ chars, upper/lower/digit/special
- SQL injection protection (parameterized queries)
- XSS protection (input sanitization)

**Rate Limiting:**
- Login: 10 req/min
- Register: 5 req/min
- Default: 60 req/min
- Настройка в `services/gateway/cmd/server/main.go`

**Production Checklist:**
1. Изменить `JWT_SECRET` (минимум 32 символа)
2. Включить PostgreSQL (`USE_POSTGRES=true`)
3. Настроить HTTPS/SSL
4. Rate limiting уже включен
5. Включить PostgreSQL SSL (`DB_SSL_MODE=require`)

## Environment Variables

См. `.env.example` для полного списка. Ключевые переменные:

```bash
# Database
USE_POSTGRES=false              # true для PostgreSQL, false для in-memory
DB_HOST=localhost
DB_NAME=marimo_dev

# Services
JWT_SECRET=your-secret-key      # ОБЯЗАТЕЛЬНО изменить в production!
CONSUL_ADDR=consul:8500
REDIS_ADDR=redis:6379
RABBITMQ_URL=amqp://admin:admin@rabbitmq:5672/

# Integrations
STRIPE_API_KEY=sk_test_...
SENDGRID_API_KEY=SG...

# Monitoring
PROMETHEUS_ENABLED=true
JAEGER_ENABLED=true
```

## Service Communication

**Synchronous (HTTP):**
- Gateway -> Services: REST API
- Service discovery через Consul или прямые URL
- Retry logic + Circuit breaker для resilience

**Asynchronous (Message Queue):**
- RabbitMQ для фоновых задач
- Очереди: `email_queue`, `audit_queue`, `notification_queue`, `events_queue`
- Workers обрабатывают события асинхронно

**Future: gRPC**
- Proto файлы уже есть в `shared/proto/`
- Планируется для inter-service communication
- Более эффективный, чем REST

## Default Credentials

```
Email: admin@example.com
Password: admin123
```

**ВАЖНО:** Изменить в production! Создается автоматически при первом запуске.

## Troubleshooting

**Порт занят:**
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -i :8080
kill -9 <PID>
```

**БД не подключается:**
```bash
docker-compose ps postgres
docker-compose restart postgres
docker-compose logs postgres
```

**Сервис не регистрируется в Consul:**
```bash
# Проверить Consul UI
open http://localhost:8500

# Проверить зарегистрированные сервисы
curl http://localhost:8500/v1/catalog/services
```

**Module import errors:**
```bash
go clean -modcache
go mod download
go mod verify
```

## Important Notes

1. **Всегда обновляйте `shared/` зависимости** после изменений в shared библиотеках
2. **Graceful shutdown реализован** во всех сервисах через `shared/utils/shutdown.go`
3. **Structured logging** используется везде - избегайте `log.Println()`, используйте `logger.Info()` и т.д.
4. **Database adapter pattern** - сервисы не должны напрямую зависеть от PostgreSQL/in-memory
5. **Multi-tenancy готов** - см. `shared/tenancy/` для использования
6. **Webhooks система** - см. `shared/webhooks/` для интеграций
7. **Frontend использует React Query** для API calls и кеширования
8. **i18n готов** во frontend (react-i18next)

## Next Steps / Roadmap

См. `NEXT_STEPS.md` для полного roadmap. Приоритеты:

1. ✅ PostgreSQL интеграция
2. ✅ Structured logging
3. ✅ Graceful shutdown
4. ⏳ Unit & Integration тесты (target >70% coverage)
5. ⏳ Kubernetes deployment
6. ⏳ CI/CD pipeline (GitHub Actions)
7. ⏳ GraphQL gateway
8. ⏳ gRPC для inter-service communication

## Useful Links

- Test Plan: `TEST_PLAN.md`
- Architecture: `docs/ARCHITECTURE.md`
- Developer Onboarding: `docs/DEVELOPER_ONBOARDING.md`
- Deployment Guide: `docs/DEPLOYMENT.md`
- API Documentation: `docs/API_DOCUMENTATION.md`
- Advanced Features: `docs/ADVANCED_FEATURES.md`

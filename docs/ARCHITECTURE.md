# Marimo ERP - Архитектура системы

## Обзор

Marimo ERP - это микросервисная система управления предприятием, построенная с использованием современных паттернов проектирования и технологий.

## Компоненты системы

### Микросервисы

1. **API Gateway** (`:8080`)
   - Единая точка входа для всех клиентских запросов
   - Маршрутизация запросов к backend сервисам
   - Аутентификация и авторизация (JWT)
   - Rate limiting
   - Circuit breaker и retry logic
   - Кеширование ответов

2. **Users Service** (`:8081`)
   - Управление пользователями
   - Аутентификация (login/logout)
   - Управление токенами (access/refresh)
   - RBAC (Role-Based Access Control)

3. **Config Service** (`:8082`)
   - Централизованное управление конфигурацией

4. **Accounting Service** (`:8083`)
   - Бухгалтерия и финансы

5. **Factory Service** (`:8084`)
   - Управление производством

6. **Shop Service** (`:8085`)
   - Управление магазинами

7. **Main Service** (`:8086`)
   - Основная бизнес-логика

### Инфраструктурные компоненты

#### Базы данных и хранилища
- **PostgreSQL** (`:5432`) - основная реляционная база данных
- **Redis** (`:6379`) - кеширование и distributed locks

#### Service Discovery
- **Consul** (`:8500`)
  - Автоматическое обнаружение сервисов
  - Health checks
  - Key-Value хранилище для конфигурации
  - DNS интерфейс для service discovery

#### Message Queue
- **RabbitMQ** (`:5672`, Management UI: `:15672`)
  - Асинхронная обработка задач
  - Event-driven архитектура
  - Email уведомления
  - Audit logging
  - Интеграции с внешними системами

#### Мониторинг и observability
- **Prometheus** (`:9090`) - сбор метрик
- **Grafana** (`:3001`) - визуализация метрик
- **Jaeger** (`:16686`) - distributed tracing
- **ELK Stack**:
  - Elasticsearch (`:9200`) - хранилище логов
  - Logstash (`:5000`) - агрегация логов
  - Kibana (`:5601`) - анализ и визуализация логов

## Паттерны проектирования

### 1. Service Discovery (Consul)

Все микросервисы регистрируются в Consul при запуске и автоматически обнаруживаются API Gateway.

```go
// Регистрация сервиса
registry, _ := discovery.NewServiceRegistry("consul:8500")
registry.Register(discovery.ServiceConfig{
    ID:              "users-service-1",
    Name:            "users",
    Address:         "users",
    Port:            8081,
    HealthCheckPath: "/health",
})

// Обнаружение сервиса
serviceURL, _ := registry.DiscoverService("users")
```

**Преимущества:**
- Динамическая маршрутизация
- Автоматическое удаление недоступных сервисов
- Load balancing через DNS
- Health monitoring

### 2. Circuit Breaker

Защита от каскадных сбоев при недоступности backend сервисов.

```go
cb := resilience.NewCircuitBreaker(resilience.Settings{
    Name:        "users-service",
    MaxRequests: 3,
    Timeout:     30 * time.Second,
    Threshold:   5,
    FailureRate: 0.5, // 50%
})

err := cb.Execute(func() error {
    return callBackendService()
})
```

**Состояния:**
- **Closed** - нормальная работа
- **Open** - сервис недоступен, запросы отклоняются немедленно
- **Half-Open** - тестирование восстановления сервиса

**Параметры:**
- Threshold: минимум 5 запросов для анализа
- FailureRate: открыть circuit при >50% ошибок
- Timeout: 30 секунд до перехода в half-open

### 3. Retry Logic

Автоматические повторные попытки при временных сбоях.

```go
policy := resilience.RetryPolicy{
    MaxAttempts:  3,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     10 * time.Second,
    Multiplier:   2.0,
    Jitter:       true,
}

result, err := resilience.Retry(ctx, policy, func() error {
    return makeHTTPRequest()
})
```

**Стратегия:**
- Exponential backoff: 100ms, 200ms, 400ms...
- Jitter для предотвращения "thundering herd"
- Максимум 3 попытки
- Только для retryable errors (5xx, timeout, network errors)

### 4. Caching (Redis)

Многоуровневое кеширование для снижения нагрузки.

```go
cache, _ := cache.NewRedisCache("redis:6379", "")

// Простое кеширование
cache.Set("user:123", userData, 5*time.Minute)

// Паттерн cache-aside
cache.GetOrSet("user:123", &user, 5*time.Minute, func() (interface{}, error) {
    return fetchUserFromDB(123)
})
```

**Стратегии кеширования:**
- Cache-aside (lazy loading)
- Write-through для критичных данных
- TTL: 5 минут для пользовательских данных
- Invalidation по паттерну для связанных ключей

### 5. Message Queue (RabbitMQ)

Асинхронная обработка для некритичных операций.

```go
publisher, _ := async.NewEventPublisher("amqp://admin:admin@rabbitmq:5672/")

// Публикация события
publisher.PublishUserRegistered(userID, email)
publisher.PublishEmail(to, subject, body)
publisher.PublishAuditLog(userID, "login", "auth")
```

**Очереди:**
- `email_queue` - отправка email уведомлений
- `audit_queue` - запись audit логов
- `notification_queue` - push уведомления
- `events_queue` - общие события системы

**Workers:**
- Email Worker - отправка писем (SendGrid, AWS SES)
- Audit Worker - запись аудит логов в БД
- Notification Worker - push/SMS уведомления

## Безопасность

### Аутентификация
- JWT токены (access: 15 мин, refresh: 7 дней)
- Bcrypt для хеширования паролей
- Refresh token rotation
- Token revocation через БД

### Авторизация
- RBAC (Role-Based Access Control)
- Роли: admin, manager, user
- Middleware проверка прав доступа

### Защита от атак
- Rate limiting (token bucket algorithm)
- SQL injection prevention (parameterized queries)
- XSS protection (input sanitization)
- HTTPS/TLS в production
- Security headers (CSP, HSTS, X-Frame-Options)

### Валидация
- Email format
- Password complexity (8+ chars, uppercase, lowercase, digit)
- Input sanitization
- Content-Type validation

## Наблюдаемость (Observability)

### Метрики (Prometheus)

**HTTP метрики:**
- `http_requests_total` - количество запросов
- `http_request_duration_seconds` - latency
- `http_request_size_bytes` - размер запросов
- `http_response_size_bytes` - размер ответов

**Business метрики:**
- `active_users_total` - активные пользователи
- `auth_attempts_total` - попытки аутентификации
- `tokens_issued_total` - выпущенные токены

**Database метрики:**
- `db_queries_total` - количество запросов к БД
- `db_query_duration_seconds` - latency запросов

### Трейсинг (Jaeger)

Распределенная трассировка для анализа latency:
- Полный путь запроса через микросервисы
- Timing каждого шага
- Идентификация узких мест
- Error tracking

### Логирование (ELK)

Централизованное логирование:
- Structured JSON logs
- Корреляция по request ID
- Log levels: DEBUG, INFO, WARN, ERROR
- Поиск и фильтрация через Kibana

### Алерты (Prometheus Alertmanager)

Критичные алерты:
- High error rate (>5% за 5 минут)
- High latency (P95 >1s за 5 минут)
- Service down (>1 минуты)
- High memory usage (>90%)
- Database connection failures
- High auth failure rate (>50% - возможная атака)

## Deployment

### Docker Compose

```bash
# Запуск основных сервисов
docker-compose up -d

# Запуск monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# Проверка статуса
docker-compose ps
```

### Порты

| Сервис | Порт | Описание |
|--------|------|----------|
| Gateway | 8080 | API Gateway |
| Users | 8081 | Users Service |
| Config | 8082 | Config Service |
| Accounting | 8083 | Accounting Service |
| Factory | 8084 | Factory Service |
| Shop | 8085 | Shop Service |
| Main | 8086 | Main Service |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache |
| Consul | 8500 | Service Discovery |
| RabbitMQ | 5672 | Message Queue |
| RabbitMQ UI | 15672 | Management Interface |
| Prometheus | 9090 | Metrics |
| Grafana | 3001 | Dashboards |
| Jaeger | 16686 | Tracing UI |
| Elasticsearch | 9200 | Logs Storage |
| Kibana | 5601 | Logs UI |

### Переменные окружения

```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=marimo_dev

# JWT
JWT_SECRET=your-secret-key-change-this-in-production

# Service Discovery
CONSUL_ADDR=consul:8500

# Cache
REDIS_ADDR=redis:6379
REDIS_PASSWORD=

# Message Queue
RABBITMQ_URL=amqp://admin:admin@rabbitmq:5672/
```

## Масштабирование

### Горизонтальное масштабирование

1. **Stateless сервисы** - легко масштабируются
   ```bash
   docker-compose up -d --scale users=3
   ```

2. **Load balancing** через Consul DNS или nginx

3. **Session management** через Redis (shared state)

### Вертикальное масштабирование

- Увеличение CPU/Memory для БД
- Connection pooling
- Query optimization

### Database sharding

- Шардирование по tenant_id
- Read replicas для read-heavy операций

## Best Practices

### Code Organization
- Shared packages для переиспользования
- Clear separation of concerns
- Interface-based design

### Error Handling
- Structured errors с контекстом
- Graceful degradation
- Circuit breaker для защиты

### Testing
- Unit tests (>70% coverage)
- Integration tests для API
- E2E tests для критичных flows
- Load testing с k6

### Security
- Principle of least privilege
- Input validation на всех уровнях
- Secrets management (не в коде!)
- Regular security audits

### Monitoring
- Алерты для всех критичных метрик
- SLO/SLA определения
- Incident response playbooks

## Roadmap

### Phase 1 (Complete)
- ✅ Базовая архитектура микросервисов
- ✅ Аутентификация и авторизация
- ✅ Security improvements
- ✅ Testing infrastructure
- ✅ Observability stack

### Phase 2 (Complete)
- ✅ Service discovery (Consul)
- ✅ Message queue (RabbitMQ)
- ✅ Caching (Redis)
- ✅ Circuit breaker
- ✅ Retry logic

### Phase 3 (Planned)
- [ ] API versioning
- [ ] GraphQL gateway
- [ ] gRPC для inter-service communication
- [ ] Kubernetes deployment
- [ ] CI/CD pipeline
- [ ] Multi-tenancy support

### Phase 4 (Future)
- [ ] Machine learning integration
- [ ] Real-time analytics
- [ ] Mobile apps
- [ ] Advanced reporting

## Документация

- [Security](./SECURITY.md) - детали безопасности
- [Observability](./OBSERVABILITY.md) - мониторинг и трейсинг
- [Testing](./TESTING.md) - стратегия тестирования
- [API Documentation](./API.md) - API endpoints

## Поддержка

Для вопросов и предложений:
- GitHub Issues
- Email: support@marimo.dev
- Documentation: https://docs.marimo.dev

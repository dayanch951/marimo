# Docker Setup Guide

Это руководство поможет вам запустить весь проект Marimo ERP в Docker контейнерах.

## Быстрый старт

### Вариант 1: Использование скрипта (рекомендуется)

```bash
# Запустить все сервисы
./start-docker.sh up

# Просмотр логов
./start-docker.sh logs

# Остановить все сервисы
./start-docker.sh down
```

### Вариант 2: Docker Compose напрямую

```bash
# Запустить все сервисы
docker-compose up -d --build

# Просмотр логов
docker-compose logs -f

# Остановить все сервисы
docker-compose down
```

## Архитектура Docker

### Сервисы

Проект включает следующие контейнеры:

#### Инфраструктурные сервисы:
- **PostgreSQL** (порт 5432) - база данных
- **Redis** (порт 6379) - кеширование
- **Consul** (порт 8500) - service discovery
- **RabbitMQ** (порт 5672, UI: 15672) - message queue

#### Микросервисы:
- **Gateway** (порт 8080) - API Gateway, единая точка входа
- **Users** (порт 8081) - управление пользователями и аутентификация
- **Config** (порт 8082) - конфигурация системы
- **Accounting** (порт 8083) - бухгалтерия
- **Factory** (порт 8084) - производство
- **Shop** (порт 8085) - интернет-магазин
- **Main** (порт 8086) - основной сервис и dashboard

#### Frontend:
- **Frontend** (порт 3000) - React приложение с nginx

### Volumes

Данные сохраняются в Docker volumes:
- `postgres_data` - данные PostgreSQL
- `redis_data` - данные Redis
- `consul_data` - данные Consul
- `rabbitmq_data` - данные RabbitMQ

### Networks

Все сервисы работают в единой сети `marimo-network` типа bridge.

## Доступные команды скрипта

```bash
./start-docker.sh [COMMAND]
```

**Команды:**

| Команда | Описание |
|---------|----------|
| `up` | Запустить все сервисы в фоновом режиме (по умолчанию) |
| `upf` или `up-foreground` | Запустить в foreground (логи видны) |
| `down` | Остановить все сервисы |
| `restart` | Перезапустить все сервисы |
| `logs [service]` | Показать логи (опционально конкретного сервиса) |
| `ps` или `status` | Показать статус всех сервисов |
| `clean` | Удалить все контейнеры, volumes и образы |
| `help` | Показать справку |

**Примеры:**

```bash
# Запустить все сервисы
./start-docker.sh up

# Просмотр логов Gateway
./start-docker.sh logs gateway

# Просмотр логов всех сервисов
./start-docker.sh logs

# Проверить статус
./start-docker.sh ps

# Остановить все
./start-docker.sh down

# Полная очистка
./start-docker.sh clean
```

## Конфигурация (.env файл)

Перед запуском убедитесь, что у вас есть файл `.env` в корне проекта. Если его нет, он будет автоматически создан из `.env.example`.

**Основные переменные:**

```bash
# Режим работы БД
USE_POSTGRES=false          # false = in-memory, true = PostgreSQL

# Database
DB_HOST=postgres
DB_NAME=marimo_dev
DB_USER=postgres
DB_PASSWORD=postgres

# Redis
REDIS_ADDR=redis:6379

# Consul
CONSUL_ADDR=consul:8500

# RabbitMQ
RABBITMQ_URL=amqp://admin:admin@rabbitmq:5672/

# JWT Secret (ВАЖНО: измените в production!)
JWT_SECRET=marimo-dev-secret-key-change-this-in-production-32chars
```

## Работа с сервисами

### Проверка здоровья сервисов

```bash
# Статус всех контейнеров
docker-compose ps

# Проверить здоровье конкретного сервиса
docker inspect --format='{{.State.Health.Status}}' marimo-postgres
```

### Логи

```bash
# Все логи
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f gateway

# Последние 100 строк
docker-compose logs --tail=100 users

# Логи за последние 10 минут
docker-compose logs --since 10m
```

### Подключение к контейнеру

```bash
# Bash в контейнере Gateway
docker-compose exec gateway sh

# Подключение к PostgreSQL
docker-compose exec postgres psql -U postgres -d marimo_dev

# Подключение к Redis
docker-compose exec redis redis-cli
```

### Перезапуск отдельного сервиса

```bash
# Перезапустить Gateway
docker-compose restart gateway

# Пересобрать и перезапустить
docker-compose up -d --build gateway
```

## Разработка с Docker

### Hot reload

Frontend поддерживает hot reload по умолчанию. Для Go сервисов нужно перезапустить контейнер после изменений.

### Быстрый цикл разработки

```bash
# 1. Внести изменения в код

# 2. Пересобрать конкретный сервис
docker-compose up -d --build users

# 3. Проверить логи
docker-compose logs -f users
```

### Отладка

```bash
# Войти в контейнер
docker-compose exec users sh

# Проверить переменные окружения
docker-compose exec users env

# Проверить процессы
docker-compose exec users ps aux
```

## Режимы работы БД

### In-Memory режим (по умолчанию)

```bash
# В .env файле
USE_POSTGRES=false
```

- Быстрый старт без PostgreSQL
- Данные теряются при перезапуске
- Идеально для разработки и тестов

### PostgreSQL режим

```bash
# В .env файле
USE_POSTGRES=true
```

- Постоянное хранение данных
- Полная поддержка транзакций
- Для production окружения

## Доступ к UI интерфейсам

После запуска сервисов доступны следующие интерфейсы:

| Сервис | URL | Credentials |
|--------|-----|-------------|
| Frontend | http://localhost:3000 | admin@example.com / admin123 |
| API Gateway | http://localhost:8080 | - |
| Consul UI | http://localhost:8500 | - |
| RabbitMQ Management | http://localhost:15672 | admin / admin |

## Порты

Убедитесь, что следующие порты свободны:

- **3000** - Frontend
- **8080** - API Gateway
- **8081-8086** - Микросервисы
- **5432** - PostgreSQL
- **6379** - Redis
- **8500** - Consul
- **5672** - RabbitMQ AMQP
- **15672** - RabbitMQ Management UI

Проверка занятых портов:

```bash
# Linux/Mac
lsof -i :8080

# Windows
netstat -ano | findstr :8080
```

## Troubleshooting

### Порт занят

```bash
# Найти процесс, занимающий порт
lsof -i :8080

# Остановить старые контейнеры
docker-compose down

# Или убить конкретный процесс
kill -9 <PID>
```

### Контейнер не запускается

```bash
# Проверить логи
docker-compose logs [service-name]

# Проверить статус
docker-compose ps

# Пересобрать с нуля
docker-compose down
docker-compose up --build
```

### PostgreSQL не подключается

```bash
# Проверить статус
docker-compose ps postgres

# Проверить логи
docker-compose logs postgres

# Перезапустить
docker-compose restart postgres

# Проверить healthcheck
docker inspect --format='{{.State.Health.Status}}' marimo-postgres
```

### Consul не регистрирует сервисы

```bash
# Проверить Consul UI
open http://localhost:8500

# Проверить зарегистрированные сервисы
curl http://localhost:8500/v1/catalog/services

# Перезапустить сервис
docker-compose restart consul
```

### Redis не подключается

```bash
# Проверить Redis
docker-compose exec redis redis-cli ping

# Должен ответить PONG
```

### Очистка всего Docker

```bash
# ОСТОРОЖНО! Удаляет ВСЕ Docker данные
docker system prune -a --volumes
```

## Production рекомендации

1. **Измените JWT_SECRET** в `.env` на случайную строку минимум 32 символа
2. **Измените пароли** для PostgreSQL, Redis, RabbitMQ
3. **Включите PostgreSQL**: `USE_POSTGRES=true`
4. **Включите SSL** для PostgreSQL: `DB_SSL_MODE=require`
5. **Настройте reverse proxy** (nginx) с HTTPS
6. **Настройте резервное копирование** volumes
7. **Мониторинг**: используйте `docker-compose.monitoring.yml` для Prometheus + Grafana

### Production запуск

```bash
# С мониторингом
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d

# С HTTPS
docker-compose -f docker-compose.yml -f docker-compose.https.yml up -d
```

## Дополнительные файлы

- `docker-compose.yml` - основной файл
- `docker-compose.monitoring.yml` - Prometheus + Grafana
- `docker-compose.https.yml` - HTTPS конфигурация
- `Dockerfile.service` - универсальный Dockerfile для Go сервисов
- `frontend/Dockerfile` - Dockerfile для React + nginx

## Полезные ссылки

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Project README](README.md)
- [CLAUDE.md](CLAUDE.md) - инструкции для разработки

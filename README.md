# Marimo ERP - Микросервисная архитектура

Полнофункциональная ERP-система с микросервисной архитектурой на Go и React.

## Архитектура

### Микросервисы (Go)

Все сервисы работают независимо и общаются через API Gateway:

1. **API Gateway** (`:8080`) - Центральная точка входа, маршрутизация запросов
2. **Users Service** (`:8081`) - Управление пользователями, аутентификация, роли
3. **Config Service** (`:8082`) - Конфигурация системы, справочники
4. **Accounting Service** (`:8083`) - Бухгалтерия, транзакции, баланс
5. **Factory Service** (`:8084`) - Производство, продукты, заказы
6. **Shop Service** (`:8085`) - Интернет-магазин, каталог, заказы
7. **Main Service** (`:8086`) - Dashboard, статистика

### Frontend (React)

Единый SPA с модульной структурой:
- Авторизация и регистрация
- Dashboard с навигацией между модулями
- Отдельные страницы для каждого сервиса
- Защищенные роуты
- Адаптивный дизайн

### Shared Library

Общие компоненты для всех сервисов:
- JWT аутентификация
- Middleware (Auth, CORS, Role-based access)
- Модели данных
- In-memory база данных
- Proto файлы для gRPC

## Структура проекта

```
marimo/
├── shared/                   # Общие библиотеки
│   ├── proto/               # Protobuf определения
│   ├── middleware/          # Middleware компоненты
│   ├── models/              # Модели данных
│   └── utils/               # Утилиты (database)
├── services/
│   ├── gateway/             # API Gateway
│   │   └── cmd/server/
│   ├── users/               # Users Service
│   │   ├── cmd/server/
│   │   └── internal/handlers/
│   ├── config/              # Config Service
│   ├── accounting/          # Accounting Service
│   ├── factory/             # Factory Service
│   ├── shop/                # Shop Service
│   └── main/                # Main Service
├── frontend/                # React приложение
│   ├── src/
│   │   ├── components/
│   │   │   ├── modules/    # Компоненты модулей
│   │   │   └── Layout.js   # Главный layout
│   │   ├── context/        # Context API
│   │   └── services/       # API сервисы
├── backend/                 # Старый монолит (для совместимости)
├── Dockerfile.service       # Generic Dockerfile для сервисов
├── docker-compose.new.yml   # Оркестрация всех сервисов
└── README.md
```

## Быстрый старт

### Docker Compose (рекомендуется)

```bash
# Запустить все сервисы
docker-compose -f docker-compose.new.yml up --build

# Доступ:
# - Frontend: http://localhost:3000
# - API Gateway: http://localhost:8080
# - Отдельные сервисы: 8081-8086
```

### Локальная разработка

Запустите каждый сервис отдельно:

```bash
# Terminal 1 - Users Service
cd services/users
go run cmd/server/main.go

# Terminal 2 - Config Service
cd services/config
go run cmd/server/main.go

# Terminal 3 - Accounting Service
cd services/accounting
go run cmd/server/main.go

# Terminal 4 - Factory Service
cd services/factory
go run cmd/server/main.go

# Terminal 5 - Shop Service
cd services/shop
go run cmd/server/main.go

# Terminal 6 - Main Service
cd services/main
go run cmd/server/main.go

# Terminal 7 - API Gateway
cd services/gateway
go run cmd/server/main.go

# Terminal 8 - Frontend
cd frontend
npm install
npm start
```

## Модули системы

### 1. Users (Пользователи)
- Регистрация и аутентификация
- Управление пользователями
- Role-based access control (RBAC)
- Роли: admin, manager, user, accountant, shop_manager

**Endpoints:**
- `POST /api/users/register` - Регистрация
- `POST /api/users/login` - Вход
- `GET /api/users/profile` - Профиль (защищено)
- `GET /api/users/list` - Список пользователей (защищено)
- `POST /api/users/admin/assign-role` - Назначить роль (admin only)

### 2. Config (Конфигурация)
- Настройки системы
- Справочники
- Параметры приложения

**Endpoints:**
- `GET /api/config` - Список настроек
- `GET /api/config/{key}` - Получить настройку
- `POST /api/config` - Создать/обновить
- `DELETE /api/config/{key}` - Удалить

### 3. Accounting (Бухгалтерия)
- Транзакции (доходы/расходы)
- Баланс
- Финансовые отчеты

**Endpoints:**
- `GET /api/accounting/transactions` - Список транзакций
- `POST /api/accounting/transactions` - Создать транзакцию
- `GET /api/accounting/transactions/{id}` - Детали транзакции
- `GET /api/accounting/balance` - Текущий баланс

### 4. Factory (Производство)
- Управление продуктами
- Производственные заказы
- Статусы производства

**Endpoints:**
- `GET /api/factory/products` - Список продуктов
- `POST /api/factory/products` - Создать продукт
- `GET /api/factory/products/{id}` - Детали продукта
- `PUT /api/factory/products/{id}/status` - Обновить статус
- `GET /api/factory/orders` - Производственные заказы
- `POST /api/factory/orders` - Создать заказ

### 5. Shop (Интернет-магазин)
- Каталог товаров
- Корзина
- Заказы

**Endpoints:**
- `GET /api/shop/products` - Каталог (публично)
- `GET /api/shop/products/{id}` - Детали товара
- `POST /api/shop/orders` - Создать заказ (защищено)
- `GET /api/shop/orders` - Мои заказы (защищено)
- `POST /api/shop/admin/products` - Управление товарами (admin)

### 6. Main (Главная)
- Dashboard
- Общая статистика
- Навигация между модулями

**Endpoints:**
- `GET /api/main/dashboard` - Данные dashboard
- `GET /api/main/stats` - Общая статистика

## Тестирование

### Вход в систему

Default admin user:
- Email: `admin@example.com`
- Password: `admin123`

### Health Checks

Проверить статус всех сервисов:
```bash
curl http://localhost:8080/health
```

### Тестирование API

```bash
# Регистрация
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123","name":"Test User"}'

# Вход
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'

# Получить токен из ответа и использовать его:
TOKEN="your-jwt-token"

# Защищенный endpoint
curl -X GET http://localhost:8080/api/users/profile \
  -H "Authorization: Bearer $TOKEN"

# Dashboard
curl -X GET http://localhost:8080/api/main/dashboard \
  -H "Authorization: Bearer $TOKEN"

# Список продуктов (публично)
curl http://localhost:8080/api/shop/products
```

## Безопасность

- JWT tokens для аутентификации
- Role-based access control
- Bcrypt для паролей
- CORS настроен
- Защищенные роуты на backend и frontend
- Middleware для валидации токенов

**В продакшен:**
1. Измените `JWT_SECRET`
2. Используйте HTTPS
3. Настройте PostgreSQL/MySQL
4. Ограничьте CORS origins
5. Добавьте rate limiting
6. Настройте логирование

## API Gateway

Gateway проксирует запросы к соответствующим сервисам:

```
/api/users/*      → Users Service (8081)
/api/config/*     → Config Service (8082)
/api/accounting/* → Accounting Service (8083)
/api/factory/*    → Factory Service (8084)
/api/shop/*       → Shop Service (8085)
/api/main/*       → Main Service (8086)
```

## Развертывание

### Docker Compose

```bash
# Production
docker-compose -f docker-compose.new.yml up -d

# Проверить логи
docker-compose -f docker-compose.new.yml logs -f

# Остановить
docker-compose -f docker-compose.new.yml down
```

### Kubernetes

Coming soon...

## Roadmap

- [x] Микросервисная архитектура
- [x] API Gateway
- [x] 6 основных сервисов
- [x] React фронтенд с модулями
- [x] JWT аутентификация
- [x] Role-based access
- [x] Docker конфигурация
- [ ] PostgreSQL интеграция
- [ ] Redis для кеширования
- [ ] Message queue (RabbitMQ/Kafka)
- [ ] Service discovery (Consul/Etcd)
- [ ] Distributed tracing (Jaeger)
- [ ] Metrics (Prometheus/Grafana)
- [ ] Kubernetes deployment
- [ ] CI/CD pipeline
- [ ] Unit & Integration tests
- [ ] API documentation (Swagger)

## Технологии

### Backend
- Go 1.21+
- gorilla/mux (HTTP routing)
- JWT authentication
- gRPC (готово в proto файлах)
- Microservices architecture

### Frontend
- React 18
- React Router v6
- Context API
- Axios
- Responsive CSS

### DevOps
- Docker & Docker Compose
- Nginx
- Multi-stage builds

## Лицензия

MIT

## Авторы

Marimo ERP Team

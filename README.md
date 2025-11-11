# Marimo ERP - Микросервисная архитектура

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)

Полнофункциональная ERP-система с микросервисной архитектурой на Go и React.

## 🎯 Возможности

- **7 микросервисов** на Go с REST API
- **API Gateway** для маршрутизации запросов
- **JWT аутентификация** с role-based access control
- **Refresh tokens** для безопасной работы с токенами
- **Rate limiting** для защиты от злоупотреблений
- **HTTPS/SSL** поддержка для production
- **Расширенная валидация** входных данных
- **React SPA** с модульной структурой
- **Docker Compose** для быстрого запуска
- **Structured logging** для всех сервисов
- **Graceful shutdown** для стабильной работы
- **PostgreSQL** для постоянного хранения (опционально)
- **In-memory database** для быстрой разработки

## 🏗️ Архитектура

### Микросервисы (Go)

| Сервис | Порт | Описание |
|--------|------|----------|
| **Gateway** | `:8080` | API Gateway, маршрутизация |
| **Users** | `:8081` | Аутентификация, RBAC |
| **Config** | `:8082` | Конфигурация, справочники |
| **Accounting** | `:8083` | Бухгалтерия, транзакции |
| **Factory** | `:8084` | Производство, заказы |
| **Shop** | `:8085` | Интернет-магазин |
| **Main** | `:8086` | Dashboard, статистика |

### Frontend (React)

- **Dashboard** - главная страница с модулями
- **Users** - управление пользователями
- **Config** - настройки системы
- **Accounting** - бухгалтерия и финансы
- **Factory** - производственные процессы
- **Shop** - каталог и заказы

## 📁 Структура проекта

```
marimo/
├── services/              # Микросервисы
│   ├── gateway/          # API Gateway (:8080)
│   ├── users/            # Users Service (:8081)
│   ├── config/           # Config Service (:8082)
│   ├── accounting/       # Accounting Service (:8083)
│   ├── factory/          # Factory Service (:8084)
│   ├── shop/             # Shop Service (:8085)
│   └── main/             # Main Service (:8086)
├── shared/               # Общие библиотеки
│   ├── database/        # Database adapters (PostgreSQL, In-memory)
│   ├── logger/          # Structured logging
│   ├── middleware/      # JWT, CORS, RBAC
│   ├── models/          # Модели данных
│   ├── proto/           # Protobuf (gRPC)
│   └── utils/           # Shutdown, helpers
├── migrations/           # SQL миграции для PostgreSQL
├── scripts/              # Utility scripts (init-db.sh, reset-db.sh)
├── frontend/             # React приложение
│   ├── src/
│   │   ├── components/
│   │   │   ├── modules/ # Страницы модулей
│   │   │   └── Layout.js
│   │   ├── context/     # Auth context
│   │   └── services/    # API calls
├── docker-compose.yml    # Оркестрация
├── Dockerfile.service    # Generic Dockerfile
├── TEST_PLAN.md         # План тестирования
└── NEXT_STEPS.md        # Roadmap развития
```

## 🗄️ База данных

Система поддерживает **два режима работы с данными**:

### In-Memory Database (по умолчанию)
- Быстрый старт без установки PostgreSQL
- Идеально для разработки и тестирования
- Данные хранятся только в памяти (теряются при перезапуске)

### PostgreSQL (для продакшена)
- Постоянное хранение данных
- Полная поддержка транзакций
- Миграции и seed данные

#### Настройка PostgreSQL

**1. Переключиться на PostgreSQL:**

Отредактируйте `.env` файл:
```bash
USE_POSTGRES=true  # Изменить на true
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=marimo_dev
DB_SSL_MODE=disable
```

**2. Инициализировать базу данных:**

```bash
# Убедитесь, что PostgreSQL запущен
# Затем выполните:
./scripts/init-db.sh
```

Скрипт создаст:
- Таблицу `users` с индексами
- Триггер для автоматического обновления `updated_at`
- Администратора по умолчанию

**3. Сбросить базу данных (ОСТОРОЖНО!):**

```bash
# Удалит и пересоздаст БД
./scripts/reset-db.sh
```

#### Структура БД

**Таблица users:**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,  -- bcrypt hash
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

**Индексы:**
- `idx_users_email` - быстрый поиск по email
- `idx_users_role` - фильтрация по ролям
- `idx_users_created_at` - сортировка по дате

## 🚀 Быстрый старт

### Docker Compose (рекомендуется)

```bash
# Клонировать репозиторий
git clone https://github.com/dayanch951/marimo.git
cd marimo

# Запустить все сервисы
docker-compose up --build

# Доступ:
# - Frontend: http://localhost:3000
# - API Gateway: http://localhost:8080
# - Health Check: http://localhost:8080/health
```

### Локальная разработка

#### Backend сервисы

```bash
# Terminal 1: Users Service
cd services/users
go mod tidy
go run cmd/server/main.go

# Terminal 2: API Gateway
cd services/gateway
go run cmd/server/main.go

# Аналогично для остальных сервисов...
```

#### Frontend

```bash
cd frontend
npm install
npm start
# Откроется на http://localhost:3000
```

## 🔐 Аутентификация

### Default Admin

```
Email: admin@example.com
Password: admin123
```

### Роли в системе

- `admin` - полный доступ
- `manager` - управление производством
- `accountant` - доступ к бухгалтерии
- `shop_manager` - управление магазином
- `user` - базовый доступ

## 📡 API Endpoints

### Users Service (`:8081`)

```bash
# Регистрация
POST /api/users/register
{
  "email": "user@example.com",
  "password": "password123",
  "name": "User Name"
}

# Вход (получение JWT токена)
POST /api/users/login
{
  "email": "user@example.com",
  "password": "password123"
}

# Профиль (требует токен)
GET /api/users/profile
Headers: Authorization: Bearer <token>

# Список пользователей (требует токен)
GET /api/users/list
Headers: Authorization: Bearer <token>

# Назначить роль (только admin)
POST /api/users/admin/assign-role
Headers: Authorization: Bearer <token>
{
  "user_id": "uuid",
  "role": "manager"
}
```

### Config Service (`:8082`)

```bash
# Получить все настройки
GET /api/config
Headers: Authorization: Bearer <token>

# Получить конкретную настройку
GET /api/config/{key}

# Установить настройку
POST /api/config
{
  "key": "app_name",
  "value": "Marimo ERP",
  "type": "system"
}
```

### Accounting Service (`:8083`)

```bash
# Баланс (только accountant/admin)
GET /api/accounting/balance
Headers: Authorization: Bearer <token>

# Транзакции
GET /api/accounting/transactions

# Создать транзакцию
POST /api/accounting/transactions
{
  "type": "income",
  "amount": 1000.00,
  "description": "Payment received",
  "category": "Sales"
}
```

### Factory Service (`:8084`)

```bash
# Продукты (manager/admin)
GET /api/factory/products
POST /api/factory/products

# Заказы
GET /api/factory/orders
POST /api/factory/orders
```

### Shop Service (`:8085`)

```bash
# Каталог (публично)
GET /api/shop/products

# Детали товара
GET /api/shop/products/{id}

# Создать заказ (требует токен)
POST /api/shop/orders
{
  "items": [
    {"product_id": "SHOP-1", "quantity": 2, "price": 29.99}
  ]
}

# Мои заказы
GET /api/shop/orders
```

### Main Service (`:8086`)

```bash
# Dashboard
GET /api/main/dashboard
Headers: Authorization: Bearer <token>

# Статистика
GET /api/main/stats
```

## 🧪 Тестирование

Следуйте инструкциям в [TEST_PLAN.md](TEST_PLAN.md)

```bash
# Быстрый health check
curl http://localhost:8080/health

# Вход в систему
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'

# Сохраните токен и используйте:
TOKEN="your-jwt-token"

curl http://localhost:8080/api/users/profile \
  -H "Authorization: Bearer $TOKEN"
```

## 🔧 Конфигурация

### Environment Variables

Создайте файл `.env` на основе `.env.example`:

```bash
# JWT Secret
JWT_SECRET=your-secret-key-change-in-production

# Service Ports (optional, defaults shown)
GATEWAY_PORT=8080
USERS_PORT=8081
CONFIG_PORT=8082
ACCOUNTING_PORT=8083
FACTORY_PORT=8084
SHOP_PORT=8085
MAIN_PORT=8086

# Database Configuration
USE_POSTGRES=false  # true - PostgreSQL, false - in-memory
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=marimo_dev
DB_SSL_MODE=disable

# Logging
LOG_LEVEL=info  # debug, info, warn, error
LOG_FORMAT=text  # json, text
```

## 🛡️ Безопасность

### Реализованные функции

- ✅ **JWT Access Tokens** (15 минут) + **Refresh Tokens** (7 дней)
- ✅ **Bcrypt** для хэширования паролей (cost 10)
- ✅ **Role-based access control** (RBAC) с 5 ролями
- ✅ **Rate Limiting** на уровне Gateway
  - Login: 10 req/min (burst 3)
  - Register: 5 req/min (burst 2)
  - Default: 60 req/min (burst 10)
- ✅ **HTTPS/SSL** support с nginx reverse proxy
- ✅ **Input Validation**:
  - Email format validation
  - Password strength requirements (8+ chars, upper/lower/digit/special)
  - Name validation
  - SQL injection protection
  - XSS protection
- ✅ **CORS** configured
- ✅ **Protected routes** с middleware
- ✅ **Security headers** (X-Frame-Options, X-XSS-Protection, etc.)
- ✅ **Token revocation** (logout, security breach)

### API Endpoints

```bash
# Вход (получение access + refresh tokens)
POST /api/users/login
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
Response: {
  "access_token": "eyJ...",
  "refresh_token": "random-base64-string",
  "expires_in": 900,
  "token_type": "Bearer"
}

# Обновление токена
POST /api/users/refresh
{
  "refresh_token": "previous-refresh-token"
}

# Выход (отзыв токена)
POST /api/users/logout
{
  "refresh_token": "token-to-revoke"
}
```

**⚠️ Чеклист для Production:**
1. ✅ Измените `JWT_SECRET` на случайную строку (минимум 32 символа)
2. ✅ Включите PostgreSQL (`USE_POSTGRES=true`)
3. ✅ Настройте HTTPS с Let's Encrypt (см. `docker-compose.https.yml`)
4. ✅ Настройте SSL для PostgreSQL (`DB_SSL_MODE=require`)
5. ✅ Rate limiting уже включен
6. ⚠️ Настройте мониторинг и логирование
7. ⚠️ Настройте резервное копирование БД
8. ⚠️ Включите HSTS в nginx (раскомментируйте в конфиге)

## 📊 Технологии

### Backend
- Go 1.21+
- gorilla/mux (HTTP routing)
- JWT (golang-jwt/jwt)
- PostgreSQL (lib/pq driver)
- gRPC (protobuf ready)
- bcrypt (password hashing)

### Frontend
- React 18
- React Router v6
- Context API
- Axios
- CSS3

### DevOps
- Docker & Docker Compose
- Nginx
- Multi-stage builds
- Health checks

## 📈 Мониторинг

```bash
# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f users

# Health check всех сервисов
curl http://localhost:8080/health | jq
```

## 🗺️ Roadmap

См. [NEXT_STEPS.md](NEXT_STEPS.md) для подробного плана развития.

### Ближайшие задачи:
- [x] PostgreSQL интеграция ✅
- [x] Structured logging ✅
- [x] Graceful shutdown ✅
- [ ] Unit & Integration тесты
- [ ] Prometheus metrics
- [ ] Redis caching
- [ ] CI/CD pipeline
- [ ] Kubernetes deployment

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 Лицензия

MIT License - см. [LICENSE](LICENSE)

## 👥 Авторы

Marimo ERP Team

## 🙏 Благодарности

- Go community
- React team
- Open source contributors

---

**⭐ Если проект полезен - поставьте звезду!**

**📧 Вопросы?** Создайте [Issue](https://github.com/dayanch951/marimo/issues)

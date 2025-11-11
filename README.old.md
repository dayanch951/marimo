# Marimo - Go Microservice with React Authentication

Полнофункциональное веб-приложение с микросервисной архитектурой, включающее:
- **Backend**: Go микросервис с REST API и gRPC
- **Frontend**: React SPA с аутентификацией
- **Authentication**: JWT-based аутентификация

## Архитектура

### Backend (Go)
- **REST API** на порту `8080`
- **gRPC сервис** на порту `50051`
- JWT токены для аутентификации
- In-memory база данных (легко заменяется на PostgreSQL/MySQL)
- Структура проекта следует best practices Go

### Frontend (React)
- Single Page Application (SPA)
- React Router для навигации
- Context API для управления состоянием
- Axios для HTTP запросов
- Защищенные маршруты

## Структура проекта

```
marimo/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # Точка входа
│   ├── internal/
│   │   ├── handlers/            # HTTP handlers
│   │   ├── grpc/               # gRPC сервис
│   │   ├── middleware/         # Middleware (CORS, Auth)
│   │   ├── models/             # Модели данных
│   │   └── proto/              # Protobuf определения
│   ├── pkg/
│   │   ├── auth/               # JWT утилиты
│   │   └── database/           # Слой базы данных
│   ├── Dockerfile
│   ├── Makefile
│   └── go.mod
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── components/         # React компоненты
│   │   ├── context/            # Context API
│   │   ├── services/           # API сервисы
│   │   ├── App.js
│   │   └── index.js
│   ├── Dockerfile
│   ├── nginx.conf
│   └── package.json
└── docker-compose.yml
```

## Быстрый старт

### Вариант 1: Docker Compose (рекомендуется)

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd marimo
```

2. Запустите приложение:
```bash
docker-compose up --build
```

3. Откройте браузер:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- gRPC: localhost:50051

### Вариант 2: Локальная разработка

#### Backend

1. Установите зависимости:
```bash
cd backend
go mod download
```

2. Запустите сервер:
```bash
go run cmd/server/main.go
```

Или используйте Makefile:
```bash
make run
```

#### Frontend

1. Установите зависимости:
```bash
cd frontend
npm install
```

2. Запустите dev сервер:
```bash
npm start
```

Приложение откроется на http://localhost:3000

## API Endpoints

### REST API

#### Публичные эндпоинты
- `POST /api/auth/register` - Регистрация пользователя
  ```json
  {
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }
  ```

- `POST /api/auth/login` - Вход в систему
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```

#### Защищенные эндпоинты
- `GET /api/profile` - Получить профиль пользователя
  - Headers: `Authorization: Bearer <token>`

- `GET /health` - Health check

### gRPC API

Сервис определен в `backend/internal/proto/auth.proto`:

- `Login(LoginRequest) returns (LoginResponse)`
- `Register(RegisterRequest) returns (RegisterResponse)`
- `ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse)`

## Конфигурация

### Backend

Переменные окружения (см. `.env.example`):
```
JWT_SECRET=your-secret-key-change-this-in-production
HTTP_PORT=8080
GRPC_PORT=50051
```

### Frontend

Переменные окружения:
```
REACT_APP_API_URL=http://localhost:8080/api
```

## Разработка

### Генерация Protobuf кода

```bash
cd backend
make proto
```

### Сборка

Backend:
```bash
cd backend
make build
```

Frontend:
```bash
cd frontend
npm run build
```

## Тестирование

### Регистрация и вход

1. Откройте http://localhost:3000
2. Нажмите "Register here"
3. Заполните форму регистрации
4. После успешной регистрации войдите с вашими credentials
5. Вы будете перенаправлены на Dashboard

### Тестирование API с curl

Регистрация:
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123","name":"Test User"}'
```

Вход:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

Получить профиль:
```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer <your-token>"
```

## Безопасность

- Пароли хэшируются с использованием bcrypt
- JWT токены для stateless аутентификации
- CORS настроен для безопасности
- Защищенные роуты на фронтенде и бэкенде
- В продакшене обязательно измените `JWT_SECRET`!

## Production Deployment

1. Измените `JWT_SECRET` в `.env`
2. Настройте HTTPS
3. Используйте реальную базу данных (PostgreSQL/MySQL)
4. Настройте правильные CORS origins
5. Добавьте rate limiting
6. Настройте логирование и мониторинг

## Roadmap

- [ ] PostgreSQL/MySQL интеграция
- [ ] Refresh tokens
- [ ] Email верификация
- [ ] Forgot password функционал
- [ ] OAuth2 (Google, GitHub)
- [ ] Rate limiting
- [ ] Тесты (unit, integration)
- [ ] CI/CD pipeline
- [ ] Kubernetes deployment

## Технологии

### Backend
- Go 1.21+
- gorilla/mux (HTTP router)
- gRPC
- JWT (golang-jwt)
- bcrypt (password hashing)

### Frontend
- React 18
- React Router v6
- Axios
- Context API

### DevOps
- Docker
- Docker Compose
- Nginx

## Лицензия

MIT

## Контакты

Для вопросов и предложений создавайте issue в репозитории.

# План тестирования Marimo ERP

## Быстрый тест всех сервисов

### Шаг 1: Запуск сервисов
```bash
# Вариант A: Docker Compose (рекомендуется)
docker-compose up --build

# Вариант B: Локально (для разработки)
# Терминал 1: Users Service
cd services/users && go run cmd/server/main.go

# Терминал 2: Gateway
cd services/gateway && go run cmd/server/main.go

# Терминал 3: Frontend
cd frontend && npm install && npm start
```

### Шаг 2: Health Check
```bash
# Проверка всех сервисов
curl http://localhost:8080/health

# Должен вернуть статус всех сервисов
```

### Шаг 3: Тест аутентификации
```bash
# Регистрация нового пользователя
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "test123",
    "name": "Test User"
  }'

# Вход (получите токен)
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }'

# Сохраните токен из ответа
TOKEN="ваш-jwt-токен"

# Тест защищенного endpoint
curl http://localhost:8080/api/users/profile \
  -H "Authorization: Bearer $TOKEN"
```

### Шаг 4: Тест модулей через браузер
1. Откройте http://localhost:3000
2. Войдите: admin@example.com / admin123
3. Проверьте каждый модуль:
   - ✅ Dashboard - статистика
   - ✅ Users - список пользователей
   - ✅ Config - настройки системы
   - ✅ Accounting - баланс и транзакции
   - ✅ Factory - продукты и заказы
   - ✅ Shop - каталог товаров

### Шаг 5: Тест каждого микросервиса

#### Config Service
```bash
curl http://localhost:8080/api/config \
  -H "Authorization: Bearer $TOKEN"
```

#### Shop Service (публичный)
```bash
curl http://localhost:8080/api/shop/products
```

#### Accounting Service
```bash
curl http://localhost:8080/api/accounting/balance \
  -H "Authorization: Bearer $TOKEN"
```

#### Factory Service
```bash
curl http://localhost:8080/api/factory/products \
  -H "Authorization: Bearer $TOKEN"
```

## Ожидаемые результаты
- ✅ Все сервисы отвечают без ошибок
- ✅ Аутентификация работает
- ✅ Protected routes требуют токен
- ✅ Frontend отображает все модули
- ✅ Навигация между страницами работает

# Quick Start с Docker

## 🚀 Самый быстрый способ запустить Marimo ERP

### Шаг 1: Убедитесь, что Docker установлен

```bash
docker --version
docker-compose --version
```

Если не установлен:
- **Windows/Mac**: [Docker Desktop](https://www.docker.com/products/docker-desktop)
- **Linux**:
  ```bash
  # Docker
  curl -fsSL https://get.docker.com -o get-docker.sh
  sh get-docker.sh

  # Docker Compose
  sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  sudo chmod +x /usr/local/bin/docker-compose
  ```

### Шаг 2: Запустите проект

```bash
# Простой способ (рекомендуется)
./start-docker.sh up

# Или используйте docker-compose напрямую
docker-compose up -d --build
```

### Шаг 3: Откройте браузер

**Frontend:** http://localhost:3000

**Учетные данные по умолчанию:**
- Email: `admin@example.com`
- Password: `admin123`

---

## 📋 Что запустится?

### Основные сервисы:
✅ **Frontend** (React) - http://localhost:3000
✅ **API Gateway** - http://localhost:8080
✅ **7 микросервисов** (Users, Config, Accounting, Factory, Shop, Main)

### Инфраструктура:
✅ **PostgreSQL** - база данных (порт 5432)
✅ **Redis** - кеширование (порт 6379)
✅ **Consul** - service discovery (UI: http://localhost:8500)
✅ **RabbitMQ** - message queue (UI: http://localhost:15672)

---

## 🛠 Полезные команды

```bash
# Просмотр логов всех сервисов
./start-docker.sh logs

# Просмотр логов конкретного сервиса
./start-docker.sh logs gateway

# Статус сервисов
./start-docker.sh ps

# Остановить все
./start-docker.sh down

# Перезапустить
./start-docker.sh restart

# Помощь
./start-docker.sh help
```

---

## 🎯 Что доступно после запуска?

| Сервис | URL | Описание |
|--------|-----|----------|
| **Frontend** | http://localhost:3000 | Веб-интерфейс |
| **API Gateway** | http://localhost:8080 | Основной API endpoint |
| **Consul UI** | http://localhost:8500 | Мониторинг сервисов |
| **RabbitMQ UI** | http://localhost:15672 | Очереди сообщений (admin/admin) |

### Микросервисы:
- **Users Service**: http://localhost:8081 - Аутентификация, пользователи
- **Config Service**: http://localhost:8082 - Конфигурация системы
- **Accounting**: http://localhost:8083 - Бухгалтерия
- **Factory**: http://localhost:8084 - Производство
- **Shop**: http://localhost:8085 - Интернет-магазин
- **Main Service**: http://localhost:8086 - Dashboard, статистика

---

## ⚙️ Конфигурация

Настройки находятся в файле `.env`. Он автоматически создается из `.env.example` при первом запуске.

**Важные параметры:**

```bash
# Режим работы БД (false = in-memory, true = PostgreSQL)
USE_POSTGRES=false

# JWT Secret (ОБЯЗАТЕЛЬНО измените в production!)
JWT_SECRET=marimo-dev-secret-key-change-this-in-production-32chars
```

---

## 🐛 Проблемы?

### Порт занят
```bash
# Проверить занятые порты
lsof -i :8080        # Linux/Mac
netstat -ano | findstr :8080  # Windows

# Остановить все контейнеры
./start-docker.sh down
```

### Сервис не запускается
```bash
# Проверить логи
./start-docker.sh logs [service-name]

# Пересобрать
docker-compose down
docker-compose up --build -d
```

### Полная очистка
```bash
# ОСТОРОЖНО! Удалит все контейнеры и данные
./start-docker.sh clean
```

---

## 📚 Дополнительная документация

- **[DOCKER.md](DOCKER.md)** - Полное руководство по Docker
- **[CLAUDE.md](CLAUDE.md)** - Инструкции для разработки
- **[README.md](README.md)** - Общая информация о проекте

---

## 🎉 Готово!

После запуска откройте http://localhost:3000 и войдите с учетными данными:
- Email: `admin@example.com`
- Password: `admin123`

**Наслаждайтесь работой с Marimo ERP!** 🚀

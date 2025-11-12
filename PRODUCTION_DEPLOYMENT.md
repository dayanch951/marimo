# Marimo ERP - Production Deployment Guide

Полное руководство по развертыванию Marimo ERP в production окружении.

## 📋 Содержание

1. [Требования](#требования)
2. [Быстрый старт](#быстрый-старт)
3. [Пошаговое развертывание](#пошаговое-развертывание)
4. [Безопасность](#безопасность)
5. [Мониторинг](#мониторинг)
6. [Резервное копирование](#резервное-копирование)
7. [Обслуживание](#обслуживание)
8. [Troubleshooting](#troubleshooting)

---

## Требования

### Минимальные системные требования

- **CPU:** 4 cores
- **RAM:** 8 GB
- **Disk:** 100 GB SSD
- **OS:** Ubuntu 20.04 LTS / Debian 11 / CentOS 8

### Рекомендуемые требования

- **CPU:** 8+ cores
- **RAM:** 16+ GB
- **Disk:** 200+ GB SSD (NVMe предпочтительно)
- **OS:** Ubuntu 22.04 LTS
- **Network:** 1 Gbps+

### Программное обеспечение

- Docker 24.0+
- Docker Compose 2.20+
- Git
- OpenSSL
- (Опционально) Certbot для Let's Encrypt

---

## Быстрый старт

### Автоматическая установка

```bash
# Клонировать репозиторий
git clone https://github.com/dayanch951/marimo.git
cd marimo

# Запустить production setup
sudo ./scripts/setup-production.sh
```

Скрипт автоматически:
- Проверит prerequisites
- Создаст .env файл
- Сгенерирует безопасные пароли
- Настроит SSL (опционально)
- Запустит все сервисы

---

## Пошаговое развертывание

### Шаг 1: Подготовка сервера

#### 1.1 Обновление системы

```bash
sudo apt update && sudo apt upgrade -y
```

#### 1.2 Установка Docker

```bash
# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Добавить пользователя в группу docker
sudo usermod -aG docker $USER
newgrp docker

# Проверить установку
docker --version
```

#### 1.3 Установка Docker Compose

```bash
# Установить Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Проверить установку
docker-compose --version
```

### Шаг 2: Клонирование репозитория

```bash
cd /opt
git clone https://github.com/dayanch951/marimo.git
cd marimo
```

### Шаг 3: Конфигурация окружения

#### 3.1 Создание .env файла

```bash
cp .env.production .env
nano .env
```

#### 3.2 Критически важные параметры для изменения

```bash
# JWT Secret (ОБЯЗАТЕЛЬНО!)
JWT_SECRET=$(openssl rand -base64 48)

# Database
DB_PASSWORD=$(openssl rand -base64 32)
DB_NAME=marimo_prod
DB_USER=marimo_user

# Redis
REDIS_PASSWORD=$(openssl rand -base64 24)

# RabbitMQ
RABBITMQ_PASS=$(openssl rand -base64 24)

# Домен
CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

#### 3.3 Генерация всех секретов автоматически

```bash
# Запустить скрипт генерации
./scripts/generate-secrets.sh
```

### Шаг 4: SSL/TLS Сертификаты

#### Вариант A: Let's Encrypt (Рекомендуется)

```bash
# Установить certbot
sudo apt install certbot python3-certbot-nginx -y

# Получить сертификат
sudo certbot certonly --standalone \
  -d yourdomain.com \
  -d www.yourdomain.com \
  -d api.yourdomain.com \
  --email admin@yourdomain.com \
  --agree-tos \
  --non-interactive

# Скопировать сертификаты
sudo mkdir -p ./ssl
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ./ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ./ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/chain.pem ./ssl/
sudo chmod 600 ./ssl/privkey.pem

# Автопродление
sudo systemctl enable certbot.timer
```

#### Вариант B: Собственный сертификат

```bash
./scripts/generate-ssl-certs.sh yourdomain.com
```

### Шаг 5: Настройка DNS

Добавьте A-записи в вашем DNS провайдере:

```
yourdomain.com          A    YOUR_SERVER_IP
www.yourdomain.com      A    YOUR_SERVER_IP
api.yourdomain.com      A    YOUR_SERVER_IP
```

Проверка:
```bash
dig yourdomain.com +short
```

### Шаг 6: Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# Или iptables
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
sudo iptables -A INPUT -j DROP
```

### Шаг 7: Запуск сервисов

```bash
# Сборка образов
docker-compose -f docker-compose.yml -f docker-compose.production.yml build

# Запуск в фоновом режиме
docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d

# Проверка статуса
docker-compose ps
```

### Шаг 8: Инициализация базы данных

```bash
# Запустить миграции
docker-compose exec users ./scripts/run-migrations.sh

# Создать admin пользователя (если не существует)
docker-compose exec postgres psql -U marimo_user -d marimo_prod << EOF
INSERT INTO users (email, password, role, created_at)
VALUES ('admin@yourdomain.com', 'CHANGE_THIS_PASSWORD', 'admin', NOW())
ON CONFLICT DO NOTHING;
EOF
```

### Шаг 9: Проверка работоспособности

```bash
# Проверить логи
docker-compose logs -f

# Проверить health checks
curl https://api.yourdomain.com/health
curl https://yourdomain.com

# Проверить Consul
curl http://localhost:8500/v1/catalog/services
```

---

## Безопасность

### Checklist безопасности

- [ ] Изменены все пароли по умолчанию
- [ ] JWT_SECRET установлен (64+ символов)
- [ ] SSL сертификаты настроены
- [ ] Firewall настроен
- [ ] HSTS включен
- [ ] Rate limiting включен
- [ ] Database SSL включен (`DB_SSL_MODE=require`)
- [ ] Redis защищен паролем
- [ ] RabbitMQ защищен паролем
- [ ] Consul ACL включен
- [ ] Backup encryption настроен
- [ ] Audit logging включен
- [ ] Fail2ban настроен (опционально)

### Дополнительные меры безопасности

#### 1. Настройка Fail2ban

```bash
# Установить fail2ban
sudo apt install fail2ban -y

# Создать jail для nginx
sudo cat > /etc/fail2ban/jail.d/nginx.conf << EOF
[nginx-limit-req]
enabled = true
filter = nginx-limit-req
action = iptables-multiport[name=ReqLimit, port="http,https"]
logpath = /var/log/nginx/error.log
findtime = 600
bantime = 7200
maxretry = 10
EOF

# Перезапустить fail2ban
sudo systemctl restart fail2ban
```

#### 2. Настройка 2FA (опционально)

```bash
# В .env добавить
TWO_FACTOR_AUTH_ENABLED=true
```

#### 3. IP Whitelisting для admin панели

В `nginx/sites-enabled/marimo.conf`:

```nginx
location /admin {
    allow 1.2.3.4;  # Ваш офисный IP
    deny all;
    proxy_pass http://api_gateway;
}
```

---

## Мониторинг

### Prometheus + Grafana

```bash
# Запустить monitoring stack
docker-compose -f docker-compose.yml \
               -f docker-compose.production.yml \
               -f docker-compose.monitoring.yml up -d

# Доступ к Grafana
open http://localhost:3001
# Логин: admin / admin (изменить!)
```

### Метрики для отслеживания

- **CPU Usage:** <80%
- **Memory Usage:** <85%
- **Disk Usage:** <80%
- **Response Time:** <500ms (p95)
- **Error Rate:** <1%
- **Database Connections:** <80% max
- **Cache Hit Rate:** >80%

### Alerting

Настроить алерты в Prometheus (`prometheus/alerts.yml`):

```yaml
groups:
- name: critical
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
    for: 5m
    annotations:
      summary: "High error rate detected"
```

### Логирование

```bash
# Централизованные логи
docker-compose logs -f --tail=100

# Логи конкретного сервиса
docker-compose logs -f gateway

# Поиск ошибок
docker-compose logs | grep ERROR

# Экспорт логов
docker-compose logs > logs/app-$(date +%Y%m%d).log
```

---

## Резервное копирование

### Автоматическое backup

#### 1. Настройка cron

```bash
# Добавить в crontab
crontab -e

# Backup каждый день в 2:00 AM
0 2 * * * /opt/marimo/scripts/backup.sh
```

#### 2. Backup скрипт

Уже включен в проект: `./scripts/backup.sh`

```bash
# Ручной backup
./scripts/backup.sh

# Backup с кастомным именем
./scripts/backup.sh production-backup-$(date +%Y%m%d)
```

#### 3. Восстановление из backup

```bash
./scripts/restore.sh path/to/backup-file.sql
```

### Backup в S3

```bash
# В .env добавить
BACKUP_S3_BUCKET=marimo-prod-backups
AWS_ACCESS_KEY_ID=YOUR_KEY
AWS_SECRET_ACCESS_KEY=YOUR_SECRET
AWS_REGION=us-east-1

# Запустить backup с S3 upload
./scripts/backup-to-s3.sh
```

### Disaster Recovery

Полное руководство: `./scripts/disaster-recovery.sh`

---

## Обслуживание

### Обновление приложения

```bash
# 1. Создать backup
./scripts/backup.sh

# 2. Получить последнюю версию
git pull origin main

# 3. Пересобрать образы
docker-compose -f docker-compose.yml -f docker-compose.production.yml build

# 4. Перезапустить сервисы (rolling update)
docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d

# 5. Проверить логи
docker-compose logs -f

# 6. Проверить health
curl https://api.yourdomain.com/health
```

### Zero-downtime Deployment (Blue-Green)

```bash
./scripts/blue-green-deploy.sh
```

### Масштабирование

```bash
# Масштабировать Gateway (2 → 4 replicas)
docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d --scale gateway=4

# Масштабировать Users service
docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d --scale users=3
```

### Мониторинг ресурсов

```bash
# Использование ресурсов контейнерами
docker stats

# Disk usage
docker system df

# Очистка неиспользуемых ресурсов
docker system prune -a
```

### Обновление зависимостей

```bash
# Go зависимости
./scripts/update-dependencies.sh

# Frontend зависимости
cd frontend && npm update && npm audit fix
```

---

## Troubleshooting

### Проблема: Сервис не запускается

```bash
# Проверить логи
docker-compose logs service_name

# Проверить конфигурацию
docker-compose config

# Проверить healthcheck
docker inspect service_name | grep -A 10 Health
```

### Проблема: База данных недоступна

```bash
# Проверить статус PostgreSQL
docker-compose exec postgres pg_isready

# Проверить подключение
docker-compose exec postgres psql -U marimo_user -d marimo_prod -c "SELECT 1"

# Проверить логи
docker-compose logs postgres
```

### Проблема: High memory usage

```bash
# Проверить использование памяти
docker stats

# Перезапустить сервис
docker-compose restart service_name

# Очистить Redis cache
docker-compose exec redis redis-cli FLUSHALL
```

### Проблема: SSL сертификат expired

```bash
# Обновить Let's Encrypt сертификат
sudo certbot renew

# Скопировать новые сертификаты
sudo cp /etc/letsencrypt/live/yourdomain.com/*.pem ./ssl/

# Перезапустить nginx
docker-compose restart nginx
```

### Проблема: Медленная работа

```bash
# 1. Проверить метрики
curl http://localhost:9090/metrics

# 2. Проверить database slow queries
docker-compose exec postgres psql -U marimo_user -d marimo_prod << EOF
SELECT query, calls, mean_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
EOF

# 3. Проверить Redis cache hit rate
docker-compose exec redis redis-cli INFO stats

# 4. Включить query caching
# В .env: CACHE_ENABLED=true
```

---

## Production Checklist

### Перед запуском

- [ ] Все `CHANGE_ME` значения заменены
- [ ] SSL сертификаты настроены
- [ ] DNS правильно настроен
- [ ] Firewall настроен
- [ ] Backup настроен
- [ ] Мониторинг настроен
- [ ] Логирование настроено
- [ ] Rate limiting настроен
- [ ] CORS правильно настроен
- [ ] Email сервис настроен (SendGrid)
- [ ] Payment gateway настроен (Stripe)
- [ ] Admin пароль изменен

### После запуска

- [ ] Проверены health checks всех сервисов
- [ ] Проверены логи на ошибки
- [ ] Проверена работа frontend
- [ ] Проверена работа API
- [ ] Создан тестовый заказ
- [ ] Проверена отправка email
- [ ] Проверена работа платежей
- [ ] Настроены алерты
- [ ] Создан первый backup
- [ ] Документирован процесс восстановления

---

## Дополнительные ресурсы

- [Docker Documentation](https://docs.docker.com/)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [PostgreSQL Best Practices](https://wiki.postgresql.org/wiki/Don%27t_Do_This)
- [Nginx Performance Tuning](https://www.nginx.com/blog/tuning-nginx/)
- [Redis Best Practices](https://redis.io/topics/admin)

---

## Поддержка

Если возникли проблемы:

1. Проверьте логи: `docker-compose logs -f`
2. Проверьте документацию: [DOCKER.md](DOCKER.md), [CLAUDE.md](CLAUDE.md)
3. Создайте issue: https://github.com/dayanch951/marimo/issues

---

**Последнее обновление:** 2025-11-12
**Версия:** 1.0.0

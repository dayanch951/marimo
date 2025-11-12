# Quick Start Guide

Get Marimo ERP up and running in minutes.

## Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- PostgreSQL 16
- (Optional) Kubernetes cluster

## 1. Clone and Setup

```bash
# Clone repository
git clone https://github.com/your-org/marimo.git
cd marimo

# Copy environment file
cp .env.example .env

# Edit .env with your configuration
nano .env
```

## 2. Start Services

### Option A: Docker Compose (Development)

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### Option B: Local Development

```bash
# Start infrastructure
docker-compose up -d postgres redis consul rabbitmq

# Run migrations
./scripts/run-migrations.sh

# Start API Gateway
cd services/api-gateway
go run main.go

# Start Auth Service (in another terminal)
cd services/auth-service
go run main.go

# Start Frontend
cd frontend
npm install
npm run dev
```

## 3. Verify Installation

```bash
# Check API Gateway health
curl http://localhost:8080/health

# Check Auth Service
curl http://localhost:8081/health

# Access frontend
open http://localhost:3000
```

## 4. Create First Tenant

```bash
# Using API
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "slug": "acme",
    "email": "admin@acme.com"
  }'
```

## 5. Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Slug: acme" \
  -d '{
    "email": "user@acme.com",
    "password": "SecurePass123!",
    "name": "John Doe"
  }'
```

## 6. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Slug: acme" \
  -d '{
    "email": "user@acme.com",
    "password": "SecurePass123!"
  }'

# Response includes JWT token
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": { ... }
}
```

## 7. Make Authenticated Request

```bash
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "X-Tenant-Slug: acme"
```

## Next Steps

### Configure Integrations

**Stripe** (for payments):
```bash
# Add to .env
STRIPE_API_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

**SendGrid** (for emails):
```bash
# Add to .env
SENDGRID_API_KEY=SG....
SENDGRID_FROM_EMAIL=noreply@yourapp.com
```

### Setup Webhooks

```bash
# Create webhook
curl -X POST http://localhost:8080/api/v1/webhooks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/marimo",
    "events": ["user.created", "payment.succeeded"],
    "secret": "your-webhook-secret"
  }'
```

### Configure Analytics

```bash
# Create custom query
curl -X POST http://localhost:8080/api/v1/analytics/queries \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Monthly Revenue",
    "source": "transactions",
    "metrics": [
      {"name": "total", "type": "sum", "field": "amount"}
    ],
    "dimensions": [
      {"name": "month", "field": "DATE_TRUNC('\'month'\'', created_at)"}
    ]
  }'
```

### Mobile App Setup

```bash
cd mobile

# Install dependencies
npm install

# iOS
cd ios && pod install && cd ..
npm run ios

# Android
npm run android
```

## Troubleshooting

**Services won't start:**
```bash
# Clean and restart
docker-compose down -v
docker-compose up -d
```

**Database connection error:**
```bash
# Check PostgreSQL
docker-compose logs postgres

# Verify connection
psql -h localhost -U postgres -d marimo_erp
```

**Port conflicts:**
```bash
# Check what's using port
lsof -i :8080

# Change port in .env
API_GATEWAY_PORT=8081
```

## Production Deployment

For production deployment, see:
- [Kubernetes Deployment](./DEPLOYMENT.md)
- [CI/CD Setup](../.github/workflows/README.md)
- [Disaster Recovery](./DISASTER_RECOVERY.md)

## Support

- Documentation: `docs/`
- Issues: GitHub Issues
- Community: Discord/Slack

# Marimo ERP API Documentation

Comprehensive guide for integrating with Marimo ERP APIs.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [Multi-tenancy](#multi-tenancy)
4. [Core APIs](#core-apis)
5. [Analytics APIs](#analytics-apis)
6. [Webhook APIs](#webhook-apis)
7. [Integration APIs](#integration-apis)
8. [Error Handling](#error-handling)
9. [Rate Limiting](#rate-limiting)
10. [SDKs and Libraries](#sdks-and-libraries)

## Getting Started

### Base URLs

- **Development**: `http://localhost:8080/api`
- **Staging**: `https://staging-api.marimo-erp.com`
- **Production**: `https://api.marimo-erp.com`

### API Versioning

Currently on version 1. All endpoints are prefixed with `/api`.

Future versions will use `/api/v2`, `/api/v3`, etc.

### OpenAPI Specification

Full OpenAPI 3.0 specification available at:
- **YAML**: [docs/openapi.yaml](./openapi.yaml)
- **Swagger UI**: `https://api.marimo-erp.com/docs`
- **ReDoc**: `https://api.marimo-erp.com/redoc`

## Authentication

### Overview

Marimo ERP uses **JWT (JSON Web Tokens)** for authentication.

### Login Flow

#### 1. Login

**Endpoint**: `POST /auth/login`

**Request**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "admin",
    "tenant_id": "123e4567-e89b-12d3-a456-426614174001"
  }
}
```

**cURL Example**:
```bash
curl -X POST https://api.marimo-erp.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

#### 2. Use Access Token

Include the token in the `Authorization` header:

```bash
curl https://api.marimo-erp.com/users \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 3. Refresh Token

When the access token expires (24 hours), use the refresh token to get a new one.

**Endpoint**: `POST /auth/refresh`

**Request**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400
}
```

### Registration

**Endpoint**: `POST /auth/register`

**Request**:
```json
{
  "email": "newuser@example.com",
  "password": "SecurePass123!",
  "name": "Jane Smith",
  "company": "Acme Corp"
}
```

**Response** (201 Created):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174002",
    "email": "newuser@example.com",
    "name": "Jane Smith",
    "role": "admin",
    "tenant_id": "123e4567-e89b-12d3-a456-426614174003"
  }
}
```

**Note**: Registration automatically creates a new tenant with a 14-day trial.

### Logout

**Endpoint**: `POST /auth/logout`

**Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

## Multi-tenancy

### Tenant Identification

Specify the tenant in one of three ways (in priority order):

#### 1. X-Tenant-ID Header (Recommended)

```bash
curl https://api.marimo-erp.com/users \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: 123e4567-e89b-12d3-a456-426614174001"
```

#### 2. X-Tenant-Slug Header

```bash
curl https://api.marimo-erp.com/users \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-Slug: acme-corp"
```

#### 3. Subdomain (for web apps)

```bash
curl https://acme-corp.marimo-erp.com/api/users \
  -H "Authorization: Bearer <token>"
```

### Tenant Management

#### Get Current Tenant

**Endpoint**: `GET /tenants/current`

**Response** (200 OK):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001",
  "name": "Acme Corp",
  "slug": "acme-corp",
  "domain": "erp.acme.com",
  "status": "active",
  "settings": {
    "max_users": 25,
    "max_storage": 10737418240,
    "allowed_features": ["analytics", "webhooks", "integrations"],
    "timezone": "UTC",
    "currency": "USD"
  },
  "subscription": {
    "plan": "professional",
    "status": "active",
    "current_period_start": "2024-01-01T00:00:00Z",
    "current_period_end": "2024-02-01T00:00:00Z"
  },
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### Update Tenant Settings

**Endpoint**: `PUT /tenants/current/settings`

**Request**:
```json
{
  "timezone": "America/New_York",
  "currency": "EUR"
}
```

#### Upgrade Subscription

**Endpoint**: `PUT /tenants/current/subscription`

**Request**:
```json
{
  "plan": "enterprise"
}
```

**Response** (200 OK):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001",
  "subscription": {
    "plan": "enterprise",
    "status": "active",
    "current_period_start": "2024-01-15T00:00:00Z",
    "current_period_end": "2024-02-15T00:00:00Z"
  }
}
```

## Core APIs

### Users API

#### List Users

**Endpoint**: `GET /users`

**Query Parameters**:
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 20, max: 100)
- `role` (string): Filter by role (admin, user, viewer)

**Response** (200 OK):
```json
{
  "users": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "admin",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "pages": 3
  }
}
```

#### Get User

**Endpoint**: `GET /users/{id}`

**Response** (200 OK):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "admin",
  "tenant_id": "123e4567-e89b-12d3-a456-426614174001",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Create User

**Endpoint**: `POST /users`

**Request**:
```json
{
  "email": "newuser@example.com",
  "name": "Jane Smith",
  "role": "user",
  "password": "SecurePass123!"
}
```

**Response** (201 Created):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174002",
  "email": "newuser@example.com",
  "name": "Jane Smith",
  "role": "user",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Update User

**Endpoint**: `PUT /users/{id}`

**Request**:
```json
{
  "name": "Jane Doe",
  "role": "admin"
}
```

**Response** (200 OK):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174002",
  "email": "newuser@example.com",
  "name": "Jane Doe",
  "role": "admin",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

#### Delete User

**Endpoint**: `DELETE /users/{id}`

**Response** (204 No Content)

## Analytics APIs

### Analytics Queries

#### Create Query

**Endpoint**: `POST /analytics/queries`

**Request**:
```json
{
  "name": "Revenue by Month",
  "source": "payments",
  "metrics": [
    {
      "name": "total_revenue",
      "type": "sum",
      "field": "amount"
    },
    {
      "name": "transaction_count",
      "type": "count",
      "field": "*"
    }
  ],
  "dimensions": [
    {
      "name": "month",
      "field": "DATE_TRUNC('month', created_at)"
    }
  ],
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  }
}
```

**Response** (201 Created):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174010",
  "name": "Revenue by Month",
  "source": "payments",
  "metrics": [...],
  "dimensions": [...],
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Execute Query

**Endpoint**: `POST /analytics/queries/{id}/execute`

**Response** (200 OK):
```json
{
  "columns": ["month", "total_revenue", "transaction_count"],
  "rows": [
    {
      "month": "2024-01",
      "total_revenue": 125000,
      "transaction_count": 342
    },
    {
      "month": "2024-02",
      "total_revenue": 138500,
      "transaction_count": 389
    }
  ],
  "summary": {
    "total_revenue": 263500,
    "transaction_count": 731
  },
  "executed_at": "2024-01-15T10:35:00Z"
}
```

### Dashboards

#### Create Dashboard

**Endpoint**: `POST /analytics/dashboards`

**Request**:
```json
{
  "name": "Executive Dashboard",
  "description": "Key metrics for leadership",
  "widgets": [
    {
      "id": "widget-1",
      "type": "metric",
      "title": "Total Revenue",
      "query_id": "123e4567-e89b-12d3-a456-426614174010",
      "config": {
        "metric": "total_revenue",
        "format": "currency"
      }
    },
    {
      "id": "widget-2",
      "type": "chart",
      "title": "Revenue Trend",
      "query_id": "123e4567-e89b-12d3-a456-426614174010",
      "config": {
        "chart_type": "line",
        "x_axis": "month",
        "y_axis": "total_revenue"
      }
    }
  ]
}
```

#### Get Dashboard

**Endpoint**: `GET /analytics/dashboards/{id}`

**Response** (200 OK):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174011",
  "name": "Executive Dashboard",
  "description": "Key metrics for leadership",
  "widgets": [...],
  "created_at": "2024-01-15T10:40:00Z"
}
```

## Webhook APIs

### Webhook Management

#### Create Webhook

**Endpoint**: `POST /webhooks`

**Request**:
```json
{
  "url": "https://your-app.com/webhooks/marimo",
  "events": [
    "user.created",
    "user.updated",
    "payment.succeeded",
    "payment.failed"
  ],
  "headers": {
    "X-Custom-Header": "value"
  }
}
```

**Response** (201 Created):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174020",
  "url": "https://your-app.com/webhooks/marimo",
  "secret": "whsec_abcdef123456",
  "events": ["user.created", "user.updated", "payment.succeeded", "payment.failed"],
  "status": "active",
  "created_at": "2024-01-15T10:45:00Z"
}
```

**Important**: Save the `secret` - you'll need it to verify webhook signatures.

#### List Webhooks

**Endpoint**: `GET /webhooks`

**Response** (200 OK):
```json
{
  "webhooks": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174020",
      "url": "https://your-app.com/webhooks/marimo",
      "events": ["user.created", "user.updated"],
      "status": "active",
      "created_at": "2024-01-15T10:45:00Z"
    }
  ]
}
```

#### Delete Webhook

**Endpoint**: `DELETE /webhooks/{id}`

**Response** (204 No Content)

### Webhook Deliveries

#### Get Delivery History

**Endpoint**: `GET /webhooks/{id}/deliveries`

**Query Parameters**:
- `status` (string): Filter by status (pending, success, failed)

**Response** (200 OK):
```json
{
  "deliveries": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174021",
      "webhook_id": "123e4567-e89b-12d3-a456-426614174020",
      "event": "user.created",
      "status": "success",
      "response_code": 200,
      "attempts": 1,
      "created_at": "2024-01-15T10:50:00Z"
    },
    {
      "id": "123e4567-e89b-12d3-a456-426614174022",
      "webhook_id": "123e4567-e89b-12d3-a456-426614174020",
      "event": "payment.succeeded",
      "status": "failed",
      "response_code": 500,
      "attempts": 3,
      "next_retry_at": "2024-01-15T11:50:00Z",
      "created_at": "2024-01-15T10:55:00Z"
    }
  ]
}
```

### Webhook Payload Format

All webhooks are sent as POST requests with this structure:

```json
{
  "id": "evt_123e4567-e89b-12d3-a456-426614174030",
  "type": "user.created",
  "timestamp": "2024-01-15T10:50:00Z",
  "tenant_id": "123e4567-e89b-12d3-a456-426614174001",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "newuser@example.com",
    "name": "John Doe",
    "role": "user"
  }
}
```

### Webhook Signature Verification

Each webhook includes a `X-Marimo-Signature` header for verification:

**Node.js Example**:
```javascript
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  const digest = hmac.update(payload).digest('hex');
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(digest)
  );
}

// Express.js middleware
app.post('/webhooks/marimo', express.raw({type: 'application/json'}), (req, res) => {
  const signature = req.headers['x-marimo-signature'];
  const secret = process.env.WEBHOOK_SECRET;

  if (!verifyWebhookSignature(req.body, signature, secret)) {
    return res.status(401).send('Invalid signature');
  }

  // Process webhook
  const event = JSON.parse(req.body);
  console.log('Received event:', event.type);

  res.status(200).send('OK');
});
```

**Go Example**:
```go
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

## Integration APIs

### Stripe Integration

#### Create Payment

**Endpoint**: `POST /integrations/stripe/payment`

**Request**:
```json
{
  "amount": 9999,
  "currency": "usd",
  "description": "Professional plan subscription",
  "customer_id": "cus_abc123",
  "metadata": {
    "tenant_id": "123e4567-e89b-12d3-a456-426614174001",
    "subscription_plan": "professional"
  }
}
```

**Response** (200 OK):
```json
{
  "client_secret": "pi_abc123_secret_xyz",
  "payment_intent_id": "pi_abc123"
}
```

#### Create Subscription

**Endpoint**: `POST /integrations/stripe/subscription`

**Request**:
```json
{
  "customer_id": "cus_abc123",
  "price_id": "price_professional_monthly",
  "trial_days": 14
}
```

**Response** (200 OK):
```json
{
  "subscription_id": "sub_abc123",
  "status": "active"
}
```

### SendGrid Integration

#### Send Email

**Endpoint**: `POST /integrations/sendgrid/email`

**Request**:
```json
{
  "to": ["user@example.com"],
  "subject": "Welcome to Marimo ERP",
  "html": "<h1>Welcome!</h1><p>Thank you for signing up.</p>",
  "text": "Welcome! Thank you for signing up.",
  "from": "noreply@marimo-erp.com"
}
```

**Response** (200 OK):
```json
{
  "message_id": "msg_abc123"
}
```

#### Send Template Email

**Endpoint**: `POST /integrations/sendgrid/template`

**Request**:
```json
{
  "to": ["user@example.com"],
  "template_id": "d-abc123",
  "dynamic_data": {
    "user_name": "John Doe",
    "company_name": "Acme Corp"
  }
}
```

## Error Handling

### Error Response Format

All errors follow this structure:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "specific error details"
  }
}
```

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request succeeded |
| 201 | Created | Resource created successfully |
| 204 | No Content | Request succeeded, no content to return |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict (e.g., duplicate) |
| 422 | Unprocessable Entity | Validation errors |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

### Common Error Codes

| Code | Description |
|------|-------------|
| `UNAUTHORIZED` | Missing or invalid authentication token |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `VALIDATION_ERROR` | Request validation failed |
| `DUPLICATE_RESOURCE` | Resource already exists |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `TENANT_NOT_FOUND` | Tenant not found or invalid |
| `SUBSCRIPTION_EXPIRED` | Tenant subscription expired |
| `FEATURE_NOT_AVAILABLE` | Feature not available in current plan |

### Error Examples

**401 Unauthorized**:
```json
{
  "error": "Invalid or expired token",
  "code": "UNAUTHORIZED"
}
```

**422 Validation Error**:
```json
{
  "error": "Validation failed",
  "code": "VALIDATION_ERROR",
  "details": {
    "email": "Invalid email format",
    "password": "Password must be at least 8 characters"
  }
}
```

**429 Rate Limit**:
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "details": {
    "retry_after": 60
  }
}
```

## Rate Limiting

### Limits

| Plan | Requests per minute | Burst |
|------|---------------------|-------|
| Trial | 60 | 100 |
| Starter | 120 | 200 |
| Professional | 300 | 500 |
| Enterprise | 1000 | 2000 |

### Rate Limit Headers

Every response includes rate limit information:

```
X-RateLimit-Limit: 300
X-RateLimit-Remaining: 285
X-RateLimit-Reset: 1642258800
```

### Handling Rate Limits

When rate limited (429), wait for the time specified in `Retry-After` header:

```javascript
async function apiCall() {
  const response = await fetch('https://api.marimo-erp.com/users');

  if (response.status === 429) {
    const retryAfter = parseInt(response.headers.get('Retry-After'));
    await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
    return apiCall(); // Retry
  }

  return response.json();
}
```

## SDKs and Libraries

### Official SDKs

#### JavaScript/TypeScript

```bash
npm install @marimo-erp/sdk
```

```typescript
import { MarimoClient } from '@marimo-erp/sdk';

const client = new MarimoClient({
  apiKey: 'your-api-key',
  tenantId: 'your-tenant-id'
});

// List users
const users = await client.users.list({ limit: 20 });

// Create user
const user = await client.users.create({
  email: 'user@example.com',
  name: 'John Doe',
  role: 'user'
});
```

#### Go

```bash
go get github.com/marimo-erp/go-sdk
```

```go
import "github.com/marimo-erp/go-sdk"

client := marimo.NewClient("your-api-key")
client.SetTenantID("your-tenant-id")

// List users
users, err := client.Users.List(&marimo.UserListParams{
    Limit: 20,
})

// Create user
user, err := client.Users.Create(&marimo.UserCreateParams{
    Email: "user@example.com",
    Name:  "John Doe",
    Role:  "user",
})
```

#### Python

```bash
pip install marimo-sdk
```

```python
from marimo import MarimoClient

client = MarimoClient(
    api_key='your-api-key',
    tenant_id='your-tenant-id'
)

# List users
users = client.users.list(limit=20)

# Create user
user = client.users.create(
    email='user@example.com',
    name='John Doe',
    role='user'
)
```

### Community SDKs

- **Ruby**: [marimo-ruby](https://github.com/marimo-erp/marimo-ruby)
- **PHP**: [marimo-php](https://github.com/marimo-erp/marimo-php)
- **.NET**: [Marimo.SDK](https://github.com/marimo-erp/marimo-dotnet)

## Additional Resources

- [OpenAPI Specification](./openapi.yaml)
- [Postman Collection](./postman_collection.json)
- [Webhook Events Reference](./WEBHOOKS.md)
- [Integration Guides](./ADVANCED_FEATURES.md)
- [Developer Onboarding](./DEVELOPER_ONBOARDING.md)

## Support

- **Documentation**: https://docs.marimo-erp.com
- **API Status**: https://status.marimo-erp.com
- **Support Email**: support@marimo-erp.com
- **Slack Community**: https://slack.marimo-erp.com
- **GitHub Issues**: https://github.com/marimo-erp/marimo/issues

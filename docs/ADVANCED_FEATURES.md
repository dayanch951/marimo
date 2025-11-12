# Advanced Features Documentation

Comprehensive guide for advanced features including Multi-tenancy, Analytics, Integrations, Webhooks, and Mobile App.

## Table of Contents

- [Multi-Tenancy](#multi-tenancy)
- [Analytics & Reporting](#analytics--reporting)
- [Third-Party Integrations](#third-party-integrations)
- [API Webhooks](#api-webhooks)
- [Mobile Application](#mobile-application)

---

## Multi-Tenancy

Multi-tenancy allows multiple organizations to use the same application instance while keeping their data completely isolated.

### Architecture

**Data Isolation Strategy**: Shared schema with `tenant_id` column

Each table includes:
- `tenant_id` UUID - References the tenant
- Row-Level Security enforcement
- Automatic filtering in all queries

### Tenant Model

```go
type Tenant struct {
    ID             uuid.UUID
    Name           string
    Slug           string      // Used in subdomain
    Domain         *string     // Custom domain support
    Status         TenantStatus // active, inactive, suspended, trial
    Settings       Settings
    Subscription   Subscription
}
```

### Tenant Resolution

Priority order:
1. `X-Tenant-ID` header (API requests)
2. `X-Tenant-Slug` header
3. Custom domain (e.g., `company.com`)
4. Subdomain (e.g., `company.marimo-erp.com`)

### Usage Example

**Creating a Tenant:**

```go
service := tenancy.NewTenantService(repo, db)

tenant, err := service.CreateTenant(ctx, "Acme Corp", "acme")
// Creates tenant with default trial settings
```

**Middleware Integration:**

```go
resolver := tenancy.NewTenantResolver(repo)
middleware := tenancy.NewTenantMiddleware(resolver)

router.Use(middleware.Middleware)
```

**Accessing Tenant in Request:**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    tenant, err := tenancy.ResolveFromContext(r.Context())
    if err != nil {
        // Handle error
    }

    // Use tenant.ID for database queries
}
```

### Subscription Plans

| Plan | Max Users | Storage | Features |
|------|-----------|---------|----------|
| Trial | 10 | 10 GB | Basic |
| Starter | 25 | 50 GB | Basic, Search, Export |
| Professional | 100 | 250 GB | + Analytics, Webhooks |
| Enterprise | Unlimited | Unlimited | + Custom Domain, SSO |

### Best Practices

1. **Always filter by tenant_id** in database queries
2. **Validate tenant status** before processing requests
3. **Use TenantAwareDB** wrapper for automatic filtering
4. **Test tenant isolation** thoroughly
5. **Monitor cross-tenant queries** (should never happen)

---

## Analytics & Reporting

Advanced analytics engine with custom queries, dashboards, and automated reports.

### Analytics Engine

**Query Structure:**

```go
query := &analytics.Query{
    Source: "users",
    Metrics: []analytics.Metric{
        {Name: "total", Type: MetricTypeCount, Field: "*"},
        {Name: "avg_age", Type: MetricTypeAverage, Field: "age"},
    },
    Dimensions: []analytics.Dimension{
        {Name: "country", Field: "country"},
    },
    TimeRange: &analytics.TimeRange{
        Start: time.Now().AddDate(0, -1, 0),
        End:   time.Now(),
    },
    GroupBy: []string{"country"},
}

result, err := engine.Execute(ctx, query)
```

### Metric Types

- **Count**: Count rows
- **Sum**: Sum values
- **Average**: Calculate average
- **Min/Max**: Find extremes
- **Percentage**: Calculate percentages

### Pre-built Reports

**User Activity Report:**

```go
builder := analytics.NewReportBuilder(engine)

report, err := builder.BuildUserActivityReport(ctx, tenantID, timeRange)
// Returns: total users, actions, session duration by date and action type
```

**Revenue Report:**

```go
report, err := builder.BuildRevenueReport(ctx, tenantID, timeRange)
// Returns: total revenue, transaction count, averages by date and payment method
```

**Usage Report:**

```go
report, err := builder.BuildUsageReport(ctx, tenantID, timeRange)
// Returns: API calls, storage used, active users over time
```

**Performance Report:**

```go
report, err := builder.BuildPerformanceReport(ctx, tenantID, timeRange)
// Returns: response times, error rates by endpoint
```

### Dashboards

**Creating Custom Dashboard:**

```go
dashboard := &analytics.Dashboard{
    ID:       uuid.New(),
    TenantID: tenantID,
    Name:     "Sales Dashboard",
    Widgets: []analytics.Widget{
        {
            ID:    "revenue-chart",
            Type:  WidgetTypeChart,
            Title: "Monthly Revenue",
            Query: revenueQuery,
            Visualization: "line",
        },
    },
}
```

**Widget Types:**

- **Metric**: Single number display
- **Chart**: Line, bar, pie charts
- **Table**: Tabular data display
- **Custom**: Custom visualizations

### Scheduled Reports

```go
report := &analytics.Report{
    Type:       ReportTypeRevenue,
    Schedule:   ScheduleWeekly,
    Recipients: []string{"manager@company.com"},
    Format:     "pdf",
}
```

---

## Third-Party Integrations

### Stripe Integration

**Payment Processing:**

```go
stripe := integrations.NewStripeClient(config)

// Create customer
customer, err := stripe.CreateCustomer(ctx, integrations.CustomerCreateParams{
    Email: "customer@example.com",
    Name:  "John Doe",
})

// Create payment intent
intent, err := stripe.CreatePaymentIntent(ctx, integrations.PaymentIntentCreateParams{
    Amount:     10000, // $100.00 in cents
    Currency:   "usd",
    CustomerID: customer.ID,
})

// Create subscription
subscription, err := stripe.CreateSubscription(ctx, integrations.SubscriptionCreateParams{
    CustomerID: customer.ID,
    PriceID:    "price_123",
    TrialDays:  14,
})
```

**Webhook Handling:**

```go
event, err := stripe.VerifyWebhookSignature(payload, signature)
if err != nil {
    // Invalid signature
}

err = stripe.HandleWebhook(ctx, event)
// Processes: payment_intent.succeeded, invoice.paid, subscription events, etc.
```

### SendGrid Integration

**Sending Emails:**

```go
sendgrid := integrations.NewSendGridClient(config)

// Simple email
response, err := sendgrid.SendEmail(ctx, &integrations.EmailMessage{
    To: []integrations.EmailAddress{
        {Email: "user@example.com", Name: "User"},
    },
    Subject:     "Welcome to Marimo ERP",
    HTMLContent: "<h1>Welcome!</h1>",
})

// Template email
response, err := sendgrid.SendTemplateEmail(ctx, templateID, to, dynamicData)
```

**Managing Contacts:**

```go
// Create contact list
list, err := sendgrid.CreateContactList(ctx, integrations.ContactListCreateParams{
    Name: "Newsletter Subscribers",
})

// Add contact
err = sendgrid.AddContactToList(ctx, list.ID, &integrations.Contact{
    Email:     "user@example.com",
    FirstName: "John",
    LastName:  "Doe",
})
```

**Email Campaigns:**

```go
campaign, err := sendgrid.CreateCampaign(ctx, &integrations.EmailCampaign{
    Title:    "Monthly Newsletter",
    Subject:  "What's New This Month",
    ListIDs:  []string{listID},
})

// Send immediately or schedule
err = sendgrid.SendCampaign(ctx, campaign.ID, scheduleTime)
```

**Analytics:**

```go
stats, err := sendgrid.GetStats(ctx, startDate, endDate)
// Returns: delivered, opens, clicks, bounces, etc.
```

---

## API Webhooks

Webhooks allow external systems to receive real-time notifications about events in your application.

### Webhook Configuration

**Creating a Webhook:**

```go
webhook := &webhooks.Webhook{
    ID:       uuid.New(),
    TenantID: tenantID,
    URL:      "https://your-app.com/webhooks/marimo",
    Secret:   "your-secret-key",
    Events: []webhooks.EventType{
        webhooks.EventUserCreated,
        webhooks.EventPaymentSucceeded,
    },
    Active:  true,
}

repo := webhooks.NewRepository(db)
err := repo.Create(ctx, webhook)
```

### Event Types

- `user.created` - New user registered
- `user.updated` - User profile updated
- `user.deleted` - User deleted
- `payment.succeeded` - Payment completed successfully
- `payment.failed` - Payment failed
- `subscription.created` - Subscription started
- `subscription.updated` - Subscription changed
- `subscription.canceled` - Subscription ended
- `custom` - Custom events

### Dispatching Events

```go
service := webhooks.NewService(repo)

event := &webhooks.Event{
    ID:       uuid.New(),
    TenantID: tenantID,
    Type:     webhooks.EventUserCreated,
    Data: map[string]interface{}{
        "user_id": userID,
        "email":   email,
        "name":    name,
    },
}

err := service.Dispatch(ctx, event)
// Sends to all subscribed webhooks asynchronously
```

### Webhook Payload

```json
{
  "id": "evt_123abc",
  "type": "user.created",
  "data": {
    "user_id": "usr_456def",
    "email": "user@example.com",
    "name": "John Doe"
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Security

**Signature Verification:**

All webhooks include `X-Webhook-Signature` header with HMAC-SHA256 signature.

```go
// Receiving webhooks
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("X-Webhook-Signature")

    if !webhooks.VerifySignature(payload, signature, webhookSecret) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // Process webhook
}
```

### Retry Logic

Failed webhooks are automatically retried with exponential backoff:

1. 1 minute
2. 5 minutes
3. 15 minutes
4. 1 hour
5. 6 hours

After 5 failed attempts, the delivery is marked as failed.

### Monitoring

```go
// Get delivery history
deliveries, err := repo.GetPendingDeliveries(ctx)

// Check delivery status
for _, delivery := range deliveries {
    fmt.Printf("Status: %s, Attempt: %d\n", delivery.Status, delivery.Attempt)
}
```

---

## Mobile Application

React Native mobile app for iOS and Android with full feature parity.

### Features

✅ Authentication (Login, Register)
✅ Dashboard with real-time stats
✅ User management
✅ Analytics and reports
✅ Push notifications
✅ Offline mode support
✅ Modern UI with dark mode
✅ Real-time updates via WebSocket
✅ Export data (CSV, Excel, PDF)

### Tech Stack

- React Native 0.73
- TypeScript 5.3
- React Query (TanStack Query)
- React Navigation
- React Hook Form + Zod
- Axios with interceptors
- AsyncStorage

### Getting Started

```bash
# Install dependencies
cd mobile && npm install

# iOS
cd ios && pod install && cd ..
npm run ios

# Android
npm run android
```

### API Integration

**Configuration:**

```typescript
// src/config/api.ts
export const API_CONFIG = {
  BASE_URL: 'https://api.marimo-erp.com',
  TIMEOUT: 30000,
};
```

**Authentication:**

```typescript
import { authService } from '@/services/authService';

// Login
await authService.login({ email, password });

// Auto token refresh
// Handled automatically by axios interceptors
```

**Making API Calls:**

```typescript
import { useQuery } from '@tanstack/react-query';
import apiClient from '@/config/api';

const { data, isLoading } = useQuery({
  queryKey: ['users'],
  queryFn: async () => {
    const response = await apiClient.get('/users');
    return response.data;
  },
});
```

### Form Validation

```typescript
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
});

const { control, handleSubmit } = useForm({
  resolver: zodResolver(schema),
});
```

### Push Notifications

```typescript
import messaging from '@react-native-firebase/messaging';

// Request permission
await messaging().requestPermission();

// Get token
const token = await messaging().getToken();

// Register token with backend
await apiClient.post('/notifications/register', { token });

// Handle notifications
messaging().onMessage(async (message) => {
  // Show notification
});
```

### Offline Support

- Automatic request queueing
- Local data caching
- Sync when online
- Conflict resolution

### Deployment

**Android:**
```bash
cd android
./gradlew bundleRelease
# Upload to Google Play Console
```

**iOS:**
```bash
cd ios
xcodebuild archive
# Upload to App Store Connect
```

---

## Best Practices

### Security

1. **Multi-tenancy**: Always validate tenant access
2. **Webhooks**: Verify signatures on all incoming webhooks
3. **API Keys**: Store securely, rotate regularly
4. **Data Encryption**: Encrypt sensitive data at rest
5. **HTTPS Only**: All communication over TLS

### Performance

1. **Analytics**: Cache frequently accessed reports
2. **Webhooks**: Process asynchronously
3. **Mobile**: Implement pagination and infinite scroll
4. **Database**: Index tenant_id on all tables
5. **Monitoring**: Track query performance

### Scalability

1. **Horizontal Scaling**: Stateless services
2. **Database**: Read replicas for analytics
3. **Caching**: Redis for sessions and queries
4. **Queue**: RabbitMQ for async processing
5. **CDN**: Static assets and images

---

## Troubleshooting

### Multi-Tenancy Issues

**Problem**: Cross-tenant data leakage
**Solution**: Use TenantAwareDB wrapper, audit queries

**Problem**: Tenant not resolved
**Solution**: Check headers, subdomain configuration

### Analytics Performance

**Problem**: Slow queries
**Solution**: Add indexes, use materialized views

**Problem**: Memory issues with large datasets
**Solution**: Implement pagination, streaming

### Integration Failures

**Problem**: Webhook timeouts
**Solution**: Implement async processing, increase timeout

**Problem**: Payment failures
**Solution**: Check API keys, test mode vs live mode

### Mobile App Issues

**Problem**: Token expiration
**Solution**: Implemented automatic refresh

**Problem**: Offline sync conflicts
**Solution**: Last-write-wins or custom resolution

---

## Additional Resources

- [Multi-Tenancy Patterns](https://docs.microsoft.com/en-us/azure/architecture/patterns/sharding)
- [Stripe Documentation](https://stripe.com/docs/api)
- [SendGrid Documentation](https://docs.sendgrid.com/)
- [React Native Documentation](https://reactnative.dev/docs/getting-started)
- [Webhook Security](https://webhooks.fyi/)

---

**Last Updated**: 2024-01-15
**Version**: 1.0

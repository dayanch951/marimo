# Tutorial 07: Webhook Integration

**Duration**: 12 minutes
**Level**: Advanced
**Prerequisites**:
- Tutorial 01: Introduction
- Tutorial 03: First API Call
- Basic understanding of HTTP and APIs

## Learning Objectives

By the end of this tutorial, viewers will:
- Understand what webhooks are and why they're useful
- Know how to create and configure webhooks in Marimo ERP
- Learn to verify webhook signatures for security
- Implement a webhook receiver in Node.js and Go
- Debug webhook deliveries

## Video Structure

### Intro (0:00 - 0:30)
**Visual**: Title card with webhook icon
**Narration**:
> "Webhooks are the glue that connects Marimo ERP to your other systems. In this tutorial, we'll explore how to set up webhooks, verify their authenticity, and build a complete webhook integration. Let's dive in!"

**On-Screen Text**: "Webhook Integration"

---

### What Are Webhooks? (0:30 - 1:30)
**Visual**: Diagram showing webhook flow
**Narration**:
> "Think of webhooks as reverse APIs. Instead of your application constantly polling Marimo ERP for updates, webhooks push notifications to your application when events occur. This is more efficient and provides real-time updates."

**Diagram Animation**:
1. Show traditional polling (app → server, repeat)
2. Show webhook (server → app, triggered by event)

**On-Screen Text**:
- "Polling: You ask 'What's new?' repeatedly"
- "Webhooks: We tell you 'Something happened!'"

**Narration continued**:
> "When a user is created, a payment succeeds, or any important event happens, Marimo ERP can instantly notify your application via HTTP POST request."

**Show event examples**:
- user.created
- user.updated
- payment.succeeded
- payment.failed
- report.generated

---

### Creating a Webhook (1:30 - 3:30)

#### Via API (1:30 - 2:30)
**Visual**: Split screen - left: code editor, right: Postman
**Narration**:
> "Let's create our first webhook. We'll use the API directly. First, authenticate and get your access token."

**Code shown**:
```bash
# Login
curl -X POST https://api.marimo-erp.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@acme.com","password":"password"}'
```

**Show response, highlight token**:
```json
{
  "token": "eyJhbGci...",
  "user": {...}
}
```

**Narration**:
> "Now, let's create a webhook that listens for user events."

**Code shown**:
```bash
curl -X POST https://api.marimo-erp.com/webhooks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-Slug: acme-corp" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/marimo",
    "events": ["user.created", "user.updated", "user.deleted"],
    "headers": {
      "X-Custom-Auth": "your-secret-key"
    }
  }'
```

**Show response**:
```json
{
  "id": "wh_abc123",
  "url": "https://your-app.com/webhooks/marimo",
  "secret": "whsec_def456",
  "events": ["user.created", "user.updated", "user.deleted"],
  "status": "active"
}
```

**On-Screen Text**: "⚠️ SAVE THE SECRET! You'll need it to verify signatures."

#### Via Web UI (2:30 - 3:30)
**Visual**: Screen recording of web dashboard
**Narration**:
> "You can also create webhooks through the web interface. Navigate to Settings → Integrations → Webhooks."

**Demo Steps**:
1. Click "New Webhook" button
2. Enter URL: `https://your-app.com/webhooks/marimo`
3. Select events: user.created, user.updated
4. Add custom header (optional)
5. Click "Create"
6. Show success message with secret
7. Copy secret to clipboard

**On-Screen Text**: "Pro tip: Use webhook.site to test without coding"

---

### Webhook Payload Format (3:30 - 4:30)
**Visual**: JSON payload with syntax highlighting
**Narration**:
> "When an event occurs, Marimo ERP sends a POST request to your webhook URL with this JSON payload."

**Show payload**:
```json
{
  "id": "evt_123",
  "type": "user.created",
  "timestamp": "2024-01-15T10:50:00Z",
  "tenant_id": "tenant_abc",
  "data": {
    "id": "user_xyz",
    "email": "newuser@acme.com",
    "name": "John Doe",
    "role": "user",
    "created_at": "2024-01-15T10:50:00Z"
  }
}
```

**Highlight each field as explained**:
- `id`: Unique event identifier
- `type`: Event type (user.created)
- `timestamp`: When event occurred
- `tenant_id`: Which tenant triggered it
- `data`: Event-specific payload

**Narration**:
> "The payload always includes these standard fields, plus event-specific data in the 'data' object."

---

### Security: Verifying Signatures (4:30 - 6:30)
**Visual**: Diagram showing HMAC signature flow
**Narration**:
> "Here's the critical part: verifying that webhooks actually came from Marimo ERP. Every webhook includes an X-Marimo-Signature header containing an HMAC signature."

**Diagram shows**:
1. Marimo creates HMAC from payload + secret
2. Sends payload + signature
3. Your app creates HMAC from payload + secret
4. Compares signatures
5. Accepts if match, rejects if different

#### Node.js Implementation (5:00 - 5:45)
**Visual**: VS Code with Node.js code
**Narration**:
> "Let's implement signature verification in Node.js."

**Code shown**:
```javascript
const express = require('express');
const crypto = require('crypto');

const app = express();

function verifyWebhookSignature(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  const digest = hmac.update(payload).digest('hex');

  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(digest)
  );
}

app.post('/webhooks/marimo',
  express.raw({type: 'application/json'}),
  (req, res) => {
    const signature = req.headers['x-marimo-signature'];
    const secret = process.env.WEBHOOK_SECRET;

    if (!verifyWebhookSignature(req.body, signature, secret)) {
      console.log('Invalid signature!');
      return res.status(401).send('Unauthorized');
    }

    const event = JSON.parse(req.body);
    console.log('Verified event:', event.type);

    // Process the event
    handleWebhookEvent(event);

    res.status(200).send('OK');
});
```

**On-Screen Text**: "Always verify signatures in production!"

#### Go Implementation (5:45 - 6:30)
**Visual**: VS Code with Go code
**Narration**:
> "And here's the same verification in Go."

**Code shown**:
```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "io"
    "net/http"
)

func VerifySignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("X-Marimo-Signature")
    secret := os.Getenv("WEBHOOK_SECRET")

    if !VerifySignature(payload, signature, secret) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    var event WebhookEvent
    json.Unmarshal(payload, &event)

    // Process event
    handleEvent(event)

    w.WriteHeader(http.StatusOK)
}
```

---

### Processing Events (6:30 - 8:00)
**Visual**: Code editor showing event handler
**Narration**:
> "Once verified, you can process the event based on its type. Let's build a complete event handler."

**Code shown**:
```javascript
function handleWebhookEvent(event) {
  console.log(`Processing ${event.type} event`);

  switch (event.type) {
    case 'user.created':
      handleUserCreated(event.data);
      break;

    case 'user.updated':
      handleUserUpdated(event.data);
      break;

    case 'payment.succeeded':
      handlePaymentSucceeded(event.data);
      break;

    default:
      console.log('Unknown event type:', event.type);
  }
}

function handleUserCreated(user) {
  console.log('New user created:', user.email);

  // Examples of what you might do:
  // - Send welcome email
  // - Create user in your system
  // - Trigger onboarding workflow
  // - Notify Slack channel

  sendWelcomeEmail(user.email, user.name);
  createUserInCRM(user);
}

async function sendWelcomeEmail(email, name) {
  // Your email sending logic
  console.log(`Sending welcome email to ${email}`);
}
```

**Demo**: Show live webhook being triggered and processed

---

### Testing Webhooks (8:00 - 9:30)

#### Using webhook.site (8:00 - 8:45)
**Visual**: Browser showing webhook.site
**Narration**:
> "Before building your webhook receiver, test with webhook.site. It gives you a temporary URL that captures all incoming webhooks."

**Demo**:
1. Go to webhook.site
2. Copy unique URL
3. Create webhook in Marimo with that URL
4. Trigger event (create a user)
5. Show webhook received on webhook.site
6. Inspect payload and headers

#### Local Testing with ngrok (8:45 - 9:30)
**Visual**: Terminal with ngrok running
**Narration**:
> "To test your local webhook receiver, use ngrok to expose your localhost to the internet."

**Terminal commands**:
```bash
# Install ngrok
brew install ngrok  # or download from ngrok.com

# Start your webhook server locally
node webhook-server.js
# Server running on localhost:3000

# Expose it with ngrok
ngrok http 3000
```

**Show ngrok output**:
```
Forwarding https://abc123.ngrok.io -> http://localhost:3000
```

**Demo**:
1. Create webhook with ngrok URL
2. Trigger event in Marimo
3. Show event received in local server logs
4. Show ngrok web interface with request details

---

### Monitoring and Debugging (9:30 - 11:00)

#### Webhook Delivery Logs (9:30 - 10:15)
**Visual**: Marimo dashboard showing webhook deliveries
**Narration**:
> "Marimo ERP tracks every webhook delivery. You can see successful deliveries, failures, and retry attempts."

**Demo**:
1. Navigate to Webhooks → Select webhook → Deliveries tab
2. Show delivery list with statuses
3. Click on failed delivery
4. Show details: response code, response body, retry schedule

**On-Screen Text**:
- "✅ Success: 200-299 status codes"
- "❌ Failed: 4xx, 5xx, or timeout"
- "🔄 Retry: Automatic with exponential backoff"

#### Retry Logic (10:15 - 11:00)
**Visual**: Timeline diagram showing retry schedule
**Narration**:
> "If your webhook endpoint is temporarily down, Marimo automatically retries with exponential backoff."

**Show retry schedule**:
- Attempt 1: Immediately
- Attempt 2: After 1 minute
- Attempt 3: After 5 minutes
- Attempt 4: After 15 minutes
- Attempt 5: After 1 hour
- Attempt 6: After 6 hours

**On-Screen Text**: "Maximum 5 retries over 24 hours"

**Narration**:
> "After 5 failed attempts, the webhook is marked as failed. You can manually retry from the dashboard."

**Demo**: Click "Retry" button on failed webhook

---

### Best Practices (11:00 - 11:45)
**Visual**: Checklist animation
**Narration**:
> "Let's wrap up with some best practices for production webhooks."

**List appears on screen**:

1. **Always verify signatures**
   > "Never trust incoming webhooks without signature verification. This prevents malicious actors from sending fake events."

2. **Respond quickly**
   > "Return 200 OK within 5 seconds. If processing takes longer, queue the event and process asynchronously."

3. **Handle duplicates**
   > "Due to retries, you might receive the same event twice. Use the event ID to deduplicate."

4. **Monitor failures**
   > "Set up alerts for repeated webhook failures. Check the delivery logs regularly."

5. **Use HTTPS**
   > "Webhook URLs must use HTTPS in production. We won't send webhooks to HTTP URLs."

**Code example**:
```javascript
// Queue for async processing
app.post('/webhooks/marimo', async (req, res) => {
  // Verify signature
  if (!verifySignature(...)) {
    return res.status(401).send('Unauthorized');
  }

  // Respond immediately
  res.status(200).send('OK');

  // Process asynchronously
  const event = JSON.parse(req.body);
  await queue.add('webhook-event', event);
});
```

---

### Outro (11:45 - 12:00)
**Visual**: Code samples on screen
**Narration**:
> "You now know how to integrate Marimo ERP with any system using webhooks. The full code examples from this tutorial are available on GitHub. Next up, we'll explore third-party integrations with Stripe and SendGrid. See you there!"

**On-Screen Text**:
- "Code: github.com/marimo-erp/webhook-examples"
- "Docs: docs.marimo-erp.com/webhooks"
- "Next: Third-party Integrations"

---

## B-Roll Footage

1. Webhook creation in UI (various angles)
2. Code scrolling (Node.js and Go)
3. Terminal with successful webhook logs
4. Failed webhook being retried
5. Ngrok tunnel startup
6. webhook.site interface
7. Network requests in browser DevTools
8. Signature verification diagram

## Code Examples to Prepare

### Node.js Complete Example
File: `webhook-server.js`
```javascript
// Complete, working webhook server
// Include all security, logging, error handling
```

### Go Complete Example
File: `webhook-server.go`
```go
// Complete, working webhook server
// Include all security, logging, error handling
```

### Test Script
File: `test-webhook.js`
```javascript
// Script to trigger test events
// For demo purposes
```

## Demo Setup

### Prerequisites
- Marimo ERP running locally or on staging
- Node.js installed
- ngrok installed
- Postman or curl
- Sample tenant with data

### Test Scenarios
1. Successful webhook delivery
2. Failed webhook (server returns 500)
3. Webhook with invalid signature
4. Retry of failed webhook
5. Duplicate event handling

### Accounts Needed
- webhook.site account (free)
- ngrok account (free tier)

## Screen Recording Notes

### Terminal Setup
- Font: Monaco or Menlo, 16pt
- Color scheme: Dark background with syntax highlighting
- Width: Readable but not full screen
- Show command prompts clearly

### Browser Setup
- Install JSON Formatter extension
- Clear cookies/cache before recording
- Zoom level: 125% for readability
- Hide bookmarks bar

### Code Editor Setup
- Theme: Dark theme (e.g., One Dark)
- Font: Fira Code or JetBrains Mono, 16pt
- Show line numbers
- Enable ligatures
- Hide minimap for clarity

## Post-Production

### Graphics Needed
1. Webhook flow diagram (polling vs webhooks)
2. HMAC signature verification flowchart
3. Retry timeline visualization
4. Best practices checklist
5. Event type icons

### Code Highlighting
- Highlight important lines during narration
- Use zoom-in effect for complex code
- Add annotations for key concepts

### Transitions
- Use quick fade between sections
- Code editor → Terminal: slide transition
- Browser → Code: fade
- Keep transitions under 0.5 seconds

## YouTube Metadata

### Title
"Webhook Integration Tutorial - Marimo ERP Advanced Features"

### Description
```
Learn how to integrate Marimo ERP with your applications using webhooks for real-time event notifications.

In this tutorial, you'll learn:
- What webhooks are and why they're useful
- Creating webhooks via API and UI
- Verifying webhook signatures (security!)
- Building webhook receivers in Node.js and Go
- Testing with webhook.site and ngrok
- Monitoring and debugging deliveries
- Production best practices

🔗 Resources:
- Code Examples: https://github.com/marimo-erp/webhook-examples
- Webhook Docs: https://docs.marimo-erp.com/webhooks
- API Reference: https://docs.marimo-erp.com/api

⏱️ Chapters:
0:00 - Introduction
0:30 - What Are Webhooks?
1:30 - Creating a Webhook
3:30 - Payload Format
4:30 - Signature Verification
6:30 - Processing Events
8:00 - Testing Webhooks
9:30 - Monitoring & Debugging
11:00 - Best Practices
11:45 - Outro

#MarimoERP #Webhooks #API #Integration #NodeJS #Golang
```

### Tags
marimo-erp, webhooks, api-integration, nodejs, golang, express, hmac, signature-verification, real-time, events, microservices, rest-api

### Thumbnail
- Text: "Webhook Integration"
- Icon: Webhook symbol (arrows pointing from box to box)
- Code snippet in background
- "Advanced" badge

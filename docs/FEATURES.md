# Marimo ERP - Advanced Features

## Overview

Документация по расширенным функциям системы Marimo ERP, включая email notifications, file storage, data export, advanced search, pagination и WebSockets.

## 1. Email Notifications

### Description

Служба email уведомлений для отправки транзакционных писем пользователям через SMTP.

### Configuration

```bash
# Environment variables
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@marimo.dev
FROM_NAME="Marimo ERP"
```

### Usage

```go
import "github.com/dayanch951/marimo/shared/email"

// Create email service
emailService := email.NewEmailService()

// Send welcome email
err := emailService.SendWelcomeEmail("user@example.com", "John Doe")

// Send password reset
err := emailService.SendPasswordResetEmail("user@example.com", "https://marimo.dev/reset/token123")

// Send generic notification
err := emailService.SendNotificationEmail("user@example.com", "Subject", "Message body")

// Send custom email
err := emailService.SendEmail(email.EmailMessage{
    To:       []string{"user@example.com"},
    Subject:  "Custom Subject",
    HTMLBody: "<h1>HTML Content</h1>",
})
```

### Features

- **HTML Templates**: Pre-built templates for common emails
- **SMTP Support**: Works with any SMTP server (Gmail, SendGrid, AWS SES)
- **Attachments**: Support for email attachments
- **Queue Integration**: Can be integrated with RabbitMQ for async sending

### Email Templates

#### Welcome Email
- Professional gradient design
- Call-to-action button
- Feature highlights
- Brand consistent

#### Password Reset
- Security warnings
- Expiring link
- Clear instructions
- Contact support info

## 2. File Upload & Storage

### Description

Гибкая система хранения файлов с поддержкой локального хранилища и MinIO/S3.

### Configuration

```bash
# Local storage
USE_LOCAL_STORAGE=true
LOCAL_STORAGE_PATH=./uploads

# MinIO/S3 storage
USE_LOCAL_STORAGE=false
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET=marimo-files
```

### Usage

```go
import (
    "context"
    "github.com/dayanch951/marimo/shared/storage"
)

// Create storage service
storageService, err := storage.NewStorageService()
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()

// Upload file
fileInfo, err := storageService.UploadFile(
    ctx,
    fileReader,
    "document.pdf",
    "application/pdf",
    fileSize,
)

// Download file
reader, info, err := storageService.DownloadFile(ctx, fileInfo.ID)
defer reader.Close()

// Get temporary URL (presigned)
url, err := storageService.GetFileURL(ctx, fileInfo.ID, 24*time.Hour)

// List files
files, err := storageService.ListFiles(ctx, "prefix/")

// Delete file
err = storageService.DeleteFile(ctx, fileInfo.ID)
```

### Features

- **Dual Storage**: Local filesystem or MinIO/S3
- **Presigned URLs**: Secure temporary file access
- **Metadata**: Store custom file metadata
- **Auto-generated IDs**: UUID-based filenames
- **Content Type Detection**: Automatic MIME type handling

### Docker Setup (MinIO)

```yaml
# docker-compose.yml
minio:
  image: minio/minio:latest
  ports:
    - "9000:9000"
    - "9001:9001"
  environment:
    MINIO_ROOT_USER: minioadmin
    MINIO_ROOT_PASSWORD: minioadmin
  command: server /data --console-address ":9001"
  volumes:
    - minio_data:/data
```

## 3. Data Export

### Description

Экспорт данных в различные форматы (CSV, Excel, PDF) для отчетов и анализа.

### Usage

```go
import "github.com/dayanch951/marimo/shared/export"

// Create export service
exportService := export.NewExportService()

// Prepare data
data := export.ExportData{
    Title:   "Sales Report",
    Headers: []string{"Date", "Product", "Quantity", "Amount"},
    Rows: [][]string{
        {"2024-01-01", "Product A", "10", "$100.00"},
        {"2024-01-02", "Product B", "5", "$50.00"},
    },
}

// Export to CSV
csvData, contentType, err := exportService.Export(data, export.FormatCSV)

// Export to Excel
excelData, contentType, err := exportService.Export(data, export.FormatExcel)

// Export to PDF
pdfData, contentType, err := exportService.Export(data, export.FormatPDF)

// Generate filename
filename := exportService.GetFilename("sales_report", export.FormatExcel)
// Output: sales_report_20240101_150405.xlsx
```

### HTTP Handler Example

```go
func ExportHandler(w http.ResponseWriter, r *http.Request) {
    format := r.URL.Query().Get("format") // csv, xlsx, pdf

    // Prepare data
    data := prepareExportData()

    // Export
    content, contentType, err := exportService.Export(data, export.ExportFormat(format))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Set headers
    filename := exportService.GetFilename("report", export.ExportFormat(format))
    w.Header().Set("Content-Type", contentType)
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

    // Write response
    w.Write(content)
}
```

### Features

#### CSV Export
- Standard RFC 4180 format
- UTF-8 encoding
- Excel compatible

#### Excel Export
- Professional styling
- Colored headers
- Auto-fit columns
- Borders and formatting
- Multiple sheets support (future)

#### PDF Export
- A4 page size
- Table layout
- Alternating row colors
- Custom headers/footers
- Page numbers
- Timestamp

## 4. Pagination

### Description

Утилиты для пагинации больших списков данных с поддержкой сортировки.

### Usage

```go
import "github.com/dayanch951/marimo/shared/pagination"

// Parse from query parameters
pageReq := pagination.ParsePageRequest(
    r.URL.Query().Get("page"),
    r.URL.Query().Get("page_size"),
    r.URL.Query().Get("sort_by"),
    r.URL.Query().Get("sort_dir"),
)

// Or create manually
pageReq := pagination.NewPageRequest(1, 20, "created_at", "desc")

// Use in SQL query
query := fmt.Sprintf(`
    SELECT * FROM users
    WHERE active = true
    %s
    %s
`, pageReq.GetSortClause(), pageReq.GetLimitOffset())

// Get total count
var total int64
db.QueryRow("SELECT COUNT(*) FROM users WHERE active = true").Scan(&total)

// Query data
rows, err := db.Query(query)
// ... process rows

// Create response
response := pagination.NewPageResponse(users, pageReq.Page, pageReq.PageSize, total)

// Send JSON response
json.NewEncoder(w).Encode(response)
```

### Response Format

```json
{
  "data": [...],
  "page": 1,
  "page_size": 20,
  "total": 150,
  "total_pages": 8,
  "has_next": true,
  "has_prev": false
}
```

### Features

- **Default Values**: Sensible defaults (page=1, size=20)
- **Max Page Size**: Prevents overloading (max 100)
- **Sort Support**: Field and direction
- **SQL Helpers**: Generate LIMIT/OFFSET clauses
- **Type-Safe**: Generic response type

## 5. Advanced Search & Filters

### Description

Мощная система поиска и фильтрации с поддержкой сложных условий.

### Filter Operators

```go
OpEqual        // eq:  field = value
OpNotEqual     // ne:  field != value
OpGreaterThan  // gt:  field > value
OpGreaterEqual // gte: field >= value
OpLessThan     // lt:  field < value
OpLessEqual    // lte: field <= value
OpLike         // like: field LIKE '%value%'
OpIn           // in:  field IN (...)
OpNotIn        // nin: field NOT IN (...)
OpBetween      // between: field BETWEEN a AND b
OpIsNull       // null: field IS NULL
OpNotNull      // notnull: field IS NOT NULL
```

### Usage

```go
import "github.com/dayanch951/marimo/shared/search"

// Create query builder
qb := search.NewQueryBuilder()

// Create search request
searchReq := search.SearchRequest{
    Query: "john",
    SearchFields: []string{"first_name", "last_name", "email"},
    Filters: search.FilterGroup{
        Logic: "AND",
        Filters: []search.Filter{
            search.EqualFilter("status", "active"),
            search.DateRangeFilter("created_at", startDate, endDate),
            search.InFilter("role", []interface{}{"admin", "manager"}),
        },
    },
}

// Build query
baseQuery := "SELECT * FROM users"
query := search.BuildCompleteQuery(baseQuery, searchReq, qb)
params := qb.GetParams()

// Execute
rows, err := db.Query(query, params...)
```

### Filter Groups

Вложенные группы с AND/OR логикой:

```go
filters := search.FilterGroup{
    Logic: "OR",
    Groups: []search.FilterGroup{
        {
            Logic: "AND",
            Filters: []search.Filter{
                {Field: "age", Operator: search.OpGreaterEqual, Value: 18},
                {Field: "country", Operator: search.OpEqual, Value: "US"},
            },
        },
        {
            Logic: "AND",
            Filters: []search.Filter{
                {Field: "age", Operator: search.OpGreaterEqual, Value: 21},
                {Field: "country", Operator: search.OpEqual, Value: "UK"},
            },
        },
    },
}
```

### Features

- **Full-text Search**: ILIKE search across multiple fields
- **Complex Filters**: AND/OR groups with nesting
- **SQL Injection Safe**: Parameterized queries
- **Case-Insensitive**: ILIKE for text matching
- **Helper Functions**: Pre-built filter creators

## 6. WebSockets

### Description

Real-time двунаправленная коммуникация между сервером и клиентами.

### Server Setup

```go
import "github.com/dayanch951/marimo/shared/websocket"

// Create hub
hub := websocket.NewHub()

// Register default handlers
websocket.RegisterDefaultHandlers(hub)

// Register custom handlers
hub.RegisterHandler("chat", func(client *websocket.Client, msg websocket.Message) error {
    // Handle chat message
    room := msg.Payload["room"].(string)
    return hub.BroadcastToRoom(room, websocket.Message{
        Type: "chat",
        Payload: map[string]interface{}{
            "user":    client.UserID,
            "message": msg.Payload["message"],
            "time":    time.Now(),
        },
    })
})

// Run hub
go hub.Run()

// HTTP endpoint
http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
    websocket.ServeWS(hub, w, r)
})
```

### Client Usage (JavaScript)

```javascript
// Connect
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
    console.log('Connected');

    // Join room
    ws.send(JSON.stringify({
        type: 'join',
        payload: { room: 'general' }
    }));
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message);

    switch(message.type) {
        case 'welcome':
            console.log('Client ID:', message.payload.client_id);
            break;
        case 'chat':
            displayMessage(message.payload);
            break;
        case 'notification':
            showNotification(message.payload);
            break;
    }
};

// Send message
function sendMessage(room, text) {
    ws.send(JSON.stringify({
        type: 'chat',
        payload: {
            room: room,
            message: text
        }
    }));
}

// Leave room
function leaveRoom(room) {
    ws.send(JSON.stringify({
        type: 'leave',
        payload: { room: room }
    }));
}
```

### Features

- **Rooms/Channels**: Group messaging support
- **Broadcast**: Send to all or specific rooms
- **Client Management**: Track connections
- **Custom Handlers**: Extensible message types
- **Ping/Pong**: Keep-alive mechanism
- **Graceful Disconnect**: Cleanup on disconnect

### Message Types

#### Built-in Messages

- `welcome` - Server welcome on connect
- `ping/pong` - Keep-alive
- `join/joined` - Join room
- `leave/left` - Leave room
- `subscribe/subscribed` - Subscribe to channel
- `unsubscribe/unsubscribed` - Unsubscribe from channel

#### Custom Messages

Register handlers for any message type:

```go
hub.RegisterHandler("order_update", func(client *Client, msg Message) error {
    // Broadcast to all users watching this order
    orderID := msg.Payload["order_id"].(string)
    return hub.BroadcastToRoom("order:"+orderID, Message{
        Type: "order_update",
        Payload: msg.Payload,
    })
})
```

### Use Cases

1. **Real-time Notifications**
   - Order updates
   - System alerts
   - User mentions

2. **Chat/Messaging**
   - Team chat
   - Customer support
   - Comments

3. **Live Data**
   - Dashboard updates
   - Stock prices
   - Analytics

4. **Collaborative Features**
   - Presence indicators
   - Concurrent editing
   - Live cursors

## Integration Examples

### Complete Feature Integration

```go
// main.go
package main

import (
    "github.com/dayanch951/marimo/shared/email"
    "github.com/dayanch951/marimo/shared/storage"
    "github.com/dayanch951/marimo/shared/export"
    "github.com/dayanch951/marimo/shared/websocket"
)

func main() {
    // Initialize services
    emailSvc := email.NewEmailService()
    storageSvc, _ := storage.NewStorageService()
    exportSvc := export.NewExportService()
    wsHub := websocket.NewHub()

    // Start WebSocket hub
    go wsHub.Run()
    websocket.RegisterDefaultHandlers(wsHub)

    // Setup HTTP handlers
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        websocket.ServeWS(wsHub, w, r)
    })

    http.HandleFunc("/upload", uploadHandler(storageSvc, wsHub))
    http.HandleFunc("/export", exportHandler(exportSvc))
    http.HandleFunc("/send-email", emailHandler(emailSvc))

    http.ListenAndServe(":8080", nil)
}
```

## Dependencies

Add to `shared/go.mod`:

```go
require (
    github.com/google/uuid v1.6.0
    github.com/gorilla/websocket v1.5.1
    github.com/jung-kurt/gofpdf v1.16.2
    github.com/minio/minio-go/v7 v7.0.66
    github.com/xuri/excelize/v2 v2.8.1
)
```

## Best Practices

### Email
- Use templates for consistency
- Queue emails for async sending
- Handle bounces and complaints
- Test with real email providers

### File Storage
- Validate file types and sizes
- Use presigned URLs for security
- Implement virus scanning
- Set up backup strategy

### Export
- Limit export size
- Use background jobs for large exports
- Cache frequently requested exports
- Provide progress indicators

### Search
- Index frequently searched fields
- Use prepared statements
- Limit result sets
- Implement query caching

### Pagination
- Always set max page size
- Use cursor-based for real-time data
- Cache total counts
- Provide "load more" option

### WebSockets
- Authenticate connections
- Rate limit messages
- Handle reconnection
- Clean up inactive connections
- Use rooms for scalability

## Troubleshooting

### Email Not Sending
- Check SMTP credentials
- Verify port (587 for TLS, 465 for SSL)
- Enable "Less secure apps" for Gmail
- Use app-specific passwords

### File Upload Fails
- Check disk space
- Verify permissions
- Check MaxRequestBody size
- Validate file types

### Export Timeout
- Reduce data set size
- Use background jobs
- Implement streaming
- Add progress updates

### WebSocket Disconnects
- Check firewall settings
- Increase ping timeout
- Handle reconnection
- Check load balancer config

## Security Considerations

1. **Email**: Don't expose email service errors to users
2. **Files**: Scan uploads for malware, validate extensions
3. **Export**: Rate limit exports, require authentication
4. **Search**: Sanitize inputs, use parameterized queries
5. **WebSocket**: Authenticate connections, validate messages

## Performance Tips

1. **Email**: Use connection pooling, batch sends
2. **Files**: Use CDN for serving, implement caching
3. **Export**: Stream large files, use compression
4. **Search**: Add database indexes, use full-text search
5. **Pagination**: Cache total counts, use covering indexes
6. **WebSocket**: Use message queues for scaling

## Future Enhancements

- [ ] Email templates editor
- [ ] File preview generation
- [ ] Advanced export (charts, pivot tables)
- [ ] Elasticsearch integration
- [ ] Cursor-based pagination
- [ ] WebSocket clustering (Redis pub/sub)

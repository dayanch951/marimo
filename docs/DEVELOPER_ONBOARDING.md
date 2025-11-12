# Developer Onboarding Guide

Welcome to the Marimo ERP development team! This guide will help you get up and running with our codebase.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Development Environment Setup](#development-environment-setup)
3. [Project Structure](#project-structure)
4. [Running the Application](#running-the-application)
5. [Development Workflow](#development-workflow)
6. [Testing](#testing)
7. [Code Style and Standards](#code-style-and-standards)
8. [Common Tasks](#common-tasks)
9. [Troubleshooting](#troubleshooting)
10. [Resources](#resources)

## Prerequisites

Before you begin, ensure you have the following installed:

### Required Tools

- **Git** (2.30+)
  ```bash
  git --version
  ```

- **Go** (1.21+)
  ```bash
  go version
  ```

- **Node.js** (18+) and npm
  ```bash
  node --version
  npm --version
  ```

- **Docker** (20+) and Docker Compose (2+)
  ```bash
  docker --version
  docker compose version
  ```

- **PostgreSQL Client** (16+)
  ```bash
  psql --version
  ```

### Optional but Recommended

- **kubectl** - Kubernetes CLI
- **Postman** or **Insomnia** - API testing
- **VS Code** or **GoLand** - IDE
- **pgAdmin** or **DBeaver** - Database GUI

### Account Access

Request access to:
- [ ] GitHub repository
- [ ] Slack workspace (#marimo-dev, #marimo-alerts)
- [ ] Development Kubernetes cluster
- [ ] Development database credentials
- [ ] AWS development account (if applicable)
- [ ] Jira/Project management tool

## Development Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/marimo.git
cd marimo
```

### 2. Install Go Dependencies

```bash
# Install all Go modules
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### 3. Install Frontend Dependencies

```bash
cd frontend
npm install
cd ..
```

### 4. Install Mobile Dependencies (Optional)

```bash
cd mobile
npm install
cd ..
```

### 5. Set Up Environment Variables

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your local configuration
# For development, default values usually work
nano .env
```

**Key variables to configure:**

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=marimo_erp_dev
DB_USER=postgres
DB_PASSWORD=your-local-password

# JWT
JWT_SECRET=your-dev-secret-min-32-chars

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# For local development, you can use:
ENVIRONMENT=development
LOG_LEVEL=debug
```

### 6. Start Infrastructure Services

```bash
# Start PostgreSQL, Redis, Consul, RabbitMQ using Docker Compose
docker compose up -d

# Verify services are running
docker compose ps
```

Expected output:
```
NAME                   SERVICE    STATUS      PORTS
marimo-postgres-1      postgres   running     0.0.0.0:5432->5432/tcp
marimo-redis-1         redis      running     0.0.0.0:6379->6379/tcp
marimo-consul-1        consul     running     0.0.0.0:8500->8500/tcp
marimo-rabbitmq-1      rabbitmq   running     0.0.0.0:5672->5672/tcp, 0.0.0.0:15672->15672/tcp
```

### 7. Run Database Migrations

```bash
# Make sure postgres is ready
sleep 5

# Run migrations
./scripts/run-migrations.sh

# Verify migrations
psql -h localhost -U postgres -d marimo_erp_dev -c "\dt"
```

### 8. Seed Development Data (Optional)

```bash
# Create a test tenant and user
psql -h localhost -U postgres -d marimo_erp_dev <<EOF
-- Insert test tenant
INSERT INTO tenants (id, name, slug, status, trial_ends_at)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Test Company',
  'test-company',
  'active',
  NOW() + INTERVAL '30 days'
);

-- Insert test user (password: TestPass123!)
INSERT INTO users (id, tenant_id, email, name, password_hash, role)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000001',
  'admin@test.com',
  'Test Admin',
  '\$2a\$10\$YourBcryptHashHere',
  'admin'
);
EOF
```

## Project Structure

```
marimo/
├── cmd/                          # Entry points for services
│   ├── api-gateway/             # API Gateway service
│   ├── auth-service/            # Authentication service
│   └── users-service/           # Users management service
├── frontend/                     # React frontend application
│   ├── src/
│   │   ├── components/          # Reusable React components
│   │   ├── pages/               # Page components
│   │   ├── services/            # API services
│   │   └── hooks/               # Custom React hooks
│   └── package.json
├── mobile/                       # React Native mobile app
│   ├── src/
│   │   ├── screens/             # Mobile screens
│   │   └── services/            # API services
│   └── package.json
├── shared/                       # Shared libraries
│   ├── analytics/               # Analytics engine
│   ├── auth/                    # Authentication utilities
│   ├── database/                # Database helpers
│   ├── integrations/            # Third-party integrations
│   │   ├── stripe.go           # Stripe integration
│   │   └── sendgrid.go         # SendGrid integration
│   ├── middleware/              # HTTP middleware
│   ├── monitoring/              # Prometheus metrics
│   ├── tenancy/                 # Multi-tenancy support
│   └── webhooks/                # Webhook system
├── migrations/                   # Database migrations
│   ├── 001_create_tenants.sql
│   ├── 002_create_analytics.sql
│   └── ...
├── k8s/                         # Kubernetes manifests
│   ├── namespace.yml
│   ├── configmap.yml
│   ├── postgres-statefulset.yml
│   └── ...
├── scripts/                      # Utility scripts
│   ├── backup.sh                # Database backup
│   ├── restore.sh               # Database restore
│   └── run-migrations.sh        # Run migrations
├── docs/                         # Documentation
│   ├── openapi.yaml             # API specification
│   ├── adr/                     # Architecture Decision Records
│   ├── DEPLOYMENT.md            # Deployment guide
│   └── QUICKSTART.md            # Quick start guide
├── .github/                      # GitHub workflows
│   └── workflows/
│       ├── ci.yml               # CI pipeline
│       └── cd.yml               # CD pipeline
├── docker-compose.yml           # Local development services
├── .env.example                 # Environment variables template
├── go.mod                       # Go dependencies
└── README.md                    # Project overview
```

## Running the Application

### Option 1: Run Services Individually (Recommended for Development)

Terminal 1 - API Gateway:
```bash
cd cmd/api-gateway
go run main.go
# Listens on :8080
```

Terminal 2 - Auth Service:
```bash
cd cmd/auth-service
go run main.go
# Registers with Consul
```

Terminal 3 - Users Service:
```bash
cd cmd/users-service
go run main.go
# Registers with Consul
```

Terminal 4 - Frontend:
```bash
cd frontend
npm start
# Opens browser at http://localhost:3000
```

### Option 2: Use Docker Compose (Full Stack)

```bash
# Build and run all services
docker compose up --build

# Or run in background
docker compose up -d --build
```

### Verify Services

```bash
# Check API Gateway
curl http://localhost:8080/health

# Check Consul service registry
curl http://localhost:8500/v1/catalog/services

# Check RabbitMQ management
open http://localhost:15672
# Username: guest, Password: guest
```

## Development Workflow

### 1. Create a Feature Branch

```bash
# Update main branch
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/your-feature-name

# Or bug fix branch
git checkout -b fix/issue-description
```

### 2. Make Changes

Follow these guidelines:
- Write tests for new functionality
- Update documentation
- Follow code style guidelines
- Keep commits atomic and meaningful

### 3. Test Your Changes

```bash
# Run tests
go test ./...

# Run linter
golangci-lint run

# Test frontend
cd frontend
npm test
npm run lint
```

### 4. Commit and Push

```bash
# Stage changes
git add .

# Commit with meaningful message
git commit -m "feat: add user profile endpoint

- Implement GET /users/:id/profile
- Add profile update validation
- Include unit tests
"

# Push to remote
git push origin feature/your-feature-name
```

### 5. Create Pull Request

1. Go to GitHub
2. Click "New Pull Request"
3. Fill in PR template:
   - Description of changes
   - Related issues
   - Testing performed
   - Screenshots (if UI changes)
4. Request review from team members
5. Address review comments

### 6. Merge and Deploy

- After approval, squash and merge
- Delete feature branch
- CI/CD automatically deploys to staging
- Verify in staging before production

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./shared/tenancy -v
```

### Integration Tests

```bash
# Run integration tests (requires running services)
go test -tags=integration ./...
```

### Frontend Tests

```bash
cd frontend

# Run tests
npm test

# Run tests with coverage
npm test -- --coverage

# Run E2E tests (requires running backend)
npm run test:e2e
```

### Manual API Testing

Use the provided Postman collection:

```bash
# Import docs/postman_collection.json into Postman
# Or use curl:

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"TestPass123!"}'

# Use token in subsequent requests
TOKEN="your-jwt-token"

# Get users
curl http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-Slug: test-company"
```

## Code Style and Standards

### Go Code Style

We follow standard Go conventions:

- **Formatting**: Use `gofmt` (automatically done by most IDEs)
- **Linting**: Must pass `golangci-lint`
- **Naming**:
  - Packages: lowercase, single word
  - Exported: PascalCase
  - Unexported: camelCase
- **Error Handling**: Always check errors
  ```go
  if err != nil {
      return fmt.Errorf("failed to do something: %w", err)
  }
  ```

### TypeScript/React Style

- **Formatting**: Use Prettier (`.prettierrc` configured)
- **Linting**: ESLint rules in `.eslintrc`
- **Components**: Functional components with hooks
- **File naming**: PascalCase for components, camelCase for utilities

### Commit Message Format

Follow Conventional Commits:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

Examples:
```
feat(auth): add OAuth2 login support

Implement OAuth2 authentication flow with Google and GitHub providers.
Includes token refresh and revocation.

Closes #123
```

### Documentation

- **Code Comments**: Document exported functions and complex logic
  ```go
  // CreateTenant creates a new tenant with the given name and slug.
  // It returns an error if the slug is already taken.
  func CreateTenant(ctx context.Context, name, slug string) (*Tenant, error) {
  ```

- **API Documentation**: Use Swagger annotations
  ```go
  // @Summary Get user by ID
  // @Description Get detailed user information
  // @Tags users
  // @Param id path string true "User ID"
  // @Success 200 {object} User
  // @Router /users/{id} [get]
  ```

## Common Tasks

### Add a New API Endpoint

1. **Define the handler**:
   ```go
   // cmd/users-service/handlers/user.go
   func (h *Handler) GetUserProfile(c *gin.Context) {
       userID := c.Param("id")
       // Implementation
   }
   ```

2. **Register the route**:
   ```go
   // cmd/users-service/main.go
   router.GET("/users/:id/profile", handlers.GetUserProfile)
   ```

3. **Add tests**:
   ```go
   // cmd/users-service/handlers/user_test.go
   func TestGetUserProfile(t *testing.T) {
       // Test implementation
   }
   ```

4. **Update OpenAPI spec**:
   ```yaml
   # docs/openapi.yaml
   /users/{id}/profile:
     get:
       summary: Get user profile
       # ...
   ```

### Add a Database Migration

1. **Create migration file**:
   ```bash
   # Create new migration with next number
   cat > migrations/011_add_user_preferences.sql <<EOF
   -- Up
   CREATE TABLE user_preferences (
       user_id UUID PRIMARY KEY REFERENCES users(id),
       theme VARCHAR(20) DEFAULT 'light',
       language VARCHAR(10) DEFAULT 'en',
       created_at TIMESTAMP DEFAULT NOW()
   );

   -- Down
   DROP TABLE user_preferences;
   EOF
   ```

2. **Run migration**:
   ```bash
   ./scripts/run-migrations.sh
   ```

### Add a New Integration

1. **Create integration file**:
   ```go
   // shared/integrations/slack.go
   package integrations

   type SlackClient struct {
       apiKey string
   }

   func NewSlackClient(apiKey string) *SlackClient {
       return &SlackClient{apiKey: apiKey}
   }

   func (s *SlackClient) SendMessage(channel, message string) error {
       // Implementation
   }
   ```

2. **Add tests**:
   ```go
   // shared/integrations/slack_test.go
   func TestSlackClient_SendMessage(t *testing.T) {
       // Test implementation
   }
   ```

3. **Update documentation**:
   ```markdown
   # docs/ADVANCED_FEATURES.md
   ## Slack Integration
   ...
   ```

### Debug a Service

#### Using Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug API Gateway
cd cmd/api-gateway
dlv debug

# Set breakpoint
(dlv) break main.main
(dlv) continue
```

#### Using Logs

```bash
# View logs from all services
docker compose logs -f

# View specific service logs
docker compose logs -f api-gateway

# View logs with grep
docker compose logs -f | grep ERROR
```

#### Database Queries

```bash
# Connect to database
psql -h localhost -U postgres -d marimo_erp_dev

# Check tenant data
SELECT * FROM tenants;

# Check users
SELECT id, email, role, tenant_id FROM users;

# Check webhook deliveries
SELECT * FROM webhook_deliveries WHERE status = 'failed';
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Database Connection Failed

```bash
# Check if PostgreSQL is running
docker compose ps postgres

# Restart PostgreSQL
docker compose restart postgres

# Check logs
docker compose logs postgres
```

### Module Import Issues

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Verify modules
go mod verify
```

### Frontend Build Errors

```bash
# Clear node_modules and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install

# Clear cache
npm cache clean --force
```

### Consul Service Discovery Issues

```bash
# Check Consul UI
open http://localhost:8500

# List registered services
curl http://localhost:8500/v1/catalog/services

# Restart Consul
docker compose restart consul
```

## Resources

### Internal Documentation

- [Quick Start Guide](./QUICKSTART.md)
- [Deployment Guide](./DEPLOYMENT.md)
- [Architecture Decision Records](./adr/)
- [API Documentation](./openapi.yaml)
- [Advanced Features](./ADVANCED_FEATURES.md)

### External Resources

#### Go
- [Effective Go](https://golang.org/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

#### React
- [React Documentation](https://react.dev/)
- [React Query Docs](https://tanstack.com/query/latest)
- [React Hook Form](https://react-hook-form.com/)

#### Kubernetes
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kubernetes Patterns](https://k8spatterns.io/)

#### Tools
- [PostgreSQL Manual](https://www.postgresql.org/docs/16/)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/getstarted.html)
- [Prometheus Documentation](https://prometheus.io/docs/)

### Team Communication

- **Slack Channels**:
  - `#marimo-dev` - General development discussion
  - `#marimo-alerts` - Production alerts
  - `#marimo-deployments` - Deployment notifications

- **Meetings**:
  - Daily Standup: 10:00 AM
  - Sprint Planning: Every other Monday
  - Retrospective: Every other Friday
  - Code Review Sessions: Wednesdays

- **On-Call**:
  - Rotation schedule in PagerDuty
  - Escalation procedures in runbooks
  - Incident response guide: `docs/INCIDENT_RESPONSE.md`

### Getting Help

1. **Check documentation** (you're here!)
2. **Search Slack** for similar questions
3. **Ask in #marimo-dev**
4. **Schedule pairing session** with senior developer
5. **Create GitHub issue** for bugs or feature requests

## Next Steps

Now that you're set up:

1. ✅ Complete this onboarding guide
2. ⏳ Review the architecture (see [ADRs](./adr/))
3. ⏳ Read the [Quick Start Guide](./QUICKSTART.md)
4. ⏳ Pick up a "good first issue" from GitHub
5. ⏳ Schedule intro meeting with your team
6. ⏳ Set up your development environment
7. ⏳ Make your first contribution!

Welcome aboard! 🚀

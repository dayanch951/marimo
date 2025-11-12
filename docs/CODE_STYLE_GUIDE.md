# Code Style Guide

Comprehensive code style guide for Marimo ERP project.

## Table of Contents

1. [Go Style Guide](#go-style-guide)
2. [TypeScript/React Style Guide](#typescriptreact-style-guide)
3. [Git Commit Guidelines](#git-commit-guidelines)
4. [Code Review Guidelines](#code-review-guidelines)
5. [Testing Guidelines](#testing-guidelines)

## Go Style Guide

### General Principles

1. Follow [Effective Go](https://golang.org/doc/effective_go.html)
2. Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
3. Run `gofmt` and `goimports` before committing
4. Use `golangci-lint` to catch issues

### Naming Conventions

#### Variables

```go
// Good: descriptive, camelCase
var userCount int
var httpClient *http.Client
var maxRetries = 3

// Bad: too short, unclear
var uc int
var c *http.Client
var m = 3
```

#### Constants

```go
// Good: clear, grouped by purpose
const (
    StatusActive   = "active"
    StatusInactive = "inactive"
    StatusSuspended = "suspended"
)

const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
)

// Bad: magic numbers without constants
if status == "active" { // Use constant instead
    // ...
}
```

#### Functions

```go
// Good: verb + noun, clear purpose
func GetUserByID(id string) (*User, error)
func CreateTenant(name, slug string) error
func ValidateEmail(email string) bool

// Bad: unclear, too generic
func Get(id string) (*User, error)
func Do(name string) error
func Check(email string) bool
```

#### Types

```go
// Good: descriptive, PascalCase
type UserRepository struct {
    db *gorm.DB
}

type TenantService interface {
    Create(ctx context.Context, tenant *Tenant) error
    GetByID(ctx context.Context, id string) (*Tenant, error)
}

// Bad: abbreviations, unclear
type UsrRepo struct {
    db *gorm.DB
}
```

### Error Handling

```go
// Good: wrap errors with context
if err := db.Create(&user).Error; err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Good: custom error types
if err != nil {
    return errors.Wrap(err, errors.ErrDatabaseError, "failed to query users")
}

// Bad: swallow errors
db.Create(&user) // Error ignored!

// Bad: no context
if err != nil {
    return err
}
```

### Context Usage

```go
// Good: pass context as first parameter
func GetUser(ctx context.Context, id string) (*User, error) {
    var user User
    if err := db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

// Good: use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := GetUser(ctx, userID)
```

### Struct Tags

```go
// Good: proper formatting, align tags
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Email     string    `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
    Name      string    `json:"name" gorm:"not null" validate:"required,min=2"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
```

### Comments

```go
// Package users provides user management functionality
package users

// UserService handles user-related operations
type UserService struct {
    repo UserRepository
}

// GetByID retrieves a user by ID.
// Returns ErrNotFound if user doesn't exist.
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
    // Implementation
}
```

### Testing

```go
// Good: table-driven tests
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name  string
        email string
        want  bool
    }{
        {"valid email", "user@example.com", true},
        {"missing @", "userexample.com", false},
        {"missing domain", "user@", false},
        {"empty", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ValidateEmail(tt.email)
            if got != tt.want {
                t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.want)
            }
        })
    }
}
```

### Code Organization

```
shared/
├── auth/           # Authentication
│   ├── jwt.go
│   ├── jwt_test.go
│   └── middleware.go
├── database/       # Database utilities
│   ├── connection.go
│   └── query_optimizer.go
├── errors/         # Error handling
│   └── errors.go
└── response/       # API responses
    └── response.go
```

### Linting

Run linter before committing:

```bash
# Run golangci-lint
golangci-lint run

# Auto-fix issues where possible
golangci-lint run --fix

# Run on specific directory
golangci-lint run ./shared/...
```

## TypeScript/React Style Guide

### General Principles

1. Use TypeScript for all new code
2. Use functional components with hooks
3. Follow React best practices
4. Use ESLint and Prettier

### Component Structure

```tsx
// Good: clear, typed, documented
import React, { useState, useEffect } from 'react';

interface UserCardProps {
  userId: string;
  onUserClick?: (id: string) => void;
}

/**
 * UserCard displays user information
 */
export const UserCard: React.FC<UserCardProps> = ({ userId, onUserClick }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchUser(userId).then(setUser).finally(() => setLoading(false));
  }, [userId]);

  if (loading) {
    return <LoadingSpinner />;
  }

  if (!user) {
    return <ErrorMessage>User not found</ErrorMessage>;
  }

  return (
    <div className="user-card" onClick={() => onUserClick?.(userId)}>
      <h3>{user.name}</h3>
      <p>{user.email}</p>
    </div>
  );
};
```

### Naming Conventions

```tsx
// Components: PascalCase
export const UserDashboard: React.FC = () => { /* ... */ };

// Hooks: camelCase, start with "use"
export const useUserData = (userId: string) => { /* ... */ };

// Constants: UPPER_SNAKE_CASE
const MAX_RETRY_ATTEMPTS = 3;
const API_BASE_URL = 'https://api.marimo-erp.com';

// Types/Interfaces: PascalCase
interface UserData {
  id: string;
  name: string;
}

type UserStatus = 'active' | 'inactive' | 'suspended';
```

### State Management

```tsx
// Good: typed state
const [user, setUser] = useState<User | null>(null);
const [loading, setLoading] = useState<boolean>(false);
const [error, setError] = useState<string | null>(null);

// Good: batch state updates
setState((prev) => ({
  ...prev,
  loading: false,
  data: newData,
}));

// Bad: multiple separate setState calls
setLoading(false);
setData(newData);
setError(null);
```

### Hooks

```tsx
// Good: custom hook with proper dependencies
export const useUser = (userId: string) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;

    const fetchUser = async () => {
      try {
        setLoading(true);
        const data = await api.getUser(userId);
        if (!cancelled) {
          setUser(data);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err as Error);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    fetchUser();

    return () => {
      cancelled = true;
    };
  }, [userId]);

  return { user, loading, error };
};
```

### Props

```tsx
// Good: clear interface, optional props marked
interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: 'primary' | 'secondary' | 'danger';
  disabled?: boolean;
  loading?: boolean;
}

export const Button: React.FC<ButtonProps> = ({
  label,
  onClick,
  variant = 'primary',
  disabled = false,
  loading = false,
}) => {
  return (
    <button
      className={`btn btn-${variant}`}
      onClick={onClick}
      disabled={disabled || loading}
    >
      {loading ? <Spinner /> : label}
    </button>
  );
};
```

### API Calls

```tsx
// Good: typed API responses
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

const fetchUsers = async (): Promise<User[]> => {
  const response = await fetch('/api/users');
  const data: ApiResponse<User[]> = await response.json();

  if (!data.success || !data.data) {
    throw new Error(data.error?.message || 'Failed to fetch users');
  }

  return data.data;
};
```

### Styling

```tsx
// Good: CSS modules or styled-components
import styles from './UserCard.module.css';

export const UserCard = () => {
  return <div className={styles.card}>...</div>;
};

// Or with styled-components
import styled from 'styled-components';

const Card = styled.div`
  padding: 1rem;
  border: 1px solid #ddd;
  border-radius: 4px;
`;
```

### Linting

```bash
# Run ESLint
npm run lint

# Auto-fix issues
npm run lint:fix

# Format with Prettier
npm run format

# Type check
npm run type-check
```

## Git Commit Guidelines

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, no code change)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks
- **perf**: Performance improvements

### Examples

```bash
# Good commits
git commit -m "feat(auth): add JWT refresh token support"
git commit -m "fix(users): correct email validation regex"
git commit -m "docs(api): update authentication endpoints"
git commit -m "refactor(cache): simplify Redis connection logic"

# Bad commits
git commit -m "fix stuff"
git commit -m "updates"
git commit -m "WIP"
```

### Detailed Commit

```
feat(analytics): add dashboard widget system

Implement reusable widget system for analytics dashboards:
- Widget base component with common functionality
- Metric, chart, and table widget types
- Drag-and-drop widget positioning
- Widget configuration modal

Closes #123
```

## Code Review Guidelines

### Reviewer Checklist

- [ ] Code follows style guide
- [ ] Tests are included and pass
- [ ] Documentation is updated
- [ ] No security vulnerabilities
- [ ] Error handling is proper
- [ ] Performance considerations addressed
- [ ] Code is maintainable and readable

### Review Comments

```markdown
# Good review comments
- "Consider extracting this into a separate function for better testability"
- "This could cause a memory leak - add cleanup in useEffect"
- "Great use of TypeScript generics here!"

# Bad review comments
- "This is wrong"
- "Rewrite this"
- "I don't like this"
```

## Testing Guidelines

### Unit Tests

```go
// Go: table-driven tests
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   *User
        wantErr bool
    }{
        {"valid user", &User{Email: "user@example.com"}, false},
        {"duplicate email", &User{Email: "existing@example.com"}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.Create(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

```tsx
// React: Jest + React Testing Library
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { UserCard } from './UserCard';

describe('UserCard', () => {
  it('renders user information', async () => {
    render(<UserCard userId="123" />);

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument();
    });
  });

  it('calls onUserClick when clicked', async () => {
    const handleClick = jest.fn();
    render(<UserCard userId="123" onUserClick={handleClick} />);

    await userEvent.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalledWith('123');
  });
});
```

### Test Coverage

Aim for:
- Unit tests: 80%+ coverage
- Integration tests: Critical paths
- E2E tests: Main user workflows

```bash
# Go coverage
go test -cover ./...

# Frontend coverage
npm test -- --coverage
```

## Continuous Integration

All code must pass:

1. Linting (golangci-lint, ESLint)
2. Unit tests
3. Integration tests
4. Security scanning
5. Build verification

## Resources

- [Effective Go](https://golang.org/doc/effective_go.html)
- [React TypeScript Cheatsheet](https://react-typescript-cheatsheet.netlify.app/)
- [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)
- [Conventional Commits](https://www.conventionalcommits.org/)

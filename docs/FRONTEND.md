# Marimo ERP - Frontend Documentation

## Overview

Современный frontend приложения построен на React 18 с TypeScript, использующий лучшие практики и современные инструменты разработки.

## Технологический стек

### Core
- **React 18.2** - UI библиотека
- **TypeScript 5.3** - Типизация
- **React Router DOM 6** - Маршрутизация

### State Management & API
- **React Query (TanStack Query) 5.17** - Server state management
- **Context API** - Client state management
- **Axios** - HTTP клиент

### Forms & Validation
- **React Hook Form 7.49** - Form management
- **Zod 3.22** - Schema validation
- **@hookform/resolvers** - Integration

### i18n & UX
- **i18next 23.7** - Internationalization
- **react-i18next 14.0** - React integration
- **Custom Loading/Skeleton components** - Loading states

### Development
- **React Scripts 5.0** - Build tooling
- **React Query Devtools** - Debugging

## Project Structure

```
frontend/
├── public/
│   └── index.html
├── src/
│   ├── components/         # React components
│   │   ├── ErrorBoundary.tsx
│   │   ├── Loading.tsx
│   │   ├── Loading.css
│   │   ├── Skeleton.tsx
│   │   ├── Skeleton.css
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   ├── Auth.css
│   │   ├── Layout.js       # To be migrated
│   │   └── modules/        # Module components
│   ├── context/            # React contexts
│   │   ├── AuthContext.tsx
│   │   └── ThemeContext.tsx
│   ├── hooks/              # Custom hooks
│   │   ├── useAuth.ts      # Auth operations
│   │   └── useTheme.ts     # Theme management
│   ├── i18n/              # Internationalization
│   │   ├── config.ts
│   │   └── locales/
│   │       ├── en.json
│   │       └── ru.json
│   ├── services/          # API services
│   │   └── api.ts
│   ├── types/             # TypeScript types
│   │   ├── auth.types.ts
│   │   ├── api.types.ts
│   │   ├── theme.types.ts
│   │   └── index.ts
│   ├── utils/             # Utilities
│   │   └── validation.ts  # Zod schemas
│   ├── App.tsx            # Main app component
│   ├── App.css            # Global styles
│   ├── index.js           # Entry point
│   └── index.css          # Base styles
├── tsconfig.json          # TypeScript config
└── package.json           # Dependencies

```

## Features Implemented

### 1. TypeScript Migration ✅

Полная миграция на TypeScript с строгой типизацией:

```typescript
// Type definitions
interface User {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  role?: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  login: (token: string, userData: Partial<User>) => void;
  logout: () => void;
}
```

**Benefits:**
- Catch errors at compile time
- Better IDE support and autocomplete
- Self-documenting code
- Refactoring safety

### 2. React Query Integration ✅

Server state management с automatic caching и refetching:

```typescript
// Custom hooks
export const useLogin = () => {
  return useMutation({
    mutationFn: async (credentials: LoginCredentials) => {
      return authAPI.login(credentials.email, credentials.password);
    },
    onSuccess: (data) => {
      localStorage.setItem('token', data.token);
    },
  });
};

// Usage in components
const loginMutation = useLogin();

const onSubmit = async (data) => {
  await loginMutation.mutateAsync(data);
};
```

**Features:**
- Automatic caching
- Background refetching
- Optimistic updates
- Loading/error states
- Request deduplication
- DevTools integration

### 3. Form Validation (React Hook Form + Zod) ✅

Type-safe form validation с отличным DX:

```typescript
// Schema definition
const loginSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Min 8 characters'),
});

// Form usage
const {
  register,
  handleSubmit,
  formState: { errors },
} = useForm<LoginFormData>({
  resolver: zodResolver(loginSchema),
});
```

**Benefits:**
- Type-safe forms
- Declarative validation
- Better performance (minimal re-renders)
- Easy error handling
- Built-in validation messages

### 4. Error Boundaries ✅

Graceful error handling с fallback UI:

```typescript
<ErrorBoundary fallback={<CustomError />}>
  <App />
</ErrorBoundary>
```

**Features:**
- Catch React errors
- Display user-friendly messages
- Error logging capability
- Development mode details
- Recovery mechanisms

### 5. Loading States & Skeleton Screens ✅

Better UX с loading indicators:

```typescript
// Loading spinner
<Loading size="medium" fullScreen message="Loading data..." />

// Skeleton screens
<SkeletonCard />
<SkeletonTable rows={5} />
<SkeletonList items={10} />
```

**Types:**
- Spinner (small, medium, large)
- Skeleton (text, circular, rectangular)
- Pre-built components (Card, Table, List)
- Pulse and wave animations

### 6. Dark Mode ✅

System-aware theme switching:

```typescript
const { theme, effectiveTheme, setTheme, toggleTheme } = useTheme();

// Themes: 'light' | 'dark' | 'system'
setTheme('dark');
toggleTheme();
```

**Features:**
- Light/Dark/System modes
- Smooth transitions
- Persistent preference
- CSS variables
- System preference detection

### 7. Internationalization (i18n) ✅

Multi-language support:

```typescript
const { t, i18n } = useTranslation();

<h1>{t('auth.login')}</h1>
<p>{t('auth.welcome', { name: user.name })}</p>

i18n.changeLanguage('ru');
```

**Supported Languages:**
- English (en)
- Russian (ru)

**Features:**
- Easy to add new languages
- Namespace support
- Interpolation
- Pluralization ready
- Persistent language choice

## API Client

### Configuration

```typescript
// Auto-configured axios instance
const apiClient = axios.create({
  baseURL: process.env.REACT_APP_API_URL || '/api',
  timeout: 30000,
});

// Request interceptor - adds auth token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor - handles token refresh
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Try to refresh token
      // Retry original request
    }
    return Promise.reject(error);
  }
);
```

### API Methods

```typescript
// Auth
await authAPI.login(email, password);
await authAPI.register(credentials);
await authAPI.logout();
await authAPI.refreshToken(refreshToken);

// Generic
await api.get<User[]>('/users');
await api.post<User>('/users', userData);
await api.put<User>(`/users/${id}`, updates);
await api.delete(`/users/${id}`);
```

## Component Examples

### Login Form with Validation

```typescript
const Login: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const loginMutation = useLogin();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormData) => {
    const response = await loginMutation.mutateAsync(data);
    if (response.success) {
      navigate('/dashboard');
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('email')} />
      {errors.email && <span>{t(errors.email.message!)}</span>}

      <button disabled={loginMutation.isPending}>
        {loginMutation.isPending ? <Loading size="small" /> : t('auth.login')}
      </button>
    </form>
  );
};
```

### Protected Route

```typescript
const ProtectedRoute: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { token, loading } = useAuth();

  if (loading) {
    return <Loading fullScreen />;
  }

  return token ? <Layout>{children}</Layout> : <Navigate to="/login" />;
};
```

### Data Fetching with React Query

```typescript
const useUsers = () => {
  return useQuery({
    queryKey: ['users'],
    queryFn: () => api.get<User[]>('/users'),
    staleTime: 5 * 60 * 1000,
  });
};

const UsersList = () => {
  const { data: users, isLoading, error } = useUsers();

  if (isLoading) return <SkeletonTable />;
  if (error) return <ErrorMessage error={error} />;

  return (
    <table>
      {users?.map(user => (
        <tr key={user.id}>
          <td>{user.email}</td>
        </tr>
      ))}
    </table>
  );
};
```

## Styling

### CSS Variables (Theme Support)

```css
/* Light theme */
:root {
  --primary-color: #4f46e5;
  --background: #f9fafb;
  --text-primary: #111827;
  --border-color: #e5e7eb;
}

/* Dark theme */
[data-theme='dark'] {
  --primary-color: #6366f1;
  --background: #111827;
  --text-primary: #f9fafb;
  --border-color: #374151;
}
```

### Component Styles

- Modular CSS files
- CSS variables for theming
- Responsive design
- Smooth transitions
- Consistent spacing

## Development

### Installation

```bash
cd frontend
npm install
```

### Running

```bash
# Development server
npm start

# Production build
npm run build

# Tests
npm test
```

### Environment Variables

```env
REACT_APP_API_URL=http://localhost:8080/api
NODE_ENV=development
```

## Best Practices

### 1. Component Structure

```typescript
// Component with proper typing
import React from 'react';
import { SomeProps } from '../types';

interface ComponentProps {
  title: string;
  onAction?: () => void;
}

const Component: React.FC<ComponentProps> = ({ title, onAction }) => {
  return <div>{title}</div>;
};

export default Component;
```

### 2. Custom Hooks

```typescript
// Reusable hook
export const useDebounce = <T,>(value: T, delay: number): T => {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const handler = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(handler);
  }, [value, delay]);

  return debouncedValue;
};
```

### 3. Error Handling

```typescript
try {
  const data = await api.get('/endpoint');
  // Handle success
} catch (error) {
  if (error instanceof ApiError) {
    // Handle API error
    toast.error(error.message);
  } else {
    // Handle unexpected error
    console.error(error);
  }
}
```

### 4. Performance

- Use `React.memo` for expensive components
- Implement code splitting with `React.lazy`
- Optimize re-renders with `useCallback`/`useMemo`
- Use React Query for caching
- Implement virtual scrolling for long lists

## Migration Guide

### From JavaScript to TypeScript

1. Rename `.js` files to `.tsx`
2. Add type annotations
3. Define interfaces for props
4. Fix type errors
5. Enable strict mode

### Adding New Features

1. Define types in `types/`
2. Create API methods in `services/`
3. Build React Query hooks in `hooks/`
4. Create components in `components/`
5. Add translations in `i18n/locales/`

## Testing (Future)

```typescript
// Component testing
import { render, screen } from '@testing-library/react';
import Login from './Login';

test('renders login form', () => {
  render(<Login />);
  expect(screen.getByRole('button', { name: /login/i })).toBeInTheDocument();
});

// Hook testing
import { renderHook } from '@testing-library/react-hooks';
import { useAuth } from './useAuth';

test('login mutation', async () => {
  const { result } = renderHook(() => useAuth());
  await result.current.login.mutateAsync(credentials);
  expect(result.current.isAuthenticated).toBe(true);
});
```

## Performance Optimization

### Bundle Analysis

```bash
npm run build
npx webpack-bundle-analyzer build/static/js/*.js
```

### Code Splitting

```typescript
const Dashboard = React.lazy(() => import('./components/Dashboard'));

<Suspense fallback={<Loading fullScreen />}>
  <Dashboard />
</Suspense>
```

### Image Optimization

- Use WebP format
- Implement lazy loading
- Responsive images

### Caching Strategy

- React Query for API data
- Service Worker for assets
- LocalStorage for preferences

## Deployment

### Build for Production

```bash
npm run build
```

### Docker Deployment

```dockerfile
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Roadmap

### Phase 1 (Complete)
- ✅ TypeScript migration
- ✅ React Query integration
- ✅ Form validation (React Hook Form + Zod)
- ✅ Error boundaries
- ✅ Loading states & skeleton screens
- ✅ Dark mode
- ✅ i18n

### Phase 2 (Planned)
- [ ] Unit tests (Jest + React Testing Library)
- [ ] E2E tests (Cypress/Playwright)
- [ ] Storybook for components
- [ ] Performance monitoring
- [ ] PWA support
- [ ] Accessibility improvements (WCAG 2.1)

### Phase 3 (Future)
- [ ] GraphQL integration
- [ ] WebSocket support
- [ ] Advanced animations (Framer Motion)
- [ ] Design system
- [ ] Mobile responsive improvements

## Troubleshooting

### Common Issues

**TypeScript errors after update:**
```bash
rm -rf node_modules package-lock.json
npm install
```

**Build fails:**
- Check Node version (requires 18+)
- Clear cache: `npm cache clean --force`
- Delete build folder

**API calls failing:**
- Check REACT_APP_API_URL
- Verify CORS settings
- Check network tab in DevTools

## Resources

- [React Documentation](https://react.dev)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [React Query Docs](https://tanstack.com/query/latest)
- [React Hook Form](https://react-hook-form.com)
- [Zod Documentation](https://zod.dev)

## Support

For issues and questions:
- GitHub Issues
- Documentation: https://docs.marimo.dev/frontend

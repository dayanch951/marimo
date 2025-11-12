# Marimo ERP Mobile App

React Native mobile application for Marimo ERP system.

## Features

- 🔐 Authentication (Login, Register, Biometric)
- 📊 Dashboard with real-time stats
- 👥 User management
- 📈 Analytics and reports
- 📱 Push notifications
- 🌐 Offline mode support
- 🎨 Modern UI with dark mode
- 🔄 Real-time updates via WebSocket
- 📤 Export data (CSV, Excel, PDF)

## Tech Stack

- **Framework**: React Native 0.73
- **Language**: TypeScript
- **State Management**: React Query (TanStack Query)
- **Navigation**: React Navigation
- **Forms**: React Hook Form + Zod
- **HTTP Client**: Axios
- **Storage**: AsyncStorage
- **UI**: React Native Paper / Custom Components

## Prerequisites

- Node.js >= 18
- React Native CLI
- Xcode (for iOS)
- Android Studio (for Android)

## Installation

```bash
# Install dependencies
npm install

# iOS only - install pods
cd ios && pod install && cd ..

# Start Metro bundler
npm start

# Run on Android
npm run android

# Run on iOS
npm run ios
```

## Project Structure

```
mobile/
├── src/
│   ├── components/       # Reusable components
│   │   ├── common/       # Common UI components
│   │   └── forms/        # Form components
│   ├── screens/          # Screen components
│   │   ├── auth/         # Authentication screens
│   │   ├── dashboard/    # Dashboard screen
│   │   ├── users/        # User management
│   │   └── reports/      # Reports & analytics
│   ├── navigation/       # Navigation configuration
│   │   ├── AuthNavigator.tsx
│   │   ├── MainNavigator.tsx
│   │   └── RootNavigator.tsx
│   ├── services/         # API services
│   │   ├── authService.ts
│   │   ├── userService.ts
│   │   └── analyticsService.ts
│   ├── hooks/            # Custom hooks
│   │   ├── useAuth.ts
│   │   ├── useWebSocket.ts
│   │   └── useOffline.ts
│   ├── utils/            # Utility functions
│   │   ├── validation.ts
│   │   ├── formatting.ts
│   │   └── storage.ts
│   ├── types/            # TypeScript types
│   │   ├── auth.types.ts
│   │   ├── user.types.ts
│   │   └── api.types.ts
│   ├── config/           # Configuration
│   │   ├── api.ts
│   │   └── constants.ts
│   └── theme/            # Theme configuration
│       ├── colors.ts
│       └── styles.ts
├── android/              # Android native code
├── ios/                  # iOS native code
├── __tests__/            # Tests
├── package.json
└── tsconfig.json
```

## Configuration

Create `.env` file:

```env
API_URL=https://api.marimo-erp.com
WS_URL=wss://api.marimo-erp.com/ws
ENVIRONMENT=development
```

## Available Scripts

```bash
# Development
npm start              # Start Metro bundler
npm run android        # Run on Android
npm run ios            # Run on iOS

# Testing
npm test               # Run tests
npm run test:watch     # Run tests in watch mode
npm run test:coverage  # Generate coverage report

# Linting
npm run lint           # Run ESLint
npm run lint:fix       # Fix ESLint errors

# Type checking
npm run type-check     # Run TypeScript compiler

# Build
# Android
cd android && ./gradlew assembleRelease

# iOS
cd ios && xcodebuild -workspace MarimoMobile.xcworkspace -scheme MarimoMobile -configuration Release
```

## Key Features Implementation

### Authentication

```typescript
import { authService } from '@/services/authService';

// Login
const response = await authService.login({
  email: 'user@example.com',
  password: 'password',
});

// Register
await authService.register({
  email: 'user@example.com',
  password: 'password',
  name: 'John Doe',
});

// Logout
await authService.logout();
```

### API Calls with React Query

```typescript
import { useQuery } from '@tanstack/react-query';
import apiClient from '@/config/api';

const { data, isLoading, error } = useQuery({
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

### Offline Support

```typescript
import { useOffline } from '@/hooks/useOffline';

const { isOffline, queuedRequests } = useOffline();

if (isOffline) {
  // Queue request for when online
  queueRequest({ method: 'POST', url: '/api/data', data });
}
```

### Push Notifications

```typescript
import messaging from '@react-native-firebase/messaging';

// Request permission
await messaging().requestPermission();

// Get FCM token
const token = await messaging().getToken();

// Listen for messages
messaging().onMessage(async (remoteMessage) => {
  console.log('Notification:', remoteMessage);
});
```

## Testing

```bash
# Run all tests
npm test

# Run specific test file
npm test -- LoginScreen.test.tsx

# Update snapshots
npm test -- -u
```

## Debugging

### React Native Debugger

1. Install React Native Debugger
2. Run app in debug mode
3. Open debugger: Cmd+D (iOS) / Cmd+M (Android)
4. Select "Debug"

### Flipper

Flipper is enabled by default in development builds.

## Deployment

### Android

```bash
cd android
./gradlew bundleRelease
```

Upload `android/app/build/outputs/bundle/release/app-release.aab` to Google Play Console.

### iOS

```bash
cd ios
xcodebuild -workspace MarimoMobile.xcworkspace -scheme MarimoMobile -configuration Release archive
```

Use Xcode to upload to App Store Connect.

## Performance Optimization

- Use `React.memo` for expensive components
- Implement list virtualization with `FlatList`
- Optimize images with `react-native-fast-image`
- Use Hermes engine (enabled by default)
- Profile with Flipper

## Security

- Store sensitive data in Keychain (iOS) / Keystore (Android)
- Use SSL pinning for API calls
- Implement biometric authentication
- Validate all user input
- Use HTTPS only

## Troubleshooting

### Common Issues

**Metro bundler won't start:**
```bash
npm start -- --reset-cache
```

**Build fails:**
```bash
# Clean Android
cd android && ./gradlew clean && cd ..

# Clean iOS
cd ios && pod deintegrate && pod install && cd ..
```

**App crashes on launch:**
- Check native logs
- Verify dependencies are installed
- Clear cache and rebuild

## Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## License

MIT

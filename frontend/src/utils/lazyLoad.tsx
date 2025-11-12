import React, { lazy, Suspense, ComponentType } from 'react';

/**
 * Loading component shown while lazy component loads
 */
interface LoadingSpinnerProps {
  fullScreen?: boolean;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ fullScreen = false }) => {
  const containerStyle = fullScreen
    ? {
        position: 'fixed' as const,
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'rgba(255, 255, 255, 0.9)',
        zIndex: 9999,
      }
    : {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '2rem',
      };

  return (
    <div style={containerStyle}>
      <div className="spinner" style={{
        border: '4px solid #f3f3f3',
        borderTop: '4px solid #3498db',
        borderRadius: '50%',
        width: '40px',
        height: '40px',
        animation: 'spin 1s linear infinite',
      }} />
    </div>
  );
};

/**
 * Retry loading component with error boundary
 */
interface RetryableComponentProps {
  error?: Error;
  retry?: () => void;
}

const ErrorFallback: React.FC<RetryableComponentProps> = ({ error, retry }) => {
  return (
    <div style={{
      padding: '2rem',
      textAlign: 'center',
      backgroundColor: '#fff3cd',
      border: '1px solid #ffc107',
      borderRadius: '4px',
      margin: '1rem',
    }}>
      <h3>Failed to load component</h3>
      <p style={{ color: '#666' }}>{error?.message || 'Unknown error'}</p>
      {retry && (
        <button
          onClick={retry}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: '#3498db',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            marginTop: '1rem',
          }}
        >
          Retry
        </button>
      )}
    </div>
  );
};

/**
 * Lazy load component with automatic retry
 */
interface LazyOptions {
  fallback?: React.ReactNode;
  delay?: number;
  maxRetries?: number;
}

export function lazyLoadComponent<T extends ComponentType<any>>(
  importFunc: () => Promise<{ default: T }>,
  options: LazyOptions = {}
): React.LazyExoticComponent<T> {
  const {
    fallback = <LoadingSpinner />,
    delay = 200,
    maxRetries = 3,
  } = options;

  let retries = 0;

  const retry = (): Promise<{ default: T }> => {
    return new Promise((resolve, reject) => {
      importFunc()
        .then(resolve)
        .catch((error) => {
          if (retries < maxRetries) {
            retries++;
            console.log(`Retrying component load (${retries}/${maxRetries})...`);
            setTimeout(() => {
              retry().then(resolve).catch(reject);
            }, 1000 * retries); // Exponential backoff
          } else {
            reject(error);
          }
        });
    });
  };

  return lazy(retry);
}

/**
 * Preload lazy component
 */
export const preloadComponent = <T extends ComponentType<any>>(
  lazyComponent: React.LazyExoticComponent<T>
): void => {
  // @ts-ignore - accessing internal preload
  if (lazyComponent._payload && lazyComponent._payload._status === -1) {
    // @ts-ignore
    lazyComponent._payload._result();
  }
};

/**
 * Lazy route component with code splitting
 */
interface LazyRouteProps {
  component: React.LazyExoticComponent<any>;
  fallback?: React.ReactNode;
}

export const LazyRoute: React.FC<LazyRouteProps> = ({
  component: Component,
  fallback = <LoadingSpinner fullScreen />,
}) => {
  return (
    <Suspense fallback={fallback}>
      <Component />
    </Suspense>
  );
};

/**
 * Route-based code splitting
 */

// Lazy load pages
export const DashboardPage = lazyLoadComponent(
  () => import('../pages/Dashboard'),
  { fallback: <LoadingSpinner fullScreen /> }
);

export const UsersPage = lazyLoadComponent(
  () => import('../pages/Users'),
  { fallback: <LoadingSpinner fullScreen /> }
);

export const AnalyticsPage = lazyLoadComponent(
  () => import('../pages/Analytics'),
  { fallback: <LoadingSpinner fullScreen /> }
);

export const SettingsPage = lazyLoadComponent(
  () => import('../pages/Settings'),
  { fallback: <LoadingSpinner fullScreen /> }
);

export const ReportsPage = lazyLoadComponent(
  () => import('../pages/Reports'),
  { fallback: <LoadingSpinner fullScreen /> }
);

/**
 * Component-based code splitting for heavy components
 */
export const ChartComponent = lazyLoadComponent(
  () => import('../components/Chart')
);

export const DataTableComponent = lazyLoadComponent(
  () => import('../components/DataTable')
);

export const RichTextEditor = lazyLoadComponent(
  () => import('../components/RichTextEditor')
);

/**
 * Prefetch routes on hover
 */
export const usePrefetchRoute = (routePath: string) => {
  const prefetch = React.useCallback(() => {
    // Get route component and preload it
    // Implementation depends on routing library
    console.log(`Prefetching route: ${routePath}`);
  }, [routePath]);

  return prefetch;
};

/**
 * Link component with prefetching
 */
interface PrefetchLinkProps {
  to: string;
  children: React.ReactNode;
  className?: string;
  prefetchComponent?: React.LazyExoticComponent<any>;
}

export const PrefetchLink: React.FC<PrefetchLinkProps> = ({
  to,
  children,
  className,
  prefetchComponent,
}) => {
  const handleMouseEnter = () => {
    if (prefetchComponent) {
      preloadComponent(prefetchComponent);
    }
  };

  return (
    <a
      href={to}
      className={className}
      onMouseEnter={handleMouseEnter}
      onTouchStart={handleMouseEnter}
    >
      {children}
    </a>
  );
};

/**
 * Dynamic import with timeout
 */
export const dynamicImportWithTimeout = <T extends any>(
  importFunc: () => Promise<T>,
  timeout: number = 10000
): Promise<T> => {
  return Promise.race([
    importFunc(),
    new Promise<T>((_, reject) =>
      setTimeout(() => reject(new Error('Import timeout')), timeout)
    ),
  ]);
};

/**
 * Intersection Observer based lazy loading for components
 */
interface LazyRenderProps {
  children: React.ReactNode;
  placeholder?: React.ReactNode;
  threshold?: number;
  rootMargin?: string;
}

export const LazyRender: React.FC<LazyRenderProps> = ({
  children,
  placeholder = <div style={{ minHeight: '200px' }} />,
  threshold = 0.1,
  rootMargin = '100px',
}) => {
  const [isVisible, setIsVisible] = React.useState(false);
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const element = ref.current;
    if (!element || !('IntersectionObserver' in window)) {
      setIsVisible(true);
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setIsVisible(true);
            observer.unobserve(entry.target);
          }
        });
      },
      { threshold, rootMargin }
    );

    observer.observe(element);

    return () => {
      if (element) {
        observer.unobserve(element);
      }
    };
  }, [threshold, rootMargin]);

  return <div ref={ref}>{isVisible ? children : placeholder}</div>;
};

/**
 * Bundle analyzer helpers
 */
export const logComponentLoad = (componentName: string) => {
  if (process.env.NODE_ENV === 'development') {
    console.log(`[Lazy Load] ${componentName} loaded at ${new Date().toISOString()}`);
  }
};

/**
 * Webpack magic comments for chunk naming
 * Usage:
 * const MyComponent = lazy(() => import(
 *   webpackChunkName: "my-component"
 *   webpackPrefetch: true
 *   './MyComponent'
 * ));
 */

/**
 * Performance monitoring for lazy loaded components
 */
export const withLoadTracking = <P extends object>(
  Component: React.ComponentType<P>,
  componentName: string
) => {
  return (props: P) => {
    React.useEffect(() => {
      if (window.performance && window.performance.mark) {
        performance.mark(`${componentName}-render-start`);

        return () => {
          performance.mark(`${componentName}-render-end`);
          performance.measure(
            `${componentName}-render`,
            `${componentName}-render-start`,
            `${componentName}-render-end`
          );

          const measure = performance.getEntriesByName(`${componentName}-render`)[0];
          console.log(`${componentName} render time: ${measure.duration}ms`);
        };
      }
    }, []);

    return <Component {...props} />;
  };
};

/**
 * Code splitting by feature
 */
export const featureFlags = {
  analytics: true,
  reports: true,
  advancedSettings: false,
};

export const loadFeature = (featureName: keyof typeof featureFlags) => {
  if (!featureFlags[featureName]) {
    return Promise.resolve({ default: () => <div>Feature not available</div> });
  }

  switch (featureName) {
    case 'analytics':
      return import('../features/Analytics');
    case 'reports':
      return import('../features/Reports');
    case 'advancedSettings':
      return import('../features/AdvancedSettings');
    default:
      return Promise.reject(new Error('Unknown feature'));
  }
};

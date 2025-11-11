import React, { Component, ErrorInfo, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

class ErrorBoundaryClass extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
      errorInfo: null,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);

    this.setState({
      error,
      errorInfo,
    });

    // Here you can log the error to an error reporting service
    // logErrorToService(error, errorInfo);
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <ErrorFallback
          error={this.state.error}
          errorInfo={this.state.errorInfo}
          onReset={this.handleReset}
        />
      );
    }

    return this.props.children;
  }
}

interface ErrorFallbackProps {
  error: Error | null;
  errorInfo: ErrorInfo | null;
  onReset: () => void;
}

const ErrorFallback: React.FC<ErrorFallbackProps> = ({ error, errorInfo, onReset }) => {
  const { t } = useTranslation();

  return (
    <div className="error-boundary-container">
      <div className="error-boundary-content">
        <div className="error-boundary-icon">⚠️</div>
        <h1 className="error-boundary-title">{t('errors.somethingWentWrong')}</h1>
        <p className="error-boundary-message">
          {error?.message || t('errors.serverError')}
        </p>

        {process.env.NODE_ENV === 'development' && errorInfo && (
          <details className="error-boundary-details">
            <summary>Error Details</summary>
            <pre className="error-boundary-stack">
              {error?.stack}
              {'\n\n'}
              {errorInfo.componentStack}
            </pre>
          </details>
        )}

        <div className="error-boundary-actions">
          <button onClick={onReset} className="btn btn-primary">
            {t('errors.tryAgain')}
          </button>
          <button onClick={() => window.location.href = '/'} className="btn btn-secondary">
            {t('errors.goHome')}
          </button>
        </div>
      </div>
    </div>
  );
};

// Functional wrapper component to use hooks
const ErrorBoundary: React.FC<Props> = (props) => {
  return <ErrorBoundaryClass {...props} />;
};

export default ErrorBoundary;

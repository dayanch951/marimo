import React from 'react';
import { useTranslation } from 'react-i18next';
import './Loading.css';

interface LoadingProps {
  size?: 'small' | 'medium' | 'large';
  fullScreen?: boolean;
  message?: string;
}

const Loading: React.FC<LoadingProps> = ({
  size = 'medium',
  fullScreen = false,
  message
}) => {
  const { t } = useTranslation();

  const sizeClasses = {
    small: 'loading-spinner-small',
    medium: 'loading-spinner-medium',
    large: 'loading-spinner-large',
  };

  const spinner = (
    <div className="loading-container">
      <div className={`loading-spinner ${sizeClasses[size]}`}>
        <div className="spinner"></div>
      </div>
      {message && <p className="loading-message">{message}</p>}
      {!message && <p className="loading-message">{t('common.loading')}</p>}
    </div>
  );

  if (fullScreen) {
    return (
      <div className="loading-fullscreen">
        {spinner}
      </div>
    );
  }

  return spinner;
};

export default Loading;

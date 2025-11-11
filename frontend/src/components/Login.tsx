import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useTranslation } from 'react-i18next';
import { useLogin } from '../hooks/useAuth';
import { loginSchema, LoginFormData } from '../utils/validation';
import Loading from './Loading';
import './Auth.css';

const Login: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const loginMutation = useLogin();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormData) => {
    try {
      const response = await loginMutation.mutateAsync(data);

      if (response.success) {
        navigate('/dashboard');
      }
    } catch (error: any) {
      console.error('Login error:', error);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h2>{t('auth.login')}</h2>
        </div>

        {loginMutation.isError && (
          <div className="error-message">
            {(loginMutation.error as any)?.message || t('auth.invalidCredentials')}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="auth-form">
          <div className="form-group">
            <label htmlFor="email">{t('auth.email')}</label>
            <input
              type="email"
              id="email"
              {...register('email')}
              placeholder={t('auth.email')}
              className={errors.email ? 'input-error' : ''}
              disabled={isSubmitting}
            />
            {errors.email && (
              <span className="field-error">{t(errors.email.message!)}</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="password">{t('auth.password')}</label>
            <input
              type="password"
              id="password"
              {...register('password')}
              placeholder={t('auth.password')}
              className={errors.password ? 'input-error' : ''}
              disabled={isSubmitting}
            />
            {errors.password && (
              <span className="field-error">{t(errors.password.message!)}</span>
            )}
          </div>

          <button
            type="submit"
            className="btn btn-primary btn-block"
            disabled={isSubmitting || loginMutation.isPending}
          >
            {isSubmitting || loginMutation.isPending ? (
              <Loading size="small" />
            ) : (
              t('auth.login')
            )}
          </button>
        </form>

        <div className="auth-footer">
          <p>
            {t('auth.dontHaveAccount')}{' '}
            <Link to="/register">{t('auth.registerHere')}</Link>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;

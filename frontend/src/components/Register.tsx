import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useTranslation } from 'react-i18next';
import { useRegister } from '../hooks/useAuth';
import { registerSchema, RegisterFormData } from '../utils/validation';
import Loading from './Loading';
import './Auth.css';

const Register: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const registerMutation = useRegister();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterFormData) => {
    try {
      const { confirmPassword, ...credentials } = data;
      const response = await registerMutation.mutateAsync(credentials);

      if (response.success) {
        navigate('/login');
      }
    } catch (error: any) {
      console.error('Registration error:', error);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h2>{t('auth.register')}</h2>
        </div>

        {registerMutation.isError && (
          <div className="error-message">
            {(registerMutation.error as any)?.message || 'Registration failed'}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="auth-form">
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="firstName">{t('auth.firstName')}</label>
              <input
                type="text"
                id="firstName"
                {...register('firstName')}
                placeholder={t('auth.firstName')}
                disabled={isSubmitting}
              />
            </div>

            <div className="form-group">
              <label htmlFor="lastName">{t('auth.lastName')}</label>
              <input
                type="text"
                id="lastName"
                {...register('lastName')}
                placeholder={t('auth.lastName')}
                disabled={isSubmitting}
              />
            </div>
          </div>

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
              <span className="field-error">{errors.password.message}</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword">{t('auth.confirmPassword')}</label>
            <input
              type="password"
              id="confirmPassword"
              {...register('confirmPassword')}
              placeholder={t('auth.confirmPassword')}
              className={errors.confirmPassword ? 'input-error' : ''}
              disabled={isSubmitting}
            />
            {errors.confirmPassword && (
              <span className="field-error">{t(errors.confirmPassword.message!)}</span>
            )}
          </div>

          <button
            type="submit"
            className="btn btn-primary btn-block"
            disabled={isSubmitting || registerMutation.isPending}
          >
            {isSubmitting || registerMutation.isPending ? (
              <Loading size="small" />
            ) : (
              t('auth.register')
            )}
          </button>
        </form>

        <div className="auth-footer">
          <p>
            {t('auth.alreadyHaveAccount')}{' '}
            <Link to="/login">{t('auth.loginHere')}</Link>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Register;

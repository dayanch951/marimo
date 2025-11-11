import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { authAPI } from '../services/api';
import { LoginCredentials, RegisterCredentials, AuthResponse, User } from '../types/auth.types';

export const useLogin = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (credentials: LoginCredentials) => {
      const response = await authAPI.login(credentials.email, credentials.password);
      return response;
    },
    onSuccess: (data) => {
      if (data.success && data.token) {
        localStorage.setItem('token', data.token);
        if (data.refresh_token) {
          localStorage.setItem('refreshToken', data.refresh_token);
        }
        queryClient.invalidateQueries({ queryKey: ['currentUser'] });
      }
    },
  });
};

export const useRegister = () => {
  return useMutation({
    mutationFn: async (credentials: RegisterCredentials) => {
      return authAPI.register(credentials);
    },
  });
};

export const useLogout = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await authAPI.logout();
    },
    onSuccess: () => {
      localStorage.removeItem('token');
      localStorage.removeItem('refreshToken');
      queryClient.clear();
    },
  });
};

export const useCurrentUser = () => {
  return useQuery({
    queryKey: ['currentUser'],
    queryFn: () => authAPI.getCurrentUser(),
    enabled: !!localStorage.getItem('token'),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: false,
  });
};

import AsyncStorage from '@react-native-async-storage/async-storage';
import apiClient from '@/config/api';

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  email: string;
  password: string;
  name: string;
  company?: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: {
    id: string;
    email: string;
    name: string;
    tenant_id: string;
  };
}

export const authService = {
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/login', credentials);
    const { access_token, refresh_token, user } = response.data;

    // Store tokens
    await AsyncStorage.multiSet([
      ['auth_token', access_token],
      ['refresh_token', refresh_token],
      ['tenant_id', user.tenant_id],
      ['user', JSON.stringify(user)],
    ]);

    return response.data;
  },

  async register(credentials: RegisterCredentials): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/register', credentials);
    const { access_token, refresh_token, user } = response.data;

    await AsyncStorage.multiSet([
      ['auth_token', access_token],
      ['refresh_token', refresh_token],
      ['tenant_id', user.tenant_id],
      ['user', JSON.stringify(user)],
    ]);

    return response.data;
  },

  async logout(): Promise<void> {
    try {
      await apiClient.post('/auth/logout');
    } catch (error) {
      // Logout even if request fails
    } finally {
      await AsyncStorage.multiRemove(['auth_token', 'refresh_token', 'tenant_id', 'user']);
    }
  },

  async getCurrentUser() {
    const userStr = await AsyncStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  },

  async isAuthenticated(): Promise<boolean> {
    const token = await AsyncStorage.getItem('auth_token');
    return !!token;
  },
};

export default authService;

import { createContext, useContext, useState, useEffect } from 'react';
import { authService } from '../services/api.service';
import toast from 'react-hot-toast';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    const token = localStorage.getItem('access_token');
    if (token) {
      try {
        const response = await authService.getMe();
        // Handle both nested and flat response structures
        const userData = response.data?.user || response.data;
        setUser(userData);
        setIsAuthenticated(true);
        console.log('[Auth] User authenticated:', userData?.email);
      } catch (error) {
        console.error('[Auth] Check auth failed:', error);
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        setIsAuthenticated(false);
      }
    }
    setLoading(false);
  };

  const login = async (email, password) => {
    try {
      console.log('[Auth] Attempting login for:', email);
      const response = await authService.login(email, password);
      const data = response.data;
      
      // Extract tokens and user from response
      const accessToken = data.access_token;
      const refreshToken = data.refresh_token;
      const userData = data.user;
      
      if (!accessToken) {
        throw new Error('No access token received');
      }
      
      localStorage.setItem('access_token', accessToken);
      if (refreshToken) {
        localStorage.setItem('refresh_token', refreshToken);
      }
      
      setUser(userData);
      setIsAuthenticated(true);
      console.log('[Auth] Login successful for:', userData?.email);
      toast.success('Login successful!');
      return true;
    } catch (error) {
      console.error('[Auth] Login failed:', error);
      const message = error.response?.data?.message || error.response?.data?.error || 'Login failed. Please check your credentials.';
      toast.error(message);
      return false;
    }
  };

  const register = async (data) => {
    try {
      console.log('[Auth] Attempting registration for:', data.email);
      const response = await authService.register(data);
      const responseData = response.data;
      
      const accessToken = responseData.access_token;
      const refreshToken = responseData.refresh_token;
      const userData = responseData.user;
      
      if (!accessToken) {
        throw new Error('No access token received');
      }
      
      localStorage.setItem('access_token', accessToken);
      if (refreshToken) {
        localStorage.setItem('refresh_token', refreshToken);
      }
      
      setUser(userData);
      setIsAuthenticated(true);
      console.log('[Auth] Registration successful for:', userData?.email);
      toast.success('Registration successful!');
      return true;
    } catch (error) {
      console.error('[Auth] Registration failed:', error);
      const message = error.response?.data?.message || error.response?.data?.error || 'Registration failed';
      toast.error(message);
      return false;
    }
  };

  const logout = async () => {
    try {
      await authService.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      setUser(null);
      setIsAuthenticated(false);
      toast.success('Logged out successfully');
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        isAuthenticated,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

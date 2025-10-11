import { useAppDispatch, useAppSelector } from "../store/hooks";
import {
  login,
  register,
  logout,
  refreshUser,
  clearError as clearAuthError,
  selectAuthState,
} from "../store/slices/authSlice";
import { LoginRequest, RegisterRequest, UserProfile } from "../types";

export const useAuth = () => {
  const dispatch = useAppDispatch();
  const auth = useAppSelector(selectAuthState);

  const handleLogin = async (credentials: LoginRequest) => {
    await dispatch(login(credentials)).unwrap();
  };

  const handleRegister = async (userData: RegisterRequest) => {
    await dispatch(register(userData)).unwrap();
  };

  const handleLogout = async () => {
    await dispatch(logout()).unwrap();
  };

  const handleRefreshUser = async () => {
    await dispatch(refreshUser()).unwrap();
  };

  const clearError = () => {
    dispatch(clearAuthError());
  };

  return {
    user: auth.user,
    token: auth.token,
    loading: auth.loading,
    error: auth.error,
    isAuthenticated: auth.isAuthenticated,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
    refreshUser: handleRefreshUser,
    clearError,
  };
};

export interface UseAuthReturn {
  user: UserProfile | null;
  token: string | null;
  loading: boolean;
  error: string | null;
  isAuthenticated: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  clearError: () => void;
}

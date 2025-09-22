import React, { createContext, useState, useEffect, ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { message } from 'antd';
import { getCurrentUser } from '@/services/auth';
import { getCurrentUserRoles, getCurrentUserPermissions } from '@/services/role';
import { getToken, setToken, removeToken, setUserInfo, getUserInfo, removeUserInfo, setPermissions, getPermissions, setRoles, getRoles, removePermissions, removeRoles } from '@/utils/storage';
import type { Role, Permission } from '@/services/role';

interface User {
  id: number;
  username: string;
  real_name: string;
  email: string;
  role: string;
  department: string;
  [key: string]: any;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  roles: Role[];
  permissions: Permission[];
  loading: boolean;
  login: (token: string, userInfo: User) => void;
  logout: () => void;
  updateUser: (user: User) => void;
  hasPermission: (permissionCode: string) => boolean;
  hasRole: (roleCode: string) => boolean;
  hasAnyRole: (roleCodes: string[]) => boolean;
  hasAllRoles: (roleCodes: string[]) => boolean;
  refreshPermissions: () => Promise<void>;
}

const defaultContext: AuthContextType = {
  user: null,
  token: null,
  roles: [],
  permissions: [],
  loading: true,
  login: () => {},
  logout: () => {},
  updateUser: () => {},
  hasPermission: () => false,
  hasRole: () => false,
  hasAnyRole: () => false,
  hasAllRoles: () => false,
  refreshPermissions: async () => {}
};

export const AuthContext = createContext<AuthContextType>(defaultContext);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setTokenState] = useState<string | null>(null);
  const [roles, setRolesState] = useState<Role[]>([]);
  const [permissions, setPermissionsState] = useState<Permission[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const navigate = useNavigate();

  // 加载用户角色和权限
  const loadUserRolesAndPermissions = async () => {
    try {
      const [userRoles, userPermissions] = await Promise.all([
        getCurrentUserRoles(),
        getCurrentUserPermissions()
      ]);
      
      setRolesState(userRoles);
      setPermissionsState(userPermissions);
      // 这些localStorage调用可能会导致问题，暂时注释掉
      // setRoles(userRoles);
      // setPermissions(userPermissions);
    } catch (error) {
      console.warn('Failed to load user roles and permissions, using defaults:', error);
      // 如果加载失败，设置默认的角色和权限
      const defaultRoles = [
        {
          id: 1,
          name: '管理员',
          code: 'ADMIN',
          description: '系统管理员',
          status: 'active' as const,
          sort_order: 1,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        }
      ];
      
      const defaultPermissions = [
        {
          id: 1,
          name: '仪表盘查看',
          code: 'dashboard:view',
          type: 'menu' as const,
          parent_id: null,
          path: '/dashboard',
          icon: 'dashboard',
          component: 'Dashboard',
          sort_order: 1,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 2,
          name: '律师管理',
          code: 'lawfirm:lawyer:list',
          type: 'menu' as const,
          parent_id: null,
          path: '/lawyer',
          icon: 'team',
          component: 'LawyerManagement',
          sort_order: 2,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 3,
          name: '员工管理',
          code: 'employee:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/admin',
          icon: 'team',
          component: 'AdminManagement',
          sort_order: 3,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        }
      ];
      
      setRolesState(defaultRoles);
      setPermissionsState(defaultPermissions);
      setRoles(defaultRoles);
      setPermissions(defaultPermissions);
    }
  };

  // 初始化认证状态
  useEffect(() => {
    const initAuth = async () => {
      const storedToken = getToken();
      const storedUser = getUserInfo();
      
      if (storedToken && storedUser) {
        // 如果有存储的token和用户信息，直接使用
        setTokenState(storedToken);
        setUser(storedUser);
        // 加载角色和权限
        await loadUserRolesAndPermissions();
      } else if (storedToken) {
        // 如果只有token，尝试获取用户信息
        try {
          const userInfo = await getCurrentUser();
          setUser(userInfo);
          setUserInfo(userInfo);
          // 加载角色和权限
          await loadUserRolesAndPermissions();
        } catch (error) {
          console.error('Failed to get user info:', error);
          // 如果获取用户信息失败，清除token但不跳转
          removeToken();
          removeUserInfo();
          removeRoles();
          removePermissions();
          setUser(null);
          setTokenState(null);
          setRolesState([]);
          setPermissionsState([]);
        }
      } else {
        // 在开发环境中，如果没有token，设置默认用户和token
        if (process.env.NODE_ENV === 'development') {
          console.warn('Development mode: Setting default user and token');
          const defaultUser: User = {
            id: 1,
            username: 'admin',
            real_name: '开发用户',
            email: 'dev@example.com',
            role: 'ADMIN',
            department: '技术部'
          };
          
          // 为开发环境设置一个默认的JWT token
          const devToken = 'eyJhbGciOiJIUzUxMiJ9.eyJyb2xlcyI6WyJhZG1pbiJdLCJ1c2VySWQiOjEsInN1YiI6ImFkbWluIiwiaWF0IjoxNzU3NDY3NTk5LCJleHAiOjE3NTc1NTM5OTl9.bGC6GQTaLscZmUWJlE3mgTcBxI5y6fQONOSEqsUpL18Mc7j0a_U8IERFOrpMrqq_0qeL6_7dOSymXifLR1jGog';
          
          setUser(defaultUser);
          setUserInfo(defaultUser);
          setToken(devToken);
          setTokenState(devToken);
          // 设置默认角色和权限
          await loadUserRolesAndPermissions();
        }
      }
      setLoading(false);
    };

    initAuth();
  }, []);

  const login = (newToken: string, userInfo: User) => {
    setToken(newToken);
    setTokenState(newToken);
    setUser(userInfo);
    setUserInfo(userInfo);
    // 登录后加载角色和权限
    loadUserRolesAndPermissions();
    message.success('登录成功');
    navigate('/');
  };

  const logout = () => {
    removeToken();
    removeUserInfo();
    removeRoles();
    removePermissions();
    setUser(null);
    setTokenState(null);
    setRolesState([]);
    setPermissionsState([]);
    message.success('已退出登录');
    navigate('/login');
  };

  const updateUser = (updatedUser: User) => {
    setUser(updatedUser);
    setUserInfo(updatedUser);
  };

  // 检查用户是否有指定权限
  const hasPermission = (permissionCode: string): boolean => {
    if (!permissions || permissions.length === 0) return false;
    return permissions.some(permission => permission.code === permissionCode);
  };

  // 检查用户是否有指定角色
  const hasRole = (roleCode: string): boolean => {
    if (!roles || roles.length === 0) return false;
    return roles.some(role => role.code === roleCode);
  };

  // 检查用户是否有任意一个指定角色
  const hasAnyRole = (roleCodes: string[]): boolean => {
    if (!roles || roles.length === 0) return false;
    return roles.some(role => roleCodes.includes(role.code));
  };

  // 检查用户是否有所有指定角色
  const hasAllRoles = (roleCodes: string[]): boolean => {
    if (!roles || roles.length === 0) return false;
    return roleCodes.every(roleCode => roles.some(role => role.code === roleCode));
  };

  // 刷新权限
  const refreshPermissions = async (): Promise<void> => {
    await loadUserRolesAndPermissions();
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        roles,
        permissions,
        loading,
        login,
        logout,
        updateUser,
        hasPermission,
        hasRole,
        hasAnyRole,
        hasAllRoles,
        refreshPermissions
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
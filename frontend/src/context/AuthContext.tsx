import React, { createContext, useState, useEffect, ReactNode } from 'react';
import { useNavigate } from 'react-router';
import { message } from '@/utils/messageHelper';
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
  const loadUserRolesAndPermissions = async (currentUser?: User | null) => {
    // 后端暂时没有角色权限管理API，根据用户角色设置默认值
    console.warn('Backend does not have role/permission APIs, using defaults based on user role');
    
    // 根据用户角色设置相应的权限
    const userRole = currentUser?.role || user?.role || 'admin';
    console.log('Setting permissions for user role:', userRole);
    
    let defaultRoles = [];
    let defaultPermissions = [];
    
    if (userRole === 'admin') {
      // 管理员角色和权限
      defaultRoles = [
        {
          id: 1,
          name: '系统管理员',
          code: 'ADMIN',
          description: '系统管理员',
          status: 'active' as const,
          sort_order: 1,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        }
      ];
      
      defaultPermissions = [
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
          name: '项目管理',
          code: 'project:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/project',
          icon: 'appstore',
          component: 'ProjectManagement',
          sort_order: 2,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 3,
          name: '案件管理',
          code: 'case:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/case',
          icon: 'solution',
          component: 'CaseManagement',
          sort_order: 3,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 4,
          name: '客户管理',
          code: 'client:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/client',
          icon: 'team',
          component: 'ClientManagement',
          sort_order: 4,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 5,
          name: '律师管理',
          code: 'lawyer:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/lawyer',
          icon: 'user',
          component: 'LawyerManagement',
          sort_order: 5,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 6,
          name: '利益冲突检查',
          code: 'conflict:check',
          type: 'menu' as const,
          parent_id: null,
          path: '/conflict',
          icon: 'file-search',
          component: 'ConflictCheck',
          sort_order: 6,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 7,
          name: '文件管理',
          code: 'file:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/file',
          icon: 'cloud-upload',
          component: 'FileManagement',
          sort_order: 7,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 8,
          name: '审批中心',
          code: 'approval:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/approval',
          icon: 'file-done',
          component: 'ApprovalCenter',
          sort_order: 8,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 9,
          name: '法条查询',
          code: 'law:search',
          type: 'menu' as const,
          parent_id: null,
          path: '/tools/law-search',
          icon: 'search',
          component: 'LawSearch',
          sort_order: 9,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 10,
          name: '案例检索',
          code: 'case:search',
          type: 'menu' as const,
          parent_id: null,
          path: '/tools/case-search',
          icon: 'database',
          component: 'CaseSearch',
          sort_order: 10,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 11,
          name: '企业信息查询',
          code: 'company:search',
          type: 'menu' as const,
          parent_id: null,
          path: '/tools/company-search',
          icon: 'bank',
          component: 'CompanySearch',
          sort_order: 11,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 12,
          name: '日程安排',
          code: 'calendar:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/calendar',
          icon: 'schedule',
          component: 'CalendarManagement',
          sort_order: 12,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 13,
          name: '文档模板',
          code: 'document:template',
          type: 'menu' as const,
          parent_id: null,
          path: '/documents',
          icon: 'file-text',
          component: 'DocumentTemplates',
          sort_order: 13,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 14,
          name: '财务管理',
          code: 'finance:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/finance',
          icon: 'calculator',
          component: 'FinanceManagement',
          sort_order: 14,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 15,
          name: '统计报表',
          code: 'report:view',
          type: 'menu' as const,
          parent_id: null,
          path: '/reports',
          icon: 'bar-chart',
          component: 'Reports',
          sort_order: 15,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 16,
          name: '用户管理',
          code: 'user:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/user',
          icon: 'user',
          component: 'UserManagement',
          sort_order: 16,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 17,
          name: '系统设置',
          code: 'system:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/settings',
          icon: 'setting',
          component: 'SystemSettings',
          sort_order: 17,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        }
      ];
    } else if (userRole === 'lawyer') {
      // 律师角色和权限
      defaultRoles = [
        {
          id: 2,
          name: '律师',
          code: 'LAWYER',
          description: '执业律师',
          status: 'active' as const,
          sort_order: 2,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        }
      ];
      
      defaultPermissions = [
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
          name: '客户管理',
          code: 'client:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/client',
          icon: 'team',
          component: 'ClientManagement',
          sort_order: 2,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 3,
          name: '案件管理',
          code: 'case:manage',
          type: 'menu' as const,
          parent_id: null,
          path: '/case',
          icon: 'file',
          component: 'CaseManagement',
          sort_order: 3,
          status: 'active' as const,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        }
      ];
    } else {
      // 其他用户（如助理等）
      defaultRoles = [
        {
          id: 3,
          name: '助理',
          code: 'ASSISTANT',
          description: '律师助理',
          status: 'active' as const,
          sort_order: 3,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        }
      ];
      
      defaultPermissions = [
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
        }
      ];
    }
    
    console.log('Setting roles:', defaultRoles);
    console.log('Setting permissions:', defaultPermissions);
    
    setRolesState(defaultRoles);
    setPermissionsState(defaultPermissions);
    setRoles(defaultRoles);
    setPermissions(defaultPermissions);
  };

  // 初始化认证状态
  useEffect(() => {
    const initAuth = async () => {
      const storedToken = getToken();
      const storedUser = getUserInfo();

      // 检查是否为开发模式
      const isDevMode = process.env.NODE_ENV === 'development';

      if (isDevMode) {
        console.log('🛠️ 开发者模式：使用简化认证流程');

        // 在开发模式下，如果有token和用户信息就直接使用，不调用API
        if (storedToken && storedUser) {
          console.log('🛠️ 开发者模式：使用存储的token和用户信息，跳过API验证');
          const devUser = storedUser;

          setTokenState(storedToken);
          setUser(devUser);
          setUserInfo(devUser);
          // 加载角色和权限
          await loadUserRolesAndPermissions(devUser);
        } else if (storedToken) {
          console.log('🛠️ 开发者模式：有token但缺少用户信息，尝试从API获取');
          try {
            const userInfo = await getCurrentUser();
            setUser(userInfo);
            setUserInfo(userInfo);
            // 加载角色和权限
            await loadUserRolesAndPermissions(userInfo);
          } catch (error) {
            console.error('开发模式下获取用户信息失败:', error);
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
          console.log('🛠️ 开发者模式：没有token，保持未登录状态');
        }
      } else if (storedToken && storedUser) {
        // 生产模式：如果有存储的token和用户信息，直接使用
        setTokenState(storedToken);
        setUser(storedUser);
        // 加载角色和权限
        await loadUserRolesAndPermissions(storedUser);
      } else if (storedToken) {
        // 生产模式：如果只有token，尝试获取用户信息
        try {
          const userInfo = await getCurrentUser();
          setUser(userInfo);
          setUserInfo(userInfo);
          // 加载角色和权限
          await loadUserRolesAndPermissions(userInfo);
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
      }
      // 如果没有token，不设置默认用户，保持未登录状态
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
    loadUserRolesAndPermissions(userInfo);
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
    // 重新加载权限（如果角色发生变化）
    loadUserRolesAndPermissions(updatedUser);
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

// 导出useAuth hook供组件使用
export const useAuth = () => {
  const context = React.useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export default AuthProvider;
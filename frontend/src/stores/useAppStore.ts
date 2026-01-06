/**
 * 应用全局状态管理 - 基于Zustand v5
 * 提供轻量级、类型安全的状态管理解决方案
 */

import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { subscribeWithSelector } from 'zustand/middleware'

// 用户状态接口
export interface User {
  id: string
  username: string
  email: string
  realName: string
  avatar?: string
  roles: string[]
  permissions: string[]
  department?: string
  position?: string
  phone?: string
  isActive: boolean
  lastLoginAt?: string
  createdAt: string
}

// 主题配置
export type Theme = 'light' | 'dark' | 'auto'

// 语言配置
export type Language = 'zh-CN' | 'en-US'

// 通知配置
export interface NotificationSettings {
  email: boolean
  push: boolean
  sms: boolean
  desktop: boolean
  caseUpdates: boolean
  systemMessages: boolean
  deadlineReminders: boolean
}

// 用户偏好设置
export interface UserPreferences {
  theme: Theme
  language: Language
  defaultPageSize: number
  autoSave: boolean
  showTips: boolean
  notifications: NotificationSettings
  sidebarCollapsed: boolean
  tableSettings: {
    size: 'small' | 'middle' | 'large'
    bordered: boolean
    striped: boolean
    showHeader: boolean
  }
}

// 应用状态接口
export interface AppState {
  // 用户认证状态
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean

  // UI状态
  sidebarCollapsed: boolean
  currentPath: string
  breadcrumb: Array<{ title: string; path?: string }>

  // 应用配置
  preferences: UserPreferences

  // 系统信息
  systemInfo: {
    version: string
    buildTime: string
    environment: string
    features: string[]
  }

  // Actions
  // 用户认证相关
  login: (user: User, token: string) => void
  logout: () => void
  updateUser: (user: Partial<User>) => void
  setLoading: (loading: boolean) => void

  // UI状态相关
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setCurrentPath: (path: string) => void
  setBreadcrumb: (breadcrumb: Array<{ title: string; path?: string }>) => void

  // 偏好设置相关
  updatePreferences: (preferences: Partial<UserPreferences>) => void
  updateTheme: (theme: Theme) => void
  updateLanguage: (language: Language) => void
  updateNotifications: (notifications: Partial<NotificationSettings>) => void

  // 系统信息相关
  setSystemInfo: (info: Partial<AppState['systemInfo']>) => void

  // 重置状态
  reset: () => void
}

// 创建应用状态Store
export const useAppStore = create<AppState>()(
  devtools(
    persist(
      (set, get) => ({
        // 初始状态
        user: null,
        isAuthenticated: false,
        isLoading: false,

        sidebarCollapsed: false,
        currentPath: '/',
        breadcrumb: [],

        preferences: {
          theme: 'light',
          language: 'zh-CN',
          defaultPageSize: 20,
          autoSave: true,
          showTips: true,
          notifications: {
            email: true,
            push: true,
            sms: false,
            desktop: true,
            caseUpdates: true,
            systemMessages: true,
            deadlineReminders: true,
          },
          sidebarCollapsed: false,
          tableSettings: {
            size: 'middle',
            bordered: true,
            striped: false,
            showHeader: true,
          },
        },

        systemInfo: {
          version: '2.1.0',
          buildTime: new Date().toISOString(),
          environment: 'development',
          features: [
            'case-management',
            'document-management',
            'conflict-detection',
            'team-collaboration',
            'report-generation',
          ],
        },

        // Actions
        login: (user: User, token: string) => {
          localStorage.setItem('auth_token', token)
          localStorage.setItem('law_oa_user_info', JSON.stringify(user))
          set({
            user,
            isAuthenticated: true,
            isLoading: false,
          })
        },

        logout: () => {
          localStorage.removeItem('auth_token')
          localStorage.removeItem('law_oa_user_info')
          localStorage.removeItem('law_oa_roles')
          localStorage.removeItem('law_oa_permissions')
          set({
            user: null,
            isAuthenticated: false,
            currentPath: '/',
            breadcrumb: [],
            isLoading: false,
          })
        },

        updateUser: (userData: Partial<User>) => {
          const currentUser = get().user
          if (currentUser) {
            const updatedUser = { ...currentUser, ...userData }
            localStorage.setItem('law_oa_user_info', JSON.stringify(updatedUser))
            set({ user: updatedUser })
          }
        },

        setLoading: (loading: boolean) => set({ isLoading: loading }),

        toggleSidebar: () =>
          set((state) => ({
            sidebarCollapsed: !state.sidebarCollapsed,
            preferences: {
              ...state.preferences,
              sidebarCollapsed: !state.sidebarCollapsed,
            },
          })),

        setSidebarCollapsed: (collapsed: boolean) =>
          set((state) => ({
            sidebarCollapsed: collapsed,
            preferences: {
              ...state.preferences,
              sidebarCollapsed: collapsed,
            },
          })),

        setCurrentPath: (path: string) => set({ currentPath: path }),

        setBreadcrumb: (breadcrumb: Array<{ title: string; path?: string }>) => set({ breadcrumb }),

        updatePreferences: (newPreferences: Partial<UserPreferences>) =>
          set((state) => ({
            preferences: { ...state.preferences, ...newPreferences },
          })),

        updateTheme: (theme: Theme) =>
          set((state) => ({
            preferences: { ...state.preferences, theme },
          })),

        updateLanguage: (language: Language) =>
          set((state) => ({
            preferences: { ...state.preferences, language },
          })),

        updateNotifications: (notifications: Partial<NotificationSettings>) =>
          set((state) => ({
            preferences: {
              ...state.preferences,
              notifications: { ...state.preferences.notifications, ...notifications },
            },
          })),

        setSystemInfo: (info: Partial<AppState['systemInfo']>) =>
          set((state) => ({
            systemInfo: { ...state.systemInfo, ...info },
          })),

        reset: () => {
          localStorage.removeItem('auth_token')
          localStorage.removeItem('law_oa_user_info')
          localStorage.removeItem('law_oa_roles')
          localStorage.removeItem('law_oa_permissions')
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
            sidebarCollapsed: false,
            currentPath: '/',
            breadcrumb: [],
            // 保留系统信息和默认偏好设置
          })
        },
      }),
      {
        name: 'law-oa-app-store',
        partialize: (state) => ({
          // 只持久化这些字段
          user: state.user,
          preferences: state.preferences,
          systemInfo: state.systemInfo,
          sidebarCollapsed: state.sidebarCollapsed,
        }),
      },
    ),
    {
      name: 'LawOA App Store',
    },
  ),
)

// 订阅状态变化
export const subscribeToAuthChanges = (
  callback: (isAuthenticated: boolean, user: User | null) => void,
) => {
  return useAppStore.subscribe(
    (state) => state.isAuthenticated,
    (isAuthenticated) => callback(isAuthenticated, useAppStore.getState().user),
  )
}

// 订阅主题变化
export const subscribeToThemeChanges = (callback: (theme: Theme) => void) => {
  return useAppStore.subscribe((state) => state.preferences.theme, callback)
}

// 订阅语言变化
export const subscribeToLanguageChanges = (callback: (language: Language) => void) => {
  return useAppStore.subscribe((state) => state.preferences.language, callback)
}

// 选择器Hook
export const useAuth = () => {
  const user = useAppStore((state) => state.user)
  const isAuthenticated = useAppStore((state) => state.isAuthenticated)
  const isLoading = useAppStore((state) => state.isLoading)

  return { user, isAuthenticated, isLoading }
}

export const usePreferences = () => {
  return useAppStore((state) => state.preferences)
}

export const useUI = () => {
  const sidebarCollapsed = useAppStore((state) => state.sidebarCollapsed)
  const currentPath = useAppStore((state) => state.currentPath)
  const breadcrumb = useAppStore((state) => state.breadcrumb)

  return { sidebarCollapsed, currentPath, breadcrumb }
}

export const useSystemInfo = () => {
  return useAppStore((state) => state.systemInfo)
}

// 工具函数
export const hasPermission = (permission: string): boolean => {
  const user = useAppStore.getState().user
  return user ? user.permissions.includes(permission) : false
}

export const hasRole = (role: string): boolean => {
  const user = useAppStore.getState().user
  return user ? user.roles.includes(role) : false
}

export const hasAnyRole = (roles: string[]): boolean => {
  const user = useAppStore.getState().user
  return user ? roles.some((role) => user.roles.includes(role)) : false
}

export const hasAnyPermission = (permissions: string[]): boolean => {
  const user = useAppStore.getState().user
  return user ? permissions.some((permission) => user.permissions.includes(permission)) : false
}

// 初始化函数
export const initializeApp = () => {
  const { login } = useAppStore.getState()

  // 从localStorage恢复用户状态
  try {
    const token = localStorage.getItem('auth_token')
    const userInfo = localStorage.getItem('law_oa_user_info')

    if (token && userInfo) {
      const user = JSON.parse(userInfo)
      login(user, token)
    }
  } catch (error) {
    console.error('Failed to restore user session:', error)
    localStorage.removeItem('auth_token')
    localStorage.removeItem('law_oa_user_info')
    localStorage.removeItem('law_oa_roles')
    localStorage.removeItem('law_oa_permissions')
  }
}

// 导出Store类型
export type AppStore = ReturnType<typeof useAppStore>

/**
 * 应用主入口 - React 18+ with Concurrent Features
 * 集成现代化状态管理和错误边界
 */

import React, { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'

// 导入调试工具
import './utils/debugHelper'

// 导入状态管理
import { QueryProvider } from './hooks/useQueryClient'
import { useAppStore } from './stores/useAppStore'
import { initializeApp } from './stores/useAppStore'

// 导入样式
import './index.css'
import './assets/styles/design-tokens.css'
import './assets/styles/modal-fix.css' // Modal层级修复样式 - 必须在设计系统token之后加载

// 导入主应用组件
import App from './App'

// 导入错误边界和加载状态
import ErrorBoundary from './components/common/ErrorBoundary'

// 导入开发工具（仅在开发环境）
import DevTools from './components/devtools/DevTools'

// 导入性能监控
import './utils/performance'

// 配置dayjs
dayjs.locale('zh-cn')

// 开发环境配置React Query
if (import.meta.env.DEV) {
  // 全局设置React Query开发模式
  const testGlobal = globalThis as typeof globalThis & { IS_REACT_ACT_ENVIRONMENT?: boolean }
  testGlobal.IS_REACT_ACT_ENVIRONMENT = true
}

// Suspense回退组件
const SuspenseFallback: React.FC = () => (
  <div className='flex items-center justify-center min-h-screen'>
    <div className='text-center'>
      <div className='animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4' />
      <p className='text-gray-600'>正在加载...</p>
    </div>
  </div>
)

const RootComponent: React.FC = () => {
  const { preferences, updateTheme } = useAppStore()
  const theme = preferences.theme

  // 根据主题切换样式
  React.useEffect(() => {
    const root = document.documentElement
    if (
      theme === 'dark' ||
      (theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)
    ) {
      root.classList.add('dark')
      root.classList.remove('light')
    } else {
      root.classList.add('light')
      root.classList.remove('dark')
    }
  }, [theme])

  // 监听系统主题变化
  React.useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (e: MediaQueryListEvent) => {
      if (preferences.theme === 'auto') {
        updateTheme(e.matches ? 'dark' : 'light')
      }
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [preferences.theme, updateTheme])

  return (
    <StrictMode>
      <ErrorBoundary>
        <QueryProvider>
          <ConfigProvider
            locale={zhCN}
            theme={{
              token: {
                colorPrimary: '#1890ff',
                borderRadius: 6,
                fontSize: 14,
                fontFamily:
                  '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
              },
              components: {
                Layout: {
                  headerBg: '#ffffff',
                  siderBg: '#f6f6f6',
                  bodyBg: '#f0f2f5',
                },
              },
            }}
          >
            <BrowserRouter>
              <React.Suspense fallback={<SuspenseFallback />}>
                <App />
                {/* 开发工具 - 仅在开发环境显示 */}
                {import.meta.env.DEV && <DevTools />}
              </React.Suspense>
            </BrowserRouter>
          </ConfigProvider>
        </QueryProvider>
      </ErrorBoundary>
    </StrictMode>
  )
}

// 获取root元素
const container = document.getElementById('root')
if (!container) {
  throw new Error('Failed to find the root element')
}

// 创建React 18+的root
const root = createRoot(container)

// 初始化应用状态
initializeApp()

// 渲染应用
root.render(<RootComponent />)

// 性能监控
if (import.meta.env.DEV) {
  // 开发环境性能监控
  if (import.meta.env.VITE_ENABLE_PERFORMANCE_MONITORING === 'true') {
    import('./utils/performance')
      .then(({ startPerformanceMonitoring }) => startPerformanceMonitoring())
      .catch(console.error)
  }

  // 热模块替换(HMR)支持
  if (import.meta.hot) {
    import.meta.hot.accept()
  }
}

// 注册Service Worker
if ('serviceWorker' in navigator && import.meta.env.PROD) {
  window.addEventListener('load', () => {
    navigator.serviceWorker
      .register('/sw.js')
      .then((registration) => {
        console.log('SW registered: ', registration)
      })
      .catch((registrationError) => {
        console.error('SW registration failed: ', registrationError)
      })
  })
}

// 错误处理
window.addEventListener('error', (event) => {
  console.error('Global error:', event.error)
})

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason)
})

// 导出root用于调试
declare global {
  interface Window {
    __REACT_DEVTOOLS_GLOBAL_HOOK__?: any
  }
}

// 开发工具支持
if (import.meta.env.DEV) {
  window.__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
    // React开发者工具配置
    // 可以在这里添加自定义的调试信息
  }
}

export default root

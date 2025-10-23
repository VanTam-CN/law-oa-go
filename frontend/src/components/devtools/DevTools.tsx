/**
 * 开发工具组件 - 集成各种调试工具
 * 仅在开发环境启用
 */

import React from 'react'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { useAppStore } from '../../stores/useAppStore'

// 导入开发工具（仅在开发环境）
if (import.meta.env.DEV) {
  // 可以在这里添加其他开发工具
  // import { ReactFlowDevtools } from 'reactflow-devtools'
}

interface DevToolsProps {
  // 可以控制是否显示特定工具
  showReactQuery?: boolean
  showStoreInspector?: boolean
  position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right'
}

const DevTools: React.FC<DevToolsProps> = ({
  showReactQuery = true,
  showStoreInspector = true,
  position = 'bottom-left'
}) => {
  const { user, preferences } = useAppStore()

  // 只在开发环境显示
  if (!import.meta.env.DEV) {
    return null
  }

  // 判断是否为管理员开发者
  const isAdminDev = user?.roles.includes('admin') && import.meta.env.DEV

  return (
    <>
      {/* React Query 开发工具 */}
      {showReactQuery && <ReactQueryDevtools
        initialIsOpen={false}
        position={position}
        buttonPosition={position}
      />}

      {/* 自定义状态检查器 - 可以扩展 */}
      {showStoreInspector && isAdminDev && (
        <div
          style={{
            position: 'fixed',
            top: position.includes('top') ? '10px' : 'auto',
            bottom: position.includes('bottom') ? '10px' : 'auto',
            left: position.includes('left') ? '10px' : 'auto',
            right: position.includes('right') ? '10px' : 'auto',
            zIndex: 9999,
            background: 'rgba(0,0,0,0.8)',
            color: 'white',
            padding: '8px',
            borderRadius: '4px',
            fontSize: '12px',
            fontFamily: 'monospace',
            maxWidth: '300px',
          }}
        >
          <div style={{ marginBottom: '4px', fontWeight: 'bold' }}>
            🛠️ Dev Tools
          </div>
          <div>Theme: {preferences.theme}</div>
          <div>User: {user?.username || 'Not logged in'}</div>
          <div>Roles: {user?.roles.join(', ') || 'None'}</div>
          <div>Env: {import.meta.env.MODE}</div>
        </div>
      )}

      {/* 键盘快捷键提示 */}
      {isAdminDev && (
        <div
          style={{
            position: 'fixed',
            bottom: '10px',
            right: '10px',
            zIndex: 9999,
            background: 'rgba(0,0,0,0.7)',
            color: 'white',
            padding: '6px',
            borderRadius: '4px',
            fontSize: '10px',
            fontFamily: 'monospace',
          }}
        >
          Press <kbd>Ctrl+Shift+D</kbd> to toggle dev tools
        </div>
      )}
    </>
  )
}

// 全局快捷键处理
if (import.meta.env.DEV) {
  document.addEventListener('keydown', (e) => {
    // Ctrl+Shift+D 切换开发工具显示
    if (e.ctrlKey && e.shiftKey && e.key === 'D') {
      e.preventDefault()
      // 这里可以实现开发工具的显示/隐藏逻辑
      console.log('Dev tools toggle shortcut pressed')
    }
  })
}

export default DevTools
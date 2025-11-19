/**
 * 错误边界组件 - React 18+最佳实践
 * 捕获和处理React组件错误，提供友好的错误界面
 */

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { Result, Button } from 'antd'
import { useNavigate } from 'react-router'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
  errorId: string | null
}

// 错误日志服务
class ErrorLoggingService {
  static logError(error: Error, errorInfo: ErrorInfo, errorId: string, componentStack?: string) {
    const errorReport = {
      id: errorId || 'unknown',
      message: error?.message || '未知错误',
      stack: error?.stack || '无堆栈信息',
      componentStack: componentStack || '',
      errorInfo: errorInfo || null,
      timestamp: new Date().toISOString(),
      userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : 'unknown',
      url: typeof window !== 'undefined' ? window.location.href : 'unknown',
      memory:
        typeof performance !== 'undefined' && performance.memory
          ? {
              usedJSHeapSize: performance.memory.usedJSHeapSize,
              totalJSHeapSize: performance.memory.totalJSHeapSize,
            }
          : null,
    }

    // 开发环境在控制台显示
    if (import.meta.env.DEV) {
      console.error('Error Boundary Caught Error:', errorReport)
    } else {
      // 生产环境发送到错误监控服务
      this.sendToErrorService(errorReport)
    }

    // 保存到localStorage用于调试
    try {
      const errors = JSON.parse(localStorage.getItem('error-logs') || '[]')
      errors.push(errorReport)
      // 只保留最近50个错误
      if (errors.length > 50) {
        errors.splice(0, errors.length - 50)
      }
      localStorage.setItem('error-logs', JSON.stringify(errors))
    } catch (e) {
      console.error('Failed to save error to localStorage:', e)
    }
  }

  private static sendToErrorService(errorReport: any) {
    // 这里可以集成第三方错误监控服务，如Sentry
    // fetch('/api/v1/errors', {
    //   method: 'POST',
    //   headers: { 'Content-Type': 'application/json' },
    //   body: JSON.stringify(errorReport),
    // }).catch(console.error)
  }

  static getRecentErrors(): any[] {
    try {
      return JSON.parse(localStorage.getItem('error-logs') || '[]')
    } catch {
      return []
    }
  }

  static clearErrors(): void {
    localStorage.removeItem('error-logs')
  }
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: null,
    }
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
      errorInfo: null,
      errorId: `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    const errorId = `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`

    this.setState({
      hasError: true,
      error,
      errorInfo,
      errorId,
    })

    // 记录错误日志
    ErrorLoggingService.logError(error, errorInfo, errorId, errorInfo.componentStack)

    // 调用自定义错误处理
    this.props.onError?.(error, errorInfo)
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: null,
    })
  }

  handleReload = () => {
    window.location.reload()
  }

  handleGoHome = () => {
    this.handleReset()
    // 使用window.location而不是navigate，确保完全重置状态
    window.location.href = '/'
  }

  handleShowErrorDetails = () => {
    const { error, errorId } = this.state
    if (!error) {
      return
    }

    const errorDetails = {
      id: errorId || 'unknown',
      message: error?.message || '未知错误',
      stack: error?.stack || '无堆栈信息',
      timestamp: new Date().toISOString(),
    }

    // 在新窗口显示错误详情
    const errorWindow = window.open('', '_blank', 'width=800,height=600,scrollbars=yes')
    if (errorWindow) {
      errorWindow.document.write(`
        <html>
          <head>
            <title>错误详情 - 律师OA系统</title>
            <style>
              body { font-family: monospace; padding: 20px; line-height: 1.6; }
              .error-details { background: #f5f5f5; padding: 20px; border-radius: 8px; margin: 10px 0; }
              .error-id { color: #666; font-size: 12px; }
              .error-stack { white-space: pre-wrap; font-size: 12px; color: #d32f2f; }
              .actions { margin-top: 20px; }
            </style>
          </head>
          <body>
            <h2>错误详情</h2>
            <div class="error-id">错误ID: ${errorDetails.id}</div>
            <div class="error-details">
              <h3>错误信息:</h3>
              <p>${errorDetails.message}</p>

              <h3>时间戳:</h3>
              <p>${errorDetails.timestamp}</p>

              <h3>堆栈信息:</h3>
              <div class="error-stack">${errorDetails.stack || '无堆栈信息'}</div>
            </div>
            <div class="actions">
              <button onclick="window.close()">关闭</button>
            </div>
          </body>
        </html>
      `)
    }
  }

  render() {
    const { hasError, error, errorInfo, errorId } = this.state
    const { fallback, children } = this.props

    if (hasError) {
      if (fallback) {
        return fallback
      }

      // 开发环境显示详细错误信息
      if (import.meta.env.DEV) {
        return (
          <div
            style={{
              padding: '20px',
              backgroundColor: '#fff2f0',
              border: '1px solid #ffccc7',
              borderRadius: '8px',
              margin: '20px',
            }}
          >
            <h2 style={{ color: '#cf1322' }}>组件渲染错误</h2>
            <details style={{ marginBottom: '16px' }}>
              <summary style={{ cursor: 'pointer', marginBottom: '8px' }}>
                错误详情 (开发环境)
              </summary>
              <div style={{ marginTop: '12px', fontSize: '14px', fontFamily: 'monospace' }}>
                <p>
                  <strong>错误ID:</strong> {errorId}
                </p>
                <p>
                  <strong>错误信息:</strong> {error?.message}
                </p>
                <p>
                  <strong>组件堆栈:</strong>
                </p>
                <pre
                  style={{
                    backgroundColor: '#f5f5f5',
                    padding: '12px',
                    borderRadius: '4px',
                    overflow: 'auto',
                    maxHeight: '200px',
                  }}
                >
                  {errorInfo?.componentStack}
                </pre>
                <p>
                  <strong>错误堆栈:</strong>
                </p>
                <pre
                  style={{
                    backgroundColor: '#f5f5f5',
                    padding: '12px',
                    borderRadius: '4px',
                    overflow: 'auto',
                    maxHeight: '300px',
                  }}
                >
                  {error?.stack}
                </pre>
              </div>
            </details>
            <div style={{ marginTop: '16px' }}>
              <button
                onClick={this.handleReset}
                style={{
                  marginRight: '8px',
                  padding: '8px 16px',
                  backgroundColor: '#1890ff',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                }}
              >
                重试
              </button>
              <button
                onClick={this.handleReload}
                style={{
                  marginRight: '8px',
                  padding: '8px 16px',
                  backgroundColor: '#52c41a',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                }}
              >
                刷新页面
              </button>
              <button
                onClick={this.handleShowErrorDetails}
                style={{
                  padding: '8px 16px',
                  backgroundColor: '#722ed1',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                }}
              >
                查看详情
              </button>
            </div>
          </div>
        )
      }

      // 生产环境显示友好的错误界面
      return (
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '100vh',
            padding: '20px',
            backgroundColor: '#f5f5f5',
          }}
        >
          <Result
            status='500'
            title='500'
            subTitle='抱歉，系统遇到了一个错误。'
            extra={[
              <Button type='primary' onClick={this.handleGoHome}>
                返回首页
              </Button>,
              <Button onClick={this.handleReload}>重新加载</Button>,
              <Button onClick={this.handleShowErrorDetails} type='link'>
                查看详情
              </Button>,
            ]}
          />
        </div>
      )
    }

    return children
  }
}

// 错误边界HOC
export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: Omit<Props, 'children'>,
) {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </ErrorBoundary>
  )

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`

  return WrappedComponent
}

// 全局错误边界（用于最外层）
export function GlobalErrorBoundary({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary
      onError={(error, errorInfo) => {
        // 可以在这里添加全局错误处理逻辑
        console.error('Global Error Boundary caught an error:', error, errorInfo)

        // 发送错误报告到监控系统
        if (!import.meta.env.DEV) {
          // 可以集成第三方错误监控服务
          // trackError(error, errorInfo)
        }
      }}
    >
      {children}
    </ErrorBoundary>
  )
}

export default ErrorBoundary

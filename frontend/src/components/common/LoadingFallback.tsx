/**
 * 加载回退组件 - React 18+最佳实践
 * 提供优雅的加载状态显示
 */

import React, { Suspense } from 'react'
import { Spin, Result, Progress } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'

// 加载类型
export type LoadingType = 'default' | 'content' | 'spinner' | 'progress' | 'skeleton'

// 加载状态接口
interface LoadingState {
  message?: string
  description?: string
  progress?: number
  type: LoadingType
}

// 加载回退组件Props
interface LoadingFallbackProps {
  message?: string
  description?: string
  type?: LoadingType
  progress?: number
  delay?: number
  size?: 'small' | 'default' | 'large'
  tip?: string
}

// 默认加载组件
const DefaultLoading: React.FC<{ message?: string; size?: 'small' | 'default' | 'large' }> = ({
  message,
  size = 'default',
}) => (
  <div className='flex flex-col items-center justify-center min-h-screen'>
    <Spin size={size} />
    {message && <p className='mt-4 text-gray-600'>{message}</p>}
  </div>
)

// 带消息的加载组件
const ContentLoading: React.FC<LoadingFallbackProps> = ({
  message = '正在加载中...',
  description,
  size = 'large',
}) => (
  <div className='flex flex-col items-center justify-center min-h-screen p-8'>
    <div className='text-center mb-6'>
      <LoadingOutlined className='text-6xl text-blue-500 mb-4' />
      <h3 className='text-xl font-semibold text-gray-800 mb-2'>{message}</h3>
      {description && <p className='text-gray-600'>{description}</p>}
    </div>
    <Spin size={size} />
  </div>
)

// 进度条加载组件
const ProgressLoading: React.FC<{
  message?: string
  progress?: number
  size?: 'small' | 'default' | 'large'
  status?: 'normal' | 'success' | 'active' | 'exception'
}> = ({ message, progress = 0, size = 'default', status = 'active' }) => (
  <div className='flex flex-col items-center justify-center min-h-screen p-8'>
    <div className='w-full max-w-md'>
      <div className='text-center mb-6'>
        <LoadingOutlined className='text-4xl text-blue-500 mb-4' />
        <h3 className='text-xl font-semibold text-gray-800 mb-2'>{message || '正在处理中...'}</h3>
      </div>
      <Progress
        percent={progress}
        size={size}
        status={status}
        strokeColor={{
          '0%': '#108ee9',
          '100%': '#52c41a',
        }}
      />
      <div className='text-center mt-4'>
        <span className='text-sm text-gray-600'>{progress}%</span>
      </div>
    </div>
  </div>
)

// 骨架屏加载组件
const SkeletonLoading: React.FC<{
  loading?: boolean
  avatar?: boolean
  paragraph?: boolean
  title?: boolean
  active?: boolean
  round?: boolean
}> = ({
  loading = true,
  avatar = true,
  paragraph = { rows: 4 },
  title = true,
  active = true,
  round = true,
}) => (
  <div className='p-8'>
    <div className='bg-white rounded-lg p-6 shadow-sm'>
      <Spin spinning={loading}>
        <div className='space-y-4'>
          <div className='h-4 bg-gray-200 rounded animate-pulse' />
          <div className='space-y-2'>
            <div className='h-4 bg-gray-200 rounded animate-pulse' />
            <div className='h-4 bg-gray-200 rounded animate-pulse w-3/4' />
            <div className='h-4 bg-gray-200 rounded animate-pulse w-1/2' />
          </div>
        </div>
      </Spin>
    </div>
  </div>
)

// 主加载回退组件
const LoadingFallback: React.FC<LoadingFallbackProps> = ({
  message,
  description,
  type = 'default',
  progress,
  delay = 200,
  size = 'default',
}) => {
  const [showLoading, setShowLoading] = React.useState(delay === 0)

  React.useEffect(() => {
    if (delay > 0) {
      const timer = setTimeout(() => setShowLoading(true), delay)
      return () => clearTimeout(timer)
    }
  }, [delay])

  if (!showLoading) {
    return null
  }

  switch (type) {
    case 'content':
      return <ContentLoading message={message} description={description} size={size} />
    case 'progress':
      return <ProgressLoading message={message} progress={progress} size={size} />
    case 'skeleton':
      return <SkeletonLoading active loading />
    default:
      return <DefaultLoading message={message} size={size} />
  }
}

// 错误回退组件
const ErrorFallback: React.FC<{
  error?: Error
  retry?: () => void
  message?: string
}> = ({ error, retry, message = '加载失败' }) => (
  <div className='flex flex-col items-center justify-center min-h-screen p-8'>
    <Result
      status='error'
      title='加载失败'
      subTitle={message}
      extra={
        retry && (
          <div className='space-x-2'>
            <button onClick={retry} type='button'>
              重试
            </button>
            <button onClick={() => window.location.reload()} type='button'>
              刷新页面
            </button>
          </div>
        )
      }
    />
  </div>
)

// 延迟加载组件
export const DelayedLoading: React.FC<{ delay?: number; fallback?: ReactNode }> = ({
  delay = 300,
  fallback,
}) => {
  const [show, setShow] = React.useState(false)

  React.useEffect(() => {
    const timer = setTimeout(() => setShow(true), delay)
    return () => clearTimeout(timer)
  }, [delay])

  if (!show) {
    return fallback || null
  }

  return <LoadingFallback />
}

// 条件加载组件
export const ConditionalLoading: React.FC<{
  condition: boolean
  children: ReactNode
  fallback?: ReactNode
}> = ({ condition, children, fallback = <LoadingFallback /> }) => {
  if (!condition) {
    return fallback
  }
  return <>{children}</>
}

// 异步加载组件包装器
export function AsyncWrapper({
  children,
  fallback,
  errorFallback,
}: {
  children: ReactNode
  fallback?: ReactNode
  errorFallback?: ReactNode
}) {
  return (
    <Suspense fallback={fallback || <LoadingFallback />}>
      <ErrorBoundary fallback={errorFallback || <ErrorFallback />}>{children}</ErrorBoundary>
    </Suspense>
  )
}

// 页面级加载组件
export const PageLoading: React.FC<{
  message?: string
  description?: string
}> = ({ message = '页面加载中...', description }) => (
  <div className='flex items-center justify-center min-h-screen'>
    <ContentLoading message={message} description={description} />
  </div>
)

// 按钮加载状态
export const ButtonLoading: React.FC<{
  loading: boolean
  children: ReactNode
  className?: string
}> = ({ loading, children, className }) => (
  <button
    className={`inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed ${className}`}
    disabled={loading}
  >
    {loading && <Spin className='mr-2' size='small' />}
    {children}
  </button>
)

// 表格加载状态
export const TableLoading: React.FC<{
  loading: boolean
  rows?: number
  children: ReactNode
}> = ({ loading, rows = 5, children }) => {
  if (!loading) {
    return <>{children}</>
  }

  return (
    <div className='space-y-2'>
      {Array.from({ length: rows }).map((_, index) => (
        <div
          key={index}
          className='h-12 bg-gray-200 rounded animate-pulse'
          style={{ animationDelay: `${index * 0.1}s` }}
        />
      ))}
    </div>
  )
}

// 图片加载组件
export const ImageLoading: React.FC<{
  src: string
  alt: string
  className?: string
  fallback?: ReactNode
}> = ({ src, alt, className, fallback }) => {
  const [loaded, setLoaded] = React.useState(false)
  const [error, setError] = React.useState(false)

  React.useEffect(() => {
    const img = new Image()
    img.onload = () => setLoaded(true)
    img.onerror = () => setError(true)
    img.src = src
  }, [src])

  if (error) {
    return (
      <div className={`flex items-center justify-center bg-gray-200 rounded ${className}`}>
        <span className='text-gray-500 text-sm'>{alt}</span>
      </div>
    )
  }

  if (!loaded) {
    return (
      <div
        className={`flex items-center justify-center bg-gray-200 rounded animate-pulse ${className}`}
      >
        <div className='w-4 h-4 bg-gray-300 rounded' />
      </div>
    )
  }

  return <img src={src} alt={alt} className={className} />
}

export default LoadingFallback

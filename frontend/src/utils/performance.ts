/**
 * 性能优化工具
 * 专为1080p显示器优化提供性能监控和优化功能
 */

import { useState, useCallback, useEffect } from 'react'

// 性能监控配置
export interface PerformanceConfig {
  targetFrameTime: number // 目标帧时间（毫秒）
  maxRenderTime: number // 最大渲染时间（毫秒）
  memoryThreshold: number // 内存阈值（MB）
  enableMonitoring: boolean // 是否启用监控
}

// 默认配置
const DEFAULT_CONFIG: PerformanceConfig = {
  targetFrameTime: 16.67, // 60fps
  maxRenderTime: 100, // 最大渲染时间
  memoryThreshold: 100, // 100MB内存阈值
    enableMonitoring: import.meta.env.DEV,
}

// 性能监控器
export class PerformanceMonitor {
  readonly config: PerformanceConfig
  private observers: PerformanceObserver[] = []
  private metrics: Map<string, number[]> = new Map()

  constructor(config: Partial<PerformanceConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config }
    if (this.config.enableMonitoring) {
      this.initializeObservers()
    }
  }

  // 初始化性能观察器
  private initializeObservers() {
    // 监控渲染性能
    const renderObserver = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        this.recordMetric('renderTime', entry.duration)
        if (entry.duration > this.config.maxRenderTime) {
          console.warn(`Slow render detected: ${entry.duration.toFixed(2)}ms`)
        }
      }
    })
    renderObserver.observe({ entryTypes: ['measure'] })
    this.observers.push(renderObserver)

    // 监控内存使用
    const memoryObserver = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        this.recordMetric('memory', entry.duration)
      }
    })
    memoryObserver.observe({ entryTypes: ['navigation'] })
    this.observers.push(memoryObserver)
  }

  // 记录性能指标
  private recordMetric(name: string, value: number) {
    if (!this.metrics.has(name)) {
      this.metrics.set(name, [])
    }
    const values = this.metrics.get(name) || []
    values.push(value)

    // 只保留最近的100个数据点
    if (values.length > 100) {
      values.shift()
    }
  }

  // 获取性能统计
  getMetrics() {
    const result: Record<string, any> = {}
    for (const [name, values] of this.metrics.entries()) {
      result[name] = {
        average: values.reduce((a, b) => a + b, 0) / values.length,
        min: Math.min(...values),
        max: Math.max(...values),
        count: values.length,
      }
    }
    return result
  }

  // 1080p显示器性能测试
  async test1080pPerformance() {
    const startTime = performance.now()

    // 模拟1080p分辨率下的典型操作
    const testOperations = [
      () => this.testFormRendering(),
      () => this.testTableRendering(),
      () => this.testModalRendering(),
      () => this.testResponsiveLayout(),
    ]

    for (const operation of testOperations) {
      await operation()
    }

    const totalTime = performance.now() - startTime
    const metrics = this.getMetrics()

    return {
      totalTime,
      metrics,
      isOptimized: totalTime < this.config.maxRenderTime,
    }
  }

  // 测试表单渲染性能
  private async testFormRendering() {
    performance.mark('form-render-start')

    // 模拟复杂表单渲染
    const testElement = document.createElement('div')
    testElement.style.width = '1920px'
    testElement.style.height = '300px'
    testElement.innerHTML = `
      <form>
        ${Array.from(
          { length: 20 },
          (_, i) =>
            `<input type="text" placeholder="Field ${i + 1}" style="margin: 5px; padding: 8px;"/>`,
        ).join('')}
      </form>
    `
    document.body.appendChild(testElement)

    // 强制重排
    testElement.offsetHeight

    performance.mark('form-render-end')
    performance.measure('form-render', 'form-render-start', 'form-render-end')

    // 清理
    document.body.removeChild(testElement)
  }

  // 测试表格渲染性能
  private async testTableRendering() {
    performance.mark('table-render-start')

    const testElement = document.createElement('table')
    testElement.style.width = '100%'
    testElement.innerHTML = `
      <thead>
        <tr>${Array.from({ length: 10 }, (_, i) => `<th>Column ${i + 1}</th>`).join('')}</tr>
      </thead>
      <tbody>
        ${Array.from(
          { length: 50 },
          (_, i) =>
            `<tr>${Array.from({ length: 10 }, (_, j) => `<td>Row ${i + 1}, Col ${j + 1}</td>`).join('')}</tr>`,
        ).join('')}
      </tbody>
    `
    document.body.appendChild(testElement)

    // 强制重排
    testElement.offsetHeight

    performance.mark('table-render-end')
    performance.measure('table-render', 'table-render-start', 'table-render-end')

    // 清理
    document.body.removeChild(testElement)
  }

  // 测试模态框渲染性能
  private async testModalRendering() {
    performance.mark('modal-render-start')

    const testElement = document.createElement('div')
    testElement.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100vw;
      height: 100vh;
      background: rgba(0, 0, 0, 0.5);
      display: flex;
      align-items: center;
      justify-content: center;
    `
    testElement.innerHTML = `
      <div style="background: white; padding: 20px; border-radius: 8px; width: 600px; max-width: 90vw;">
        <h2>Test Modal</h2>
        <p>This is a performance test modal for 1080p displays.</p>
        <button style="padding: 8px 16px;">Close</button>
      </div>
    `
    document.body.appendChild(testElement)

    // 强制重排
    testElement.offsetHeight

    performance.mark('modal-render-end')
    performance.measure('modal-render', 'modal-render-start', 'modal-render-end')

    // 清理
    document.body.removeChild(testElement)
  }

  // 测试响应式布局性能
  private async testResponsiveLayout() {
    performance.mark('layout-render-start')

    const testElement = document.createElement('div')
    testElement.style.width = '1920px'
    testElement.style.display = 'grid'
    testElement.style.gridTemplateColumns = 'repeat(4, 1fr)'
    testElement.style.gap = '16px'

    for (let i = 0; i < 20; i++) {
      const child = document.createElement('div')
      child.style.cssText = `
        background: #f0f0f0;
        padding: 16px;
        border-radius: 4px;
        min-height: 100px;
      `
      child.textContent = `Item ${i + 1}`
      testElement.appendChild(child)
    }

    document.body.appendChild(testElement)

    // 强制重排
    testElement.offsetHeight

    performance.mark('layout-render-end')
    performance.measure('layout-render', 'layout-render-start', 'layout-render-end')

    // 清理
    document.body.removeChild(testElement)
  }

  // 清理观察器
  dispose() {
    this.observers.forEach((observer) => observer.disconnect())
    this.observers = []
    this.metrics.clear()
  }
}

// React Hook for performance monitoring
export function usePerformanceMonitor(config?: Partial<PerformanceConfig>) {
  const [monitor] = useState(() => new PerformanceMonitor(config))
  const [metrics, setMetrics] = useState<any>({})
  const [isRunning, setIsRunning] = useState(false)

  const startMonitoring = useCallback(() => {
    setIsRunning(true)
    const updateMetrics = () => {
      if (monitor.config.enableMonitoring) {
        setMetrics(monitor.getMetrics())
        if (isRunning) {
          requestAnimationFrame(updateMetrics)
        }
      }
    }
    updateMetrics()
  }, [monitor, isRunning])

  const stopMonitoring = useCallback(() => {
    setIsRunning(false)
  }, [])

  const run1080pTest = useCallback(async () => {
    const result = await monitor.test1080pPerformance()
    return result
  }, [monitor])

  useEffect(() => {
    return () => {
      monitor.dispose()
    }
  }, [monitor])

  return {
    metrics,
    isRunning,
    startMonitoring,
    stopMonitoring,
    run1080pTest,
  }
}

// 1080p优化工具函数
export const Display1080pOptimizer = {
  // 检测是否为1080p显示器
  is1080pDisplay(): boolean {
    const width = window.screen.width
    const height = window.screen.height
    return (width === 1920 && height === 1080) || (width === 1080 && height === 1920)
  },

  // 检测是否为1080p范围内的分辨率
  isWithin1080pRange(): boolean {
    const width = window.innerWidth
    return width >= 1600 && width <= 1920
  },

  // 获取优化的间距值
  getOptimizedSpacing(baseSpacing: number): number {
    if (!this.isWithin1080pRange()) {
      return baseSpacing
    }
    return Math.round(baseSpacing * 0.75) // 1080p使用75%的间距
  },

  // 获取优化的字体大小
  getOptimizedFontSize(baseFontSize: number): number {
    if (!this.isWithin1080pRange()) {
      return baseFontSize
    }
    return Math.round(baseFontSize * 0.9) // 1080p使用90%的字体大小
  },

  // 应用1080p优化样式
  apply1080pOptimizations(element: HTMLElement) {
    if (!this.isWithin1080pRange()) {
      return
    }

    const styles = {
      '--spacing-unit': `${this.getOptimizedSpacing(16)}px`,
      '--font-size-base': `${this.getOptimizedFontSize(14)}px`,
      '--border-radius-sm': '3px',
      '--border-radius-md': '4px',
      '--border-radius-lg': '6px',
    } as Record<string, string>

    Object.entries(styles).forEach(([property, value]) => {
      element.style.setProperty(property, value)
    })
  },
}

export const startPerformanceMonitoring = () => new PerformanceMonitor({ enableMonitoring: true })

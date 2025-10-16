/**
 * 性能监控组件
 * 显示缓存状态、性能指标和网络请求统计
 */

import React, { useState, useEffect } from 'react';
import { Card, Badge, Button, ProgressBar } from 'react-bootstrap';
import { globalCache } from '../hooks/useCases';

interface PerformanceStats {
  cacheHits: number;
  cacheMisses: number;
  totalRequests: number;
  averageResponseTime: number;
  errorCount: number;
}

interface PerformanceMonitorProps {
  enabled?: boolean;
  showDetails?: boolean;
}

const PerformanceMonitor: React.FC<PerformanceMonitorProps> = ({
  enabled = true,
  showDetails = false
}) => {
  const [stats, setStats] = useState<PerformanceStats>({
    cacheHits: 0,
    cacheMisses: 0,
    totalRequests: 0,
    averageResponseTime: 0,
    errorCount: 0
  });
  const [cacheSize, setCacheSize] = useState(0);
  const [isVisible, setIsVisible] = useState(false);

  // 模拟性能统计（在实际应用中，这些数据应该从真实的性能监控系统中获取）
  useEffect(() => {
    if (!enabled) return;

    const interval = setInterval(() => {
      // 获取缓存大小
      setCacheSize(globalCache.size());

      // 模拟统计数据（实际应用中应该从性能监控API获取）
      setStats(prev => ({
        ...prev,
        totalRequests: prev.totalRequests + Math.floor(Math.random() * 3),
        cacheHits: prev.cacheHits + Math.floor(Math.random() * 2),
        errorCount: prev.errorCount + (Math.random() > 0.9 ? 1 : 0)
      }));
    }, 5000);

    return () => clearInterval(interval);
  }, [enabled]);

  // 计算缓存命中率
  const cacheHitRate = stats.totalRequests > 0
    ? (stats.cacheHits / stats.totalRequests * 100).toFixed(1)
    : '0.0';

  // 获取缓存状态颜色
  const getCacheStatusColor = () => {
    if (cacheSize === 0) return 'danger';
    if (cacheSize < 5) return 'warning';
    if (cacheSize < 20) return 'info';
    return 'success';
  };

  // 获取性能状态
  const getPerformanceStatus = () => {
    if (stats.errorCount > stats.totalRequests * 0.1) return 'danger';
    if (parseFloat(cacheHitRate) < 30) return 'warning';
    return 'success';
  };

  // 清理缓存
  const handleClearCache = () => {
    globalCache.clear();
    setCacheSize(0);
  };

  if (!enabled) {
    return null;
  }

  return (
    <>
      {/* 浮动按钮 */}
      <Button
        variant="outline-secondary"
        size="sm"
        onClick={() => setIsVisible(!isVisible)}
        style={{
          position: 'fixed',
          bottom: '20px',
          right: '20px',
          zIndex: 1000,
          borderRadius: '50%',
          width: '50px',
          height: '50px',
          padding: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: '0 2px 10px rgba(0,0,0,0.2)'
        }}
      >
        📊
      </Button>

      {/* 性能监控面板 */}
      {isVisible && (
        <Card
          style={{
            position: 'fixed',
            bottom: '80px',
            right: '20px',
            width: '320px',
            zIndex: 999,
            maxHeight: '70vh',
            overflowY: 'auto',
            boxShadow: '0 4px 20px rgba(0,0,0,0.15)'
          }}
        >
          <Card.Header className="d-flex justify-content-between align-items-center">
            <h6 className="mb-0">性能监控</h6>
            <Button
              variant="outline-secondary"
              size="sm"
              onClick={() => setIsVisible(false)}
            >
              ×
            </Button>
          </Card.Header>
          <Card.Body className="p-3">
            {/* 缓存状态 */}
            <div className="mb-3">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <span className="fw-bold">缓存状态</span>
                <Badge bg={getCacheStatusColor()}>
                  {cacheSize} 项
                </Badge>
              </div>
              <ProgressBar
                variant={getCacheStatusColor()}
                now={Math.min(cacheSize * 5, 100)}
                style={{ height: '6px' }}
              />
            </div>

            {/* 性能指标 */}
            <div className="mb-3">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <span className="fw-bold">整体性能</span>
                <Badge bg={getPerformanceStatus()}>
                  {cacheHitRate}% 命中率
                </Badge>
              </div>
              <ProgressBar
                variant={getPerformanceStatus()}
                now={parseFloat(cacheHitRate)}
                style={{ height: '6px' }}
              />
            </div>

            {/* 详细统计 */}
            {showDetails && (
              <div className="small text-muted">
                <div className="d-flex justify-content-between mb-1">
                  <span>总请求数:</span>
                  <span>{stats.totalRequests}</span>
                </div>
                <div className="d-flex justify-content-between mb-1">
                  <span>缓存命中:</span>
                  <span className="text-success">{stats.cacheHits}</span>
                </div>
                <div className="d-flex justify-content-between mb-1">
                  <span>缓存未命中:</span>
                  <span className="text-warning">{stats.cacheMisses}</span>
                </div>
                <div className="d-flex justify-content-between mb-1">
                  <span>错误次数:</span>
                  <span className="text-danger">{stats.errorCount}</span>
                </div>
                <div className="d-flex justify-content-between">
                  <span>平均响应时间:</span>
                  <span>{stats.averageResponseTime}ms</span>
                </div>
              </div>
            )}

            {/* 操作按钮 */}
            <div className="d-grid gap-2 mt-3">
              <Button
                variant="outline-primary"
                size="sm"
                onClick={() => window.location.reload()}
              >
                🔄 刷新页面
              </Button>
              <Button
                variant="outline-warning"
                size="sm"
                onClick={handleClearCache}
              >
                🗑️ 清理缓存
              </Button>
              <Button
                variant="outline-info"
                size="sm"
                onClick={() => {
                  // 打开Chrome DevTools
                  if (typeof window !== 'undefined' && window.console) {
                    console.log('性能数据:', stats);
                    console.log('缓存大小:', cacheSize);
                    console.log('使用 Chrome DevTools 查看详细性能信息');
                  }
                }}
              >
                📈 DevTools
              </Button>
            </div>

            {/* 提示信息 */}
            <div className="mt-3 p-2 bg-light rounded">
              <small className="text-muted">
                💡 提示：使用 Chrome DevTools 的 Network 和 Performance 面板可以查看更详细的性能数据
              </small>
            </div>
          </Card.Body>
        </Card>
      )}
    </>
  );
};

export default PerformanceMonitor;
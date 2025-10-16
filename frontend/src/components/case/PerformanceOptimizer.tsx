import React, { useEffect, useState, useCallback } from 'react';
import { Card, Switch, Button, Space, Typography, Progress, Alert, Tag } from 'antd';
import {
  MonitorOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  InfoCircleOutlined
} from '@ant-design/icons';
import { usePerformanceMonitor, Display1080pOptimizer } from '@/utils/performance';

const { Text, Title } = Typography;

interface PerformanceOptimizerProps {
  enableAutoOptimization?: boolean;
  showMetrics?: boolean;
  className?: string;
}

interface PerformanceMetrics {
  renderTime: { average: number; min: number; max: number; count: number };
  memory: { average: number; min: number; max: number; count: number };
  [key: string]: any;
}

/**
 * 性能优化器组件
 * 专为1080p显示器提供性能监控和优化功能
 */
const PerformanceOptimizer: React.FC<PerformanceOptimizerProps> = ({
  enableAutoOptimization = true,
  showMetrics = true,
  className = ''
}) => {
  const [isOptimized, setIsOptimized] = useState(false);
  const [autoOptimizationEnabled, setAutoOptimizationEnabled] = useState(enableAutoOptimization);
  const [testResults, setTestResults] = useState<any>(null);
  const [isTesting, setIsTesting] = useState(false);

  const {
    metrics,
    isRunning,
    startMonitoring,
    stopMonitoring,
    run1080pTest
  } = usePerformanceMonitor({
    enableMonitoring: process.env.NODE_ENV === 'development'
  });

  // 自动应用1080p优化
  const apply1080pOptimizations = useCallback(() => {
    const element = document.documentElement;
    Display1080pOptimizer.apply1080pOptimizations(element);

    // 添加1080p优化类名
    if (Display1080pOptimizer.isWithin1080pRange()) {
      document.body.classList.add('display-1080p-optimized');
    } else {
      document.body.classList.remove('display-1080p-optimized');
    }

    setIsOptimized(true);
  }, []);

  // 移除1080p优化
  const remove1080pOptimizations = useCallback(() => {
    document.body.classList.remove('display-1080p-optimized');
    setIsOptimized(false);
  }, []);

  // 运行性能测试
  const runPerformanceTest = useCallback(async () => {
    setIsTesting(true);
    try {
      const results = await run1080pTest();
      setTestResults(results);

      if (results.isOptimized) {
        console.log('✅ 1080p性能测试通过');
      } else {
        console.warn('⚠️ 1080p性能需要优化');
      }
    } catch (error) {
      console.error('性能测试失败:', error);
    } finally {
      setIsTesting(false);
    }
  }, [run1080pTest]);

  // 初始化时自动应用优化
  useEffect(() => {
    if (autoOptimizationEnabled && Display1080pOptimizer.isWithin1080pRange()) {
      apply1080pOptimizations();
    }

    // 监听窗口大小变化
    const handleResize = () => {
      if (autoOptimizationEnabled && Display1080pOptimizer.isWithin1080pRange()) {
        apply1080pOptimizations();
      } else if (!Display1080pOptimizer.isWithin1080pRange()) {
        remove1080pOptimizations();
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [autoOptimizationEnabled, apply1080pOptimizations, remove1080pOptimizations]);

  // 获取性能状态
  const getPerformanceStatus = useCallback(() => {
    if (!metrics.renderTime) return 'unknown';

    const avgRenderTime = metrics.renderTime.average;
    if (avgRenderTime < 16.67) return 'excellent'; // 60fps
    if (avgRenderTime < 33.33) return 'good'; // 30fps
    if (avgRenderTime < 100) return 'acceptable';
    return 'poor';
  }, [metrics]);

  // 获取性能状态颜色
  const getStatusColor = useCallback((status: string) => {
    switch (status) {
      case 'excellent': return 'success';
      case 'good': return 'processing';
      case 'acceptable': return 'warning';
      case 'poor': return 'error';
      default: return 'default';
    }
  }, []);

  // 获取性能分数
  const getPerformanceScore = useCallback(() => {
    if (!metrics.renderTime) return 0;

    const avgRenderTime = metrics.renderTime.average;
    const maxScore = 100;

    if (avgRenderTime <= 16.67) return maxScore; // 60fps = 100分
    if (avgRenderTime <= 33.33) return 80; // 30fps = 80分
    if (avgRenderTime <= 50) return 60; // 20fps = 60分
    if (avgRenderTime <= 100) return 40; // 10fps = 40分
    return 20; // <10fps = 20分
  }, [metrics]);

  const performanceStatus = getPerformanceStatus();
  const performanceScore = getPerformanceScore();

  return (
    <Card
      className={`performance-optimizer ${className}`}
      title={
        <Space>
          <MonitorOutlined />
          <span>性能优化器</span>
          {Display1080pOptimizer.isWithin1080pRange() && (
            <Tag color="blue">1080p显示器</Tag>
          )}
        </Space>
      }
      size="small"
      extra={
        <Space>
          <Text type="secondary">自动优化</Text>
          <Switch
            checked={autoOptimizationEnabled}
            onChange={setAutoOptimizationEnabled}
            size="small"
          />
        </Space>
      }
    >
      {/* 状态指示器 */}
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Tag color={getStatusColor(performanceStatus)}>
            {performanceStatus === 'excellent' && <CheckCircleOutlined />}
            {performanceStatus === 'good' && <ThunderboltOutlined />}
            {performanceStatus === 'acceptable' && <WarningOutlined />}
            {performanceStatus === 'poor' && <WarningOutlined />}
            {performanceStatus === 'unknown' && <InfoCircleOutlined />}
            性能状态: {performanceStatus === 'excellent' ? '优秀' :
                     performanceStatus === 'good' ? '良好' :
                     performanceStatus === 'acceptable' ? '一般' :
                     performanceStatus === 'poor' ? '较差' : '未知'}
          </Tag>
          {Display1080pOptimizer.is1080pDisplay() && (
            <Tag color="green">标准1080p</Tag>
          )}
          {Display1080pOptimizer.isWithin1080pRange() && !Display1080pOptimizer.is1080pDisplay() && (
            <Tag color="orange">1080p范围</Tag>
          )}
        </Space>
      </div>

      {/* 性能评分 */}
      <div style={{ marginBottom: 16 }}>
        <Text strong>性能评分: </Text>
        <Progress
          percent={performanceScore}
          size="small"
          status={performanceScore >= 80 ? 'success' : performanceScore >= 60 ? 'normal' : 'exception'}
          format={(percent) => `${percent}分`}
        />
      </div>

      {/* 控制按钮 */}
      <Space style={{ marginBottom: 16 }}>
        <Button
          type={isOptimized ? 'primary' : 'default'}
          size="small"
          icon={<ThunderboltOutlined />}
          onClick={isOptimized ? remove1080pOptimizations : apply1080pOptimizations}
          disabled={!Display1080pOptimizer.isWithin1080pRange()}
        >
          {isOptimized ? '已优化' : '应用优化'}
        </Button>

        <Button
          size="small"
          icon={<MonitorOutlined />}
          loading={isTesting}
          onClick={runPerformanceTest}
        >
          {isTesting ? '测试中...' : '性能测试'}
        </Button>

        <Button
          size="small"
          type={isRunning ? 'default' : 'primary'}
          onClick={isRunning ? stopMonitoring : startMonitoring}
        >
          {isRunning ? '停止监控' : '开始监控'}
        </Button>
      </Space>

      {/* 显示检测结果 */}
      {testResults && (
        <Alert
          message="测试结果"
          description={
            <div>
              <div>总耗时: {testResults.totalTime.toFixed(2)}ms</div>
              <div>状态: {testResults.isOptimized ? '✅ 通过' : '⚠️ 需要优化'}</div>
              {testResults.metrics.renderTime && (
                <div>平均渲染时间: {testResults.metrics.renderTime.average.toFixed(2)}ms</div>
              )}
            </div>
          }
          type={testResults.isOptimized ? 'success' : 'warning'}
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      {/* 显示详细指标 */}
      {showMetrics && Object.keys(metrics).length > 0 && (
        <div>
          <Title level={5}>性能指标</Title>
          <div style={{ fontSize: '12px', color: '#666' }}>
            {Object.entries(metrics).map(([name, data]: [string, any]) => (
              <div key={name} style={{ marginBottom: 8 }}>
                <Text strong>{name}: </Text>
                <span>平均 {data.average?.toFixed(2)}ms, </span>
                <span>最小 {data.min?.toFixed(2)}ms, </span>
                <span>最大 {data.max?.toFixed(2)}ms</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 1080p提示信息 */}
      {Display1080pOptimizer.isWithin1080pRange() && (
        <Alert
          message="1080p显示器优化"
          description="检测到您的显示器在1080p范围内，已自动应用紧凑布局优化以提升使用体验。"
          type="info"
          showIcon
          style={{ marginTop: 16 }}
        />
      )}
    </Card>
  );
};

export default PerformanceOptimizer;
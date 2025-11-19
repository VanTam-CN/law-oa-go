import React, { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Progress, Table, Tag, Button, Modal, Typography } from 'antd'
import {
  ReloadOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

const { Text } = Typography

// 性能指标接口
interface PerformanceMetrics {
  timestamp: string
  memoryUsage: number
  memoryTotal: number
  cpuUsage: number
  responseTime: number
  throughput: number
  errorRate: number
  cacheHitRate: number
  activeConnections: number
}

// 慢查询接口
interface SlowQuery {
  id: string
  query: string
  executionTime: number
  timestamp: string
  frequency: number
}

// 系统健康状态
interface SystemHealth {
  status: 'HEALTHY' | 'WARNING' | 'CRITICAL'
  cpu: number
  memory: number
  disk: number
  network: number
}

/**
 * 性能监控仪表板
 */
const PerformanceDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null)
  const [slowQueries, setSlowQueries] = useState<SlowQuery[]>([])
  const [systemHealth, setSystemHealth] = useState<SystemHealth | null>(null)
  const [loading, setLoading] = useState(false)
  const [selectedQuery, setSelectedQuery] = useState<SlowQuery | null>(null)

  // 模拟获取性能指标
  const fetchMetrics = async () => {
    setLoading(true)
    try {
      // 这里应该调用真实的API
      const mockMetrics: PerformanceMetrics = {
        timestamp: new Date().toISOString(),
        memoryUsage: 2.5 * 1024 * 1024 * 1024, // 2.5GB
        memoryTotal: 8 * 1024 * 1024 * 1024, // 8GB
        cpuUsage: Math.random() * 100,
        responseTime: Math.random() * 1000,
        throughput: Math.random() * 1000,
        errorRate: Math.random() * 5,
        cacheHitRate: 85 + Math.random() * 10,
        activeConnections: Math.floor(Math.random() * 100),
      }

      setMetrics(mockMetrics)

      // 模拟慢查询数据
      const mockSlowQueries: SlowQuery[] = [
        {
          id: '1',
          query: 'SELECT * FROM lf_case WHERE case_name LIKE ?',
          executionTime: 2500,
          timestamp: new Date().toISOString(),
          frequency: 150,
        },
        {
          id: '2',
          query: 'SELECT COUNT(*) FROM lf_document WHERE upload_time > ?',
          executionTime: 1800,
          timestamp: new Date().toISOString(),
          frequency: 80,
        },
      ]

      setSlowQueries(mockSlowQueries)

      // 模拟系统健康状态
      const mockHealth: SystemHealth = {
        status:
          mockMetrics.cpuUsage > 80
            ? 'CRITICAL'
            : mockMetrics.cpuUsage > 60
              ? 'WARNING'
              : 'HEALTHY',
        cpu: mockMetrics.cpuUsage,
        memory: (mockMetrics.memoryUsage / mockMetrics.memoryTotal) * 100,
        disk: Math.random() * 100,
        network: Math.random() * 100,
      }

      setSystemHealth(mockHealth)
    } catch (error) {
      console.error('获取性能指标失败:', error)
    } finally {
      setLoading(false)
    }
  }

  // 初始化加载数据
  useEffect(() => {
    fetchMetrics()

    // 每30秒刷新一次
    const interval = setInterval(fetchMetrics, 30000)
    return () => clearInterval(interval)
  }, [])

  // 格式化内存大小
  const formatMemory = (bytes: number) => {
    const gb = bytes / (1024 * 1024 * 1024)
    return `${gb.toFixed(2)} GB`
  }

  // 获取健康状态颜色
  const getHealthColor = (status: SystemHealth['status']) => {
    switch (status) {
      case 'HEALTHY':
        return '#52c41a'
      case 'WARNING':
        return '#faad14'
      case 'CRITICAL':
        return '#f5222d'
      default:
        return '#d9d9d9'
    }
  }

  // 获取健康状态图标
  const getHealthIcon = (status: SystemHealth['status']) => {
    switch (status) {
      case 'HEALTHY':
        return <CheckCircleOutlined />
      case 'WARNING':
        return <ExclamationCircleOutlined />
      case 'CRITICAL':
        return <WarningOutlined />
      default:
        return <ExclamationCircleOutlined />
    }
  }

  // 慢查询表格列
  const slowQueryColumns: ColumnsType<SlowQuery> = [
    {
      title: '查询语句',
      dataIndex: 'query',
      key: 'query',
      ellipsis: true,
    },
    {
      title: '执行时间',
      dataIndex: 'executionTime',
      key: 'executionTime',
      render: (time: number) => (
        <Tag color={time > 2000 ? 'red' : time > 1000 ? 'orange' : 'green'}>{time}ms</Tag>
      ),
    },
    {
      title: '执行频率',
      dataIndex: 'frequency',
      key: 'frequency',
      render: (freq: number) => `${freq}/min`,
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Button type='link' onClick={() => setSelectedQuery(record)}>
          详情
        </Button>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card
            title='性能监控'
            extra={
              <Button icon={<ReloadOutlined />} onClick={fetchMetrics} loading={loading}>
                刷新
              </Button>
            }
          >
            <Row gutter={[16, 16]}>
              {/* 系统健康状态 */}
              <Col span={24}>
                <Card title='系统健康状态' size='small'>
                  <Row gutter={[16, 16]}>
                    <Col span={6}>
                      <Statistic
                        title='CPU使用率'
                        value={systemHealth?.cpu || 0}
                        suffix='%'
                        valueStyle={{
                          color: getHealthColor(systemHealth?.status || 'WARNING'),
                        }}
                      />
                      <Progress
                        percent={systemHealth?.cpu || 0}
                        size='small'
                        strokeColor={getHealthColor(systemHealth?.status || 'WARNING')}
                      />
                    </Col>
                    <Col span={6}>
                      <Statistic
                        title='内存使用率'
                        value={systemHealth?.memory || 0}
                        suffix='%'
                        valueStyle={{
                          color: getHealthColor(systemHealth?.status || 'WARNING'),
                        }}
                      />
                      <Progress
                        percent={systemHealth?.memory || 0}
                        size='small'
                        strokeColor={getHealthColor(systemHealth?.status || 'WARNING')}
                      />
                    </Col>
                    <Col span={6}>
                      <Statistic
                        title='磁盘使用率'
                        value={systemHealth?.disk || 0}
                        suffix='%'
                        valueStyle={{ color: '#52c41a' }}
                      />
                      <Progress
                        percent={systemHealth?.disk || 0}
                        size='small'
                        strokeColor='#52c41a'
                      />
                    </Col>
                    <Col span={6}>
                      <Statistic
                        title='网络使用率'
                        value={systemHealth?.network || 0}
                        suffix='%'
                        valueStyle={{ color: '#52c41a' }}
                      />
                      <Progress
                        percent={systemHealth?.network || 0}
                        size='small'
                        strokeColor='#52c41a'
                      />
                    </Col>
                  </Row>
                </Card>
              </Col>

              {/* 性能指标 */}
              <Col span={6}>
                <Card size='small'>
                  <Statistic
                    title='内存使用'
                    value={metrics ? formatMemory(metrics.memoryUsage) : '0'}
                    suffix={`/ ${metrics ? formatMemory(metrics.memoryTotal) : '0'}`}
                  />
                </Card>
              </Col>
              <Col span={6}>
                <Card size='small'>
                  <Statistic
                    title='平均响应时间'
                    value={metrics?.responseTime || 0}
                    suffix='ms'
                    valueStyle={{
                      color:
                        metrics?.responseTime && metrics.responseTime > 1000
                          ? '#f5222d'
                          : '#52c41a',
                    }}
                  />
                </Card>
              </Col>
              <Col span={6}>
                <Card size='small'>
                  <Statistic title='吞吐量' value={metrics?.throughput || 0} suffix='req/s' />
                </Card>
              </Col>
              <Col span={6}>
                <Card size='small'>
                  <Statistic
                    title='缓存命中率'
                    value={metrics?.cacheHitRate || 0}
                    suffix='%'
                    valueStyle={{
                      color:
                        metrics?.cacheHitRate && metrics.cacheHitRate > 80 ? '#52c41a' : '#faad14',
                    }}
                  />
                </Card>
              </Col>

              {/* 慢查询表格 */}
              <Col span={24}>
                <Card title='慢查询监控' size='small'>
                  <Table
                    columns={slowQueryColumns}
                    dataSource={slowQueries}
                    rowKey='id'
                    pagination={false}
                    size='small'
                  />
                </Card>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* 慢查询详情弹窗 */}
      <Modal
        title='慢查询详情'
        open={!!selectedQuery}
        onCancel={() => setSelectedQuery(null)}
        footer={null}
        width={800}
      >
        {selectedQuery && (
          <div>
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Text strong>查询语句：</Text>
                <div
                  style={{
                    background: '#f5f5f5',
                    padding: '8px',
                    borderRadius: '4px',
                    marginTop: '4px',
                    fontFamily: 'monospace',
                  }}
                >
                  {selectedQuery.query}
                </div>
              </Col>
              <Col span={12}>
                <Text strong>执行时间：</Text>
                <div style={{ marginTop: '4px' }}>
                  <Tag color={selectedQuery.executionTime > 2000 ? 'red' : 'orange'}>
                    {selectedQuery.executionTime}ms
                  </Tag>
                </div>
                <Text strong style={{ display: 'block', marginTop: '8px' }}>
                  执行频率：
                </Text>
                <div style={{ marginTop: '4px' }}>{selectedQuery.frequency}/min</div>
              </Col>
            </Row>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default PerformanceDashboard

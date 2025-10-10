import React, { useState, useEffect } from 'react';
import { Button, Card, Modal, Badge, Table } from 'react-bootstrap';
import { devToolsValidator } from '../utils/devToolsValidation';

interface LogEntry {
  timestamp: number;
  type: string;
  category: string;
  message: string;
  data?: any;
  duration?: number;
}

const DevToolsPanel: React.FC = () => {
  const [show, setShow] = useState(false);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [activeTab, setActiveTab] = useState('all');

  useEffect(() => {
    if (process.env.NODE_ENV === 'development') {
      // 定期更新日志
      const interval = setInterval(() => {
        setLogs(devToolsValidator.getLogs());
      }, 1000);

      return () => clearInterval(interval);
    }
  }, []);

  // 只在开发模式下显示
  if (process.env.NODE_ENV !== 'development') {
    return null;
  }

  const filteredLogs = activeTab === 'all'
    ? logs
    : logs.filter(log => log.type === activeTab);

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'api': return 'info';
      case 'data': return 'success';
      case 'error': return 'danger';
      case 'performance': return 'warning';
      case 'render': return 'secondary';
      default: return 'primary';
    }
  };

  const getApiStats = () => {
    const apiLogs = logs.filter(log => log.type === 'api');
    const avgDuration = apiLogs.length > 0
      ? apiLogs.reduce((sum, log) => sum + (log.duration || 0), 0) / apiLogs.length
      : 0;
    const slowest = apiLogs.length > 0
      ? Math.max(...apiLogs.map(log => log.duration || 0))
      : 0;

    return { count: apiLogs.length, avgDuration: avgDuration.toFixed(2), slowest };
  };

  const getDataStats = () => {
    const dataLogs = logs.filter(log => log.type === 'data');
    const totalItems = dataLogs.reduce((sum, log) => sum + (log.data?.itemCount || 0), 0);
    const errorCount = dataLogs.filter(log => !log.data?.hasRequiredFields || log.data?.isEmpty).length;

    return { count: dataLogs.length, totalItems, errorCount };
  };

  return (
    <>
      {/* 开发工具浮动按钮 */}
      <Button
        variant="dark"
        size="sm"
        className="position-fixed"
        style={{
          bottom: '20px',
          right: '20px',
          zIndex: 9999,
          borderRadius: '50%',
          width: '50px',
          height: '50px',
          fontSize: '20px'
        }}
        onClick={() => setShow(true)}
        title="开发工具面板"
      >
        🔧
      </Button>

      {/* 开发工具模态框 */}
      <Modal
        show={show}
        onHide={() => setShow(false)}
        size="xl"
        fullscreen="lg-down"
      >
        <Modal.Header closeButton>
          <Modal.Title>
            🔧 Law OA 开发工具面板
            <Badge bg="info" className="ms-2">
              {logs.length} 条日志
            </Badge>
          </Modal.Title>
        </Modal.Header>

        <Modal.Body>
          {/* 统计信息 */}
          <div className="row mb-3">
            <div className="col-md-3">
              <Card className="text-center">
                <Card.Body>
                  <h6 className="text-muted">API调用</h6>
                  <h4>{getApiStats().count}</h4>
                  <small className="text-muted">
                    平均: {getApiStats().avgDuration}ms
                  </small>
                </Card.Body>
              </Card>
            </div>

            <div className="col-md-3">
              <Card className="text-center">
                <Card.Body>
                  <h6 className="text-muted">数据验证</h6>
                  <h4>{getDataStats().count}</h4>
                  <small className="text-muted">
                    错误: {getDataStats().errorCount}
                  </small>
                </Card.Body>
              </Card>
            </div>

            <div className="col-md-3">
              <Card className="text-center">
                <Card.Body>
                  <h6 className="text-muted">性能日志</h6>
                  <h4>{logs.filter(log => log.type === 'performance').length}</h4>
                  <small className="text-muted">
                    渲染: {logs.filter(log => log.type === 'render').length}
                  </small>
                </Card.Body>
              </Card>
            </div>

            <div className="col-md-3">
              <Card className="text-center">
                <Card.Body>
                  <h6 className="text-muted">错误</h6>
                  <h4 className="text-danger">{logs.filter(log => log.type === 'error').length}</h4>
                  <small className="text-muted">
                    总计: {logs.length}
                  </small>
                </Card.Body>
              </Card>
            </div>
          </div>

          {/* 控制按钮 */}
          <div className="mb-3">
            <Button variant="outline-primary" size="sm" onClick={() => devToolsValidator.validateCurrentData()}>
              📊 验证数据
            </Button>
            <Button variant="outline-info" size="sm" onClick={() => devToolsValidator.getPerformanceReport()} className="ms-2">
              ⚡ 性能报告
            </Button>
            <Button variant="outline-warning" size="sm" onClick={() => devToolsValidator.clearLogs()} className="ms-2">
              🗑️ 清除日志
            </Button>
            <Button variant="outline-success" size="sm" onClick={() => {
              const logs = devToolsValidator.exportLogs();
              const blob = new Blob([logs], { type: 'application/json' });
              const url = URL.createObjectURL(blob);
              const a = document.createElement('a');
              a.href = url;
              a.download = `devtools-logs-${new Date().toISOString()}.json`;
              a.click();
              URL.revokeObjectURL(url);
            }} className="ms-2">
              💾 导出日志
            </Button>
          </div>

          {/* 标签页 */}
          <div className="mb-3">
            <Button
              variant={activeTab === 'all' ? 'primary' : 'outline-primary'}
              size="sm"
              onClick={() => setActiveTab('all')}
              className="me-2"
            >
              全部 ({logs.length})
            </Button>
            <Button
              variant={activeTab === 'api' ? 'info' : 'outline-info'}
              size="sm"
              onClick={() => setActiveTab('api')}
              className="me-2"
            >
              API ({logs.filter(log => log.type === 'api').length})
            </Button>
            <Button
              variant={activeTab === 'data' ? 'success' : 'outline-success'}
              size="sm"
              onClick={() => setActiveTab('data')}
              className="me-2"
            >
              数据 ({logs.filter(log => log.type === 'data').length})
            </Button>
            <Button
              variant={activeTab === 'performance' ? 'warning' : 'outline-warning'}
              size="sm"
              onClick={() => setActiveTab('performance')}
              className="me-2"
            >
              性能 ({logs.filter(log => log.type === 'performance').length})
            </Button>
            <Button
              variant={activeTab === 'error' ? 'danger' : 'outline-danger'}
              size="sm"
              onClick={() => setActiveTab('error')}
              className="me-2"
            >
              错误 ({logs.filter(log => log.type === 'error').length})
            </Button>
          </div>

          {/* 日志表格 */}
          <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
            <Table striped bordered hover size="sm">
              <thead>
                <tr>
                  <th>时间</th>
                  <th>类型</th>
                  <th>分类</th>
                  <th>消息</th>
                  <th>详情</th>
                </tr>
              </thead>
              <tbody>
                {filteredLogs.slice(-100).reverse().map((log, index) => (
                  <tr key={index}>
                    <td style={{ fontSize: '12px', whiteSpace: 'nowrap' }}>
                      {new Date(log.timestamp).toLocaleTimeString()}
                    </td>
                    <td>
                      <Badge bg={getTypeColor(log.type)}>
                        {log.type.toUpperCase()}
                      </Badge>
                    </td>
                    <td style={{ fontSize: '12px' }}>{log.category}</td>
                    <td style={{ fontSize: '12px', maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {log.message}
                    </td>
                    <td style={{ fontSize: '11px' }}>
                      {log.type === 'api' && (
                        <div>
                          <small>耗时: {log.duration}ms</small>
                          {log.data && (
                            <details className="mt-1">
                              <summary style={{ cursor: 'pointer' }}>数据</summary>
                              <pre style={{ fontSize: '10px', maxHeight: '100px', overflow: 'auto' }}>
                                {JSON.stringify(log.data, null, 2)}
                              </pre>
                            </details>
                          )}
                        </div>
                      )}
                      {log.type === 'data' && (
                        <div>
                          <small>数量: {log.data?.itemCount}</small>
                          <small className="d-block text-warning">
                            {!log.data?.hasRequiredFields && '缺少字段'}
                            {log.data?.isEmpty && '空数据'}
                          </small>
                        </div>
                      )}
                      {log.type === 'performance' && (
                        <small>
                          {log.data?.metric}: {log.data?.value}{log.data?.unit}
                        </small>
                      )}
                      {log.type === 'error' && (
                        <small className="text-danger">
                          {log.data?.context && '查看控制台'}
                        </small>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>

          {/* 开发工具说明 */}
          <Card className="mt-3">
            <Card.Header>
              <Card.Title as="h6">📖 使用说明</Card.Title>
            </Card.Header>
            <Card.Body>
              <div className="row">
                <div className="col-md-6">
                  <h6>控制台命令:</h6>
                  <ul>
                    <li><code>window.devTools.getLogs()</code> - 获取所有日志</li>
                    <li><code>window.devTools.clearLogs()</code> - 清除日志</li>
                    <li><code>window.devTools.validateData()</code> - 手动验证数据</li>
                    <li><code>window.devTools.performanceReport()</code> - 性能报告</li>
                  </ul>
                </div>
                <div className="col-md-6">
                  <h6>功能特性:</h6>
                  <ul>
                    <li>实时API调用监控</li>
                    <li>数据完整性验证</li>
                    <li>性能指标追踪</li>
                    <li>错误日志记录</li>
                    <li>组件渲染监控</li>
                  </ul>
                </div>
              </div>
            </Card.Body>
          </Card>
        </Modal.Body>
      </Modal>
    </>
  );
};

export default DevToolsPanel;
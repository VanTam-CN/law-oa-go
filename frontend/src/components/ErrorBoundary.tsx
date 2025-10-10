import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Alert, Button, Card } from 'react-bootstrap';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null
    };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return {
      hasError: true,
      error
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({
      error,
      errorInfo
    });

    // 记录错误到外部服务（如果需要）
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // 在开发环境中打印错误
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null
    });
  };

  getErrorType = (error: Error): string => {
    if (error.name === 'TypeError') return '类型错误';
    if (error.name === 'ReferenceError') return '引用错误';
    if (error.name === 'NetworkError') return '网络错误';
    if (error.message.includes('401')) return '认证错误';
    if (error.message.includes('403')) return '权限错误';
    if (error.message.includes('404')) return '资源未找到';
    if (error.message.includes('500')) return '服务器错误';
    return '未知错误';
  };

  getErrorSeverity = (error: Error): 'danger' | 'warning' | 'info' => {
    if (error.message.includes('401') || error.message.includes('403')) return 'warning';
    if (error.message.includes('500') || error.name === 'NetworkError') return 'danger';
    return 'danger';
  };

  getErrorMessage = (error: Error): string => {
    const message = error.message;

    // 网络错误
    if (error.name === 'NetworkError' || message.includes('fetch')) {
      return '网络连接失败，请检查网络设置后重试';
    }

    // HTTP状态码错误
    if (message.includes('401')) {
      return '登录已过期，请重新登录';
    }
    if (message.includes('403')) {
      return '权限不足，请联系管理员';
    }
    if (message.includes('404')) {
      return '请求的资源不存在';
    }
    if (message.includes('500')) {
      return '服务器内部错误，请稍后重试';
    }
    if (message.includes('502') || message.includes('503')) {
      return '服务暂时不可用，请稍后重试';
    }

    // 数据格式错误
    if (message.includes('JSON') || message.includes('parse')) {
      return '数据格式错误，请联系技术支持';
    }

    // 超时错误
    if (message.includes('timeout') || message.includes('超时')) {
      return '请求超时，请检查网络连接';
    }

    return message || '发生了未知错误';
  };

  getErrorSolution = (error: Error): string[] => {
    const solutions: string[] = [];
    const message = error.message;

    if (error.name === 'NetworkError' || message.includes('fetch')) {
      solutions.push('检查网络连接');
      solutions.push('确认服务器地址正确');
      solutions.push('检查防火墙设置');
    }

    if (message.includes('401')) {
      solutions.push('重新登录系统');
      solutions.push('检查账号密码');
    }

    if (message.includes('403')) {
      solutions.push('联系管理员分配权限');
      solutions.push('确认账号状态正常');
    }

    if (message.includes('500')) {
      solutions.push('稍后重试');
      solutions.push('联系技术支持');
    }

    if (message.includes('timeout')) {
      solutions.push('检查网络速度');
      solutions.push('减少同时操作的次数');
    }

    if (solutions.length === 0) {
      solutions.push('刷新页面重试');
      solutions.push('联系技术支持');
    }

    return solutions;
  };

  render() {
    if (this.state.hasError) {
      const { error } = this.state;
      if (!error) return null;
      const errorType = this.getErrorType(error);
      const severity = this.getErrorSeverity(error);
      const errorMessage = this.getErrorMessage(error);
      const solutions = this.getErrorSolution(error);

      // 如果提供了自定义fallback，使用它
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="error-boundary-container p-4">
          <Card border={severity} className="text-center">
            <Card.Header className="bg-light">
              <Card.Title as="h4" className={`text-${severity}`}>
                ⚠️ 页面出现错误
              </Card.Title>
            </Card.Header>
            <Card.Body>
              <Alert variant={severity} className="text-start">
                <Alert.Heading as="h5">错误类型: {errorType}</Alert.Heading>
                <p className="mb-2">{errorMessage}</p>

                {/* 在开发环境中显示详细错误信息 */}
                {process.env.NODE_ENV === 'development' && (
                  <details className="mt-3">
                    <summary style={{ cursor: 'pointer' }}>详细信息</summary>
                    <pre style={{ fontSize: '12px', marginTop: '10px', textAlign: 'left' }}>
                      {error.stack}
                    </pre>
                  </details>
                )}
              </Alert>

              <div className="mt-4">
                <h6>解决方案:</h6>
                <ul className="text-start">
                  {solutions.map((solution, index) => (
                    <li key={index}>{solution}</li>
                  ))}
                </ul>
              </div>

              <div className="mt-4">
                <Button variant="primary" onClick={this.handleReset} className="me-2">
                  🔄 重新加载
                </Button>
                <Button variant="outline-secondary" onClick={() => window.location.href = '/'}>
                  🏠 返回首页
                </Button>
              </div>

              <div className="mt-3">
                <small className="text-muted">
                  如果问题持续存在，请联系技术支持团队
                </small>
              </div>
            </Card.Body>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
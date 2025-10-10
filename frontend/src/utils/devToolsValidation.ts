/**
 * Chrome DevTools 验证工具
 * 在开发模式下提供详细的API调用和数据显示验证信息
 */

interface DevToolsLog {
  timestamp: number;
  type: 'api' | 'data' | 'render' | 'error' | 'performance';
  category: string;
  message: string;
  data?: any;
  duration?: number;
}

interface ApiCallLog extends DevToolsLog {
  type: 'api';
  method: string;
  url: string;
  request?: any;
  response?: any;
  statusCode?: number;
  duration: number;
}

interface DataValidationLog extends DevToolsLog {
  type: 'data';
  dataType: string;
  validation: {
    isEmpty: boolean;
    itemCount: number;
    hasRequiredFields: boolean;
    missingFields?: string[];
    sampleItem?: any;
  };
}

interface PerformanceLog extends DevToolsLog {
  type: 'performance';
  metric: string;
  value: number;
  unit: string;
}

class DevToolsValidator {
  private logs: DevToolsLog[] = [];
  private isEnabled: boolean = false;
  private maxLogs: number = 1000;
  private apiCallTimes: Map<string, number> = new Map();

  constructor() {
    this.isEnabled = process.env.NODE_ENV === 'development';
    if (this.isEnabled) {
      this.initDevTools();
    }
  }

  private initDevTools() {
    // 在控制台添加开发工具信息
    console.log('%c🔧 Law OA DevTools Validator Enabled', 'color: #4CAF50; font-size: 14px; font-weight: bold;');
    console.log('%cAvailable commands:', 'color: #2196F3; font-size: 12px;');
    console.log('%c- window.devTools.getLogs() - 获取所有日志', 'color: #666;');
    console.log('%c- window.devTools.clearLogs() - 清除日志', 'color: #666;');
    console.log('%c- window.devTools.validateData() - 手动验证数据', 'color: #666;');
    console.log('%c- window.devTools.performanceReport() - 性能报告', 'color: #666;');

    // 暴露到全局
    (window as any).devTools = {
      getLogs: this.getLogs.bind(this),
      clearLogs: this.clearLogs.bind(this),
      validateData: this.validateCurrentData.bind(this),
      performanceReport: this.getPerformanceReport.bind(this),
      exportLogs: this.exportLogs.bind(this)
    };
  }

  // 记录API调用
  logApiCall(method: string, url: string, request?: any, response?: any, statusCode?: number): void {
    if (!this.isEnabled) return;

    const startTime = this.apiCallTimes.get(`${method}:${url}`);
    const duration = startTime ? Date.now() - startTime : 0;

    const log: ApiCallLog = {
      timestamp: Date.now(),
      type: 'api',
      category: 'API Calls',
      message: `${method} ${url}`,
      method,
      url,
      request,
      response,
      statusCode,
      duration
    };

    this.addLog(log);
    this.logToConsole(log);
  }

  // 记录API调用开始
  logApiCallStart(method: string, url: string, request?: any): void {
    if (!this.isEnabled) return;

    this.apiCallTimes.set(`${method}:${url}`, Date.now());

    const log: DevToolsLog = {
      timestamp: Date.now(),
      type: 'api',
      category: 'API Calls',
      message: `Starting ${method} ${url}`,
      data: { request }
    };

    this.addLog(log);
  }

  // 记录数据验证
  logDataValidation(dataType: string, data: any, requiredFields: string[] = []): void {
    if (!this.isEnabled) return;

    const validation = this.validateData(data, requiredFields);

    const log: DataValidationLog = {
      timestamp: Date.now(),
      type: 'data',
      category: 'Data Validation',
      message: `${dataType} validation`,
      dataType,
      validation
    };

    this.addLog(log);
    this.logToConsole(log);
  }

  // 记录渲染信息
  logRender(component: string, props?: any, state?: any): void {
    if (!this.isEnabled) return;

    const log: DevToolsLog = {
      timestamp: Date.now(),
      type: 'render',
      category: 'Rendering',
      message: `${component} rendered`,
      data: { props, state }
    };

    this.addLog(log);
  }

  // 记录性能指标
  logPerformance(metric: string, value: number, unit: string = 'ms'): void {
    if (!this.isEnabled) return;

    const log: PerformanceLog = {
      timestamp: Date.now(),
      type: 'performance',
      category: 'Performance',
      message: `${metric}: ${value}${unit}`,
      metric,
      value,
      unit
    };

    this.addLog(log);
    this.logToConsole(log);
  }

  // 记录错误
  logError(category: string, error: Error | string, context?: any): void {
    if (!this.isEnabled) return;

    const errorMessage = error instanceof Error ? error.message : error;
    const stack = error instanceof Error ? error.stack : undefined;

    const log: DevToolsLog = {
      timestamp: Date.now(),
      type: 'error',
      category,
      message: errorMessage,
      data: { stack, context }
    };

    this.addLog(log);
    this.logToConsole(log);
  }

  // 验证数据
  private validateData(data: any, requiredFields: string[]) {
    const isEmpty = !data || (Array.isArray(data) ? data.length === 0 : Object.keys(data).length === 0);
    const itemCount = Array.isArray(data) ? data.length : (data ? Object.keys(data).length : 0);

    let hasRequiredFields = true;
    let missingFields: string[] = [];
    let sampleItem: any = null;

    if (Array.isArray(data) && data.length > 0) {
      sampleItem = data[0];
      requiredFields.forEach(field => {
        if (!(field in sampleItem)) {
          hasRequiredFields = false;
          missingFields.push(field);
        }
      });
    }

    return {
      isEmpty,
      itemCount,
      hasRequiredFields,
      missingFields: missingFields.length > 0 ? missingFields : undefined,
      sampleItem
    };
  }

  // 验证当前页面数据
  validateCurrentData(): void {
    if (!this.isEnabled) return;

    console.group('%c📊 Data Validation Report', 'color: #2196F3; font-size: 14px; font-weight: bold;');

    // 验证案件数据
    const casesElement = document.querySelector('[data-cases-container]');
    if (casesElement) {
      const casesData = (casesElement as any).__casesData;
      if (casesData) {
        this.logDataValidation('Cases', casesData, ['id', 'title', 'client_name', 'lawyer_name', 'status', 'priority']);
      }
    }

    // 验证客户数据
    const clientsElement = document.querySelector('[data-clients-container]');
    if (clientsElement) {
      const clientsData = (clientsElement as any).__clientsData;
      if (clientsData) {
        this.logDataValidation('Clients', clientsData, ['id', 'name', 'email', 'phone']);
      }
    }

    // 验证律师数据
    const lawyersElement = document.querySelector('[data-lawyers-container]');
    if (lawyersElement) {
      const lawyersData = (lawyersElement as any).__lawyersData;
      if (lawyersData) {
        this.logDataValidation('Lawyers', lawyersData, ['id', 'name', 'email', 'role']);
      }
    }

    console.groupEnd();
  }

  // 获取性能报告
  getPerformanceReport(): void {
    if (!this.isEnabled) return;

    const performanceLogs = this.logs.filter(log => log.type === 'performance');
    const apiLogs = this.logs.filter(log => log.type === 'api') as ApiCallLog[];

    console.group('%c⚡ Performance Report', 'color: #FF9800; font-size: 14px; font-weight: bold;');

    if (performanceLogs.length > 0) {
      console.table(performanceLogs.map(log => ({
        Metric: (log as PerformanceLog).metric,
        Value: (log as PerformanceLog).value,
        Unit: (log as PerformanceLog).unit,
        Time: new Date(log.timestamp).toLocaleTimeString()
      })));
    }

    if (apiLogs.length > 0) {
      const avgApiTime = apiLogs.reduce((sum, log) => sum + log.duration, 0) / apiLogs.length;
      const slowestApi = apiLogs.reduce((slowest, current) => current.duration > slowest.duration ? current : slowest);

      console.log('📈 API Performance Summary:');
      console.log(`   Average response time: ${avgApiTime.toFixed(2)}ms`);
      console.log(`   Slowest API: ${slowestApi.method} ${slowestApi.url} (${slowestApi.duration}ms)`);
      console.log(`   Total API calls: ${apiLogs.length}`);
    }

    console.groupEnd();
  }

  // 添加日志
  private addLog(log: DevToolsLog): void {
    this.logs.push(log);

    // 保持日志数量在限制内
    if (this.logs.length > this.maxLogs) {
      this.logs = this.logs.slice(-this.maxLogs);
    }
  }

  // 输出到控制台
  private logToConsole(log: DevToolsLog): void {
    const time = new Date(log.timestamp).toLocaleTimeString();
    const prefix = `[${time}] ${log.category}:`;

    switch (log.type) {
      case 'api':
        const apiLog = log as ApiCallLog;
        const statusColor = apiLog.statusCode && apiLog.statusCode >= 200 && apiLog.statusCode < 300 ? '#4CAF50' : '#F44336';
        console.log(
          `%c${prefix} ${apiLog.method} ${apiLog.url} (${apiLog.duration}ms)`,
          `color: ${statusColor}; font-weight: bold;`,
          apiLog.response
        );
        break;

      case 'data':
        const dataLog = log as DataValidationLog;
        const validationColor = dataLog.validation.hasRequiredFields && !dataLog.validation.isEmpty ? '#4CAF50' : '#FF9800';
        console.log(
          `%c${prefix} ${dataLog.dataType} - ${dataLog.validation.itemCount} items`,
          `color: ${validationColor}; font-weight: bold;`,
          dataLog.validation
        );
        break;

      case 'error':
        console.error(`${prefix} ${log.message}`, log.data);
        break;

      case 'performance':
        const perfLog = log as PerformanceLog;
        console.log(
          `%c${prefix} ${perfLog.metric}: ${perfLog.value}${perfLog.unit}`,
          'color: #9C27B0; font-weight: bold;'
        );
        break;

      default:
        console.log(`${prefix} ${log.message}`, log.data);
    }
  }

  // 获取所有日志
  getLogs(type?: string): DevToolsLog[] {
    if (type) {
      return this.logs.filter(log => log.type === type);
    }
    return [...this.logs];
  }

  // 清除日志
  clearLogs(): void {
    this.logs = [];
    console.clear();
    console.log('%c🔧 DevTools logs cleared', 'color: #4CAF50; font-size: 12px;');
  }

  // 导出日志
  exportLogs(): string {
    const exportData = {
      timestamp: new Date().toISOString(),
      logs: this.logs,
      summary: {
        total: this.logs.length,
        byType: {
          api: this.logs.filter(log => log.type === 'api').length,
          data: this.logs.filter(log => log.type === 'data').length,
          render: this.logs.filter(log => log.type === 'render').length,
          error: this.logs.filter(log => log.type === 'error').length,
          performance: this.logs.filter(log => log.type === 'performance').length
        }
      }
    };

    return JSON.stringify(exportData, null, 2);
  }
}

// 创建单例实例
export const devToolsValidator = new DevToolsValidator();

// 导出便捷方法
export const logApiCall = (method: string, url: string, request?: any, response?: any, statusCode?: number) =>
  devToolsValidator.logApiCall(method, url, request, response, statusCode);

export const logApiCallStart = (method: string, url: string, request?: any) =>
  devToolsValidator.logApiCallStart(method, url, request);

export const logDataValidation = (dataType: string, data: any, requiredFields?: string[]) =>
  devToolsValidator.logDataValidation(dataType, data, requiredFields);

export const logRender = (component: string, props?: any, state?: any) =>
  devToolsValidator.logRender(component, props, state);

export const logPerformance = (metric: string, value: number, unit?: string) =>
  devToolsValidator.logPerformance(metric, value, unit);

export const logError = (category: string, error: Error | string, context?: any) =>
  devToolsValidator.logError(category, error, context);

export default devToolsValidator;
/**
 * 后端服务集成测试脚本
 * 专注于API联调验证，测试后端各服务的集成情况
 */

const { default: fetch } = require('node-fetch');

const CONFIG = {
  backend: {
    url: 'http://localhost:8080',
    timeout: 10000
  }
};

// 测试结果记录
const testResults = {
  startTime: new Date(),
  tests: [],
  summary: {
    total: 0,
    passed: 0,
    failed: 0
  }
};

function log(message, type = 'info') {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] [${type.toUpperCase()}] ${message}`);
}

function recordTest(name, passed, details = '') {
  testResults.summary.total++;
  if (passed) {
    testResults.summary.passed++;
    log(`✅ ${name} - 通过`);
  } else {
    testResults.summary.failed++;
    log(`❌ ${name} - 失败: ${details}`);
  }

  testResults.tests.push({
    name,
    passed,
    details,
    timestamp: new Date()
  });
}

async function makeRequest(url, options = {}) {
  try {
    const response = await fetch(url, {
      timeout: CONFIG.backend.timeout,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      },
      ...options
    });

    const data = await response.json();

    return {
      status: response.status,
      ok: response.ok,
      data,
      headers: response.headers
    };
  } catch (error) {
    log(`请求失败: ${url} - ${error.message}`, 'error');
    throw error;
  }
}

async function testHealthEndpoints() {
  log('=== 测试健康检查端点 ===');

  const endpoints = [
    { name: '基础健康检查', url: '/health' },
    { name: '详细健康检查', url: '/api/v1/health/detailed' },
    { name: '存活检查', url: '/health/live' },
    { name: '就绪检查', url: '/health/ready' },
    { name: '健康指标', url: '/api/v1/health/metrics' },
    { name: '依赖健康检查', url: '/api/v1/health/dependencies' }
  ];

  for (const endpoint of endpoints) {
    try {
      const response = await makeRequest(`${CONFIG.backend.url}${endpoint.url}`);
      const passed = response.ok || response.status === 200;
      recordTest(endpoint.name, passed, passed ? `状态码: ${response.status}` : `状态码: ${response.status}`);
    } catch (error) {
      recordTest(endpoint.name, false, error.message);
    }
  }
}

async function testDatabaseConnectivity() {
  log('=== 测试数据库连接 ===');

  try {
    // 通过健康检查获取数据库状态
    const health = await makeRequest(`${CONFIG.backend.url}/api/v1/health/detailed`);

    if (health.data && health.data.checks) {
      const checks = health.data.checks;

      // 检查数据库连接
      const dbStatus = checks.database ? checks.database.status : 'unknown';
      recordTest('数据库连接状态', dbStatus === 'healthy' || dbStatus === 'ok', `状态: ${dbStatus}`);

      // 检查Redis连接
      const cacheStatus = checks.cache ? checks.cache.status : 'unknown';
      recordTest('Redis缓存连接', cacheStatus === 'healthy' || cacheStatus === 'ok', `状态: ${cacheStatus}`);

      // 检查存储状态
      const storageStatus = checks.storage ? checks.storage.status : 'unknown';
      recordTest('存储系统状态', storageStatus === 'healthy' || storageStatus === 'ok', `状态: ${storageStatus}`);

    } else {
      recordTest('健康检查数据格式', false, '无法解析健康检查数据');
    }
  } catch (error) {
    recordTest('数据库连接测试', false, error.message);
  }
}

async function testAPIRouting() {
  log('=== 测试API路由配置 ===');

  // 测试主要API路由是否存在（期望401或404，而不是500）
  const routes = [
    { name: '认证路由-登录', url: '/api/auth/login', method: 'POST' },
    { name: '认证路由-注册', url: '/api/auth/register', method: 'POST' },
    { name: '用户管理路由', url: '/api/users', method: 'GET' },
    { name: '律师管理路由', url: '/api/lawyers', method: 'GET' },
    { name: '客户管理路由', url: '/api/clients', method: 'GET' },
    { name: '案件管理路由', url: '/api/cases', method: 'GET' },
    { name: '仪表板路由', url: '/api/dashboard/statistics', method: 'GET' },
    { name: '文件管理路由', url: '/api/files', method: 'GET' },
    { name: '审批管理路由', url: '/api/approvals', method: 'GET' },
    { name: '冲突检查路由', url: '/api/conflict/check', method: 'POST' },
    { name: 'RBAC权限路由', url: '/api/roles', method: 'GET' },
    { name: '搜索路由', url: '/api/v1/search', method: 'GET' }
  ];

  for (const route of routes) {
    try {
      const options = route.method === 'POST' ? { method: 'POST', body: '{}' } : {};
      const response = await makeRequest(`${CONFIG.backend.url}${route.url}`, options);

      // 401表示路由存在但需要认证，404表示路由不存在，200表示公开路由
      const routeExists = response.status !== 404;
      const expectedAuth = response.status === 401;

      if (routeExists) {
        recordTest(route.name, true, `路由存在, 状态码: ${response.status}${expectedAuth ? ' (需要认证)' : ''}`);
      } else {
        recordTest(route.name, false, `路由不存在, 状态码: ${response.status}`);
      }
    } catch (error) {
      recordTest(route.name, false, `路由测试失败: ${error.message}`);
    }
  }
}

async function testMiddlewareStack() {
  log('=== 测试中间件堆栈 ===');

  try {
    // 测试CORS中间件
    const response = await makeRequest(`${CONFIG.backend.url}/health`, {
      headers: { 'Origin': 'http://localhost:3000' }
    });

    const corsHeaders = response.headers.get('access-control-allow-origin');
    recordTest('CORS中间件', !!corsHeaders, corsHeaders ? `允许来源: ${corsHeaders}` : '未检测到CORS头');

    // 测试安全头中间件
    const securityHeaders = {
      'x-content-type-options': response.headers.get('x-content-type-options'),
      'x-frame-options': response.headers.get('x-frame-options'),
      'x-xss-protection': response.headers.get('x-xss-protection')
    };

    const securityHeadersCount = Object.values(securityHeaders).filter(h => h).length;
    recordTest('安全头中间件', securityHeadersCount >= 2, `检测到 ${securityHeadersCount} 个安全头`);

    // 测试请求ID中间件
    const requestId = response.data?.request_id || response.headers.get('x-request-id');
    recordTest('请求ID中间件', !!requestId, requestId ? `请求ID: ${requestId}` : '未检测到请求ID');

  } catch (error) {
    recordTest('中间件测试', false, error.message);
  }
}

async function testErrorHandling() {
  log('=== 测试错误处理机制 ===');

  try {
    // 测试404错误处理
    const notFoundResponse = await makeRequest(`${CONFIG.backend.url}/api/nonexistent-endpoint`);
    recordTest('404错误处理', notFoundResponse.status === 404,
      notFoundResponse.data?.error ? `错误信息: ${notFoundResponse.data.error.message}` : '无错误信息');

    // 测试无效JSON处理
    try {
      const invalidJsonResponse = await fetch(`${CONFIG.backend.url}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: 'invalid json'
      });

      const invalidJsonStatus = invalidJsonResponse.status;
      recordTest('无效JSON处理', invalidJsonStatus === 400, `状态码: ${invalidJsonStatus}`);
    } catch (error) {
      recordTest('无效JSON处理', false, error.message);
    }

  } catch (error) {
    recordTest('错误处理测试', false, error.message);
  }
}

async function testPerformanceMonitoring() {
  log('=== 测试性能监控 ===');

  try {
    // 测试指标端点
    const metricsResponse = await makeRequest(`${CONFIG.backend.url}/metrics`);
    recordTest('Prometheus指标', metricsResponse.status === 200, '指标端点可访问');

    // 测试性能监控端点
    const perfCacheResponse = await makeRequest(`${CONFIG.backend.url}/performance/cache`);
    recordTest('缓存性能监控', perfCacheResponse.status === 200, '缓存监控端点可访问');

    const perfDBResponse = await makeRequest(`${CONFIG.backend.url}/performance/database`);
    recordTest('数据库性能监控', perfDBResponse.status === 200, '数据库监控端点可访问');

  } catch (error) {
    recordTest('性能监控测试', false, error.message);
  }
}

async function testServiceDependencies() {
  log('=== 测试服务依赖 ===');

  try {
    // 获取详细健康检查信息
    const healthResponse = await makeRequest(`${CONFIG.backend.url}/api/v1/health/detailed`);

    if (healthResponse.data && healthResponse.data.checks) {
      const checks = healthResponse.data.checks;

      // 检查各个服务组件
      const services = [
        { name: '数据库服务', key: 'database' },
        { name: '缓存服务', key: 'cache' },
        { name: '存储服务', key: 'storage' },
        { name: '搜索引擎', key: 'elasticsearch' }
      ];

      for (const service of services) {
        const status = checks[service.key]?.status;
        const healthy = status === 'healthy' || status === 'ok';

        if (status) {
          recordTest(`${service.name}依赖`, healthy, `状态: ${status}`);
        } else {
          recordTest(`${service.name}依赖`, false, '服务未在健康检查中注册');
        }
      }

    } else {
      recordTest('服务依赖检查', false, '无法获取健康检查数据');
    }
  } catch (error) {
    recordTest('服务依赖测试', false, error.message);
  }
}

function generateReport() {
  const endTime = new Date();
  const duration = endTime - testResults.startTime;

  const report = {
    summary: {
      startTime: testResults.startTime.toISOString(),
      endTime: endTime.toISOString(),
      duration: `${Math.round(duration / 1000)}s`,
      totalTests: testResults.summary.total,
      passedTests: testResults.summary.passed,
      failedTests: testResults.summary.failed,
      successRate: `${Math.round((testResults.summary.passed / testResults.summary.total) * 100)}%`
    },
    tests: testResults.tests
  };

  // 保存报告
  const reportPath = `backend-integration-test-report-${Date.now()}.json`;
  require('fs').writeFileSync(reportPath, JSON.stringify(report, null, 2));

  log(`\n=== 后端集成测试报告 ===`);
  log(`测试时长: ${report.summary.duration}`);
  log(`总测试数: ${report.summary.totalTests}`);
  log(`通过测试: ${report.summary.passedTests}`);
  log(`失败测试: ${report.summary.failedTests}`);
  log(`成功率: ${report.summary.successRate}`);
  log(`详细报告: ${reportPath}`);

  return report;
}

async function runIntegrationTests() {
  log('开始后端服务集成测试...');
  log(`测试目标: ${CONFIG.backend.url}`);

  try {
    await testHealthEndpoints();
    await testDatabaseConnectivity();
    await testAPIRouting();
    await testMiddlewareStack();
    await testErrorHandling();
    await testPerformanceMonitoring();
    await testServiceDependencies();

    const report = generateReport();

    log('后端服务集成测试完成！');
    return report;

  } catch (error) {
    log('集成测试执行失败:', error, 'error');
    throw error;
  }
}

// 运行测试
if (require.main === module) {
  runIntegrationTests()
    .then(() => {
      console.log('测试完成');
      process.exit(0);
    })
    .catch((error) => {
      console.error('测试失败:', error);
      process.exit(1);
    });
}

module.exports = { runIntegrationTests };
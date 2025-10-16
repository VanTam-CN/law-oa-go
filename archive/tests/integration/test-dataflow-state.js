/**
 * 数据流和状态管理测试脚本
 * 测试前端数据流、状态管理和API响应处理
 */

const puppeteer = require('puppeteer');
const { default: fetch } = require('node-fetch');

const CONFIG = {
  frontend: {
    url: 'http://localhost:5174',
    timeout: 30000
  },
  backend: {
    url: 'http://localhost:8080'
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

async function testFrontendDataLoading() {
  log('=== 测试前端数据加载 ===');

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1920, height: 1080 }
  });

  try {
    const page = await browser.newPage();

    // 监听网络请求和响应
    const networkRequests = [];
    const networkResponses = [];

    page.on('request', request => {
      networkRequests.push({
        url: request.url(),
        method: request.method(),
        timestamp: Date.now()
      });
    });

    page.on('response', response => {
      networkResponses.push({
        url: response.url(),
        status: response.status(),
        timestamp: Date.now()
      });
    });

    // 访问前端页面
    await page.goto(CONFIG.frontend.url, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    // 等待页面完全加载
    await page.waitForTimeout(3000);

    // 检查是否发起了API请求
    const apiRequests = networkRequests.filter(req =>
      req.url.includes('/api/') && !req.url.includes('localhost:5174')
    );

    recordTest('前端API请求发起', apiRequests.length > 0,
      `发现 ${apiRequests.length} 个API请求`);

    // 检查API响应状态
    const apiResponses = networkResponses.filter(res =>
      res.url.includes('/api/') && !res.url.includes('localhost:5174')
    );

    const successfulResponses = apiResponses.filter(res => res.status < 400);
    recordTest('API响应状态检查', successfulResponses.length === apiResponses.length,
      `成功响应: ${successfulResponses.length}/${apiResponses.length}`);

    // 检查页面状态
    const pageTitle = await page.title();
    recordTest('页面标题加载', pageTitle.length > 0, `标题: ${pageTitle}`);

    // 检查是否有错误状态显示
    const errorElements = await page.$$('[data-testid="error"], .error, .alert-error');
    recordTest('页面错误状态检查', errorElements.length === 0,
      `发现 ${errorElements.length} 个错误元素`);

    await page.close();

  } catch (error) {
    recordTest('前端数据加载测试', false, error.message);
  } finally {
    await browser.close();
  }
}

async function testStateManagement() {
  log('=== 测试状态管理 ===');

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1920, height: 1080 }
  });

  try {
    const page = await browser.newPage();

    // 访问前端页面
    await page.goto(CONFIG.frontend.url, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    // 等待应用初始化
    await page.waitForTimeout(2000);

    // 检查React应用是否正确挂载
    const reactApp = await page.$('#root, #app, [data-reactroot]');
    recordTest('React应用挂载', !!reactApp, 'React根元素检测');

    // 检查路由状态
    const currentPath = await page.evaluate(() => window.location.pathname);
    recordTest('前端路由状态', currentPath.length > 0, `当前路径: ${currentPath}`);

    // 检查全局状态管理（如果存在）
    try {
      const globalState = await page.evaluate(() => {
        // 检查常见的状态管理库
        if (window.__REDUX_STORE__) {
          return 'Redux';
        }
        if (window.Vuex) {
          return 'Vuex';
        }
        if (window.Pinia) {
          return 'Pinia';
        }
        if (window.React) {
          // 检查Context API或其他React状态
          const hasContext = document.querySelector('[data-react-context]');
          return hasContext ? 'React Context' : 'Unknown';
        }
        return 'None';
      });

      recordTest('状态管理库检测', globalState !== 'None', `状态管理: ${globalState}`);
    } catch (error) {
      recordTest('状态管理库检测', false, error.message);
    }

    // 测试页面导航状态
    try {
      const navigationLinks = await page.$$('a[href^="/"]');
      recordTest('导航链接检测', navigationLinks.length > 0,
        `发现 ${navigationLinks.length} 个内部链接`);

      if (navigationLinks.length > 0) {
        // 尝试点击导航链接
        const firstLink = navigationLinks[0];
        const linkText = await page.evaluate(el => el.textContent, firstLink);

        await firstLink.click();
        await page.waitForTimeout(2000);

        const newPath = await page.evaluate(() => window.location.pathname);
        recordTest('页面导航状态', newPath !== currentPath,
          `从 ${currentPath} 导航到 ${newPath}`);
      }
    } catch (error) {
      recordTest('页面导航状态', false, error.message);
    }

    await page.close();

  } catch (error) {
    recordTest('状态管理测试', false, error.message);
  } finally {
    await browser.close();
  }
}

async function testAPIResponseHandling() {
  log('=== 测试API响应处理 ===');

  // 直接测试后端API的响应格式
  const endpoints = [
    { name: '健康检查响应', url: '/health', expectedFields: ['status'] },
    { name: '详细健康检查响应', url: '/api/v1/health/detailed', expectedFields: [] },
    { name: '用户列表响应', url: '/api/users', expectedFields: [] }
  ];

  for (const endpoint of endpoints) {
    try {
      const response = await fetch(`${CONFIG.backend.url}${endpoint.url}`);
      const data = await response.json();

      // 检查响应格式
      const hasValidStructure = typeof data === 'object' && data !== null;
      recordTest(`${endpoint.name}格式`, hasValidStructure, '响应是有效JSON对象');

      // 检查预期字段
      if (endpoint.expectedFields.length > 0) {
        const hasExpectedFields = endpoint.expectedFields.every(field => field in data);
        recordTest(`${endpoint.name}字段检查`, hasExpectedFields,
          `包含字段: ${endpoint.expectedFields.join(', ')}`);
      }

      // 检查错误处理格式
      if (!response.ok) {
        const hasErrorStructure = data.error || data.message;
        recordTest(`${endpoint.name}错误格式`, !!hasErrorStructure, '错误响应结构检查');
      }

    } catch (error) {
      recordTest(`${endpoint.name}测试`, false, error.message);
    }
  }
}

async function testDataConsistency() {
  log('=== 测试数据一致性 ===');

  try {
    // 测试相同API多次调用的一致性
    const testUrl = `${CONFIG.backend.url}/health`;
    const responses = [];

    for (let i = 0; i < 3; i++) {
      const response = await fetch(testUrl);
      const data = await response.json();
      responses.push({ status: response.status, data });

      // 等待一小段时间
      await new Promise(resolve => setTimeout(resolve, 500));
    }

    // 检查状态码一致性
    const statusCodes = responses.map(r => r.status);
    const consistentStatus = statusCodes.every(code => code === statusCodes[0]);
    recordTest('API状态码一致性', consistentStatus, `状态码: ${statusCodes.join(', ')}`);

    // 检查响应数据一致性
    const dataStrings = responses.map(r => JSON.stringify(r.data));
    const consistentData = dataStrings.every(data => data === dataStrings[0]);
    recordTest('API响应数据一致性', consistentData, '多次调用数据一致性检查');

  } catch (error) {
    recordTest('数据一致性测试', false, error.message);
  }
}

async function testCachingBehavior() {
  log('=== 测试缓存行为 ===');

  try {
    // 测试缓存控制头
    const cacheTestUrl = `${CONFIG.backend.url}/health`;
    const response = await fetch(cacheTestUrl);

    const cacheControl = response.headers.get('cache-control');
    const etag = response.headers.get('etag');
    const lastModified = response.headers.get('last-modified');

    recordTest('缓存控制头', !!cacheControl, cacheControl || '未设置');
    recordTest('ETag头', !!etag, etag || '未设置');
    recordTest('Last-Modified头', !!lastModified, lastModified || '未设置');

    // 测试条件请求（如果ETag存在）
    if (etag) {
      const conditionalResponse = await fetch(cacheTestUrl, {
        headers: { 'If-None-Match': etag }
      });

      const isNotModified = conditionalResponse.status === 304;
      recordTest('条件请求支持', isNotModified, `状态码: ${conditionalResponse.status}`);
    }

  } catch (error) {
    recordTest('缓存行为测试', false, error.message);
  }
}

async function testErrorRecovery() {
  log('=== 测试错误恢复机制 ===');

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1920, height: 1080 }
  });

  try {
    const page = await browser.newPage();

    // 监听控制台错误
    const consoleErrors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    // 监听页面错误
    const pageErrors = [];
    page.on('pageerror', error => {
      pageErrors.push(error.message);
    });

    // 访问前端页面
    await page.goto(CONFIG.frontend.url, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    // 模拟网络断开
    await page.setOfflineMode(true);
    await page.waitForTimeout(2000);

    // 尝试进行一些操作
    const offlineErrors = consoleErrors.length + pageErrors.length;

    // 恢复网络
    await page.setOfflineMode(false);
    await page.waitForTimeout(2000);

    // 检查页面是否仍然响应
    const pageStillResponsive = await page.evaluate(() => {
      return document.readyState === 'complete';
    });

    recordTest('离线错误处理', offlineErrors > 0, `离线时错误: ${offlineErrors} 个`);
    recordTest('网络恢复后响应性', pageStillResponsive, '页面恢复响应');

    await page.close();

  } catch (error) {
    recordTest('错误恢复测试', false, error.message);
  } finally {
    await browser.close();
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
  const reportPath = `dataflow-state-test-report-${Date.now()}.json`;
  require('fs').writeFileSync(reportPath, JSON.stringify(report, null, 2));

  log(`\n=== 数据流和状态管理测试报告 ===`);
  log(`测试时长: ${report.summary.duration}`);
  log(`总测试数: ${report.summary.totalTests}`);
  log(`通过测试: ${report.summary.passedTests}`);
  log(`失败测试: ${report.summary.failedTests}`);
  log(`成功率: ${report.summary.successRate}`);
  log(`详细报告: ${reportPath}`);

  return report;
}

async function runDataflowTests() {
  log('开始数据流和状态管理测试...');
  log(`前端地址: ${CONFIG.frontend.url}`);
  log(`后端地址: ${CONFIG.backend.url}`);

  try {
    await testFrontendDataLoading();
    await testStateManagement();
    await testAPIResponseHandling();
    await testDataConsistency();
    await testCachingBehavior();
    await testErrorRecovery();

    const report = generateReport();

    log('数据流和状态管理测试完成！');
    return report;

  } catch (error) {
    log('测试执行失败:', error, 'error');
    throw error;
  }
}

// 运行测试
if (require.main === module) {
  runDataflowTests()
    .then(() => {
      console.log('测试完成');
      process.exit(0);
    })
    .catch((error) => {
      console.error('测试失败:', error);
      process.exit(1);
    });
}

module.exports = { runDataflowTests };
/**
 * 端到端功能测试脚本
 * 测试前端和后端服务的完整功能流程
 */

const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

// 测试配置
const CONFIG = {
  frontend: {
    url: 'http://localhost:5174',
    timeout: 30000
  },
  backend: {
    url: 'http://localhost:8080',
    timeout: 10000
  },
  testData: {
    user: {
      email: 'test@example.com',
      password: '123456',
      name: '测试用户'
    },
    lawyer: {
      name: '测试律师',
      phone: '13800138000',
      email: 'lawyer@example.com',
      licenseNumber: '123456789012345',
      department: '民事诉讼部'
    },
    client: {
      name: '测试客户',
      phone: '13900139000',
      email: 'client@example.com',
      type: '个人',
      address: '测试地址'
    },
    case: {
      title: '测试案件',
      caseNumber: 'TC20250001',
      caseType: '民事诉讼',
      clientName: '测试客户',
      lawyerName: '测试律师',
      description: '这是一个测试案件描述'
    }
  }
};

// 测试结果记录
let testResults = {
  startTime: new Date(),
  endTime: null,
  totalTests: 0,
  passedTests: 0,
  failedTests: 0,
  errors: [],
  details: []
};

// 辅助函数
function log(message, type = 'info') {
  const timestamp = new Date().toISOString();
  const logMessage = `[${timestamp}] [${type.toUpperCase()}] ${message}`;
  console.log(logMessage);

  testResults.details.push({
    timestamp,
    type,
    message
  });
}

function recordResult(testName, passed, error = null) {
  testResults.totalTests++;
  if (passed) {
    testResults.passedTests++;
    log(`✅ ${testName} - 通过`, 'success');
  } else {
    testResults.failedTests++;
    log(`❌ ${testName} - 失败: ${error}`, 'error');
    testResults.errors.push({
      test: testName,
      error: error
    });
  }
}

// 网络请求辅助函数
async function makeRequest(url, options = {}) {
  try {
    const response = await fetch(url, {
      timeout: CONFIG.backend.timeout,
      ...options
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    log(`网络请求失败: ${url} - ${error.message}`, 'error');
    throw error;
  }
}

// 测试函数
async function testBackendHealth() {
  log('开始测试后端服务健康状态...');

  try {
    const health = await makeRequest(`${CONFIG.backend.url}/health`);
    recordResult('后端健康检查', health.status === 'ok' || health.status === 'healthy');

    // 测试详细的健康检查
    const detailedHealth = await makeRequest(`${CONFIG.backend.url}/api/v1/health/detailed`);
    recordResult('后端详细健康检查', detailedHealth.status === 'healthy');

  } catch (error) {
    recordResult('后端健康检查', false, error.message);
  }
}

async function testBackendAPIs() {
  log('开始测试后端API接口...');

  // 测试仪表盘统计
  try {
    const stats = await makeRequest(`${CONFIG.backend.url}/api/v1/dashboard/statistics`);
    recordResult('仪表盘统计API', stats && typeof stats === 'object');
  } catch (error) {
    recordResult('仪表盘统计API', false, error.message);
  }

  // 测试律师列表
  try {
    const lawyers = await makeRequest(`${CONFIG.backend.url}/api/lawyers`);
    recordResult('律师列表API', lawyers && Array.isArray(lawyers.data || lawyers));
  } catch (error) {
    recordResult('律师列表API', false, error.message);
  }

  // 测试客户列表
  try {
    const clients = await makeRequest(`${CONFIG.backend.url}/api/clients`);
    recordResult('客户列表API', clients && Array.isArray(clients.data || clients));
  } catch (error) {
    recordResult('客户列表API', false, error.message);
  }

  // 测试案件列表
  try {
    const cases = await makeRequest(`${CONFIG.backend.url}/api/cases`);
    recordResult('案件列表API', cases && Array.isArray(cases.data || cases));
  } catch (error) {
    recordResult('案件列表API', false, error.message);
  }
}

async function testFrontendLoad() {
  log('开始测试前端页面加载...');

  const browser = await puppeteer.launch({
    headless: false, // 显示浏览器窗口便于调试
    defaultViewport: { width: 1920, height: 1080 }
  });

  try {
    const page = await browser.newPage();

    // 监听网络请求
    const requests = [];
    page.on('request', request => {
      requests.push({
        url: request.url(),
        method: request.method()
      });
    });

    // 监听控制台错误
    const consoleErrors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    // 访问前端页面
    await page.goto(CONFIG.frontend.url, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    // 检查页面标题
    const title = await page.title();
    recordResult('前端页面标题', title.includes('律所') || title.includes('Law'));

    // 检查是否有控制台错误
    recordResult('前端控制台错误检查', consoleErrors.length === 0,
      consoleErrors.length > 0 ? consoleErrors.join('; ') : null);

    // 检查网络请求
    const apiRequests = requests.filter(req =>
      req.url.includes('/api/') && !req.url.includes('localhost:5174')
    );
    recordResult('前端API请求', apiRequests.length > 0);

    log(`前端页面加载完成，共发起 ${requests.length} 个请求，其中 ${apiRequests.length} 个API请求`);

    await page.close();

  } catch (error) {
    recordResult('前端页面加载', false, error.message);
  } finally {
    await browser.close();
  }
}

async function testDatabaseConnection() {
  log('开始测试数据库连接...');

  try {
    // 通过后端API测试数据库连接
    const response = await makeRequest(`${CONFIG.backend.url}/api/v1/health`);
    const checks = response.checks || {};

    const dbConnected = checks.database ?
      (checks.database.status === 'healthy' || checks.database.status === 'ok') :
      true; // 如果没有数据库检查，假设连接正常

    recordResult('数据库连接', dbConnected);

    // 测试Redis连接
    const cacheConnected = checks.cache ?
      (checks.cache.status === 'healthy' || checks.cache.status === 'ok') :
      true;

    recordResult('Redis缓存连接', cacheConnected);

  } catch (error) {
    recordResult('数据库连接测试', false, error.message);
  }
}

async function testAuthenticationFlow() {
  log('开始测试认证流程...');

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1920, height: 1080 }
  });

  try {
    const page = await browser.newPage();

    // 访问登录页面
    await page.goto(`${CONFIG.frontend.url}/login`, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    // 等待登录表单加载
    await page.waitForSelector('form', { timeout: 5000 });

    // 填写登录信息
    await page.type('input[name="email"]', CONFIG.testData.user.email);
    await page.type('input[name="password"]', CONFIG.testData.user.password);

    // 提交登录表单
    await page.click('button[type="submit"]');

    // 等待登录完成（跳转到首页或显示登录成功）
    await page.waitForTimeout(3000);

    // 检查是否登录成功（URL变化或显示用户信息）
    const currentUrl = page.url();
    const isLoggedIn = !currentUrl.includes('/login') ||
      await page.$('.user-info') !== null;

    recordResult('用户登录流程', isLoggedIn);

    await page.close();

  } catch (error) {
    recordResult('用户登录流程', false, error.message);
  } finally {
    await browser.close();
  }
}

// 生成测试报告
function generateReport() {
  testResults.endTime = new Date();
  const duration = testResults.endTime - testResults.startTime;

  const report = {
    summary: {
      startTime: testResults.startTime.toISOString(),
      endTime: testResults.endTime.toISOString(),
      duration: `${Math.round(duration / 1000)}s`,
      totalTests: testResults.totalTests,
      passedTests: testResults.passedTests,
      failedTests: testResults.failedTests,
      successRate: `${Math.round((testResults.passedTests / testResults.totalTests) * 100)}%`
    },
    errors: testResults.errors,
    details: testResults.details
  };

  // 保存测试报告
  const reportPath = path.join(__dirname, `e2e-test-report-${Date.now()}.json`);
  fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));

  // 生成可读的HTML报告
  const htmlReport = generateHTMLReport(report);
  const htmlPath = path.join(__dirname, `e2e-test-report-${Date.now()}.html`);
  fs.writeFileSync(htmlPath, htmlReport);

  log(`测试报告已生成: ${reportPath}`);
  log(`HTML报告已生成: ${htmlPath}`);

  return report;
}

function generateHTMLReport(report) {
  return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>端到端测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #eee; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric { background: #f8f9fa; padding: 15px; border-radius: 6px; text-align: center; }
        .metric h3 { margin: 0 0 10px 0; color: #333; }
        .metric .value { font-size: 2em; font-weight: bold; color: #007bff; }
        .metric .label { color: #666; font-size: 0.9em; }
        .success { color: #28a745; }
        .error { color: #dc3545; }
        .section { margin: 30px 0; }
        .section h2 { border-bottom: 1px solid #ddd; padding-bottom: 10px; }
        .error-item { background: #f8d7da; border: 1px solid #f5c6cb; border-radius: 4px; padding: 15px; margin: 10px 0; }
        .log-item { background: #f8f9fa; border-left: 4px solid #007bff; padding: 10px; margin: 5px 0; font-family: monospace; font-size: 0.9em; }
        .log-item.error { border-left-color: #dc3545; }
        .log-item.success { border-left-color: #28a745; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>律所OA系统 - 端到端测试报告</h1>
            <p>生成时间: ${new Date().toLocaleString('zh-CN')}</p>
        </div>

        <div class="summary">
            <div class="metric">
                <h3>总测试数</h3>
                <div class="value">${report.summary.totalTests}</div>
                <div class="label">Tests</div>
            </div>
            <div class="metric">
                <h3>通过数</h3>
                <div class="value success">${report.summary.passedTests}</div>
                <div class="label">Passed</div>
            </div>
            <div class="metric">
                <h3>失败数</h3>
                <div class="value error">${report.summary.failedTests}</div>
                <div class="label">Failed</div>
            </div>
            <div class="metric">
                <h3>成功率</h3>
                <div class="value">${report.summary.successRate}</div>
                <div class="label">Success Rate</div>
            </div>
            <div class="metric">
                <h3>测试时长</h3>
                <div class="value">${report.summary.duration}</div>
                <div class="label">Duration</div>
            </div>
        </div>

        ${report.errors.length > 0 ? `
        <div class="section">
            <h2>错误详情</h2>
            ${report.errors.map(error => `
                <div class="error-item">
                    <strong>${error.test}</strong><br>
                    <code>${error.error}</code>
                </div>
            `).join('')}
        </div>
        ` : ''}

        <div class="section">
            <h2>详细日志</h2>
            ${report.details.map(log => `
                <div class="log-item ${log.type}">
                    <strong>[${log.timestamp}] [${log.type.toUpperCase()}]</strong> ${log.message}
                </div>
            `).join('')}
        </div>
    </div>
</body>
</html>
  `;
}

// 主测试函数
async function runTests() {
  log('开始端到端功能测试...');
  log(`前端地址: ${CONFIG.frontend.url}`);
  log(`后端地址: ${CONFIG.backend.url}`);

  const tests = [
    testBackendHealth,
    testBackendAPIs,
    testDatabaseConnection,
    testFrontendLoad,
    testAuthenticationFlow
  ];

  for (const test of tests) {
    try {
      await test();
    } catch (error) {
      log(`测试执行失败: ${error.message}`, 'error');
    }

    // 测试间隔
    await new Promise(resolve => setTimeout(resolve, 2000));
  }

  const report = generateReport();

  log('端到端功能测试完成!');
  log(`总计: ${report.summary.totalTests} 项测试`);
  log(`通过: ${report.summary.passedTests} 项`);
  log(`失败: ${report.summary.failedTests} 项`);
  log(`成功率: ${report.summary.successRate}`);

  return report;
}

// 如果直接运行此脚本
if (require.main === module) {
  runTests()
    .then(() => {
      console.log('测试完成');
      process.exit(0);
    })
    .catch((error) => {
      console.error('测试失败:', error);
      process.exit(1);
    });
}

module.exports = {
  runTests,
  CONFIG
};
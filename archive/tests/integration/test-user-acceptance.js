/**
 * 真实用户验收测试场景
 * 模拟真实用户使用场景，验证系统功能和用户体验
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
  },
  userScenarios: [
    {
      name: '新用户首次访问',
      description: '新用户首次访问系统的体验',
      role: 'guest'
    },
    {
      name: '管理员工作流程',
      description: '管理员管理系统的日常工作流程',
      role: 'admin'
    },
    {
      name: '律师工作流程',
      description: '律师处理案件的日常工作流程',
      role: 'lawyer'
    }
  ]
};

// 测试结果记录
const testResults = {
  startTime: new Date(),
  scenarios: [],
  summary: {
    totalScenarios: 0,
    passedScenarios: 0,
    failedScenarios: 0
  }
};

function log(message, type = 'info') {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] [${type.toUpperCase()}] ${message}`);
}

function recordScenario(name, passed, steps = [], issues = []) {
  testResults.summary.totalScenarios++;
  if (passed) {
    testResults.summary.passedScenarios++;
    log(`✅ 场景通过: ${name}`);
  } else {
    testResults.summary.failedScenarios++;
    log(`❌ 场景失败: ${name} - 问题: ${issues.join(', ')}`);
  }

  testResults.scenarios.push({
    name,
    passed,
    steps,
    issues,
    timestamp: new Date()
  });
}

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function runGuestScenario() {
  log('🔍 执行场景: 新用户首次访问');

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1920, height: 1080 }
  });

  const steps = [];
  const issues = [];

  try {
    const page = await browser.newPage();

    // 步骤1: 访问首页
    steps.push('访问系统首页');
    await page.goto(CONFIG.frontend.url, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    // 检查页面是否正常加载
    const pageTitle = await page.title();
    if (!pageTitle) {
      issues.push('页面标题未加载');
    }

    // 步骤2: 检查登录/注册入口
    steps.push('查找登录/注册入口');
    await sleep(2000);

    const loginButton = await page.$('a[href*="login"], button:contains("登录"), .login-btn');
    const registerButton = await page.$('a[href*="register"], button:contains("注册"), .register-btn');

    if (!loginButton && !registerButton) {
      issues.push('未找到登录或注册入口');
    }

    // 步骤3: 检查页面导航
    steps.push('检查页面导航结构');
    const navigation = await page.$('nav, .navigation, .navbar, .menu');
    if (!navigation) {
      issues.push('未找到页面导航');
    }

    // 步骤4: 检查系统介绍信息
    steps.push('查看系统介绍');
    const introElements = await page.$$('.hero, .intro, .description, h1, h2');
    if (introElements.length === 0) {
      issues.push('未找到系统介绍信息');
    }

    // 步骤5: 测试响应式设计
    steps.push('测试响应式设计');
    await page.setViewport({ width: 768, height: 1024 }); // 平板尺寸
    await sleep(1000);

    const mobileNavigation = await page.$('.mobile-menu, .hamburger, .menu-toggle');
    await page.setViewport({ width: 1920, height: 1080 }); // 恢复桌面尺寸

    await browser.close();

  } catch (error) {
    issues.push(`执行错误: ${error.message}`);
    await browser.close();
  }

  recordScenario('新用户首次访问', issues.length === 0, steps, issues);
}

async function runAdminScenario() {
  log('👨‍💼 执行场景: 管理员工作流程');

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1920, height: 1080 }
  });

  const steps = [];
  const issues = [];

  try {
    const page = await browser.newPage();

    // 步骤1: 访问管理页面
    steps.push('访问管理系统');
    await page.goto(CONFIG.frontend.url, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    await sleep(2000);

    // 步骤2: 查找管理入口
    steps.push('查找管理功能入口');

    // 常见的管理入口选择器
    const adminSelectors = [
      'a[href*="admin"]',
      'a[href*="manage"]',
      'a[href*="management"]',
      '.admin-btn',
      '.manage-btn',
      'button:contains("管理")',
      'button:contains("设置")'
    ];

    let adminButton = null;
    for (const selector of adminSelectors) {
      try {
        adminButton = await page.$(selector);
        if (adminButton) break;
      } catch (e) {
        // 继续尝试下一个选择器
      }
    }

    if (!adminButton) {
      issues.push('未找到管理功能入口');
    } else {
      await adminButton.click();
      await sleep(2000);
    }

    // 步骤3: 检查用户管理功能
    steps.push('检查用户管理功能');
    const userManagement = await page.$('a[href*="user"], .user-management, .users');
    if (!userManagement) {
      issues.push('未找到用户管理功能');
    }

    // 步骤4: 检查系统设置
    steps.push('检查系统设置');
    const systemSettings = await page.$('a[href*="setting"], .settings, .config');
    if (!systemSettings) {
      issues.push('未找到系统设置');
    }

    // 步骤5: 检查数据统计
    steps.push('查看数据统计');
    const statistics = await page.$('.dashboard, .stats, .statistics, .chart');
    if (!statistics) {
      issues.push('未找到数据统计功能');
    }

    await browser.close();

  } catch (error) {
    issues.push(`执行错误: ${error.message}`);
    await browser.close();
  }

  recordScenario('管理员工作流程', issues.length === 0, steps, issues);
}

async function runLawyerScenario() {
  log('⚖️ 执行场景: 律师工作流程');

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1920, height: 1080 }
  });

  const steps = [];
  const issues = [];

  try {
    const page = await browser.newPage();

    // 步骤1: 访问律师工作台
    steps.push('访问律师工作台');
    await page.goto(CONFIG.frontend.url, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    await sleep(2000);

    // 步骤2: 查找案件管理
    steps.push('查找案件管理功能');
    const caseManagement = await page.$('a[href*="case"], .case-management, .cases');
    if (!caseManagement) {
      issues.push('未找到案件管理功能');
    }

    // 步骤3: 查找客户管理
    steps.push('查找客户管理功能');
    const clientManagement = await page.$('a[href*="client"], .client-management, .clients');
    if (!clientManagement) {
      issues.push('未找到客户管理功能');
    }

    // 步骤4: 查找文档管理
    steps.push('查找文档管理功能');
    const documentManagement = await page.$('a[href*="document"], .document-management, .docs, .files');
    if (!documentManagement) {
      issues.push('未找到文档管理功能');
    }

    // 步骤5: 查找日程安排
    steps.push('查找日程安排功能');
    const calendar = await page.$('a[href*="calendar"], .schedule, .agenda, .appointment');
    if (!calendar) {
      issues.push('未找到日程安排功能');
    }

    // 步骤6: 检查搜索功能
    steps.push('检查搜索功能');
    const searchInput = await page.$('input[type="search"], .search-input, .search-box');
    if (!searchInput) {
      issues.push('未找到搜索功能');
    }

    await browser.close();

  } catch (error) {
    issues.push(`执行错误: ${error.message}`);
    await browser.close();
  }

  recordScenario('律师工作流程', issues.length === 0, steps, issues);
}

async function testSystemPerformance() {
  log('🚀 测试系统性能');

  const steps = [];
  const issues = [];

  try {
    // 测试页面加载时间
    const startTime = Date.now();
    const response = await fetch(CONFIG.frontend.url);
    const loadTime = Date.now() - startTime;

    steps.push('页面加载时间测试');
    if (loadTime > 5000) {
      issues.push(`页面加载时间过长: ${loadTime}ms`);
    }

    // 测试API响应时间
    const apiStartTime = Date.now();
    const apiResponse = await fetch(`${CONFIG.backend.url}/health`);
    const apiResponseTime = Date.now() - apiStartTime;

    steps.push('API响应时间测试');
    if (apiResponseTime > 2000) {
      issues.push(`API响应时间过长: ${apiResponseTime}ms`);
    }

    recordScenario('系统性能测试', issues.length === 0, steps, issues);

  } catch (error) {
    issues.push(`性能测试错误: ${error.message}`);
    recordScenario('系统性能测试', false, steps, issues);
  }
}

async function testAccessibility() {
  log('♿ 测试可访问性');

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1920, height: 1080 }
  });

  const steps = [];
  const issues = [];

  try {
    const page = await browser.newPage();

    await page.goto(CONFIG.frontend.url, {
      waitUntil: 'networkidle2',
      timeout: CONFIG.frontend.timeout
    });

    // 检查图片alt属性
    steps.push('检查图片alt属性');
    const imagesWithoutAlt = await page.$$eval('img:not([alt]), img[alt=""]',
      imgs => imgs.length);

    if (imagesWithoutAlt > 0) {
      issues.push(`${imagesWithoutAlt} 张图片缺少alt属性`);
    }

    // 检查标题层级
    steps.push('检查标题层级');
    const headings = await page.$$eval('h1, h2, h3, h4, h5, h6',
      headings => headings.map(h => ({
        tag: h.tagName.toLowerCase(),
        text: h.textContent.trim()
      })));

    const hasH1 = headings.some(h => h.tag === 'h1');
    if (!hasH1) {
      issues.push('页面缺少H1标题');
    }

    // 检查表单标签
    steps.push('检查表单标签');
    const inputsWithoutLabels = await page.$$eval('input:not([type="hidden"]):not([type="submit"]):not([type="button"])',
      inputs => inputs.filter(input => {
        const id = input.id;
        if (!id) return true;

        const label = document.querySelector(`label[for="${id}"]`);
        return !label;
      }).length);

    if (inputsWithoutLabels > 0) {
      issues.push(`${inputsWithoutLabels} 个输入框缺少标签`);
    }

    await browser.close();

  } catch (error) {
    issues.push(`可访问性测试错误: ${error.message}`);
    await browser.close();
  }

  recordScenario('可访问性测试', issues.length === 0, steps, issues);
}

function generateReport() {
  const endTime = new Date();
  const duration = endTime - testResults.startTime;

  const report = {
    summary: {
      startTime: testResults.startTime.toISOString(),
      endTime: endTime.toISOString(),
      duration: `${Math.round(duration / 1000)}s`,
      totalScenarios: testResults.summary.totalScenarios,
      passedScenarios: testResults.summary.passedScenarios,
      failedScenarios: testResults.summary.failedScenarios,
      successRate: `${Math.round((testResults.summary.passedScenarios / testResults.summary.totalScenarios) * 100)}%`
    },
    scenarios: testResults.scenarios,
    recommendations: generateRecommendations()
  };

  // 保存报告
  const reportPath = `user-acceptance-test-report-${Date.now()}.json`;
  require('fs').writeFileSync(reportPath, JSON.stringify(report, null, 2));

  // 生成HTML报告
  const htmlReport = generateHTMLReport(report);
  const htmlPath = `user-acceptance-test-report-${Date.now()}.html`;
  require('fs').writeFileSync(htmlPath, htmlReport);

  log(`\n=== 用户验收测试报告 ===`);
  log(`测试时长: ${report.summary.duration}`);
  log(`总场景数: ${report.summary.totalScenarios}`);
  log(`通过场景: ${report.summary.passedScenarios}`);
  log(`失败场景: ${report.summary.failedScenarios}`);
  log(`成功率: ${report.summary.successRate}`);
  log(`JSON报告: ${reportPath}`);
  log(`HTML报告: ${htmlPath}`);

  return report;
}

function generateRecommendations() {
  const recommendations = [];

  testResults.scenarios.forEach(scenario => {
    if (!scenario.passed) {
      recommendations.push({
        scenario: scenario.name,
        issues: scenario.issues,
        priority: scenario.issues.length > 2 ? 'high' : 'medium'
      });
    }
  });

  return recommendations;
}

function generateHTMLReport(report) {
  return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>用户验收测试报告</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 20px; background: #f5f7fa; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 40px; padding-bottom: 20px; border-bottom: 2px solid #e1e8ed; }
        .header h1 { color: #2c3e50; margin-bottom: 10px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 40px; }
        .metric { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 10px; text-align: center; }
        .metric h3 { margin: 0 0 10px 0; font-size: 1.1em; }
        .metric .value { font-size: 2.5em; font-weight: bold; margin-bottom: 5px; }
        .metric .label { font-size: 0.9em; opacity: 0.9; }
        .success { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
        .error { background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%); }
        .section { margin: 30px 0; }
        .section h2 { color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px; margin-bottom: 20px; }
        .scenario { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 8px; padding: 20px; margin: 15px 0; }
        .scenario.passed { border-left: 4px solid #28a745; }
        .scenario.failed { border-left: 4px solid #dc3545; }
        .scenario h3 { margin: 0 0 15px 0; color: #495057; }
        .scenario-title { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
        .status-badge { padding: 4px 12px; border-radius: 20px; font-size: 0.8em; font-weight: bold; }
        .status-passed { background: #d4edda; color: #155724; }
        .status-failed { background: #f8d7da; color: #721c24; }
        .steps { margin: 15px 0; }
        .step { background: #e9ecef; padding: 8px 12px; margin: 5px 0; border-radius: 4px; font-size: 0.9em; }
        .issues { margin: 15px 0; }
        .issue { background: #f8d7da; border: 1px solid #f5c6cb; border-radius: 4px; padding: 10px; margin: 5px 0; color: #721c24; }
        .recommendations { background: #fff3cd; border: 1px solid #ffeaa7; border-radius: 8px; padding: 20px; }
        .recommendation { margin: 10px 0; padding: 10px; background: white; border-radius: 4px; }
        .recommendation.high { border-left: 4px solid #dc3545; }
        .recommendation.medium { border-left: 4px solid #ffc107; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>律所OA系统 - 用户验收测试报告</h1>
            <p>生成时间: ${new Date().toLocaleString('zh-CN')}</p>
        </div>

        <div class="summary">
            <div class="metric">
                <h3>总场景数</h3>
                <div class="value">${report.summary.totalScenarios}</div>
                <div class="label">Scenarios</div>
            </div>
            <div class="metric success">
                <h3>通过场景</h3>
                <div class="value">${report.summary.passedScenarios}</div>
                <div class="label">Passed</div>
            </div>
            <div class="metric error">
                <h3>失败场景</h3>
                <div class="value">${report.summary.failedScenarios}</div>
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

        <div class="section">
            <h2>测试场景详情</h2>
            ${report.scenarios.map(scenario => `
                <div class="scenario ${scenario.passed ? 'passed' : 'failed'}">
                    <div class="scenario-title">
                        <h3>${scenario.name}</h3>
                        <span class="status-badge ${scenario.passed ? 'status-passed' : 'status-failed'}">
                            ${scenario.passed ? '通过' : '失败'}
                        </span>
                    </div>
                    <div class="steps">
                        <strong>执行步骤:</strong>
                        ${scenario.steps.map(step => `<div class="step">✓ ${step}</div>`).join('')}
                    </div>
                    ${scenario.issues.length > 0 ? `
                        <div class="issues">
                            <strong>发现问题:</strong>
                            ${scenario.issues.map(issue => `<div class="issue">⚠️ ${issue}</div>`).join('')}
                        </div>
                    ` : ''}
                </div>
            `).join('')}
        </div>

        ${report.recommendations.length > 0 ? `
        <div class="section">
            <h2>改进建议</h2>
            <div class="recommendations">
                ${report.recommendations.map(rec => `
                    <div class="recommendation ${rec.priority}">
                        <strong>${rec.scenario}</strong>
                        <ul>
                            ${rec.issues.map(issue => `<li>${issue}</li>`).join('')}
                        </ul>
                        <small>优先级: ${rec.priority === 'high' ? '高' : '中'}</small>
                    </div>
                `).join('')}
            </div>
        </div>
        ` : ''}

        <div class="section">
            <h2>测试结论</h2>
            <p>
                ${report.summary.successRate === '100%' ?
                  '🎉 所有测试场景均通过，系统满足用户验收标准。' :
                  report.summary.successRate >= '80%' ?
                  '✅ 大部分测试场景通过，系统基本满足用户需求，建议优化失败场景。' :
                  '⚠️ 部分测试场景失败，需要进一步改进才能满足用户需求。'
                }
            </p>
        </div>
    </div>
</body>
</html>
  `;
}

async function runUserAcceptanceTests() {
  log('开始用户验收测试...');
  log(`前端地址: ${CONFIG.frontend.url}`);
  log(`后端地址: ${CONFIG.backend.url}`);

  try {
    // 执行所有用户场景
    await runGuestScenario();
    await runAdminScenario();
    await runLawyerScenario();
    await testSystemPerformance();
    await testAccessibility();

    const report = generateReport();

    log('用户验收测试完成！');
    return report;

  } catch (error) {
    log('测试执行失败:', error, 'error');
    throw error;
  }
}

// 运行测试
if (require.main === module) {
  runUserAcceptanceTests()
    .then(() => {
      console.log('测试完成');
      process.exit(0);
    })
    .catch((error) => {
      console.error('测试失败:', error);
      process.exit(1);
    });
}

module.exports = { runUserAcceptanceTests };
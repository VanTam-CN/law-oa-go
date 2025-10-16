const puppeteer = require('puppeteer');
const axios = require('axios');

class SystemDiagnostics {
  constructor() {
    this.baseURL = 'http://localhost:8080';
    this.frontendURL = 'http://localhost:3003';
    this.token = null;
    this.browser = null;
    this.page = null;
    this.results = {
      backend: {},
      frontend: {},
      integration: {},
      issues: []
    };
  }

  async initialize() {
    console.log('🚀 启动系统全面诊断...');
    
    // 启动浏览器
    this.browser = await puppeteer.launch({
      headless: false,
      defaultViewport: { width: 1920, height: 1080 },
      args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    this.page = await this.browser.newPage();
    
    // 监听控制台错误
    this.page.on('console', msg => {
      if (msg.type() === 'error') {
        this.results.issues.push({
          type: 'frontend_console_error',
          message: msg.text(),
          timestamp: new Date().toISOString()
        });
      }
    });

    // 监听网络错误
    this.page.on('response', response => {
      if (response.status() >= 400) {
        this.results.issues.push({
          type: 'network_error',
          url: response.url(),
          status: response.status(),
          timestamp: new Date().toISOString()
        });
      }
    });
  }

  async testBackendAPIs() {
    console.log('🔧 测试后端API...');
    
    try {
      // 1. 测试登录API
      const loginResponse = await axios.post(`${this.baseURL}/api/auth/login`, {
        email: 'admin@example.com',
        password: 'admin123'
      });
      
      if (loginResponse.data.success) {
        this.token = loginResponse.data.data.token;
        this.results.backend.auth = { status: 'success', message: '登录成功' };
        console.log('✅ 登录API正常');
      }
    } catch (error) {
      this.results.backend.auth = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'backend_auth_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 登录API失败:', error.message);
    }

    if (!this.token) {
      console.log('⚠️ 无法获取token，跳过需要认证的API测试');
      return;
    }

    const headers = { Authorization: `Bearer ${this.token}` };

    // 2. 测试案件API
    try {
      const casesResponse = await axios.get(`${this.baseURL}/api/cases?page=1&page_size=10`, { headers });
      this.results.backend.cases = { 
        status: 'success', 
        count: casesResponse.data.data?.length || 0,
        total: casesResponse.data.pagination?.total || 0
      };
      console.log(`✅ 案件API正常，共${this.results.backend.cases.total}个案件`);
    } catch (error) {
      this.results.backend.cases = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'backend_cases_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 案件API失败:', error.message);
    }

    // 3. 测试客户API
    try {
      const clientsResponse = await axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=10`, { headers });
      this.results.backend.clients = { 
        status: 'success', 
        count: clientsResponse.data.data?.list?.length || 0
      };
      console.log(`✅ 客户API正常，共${this.results.backend.clients.count}个客户`);
    } catch (error) {
      this.results.backend.clients = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'backend_clients_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 客户API失败:', error.message);
    }

    // 4. 测试律师API
    try {
      const lawyersResponse = await axios.get(`${this.baseURL}/api/lawfirm/lawyers?pageNum=1&pageSize=10`, { headers });
      this.results.backend.lawyers = { 
        status: 'success', 
        count: lawyersResponse.data.data?.list?.length || 0
      };
      console.log(`✅ 律师API正常，共${this.results.backend.lawyers.count}个律师`);
    } catch (error) {
      this.results.backend.lawyers = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'backend_lawyers_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 律师API失败:', error.message);
    }

    // 5. 测试利益冲突检测API
    try {
      const conflictResponse = await axios.post(`${this.baseURL}/api/conflict/check`, {
        clientId: "1",
        clientName: "测试客户",
        caseName: "测试案件",
        caseType: "civil",
        clientType: "PERSON",
        otherParties: ["对方当事人"],
        searchYears: 5,
        userId: "1",
        requestTime: new Date().toISOString()
      }, { headers });
      
      this.results.backend.conflict = { 
        status: 'success', 
        hasConflict: conflictResponse.data.data?.hasConflict || false
      };
      console.log('✅ 利益冲突检测API正常');
    } catch (error) {
      this.results.backend.conflict = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'backend_conflict_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 利益冲突检测API失败:', error.message);
    }
  }

  async testFrontendPages() {
    console.log('🖥️ 测试前端页面...');

    // 1. 测试首页加载
    try {
      await this.page.goto(this.frontendURL, { waitUntil: 'networkidle0', timeout: 30000 });
      const title = await this.page.title();
      this.results.frontend.homepage = { status: 'success', title };
      console.log('✅ 首页加载正常:', title);
    } catch (error) {
      this.results.frontend.homepage = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'frontend_homepage_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 首页加载失败:', error.message);
    }

    // 2. 设置测试token
    try {
      await this.page.evaluate(() => {
        const testToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzYwMDkzMjkzLCJpYXQiOjE3NjAwMDY4OTN9.868MdMFobxA9bth5oOGvXPMVnDvkdNfAE9U9Vq29I4s';
        localStorage.setItem('auth_token', testToken);
        localStorage.setItem('law_oa_token', testToken);
      });
      console.log('✅ 测试token已设置');
    } catch (error) {
      console.log('❌ 设置token失败:', error.message);
    }

    // 3. 测试案件管理页面
    await this.testCaseManagementPage();

    // 4. 测试客户管理页面
    await this.testClientManagementPage();

    // 5. 测试律师管理页面
    await this.testLawyerManagementPage();
  }

  async testCaseManagementPage() {
    console.log('📋 测试案件管理页面...');
    
    try {
      await this.page.goto(`${this.frontendURL}/case`, { waitUntil: 'networkidle0', timeout: 30000 });
      
      // 等待页面加载
      await this.page.waitForSelector('.case-management', { timeout: 10000 });
      
      // 检查统计卡片
      const statsCards = await this.page.$$('.stats-row .ant-card');
      const hasData = await this.page.$('.ant-table-tbody tr:not(.ant-table-placeholder)');
      
      this.results.frontend.caseManagement = {
        status: 'success',
        statsCardsCount: statsCards.length,
        hasTableData: !!hasData,
        url: this.page.url()
      };
      
      console.log(`✅ 案件管理页面加载正常，统计卡片: ${statsCards.length}个，有数据: ${!!hasData}`);
      
      // 测试搜索功能
      const searchInput = await this.page.$('input[placeholder*="搜索"]');
      if (searchInput) {
        await searchInput.type('测试');
        await this.page.keyboard.press('Enter');
        await this.page.waitForTimeout(2000);
        console.log('✅ 搜索功能测试完成');
      }
      
      // 测试新建按钮
      const createButton = await this.page.$('button:has-text("新建案件")');
      if (createButton) {
        console.log('✅ 新建案件按钮存在');
      }
      
    } catch (error) {
      this.results.frontend.caseManagement = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'frontend_case_management_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 案件管理页面测试失败:', error.message);
    }
  }

  async testClientManagementPage() {
    console.log('👥 测试客户管理页面...');
    
    try {
      await this.page.goto(`${this.frontendURL}/client`, { waitUntil: 'networkidle0', timeout: 30000 });
      
      // 等待页面加载
      await this.page.waitForSelector('.client-management', { timeout: 10000 });
      
      const hasData = await this.page.$('.ant-table-tbody tr:not(.ant-table-placeholder)');
      
      this.results.frontend.clientManagement = {
        status: 'success',
        hasTableData: !!hasData,
        url: this.page.url()
      };
      
      console.log(`✅ 客户管理页面加载正常，有数据: ${!!hasData}`);
      
    } catch (error) {
      this.results.frontend.clientManagement = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'frontend_client_management_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 客户管理页面测试失败:', error.message);
    }
  }

  async testLawyerManagementPage() {
    console.log('⚖️ 测试律师管理页面...');
    
    try {
      await this.page.goto(`${this.frontendURL}/lawyer`, { waitUntil: 'networkidle0', timeout: 30000 });
      
      // 等待页面加载
      await this.page.waitForSelector('.lawyer-management', { timeout: 10000 });
      
      const hasData = await this.page.$('.ant-table-tbody tr:not(.ant-table-placeholder)');
      
      this.results.frontend.lawyerManagement = {
        status: 'success',
        hasTableData: !!hasData,
        url: this.page.url()
      };
      
      console.log(`✅ 律师管理页面加载正常，有数据: ${!!hasData}`);
      
    } catch (error) {
      this.results.frontend.lawyerManagement = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'frontend_lawyer_management_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 律师管理页面测试失败:', error.message);
    }
  }

  async testDataIntegration() {
    console.log('🔗 测试数据关联性...');
    
    if (!this.token) {
      console.log('⚠️ 无token，跳过数据关联测试');
      return;
    }

    const headers = { Authorization: `Bearer ${this.token}` };

    try {
      // 获取案件数据
      const casesResponse = await axios.get(`${this.baseURL}/api/cases?page=1&page_size=5`, { headers });
      const cases = casesResponse.data.data || [];

      // 获取客户数据
      const clientsResponse = await axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=100`, { headers });
      const clients = clientsResponse.data.data?.list || [];

      // 获取律师数据
      const lawyersResponse = await axios.get(`${this.baseURL}/api/lawfirm/lawyers?pageNum=1&pageSize=100`, { headers });
      const lawyers = lawyersResponse.data.data?.list || [];

      // 检查数据关联
      let validCaseClientLinks = 0;
      let validCaseLawyerLinks = 0;

      for (const caseItem of cases) {
        // 检查客户关联
        const clientExists = clients.find(c => c.id === caseItem.client_id);
        if (clientExists) validCaseClientLinks++;

        // 检查律师关联
        const lawyerExists = lawyers.find(l => l.id === caseItem.lawyer_id);
        if (lawyerExists) validCaseLawyerLinks++;
      }

      this.results.integration.dataLinks = {
        status: 'success',
        totalCases: cases.length,
        validCaseClientLinks,
        validCaseLawyerLinks,
        clientLinkRate: cases.length > 0 ? (validCaseClientLinks / cases.length * 100).toFixed(1) : 0,
        lawyerLinkRate: cases.length > 0 ? (validCaseLawyerLinks / cases.length * 100).toFixed(1) : 0
      };

      console.log(`✅ 数据关联检查完成:`);
      console.log(`   案件-客户关联率: ${this.results.integration.dataLinks.clientLinkRate}%`);
      console.log(`   案件-律师关联率: ${this.results.integration.dataLinks.lawyerLinkRate}%`);

    } catch (error) {
      this.results.integration.dataLinks = { status: 'error', message: error.message };
      this.results.issues.push({
        type: 'integration_data_links_error',
        message: error.message,
        timestamp: new Date().toISOString()
      });
      console.log('❌ 数据关联测试失败:', error.message);
    }
  }

  async generateReport() {
    console.log('📊 生成诊断报告...');
    
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        totalIssues: this.results.issues.length,
        backendStatus: Object.values(this.results.backend).filter(r => r.status === 'success').length,
        frontendStatus: Object.values(this.results.frontend).filter(r => r.status === 'success').length,
        integrationStatus: Object.values(this.results.integration).filter(r => r.status === 'success').length
      },
      results: this.results,
      recommendations: this.generateRecommendations()
    };

    // 保存报告
    const fs = require('fs');
    fs.writeFileSync('system_diagnosis_report.json', JSON.stringify(report, null, 2));
    
    console.log('\n📋 诊断报告摘要:');
    console.log(`   发现问题: ${report.summary.totalIssues}个`);
    console.log(`   后端API正常: ${report.summary.backendStatus}个`);
    console.log(`   前端页面正常: ${report.summary.frontendStatus}个`);
    console.log(`   数据集成正常: ${report.summary.integrationStatus}个`);
    
    return report;
  }

  generateRecommendations() {
    const recommendations = [];
    
    // 基于发现的问题生成建议
    this.results.issues.forEach(issue => {
      switch (issue.type) {
        case 'frontend_console_error':
          recommendations.push({
            priority: 'high',
            category: 'frontend',
            issue: '前端控制台错误',
            solution: '检查并修复JavaScript错误，确保代码质量'
          });
          break;
        case 'network_error':
          recommendations.push({
            priority: 'high',
            category: 'integration',
            issue: '网络请求错误',
            solution: '检查API端点和网络连接，修复HTTP错误'
          });
          break;
        case 'backend_auth_error':
          recommendations.push({
            priority: 'critical',
            category: 'backend',
            issue: '认证系统错误',
            solution: '修复登录API，确保认证流程正常'
          });
          break;
      }
    });

    // 数据关联问题
    if (this.results.integration.dataLinks?.clientLinkRate < 90) {
      recommendations.push({
        priority: 'medium',
        category: 'data',
        issue: '案件-客户关联不完整',
        solution: '检查数据完整性，修复缺失的关联关系'
      });
    }

    if (this.results.integration.dataLinks?.lawyerLinkRate < 90) {
      recommendations.push({
        priority: 'medium',
        category: 'data',
        issue: '案件-律师关联不完整',
        solution: '检查数据完整性，修复缺失的关联关系'
      });
    }

    return recommendations;
  }

  async cleanup() {
    if (this.browser) {
      await this.browser.close();
    }
  }

  async run() {
    try {
      await this.initialize();
      await this.testBackendAPIs();
      await this.testFrontendPages();
      await this.testDataIntegration();
      const report = await this.generateReport();
      return report;
    } catch (error) {
      console.error('❌ 诊断过程中发生错误:', error);
      throw error;
    } finally {
      await this.cleanup();
    }
  }
}

// 运行诊断
async function main() {
  const diagnostics = new SystemDiagnostics();
  try {
    const report = await diagnostics.run();
    console.log('\n🎉 系统诊断完成！报告已保存到 system_diagnosis_report.json');
    
    // 如果有严重问题，立即开始修复
    const criticalIssues = report.recommendations.filter(r => r.priority === 'critical');
    if (criticalIssues.length > 0) {
      console.log('\n🚨 发现严重问题，需要立即修复:');
      criticalIssues.forEach(issue => {
        console.log(`   - ${issue.issue}: ${issue.solution}`);
      });
    }
    
    return report;
  } catch (error) {
    console.error('诊断失败:', error);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = SystemDiagnostics;
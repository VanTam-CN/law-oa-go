const axios = require('axios');

const BASE_URL = 'http://localhost:8080';
const API_BASE = '/api';

// 使用之前获取的有效令牌
const VALID_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo5LCJ1c2VybmFtZSI6InRlc3RhZG1pbkBsYXdmaXJtLmNvbSIsInJvbGUiOiJhZG1pbiIsImV4cCI6MTc1OTM3MjQwMiwiaWF0IjoxNzU5Mjg2MDAyfQ.o2wHdd2HlUn3sYBRDTJVo5gblkcoKVg5Un3ZYyvAyS8';

// 测试结果收集器
const testResults = {
  passed: 0,
  failed: 0,
  results: []
};

// 创建带认证的客户端
const apiClient = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${VALID_TOKEN}`
  }
});

// 测试函数
async function testAPI(method, url, data = null, description = '') {
  try {
    const config = {
      method: method,
      url: API_BASE + url
    };

    if (data && (method === 'POST' || 'PUT')) {
      config.data = data;
    }

    const response = await apiClient(config);

    testResults.passed++;
    testResults.results.push({
      description,
      status: '✅ PASS',
      url: API_BASE + url,
      status_code: response.status,
      response: response.data
    });

    console.log(`✅ ${description}: ${response.status}`);
    return true;
  } catch (error) {
    testResults.failed++;
    testResults.results.push({
      description,
      status: '❌ FAIL',
      url: API_BASE + url,
      error: error.response?.data?.message || error.message,
      status_code: error.response?.status || 'NO_RESPONSE'
    });

    console.log(`❌ ${description}: ${error.response?.status || 'NO_RESPONSE'} - ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 主测试函数
async function runAuthenticatedTests() {
  console.log('🚀 开始认证API测试...\n');

  // 测试仪表盘API
  console.log('📊 测试仪表盘API...');
  await testAPI('GET', '/dashboard', null, '获取仪表盘数据');
  await testAPI('GET', '/dashboard/statistics', null, '获取统计数据');
  await testAPI('GET', '/todos', null, '获取待办事项');
  await testAPI('GET', '/activities', null, '获取活动记录');

  // 测试客户管理API
  console.log('\n👥 测试客户管理API...');
  await testAPI('GET', '/clients', null, '获取客户列表');
  await testAPI('GET', '/clients/stats', null, '获取客户统计');

  // 测试案件管理API
  console.log('\n⚖️ 测试案件管理API...');
  await testAPI('GET', '/cases', null, '获取案件列表');
  await testAPI('GET', '/cases/stats', null, '获取案件统计');

  // 测试文件管理API
  console.log('\n📁 测试文件管理API...');
  await testAPI('GET', '/files', null, '获取文件列表');
  await testAPI('GET', '/files/stats', null, '获取文件统计');

  // 测试律师管理API
  console.log('\n👨‍⚖️ 测试律师管理API...');
  await testAPI('GET', '/lawyers', null, '获取律师列表');
  await testAPI('GET', '/lawyers/stats', null, '获取律师统计');

  // 测试审批中心API
  console.log('\n✅ 测试审批中心API...');
  await testAPI('GET', '/approvals', null, '获取审批列表');
  await testAPI('GET', '/approvals/stats', null, '获取审批统计');
  await testAPI('GET', '/approvals/pending', null, '获取待审批事项');

  // 测试用户管理API
  console.log('\n👤 测试用户管理API...');
  await testAPI('GET', '/users', null, '获取用户列表');
  await testAPI('GET', '/users/profile', null, '获取用户资料');

  // 测试冲突检测API
  console.log('\n🔍 测试冲突检测API...');
  await testAPI('GET', '/conflict/stats', null, '获取冲突检测统计');
  await testAPI('GET', '/conflict/rules', null, '获取冲突检测规则');
  await testAPI('GET', '/conflict/standards', null, '获取MCP标准');

  // 打印测试报告
  console.log('\n📈 测试报告:');
  console.log(`✅ 通过: ${testResults.passed}`);
  console.log(`❌ 失败: ${testResults.failed}`);
  console.log(`📊 成功率: ${((testResults.passed / (testResults.passed + testResults.failed)) * 100).toFixed(1)}%`);

  if (testResults.failed > 0) {
    console.log('\n❌ 失败的测试:');
    testResults.results
      .filter(r => r.status === '❌ FAIL')
      .forEach(r => {
        console.log(`  ${r.description}: ${r.error} (${r.status_code})`);
      });
  }

  console.log('\n🎯 第一阶段认证API测试完成！');
}

runAuthenticatedTests();
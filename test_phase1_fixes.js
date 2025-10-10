#!/usr/bin/env node

/**
 * API测试脚本 - 验证第一阶段修复
 * 测试所有主要API端点的可达性
 */

const axios = require('axios');

const BASE_URL = 'http://localhost:8080';
const API_BASE = '/api';

// 测试数据
const testUser = {
  email: 'testadmin@lawfirm.com',
  password: 'TestAdmin123!'
};

// 用于存储JWT令牌
let authToken = '';

// 设置请求拦截器
const apiClient = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// 设置认证令牌
apiClient.interceptors.request.use(config => {
  if (authToken) {
    config.headers.Authorization = `Bearer ${authToken}`;
  }
  return config;
});

// 测试结果收集器
const testResults = {
  passed: 0,
  failed: 0,
  results: []
};

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
      data_keys: response.data ? Object.keys(response.data) : []
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
async function runTests() {
  console.log('🚀 开始API测试...\n');

  // 1. 测试登录
  console.log('📝 测试认证功能...');
  await testAPI('POST', '/auth/login', testUser, '用户登录');

  // 从登录响应中提取令牌（如果登录成功）
  try {
    const loginResponse = await apiClient.post('/api/auth/login', testUser);
    if (loginResponse.data && loginResponse.data.data && loginResponse.data.data.token) {
      authToken = loginResponse.data.data.token;
      console.log('🔑 认证令牌获取成功');
    }
  } catch (error) {
    console.log('⚠️  登录失败，使用无效令牌继续测试其他端点');
  }

  console.log('\n📊 测试仪表盘API...');
  await testAPI('GET', '/dashboard', null, '获取仪表盘数据');
  await testAPI('GET', '/dashboard/statistics', null, '获取统计数据');
  await testAPI('GET', '/todos', null, '获取待办事项');
  await testAPI('GET', '/activities', null, '获取活动记录');

  console.log('\n👥 测试客户管理API...');
  await testAPI('GET', '/clients', null, '获取客户列表');
  await testAPI('GET', '/clients/stats', null, '获取客户统计');

  console.log('\n⚖️ 测试案件管理API...');
  await testAPI('GET', '/cases/stats', null, '获取案件统计');

  console.log('\n📁 测试文件管理API...');
  await testAPI('GET', '/files', null, '获取文件列表');
  await testAPI('GET', '/files/stats', null, '获取文件统计');

  console.log('\n👨‍⚖️ 测试律师管理API...');
  await testAPI('GET', '/lawyers', null, '获取律师列表');
  await testAPI('GET', '/lawyers/stats', null, '获取律师统计');

  console.log('\n✅ 测试审批中心API...');
  await testAPI('GET', '/approvals', null, '获取审批列表');
  await testAPI('GET', '/approvals/stats', null, '获取审批统计');
  await testAPI('GET', '/approvals/pending', null, '获取待审批事项');

  console.log('\n👤 测试用户管理API...');
  await testAPI('GET', '/admin/users', null, '获取用户列表');

  console.log('\n🔍 测试冲突检测API...');
  await testAPI('GET', '/conflict/stats', null, '获取冲突检测统计');
  await testAPI('GET', '/conflict/rules', null, '获取冲突检测规则');
  await testAPI('GET', '/conflict/standards', null, '获取MCP标准');

  console.log('\n📈 测试报告:');
  console.log(`✅ 通过: ${testResults.passed}`);
  console.log(`❌ 失败: ${testResults.failed}`);
  console.log(`📊 成功率: ${((testResults.passed / (testResults.passed + testResults.failed)) * 100).toFixed(1)}%`);

  // 显示失败的测试
  const failedTests = testResults.results.filter(r => r.status === '❌ FAIL');
  if (failedTests.length > 0) {
    console.log('\n❌ 失败的测试:');
    failedTests.forEach(test => {
      console.log(`  ${test.description}: ${test.error} (${test.status_code})`);
    });
  }

  console.log('\n🎯 第一阶段修复验证完成！');

  // 成功退出
  process.exit(testResults.failed === 0 ? 0 : 1);
}

// 运行测试
runTests().catch(error => {
  console.error('❌ 测试执行失败:', error.message);
  process.exit(1);
});
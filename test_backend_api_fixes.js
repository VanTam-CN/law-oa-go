#!/usr/bin/env node

/**
 * Law OA Go 系统修复验证脚本
 * 验证后端API路由和前端服务配置是否正确
 */

const axios = require('axios');

// 配置
const API_BASE_URL = 'http://localhost:8080/api';

// 测试用户凭证
const TEST_USER = {
  email: 'admin@lawfirm.com',
  password: 'Admin123!'
};

let authToken = null;

// 颜色输出
const colors = {
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  reset: '\x1b[0m'
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

// API请求函数
async function apiRequest(method, endpoint, data = null) {
  try {
    const config = {
      method,
      url: `${API_BASE_URL}${endpoint}`,
      headers: {}
    };

    if (authToken) {
      config.headers.Authorization = `Bearer ${authToken}`;
    }

    if (data) {
      config.data = data;
      config.headers['Content-Type'] = 'application/json';
    }

    const response = await axios(config);
    return { success: true, data: response.data };
  } catch (error) {
    return {
      success: false,
      error: error.response?.data || error.message,
      status: error.response?.status
    };
  }
}

// 测试认证
async function testAuthentication() {
  log('\n=== 测试认证功能 ===', 'blue');

  // 测试登录
  const loginResult = await apiRequest('POST', '/auth/login', TEST_USER);
  if (loginResult.success) {
    authToken = loginResult.data.token;
    log('✅ 用户登录成功', 'green');
    return true;
  } else {
    log(`❌ 用户登录失败: ${loginResult.error}`, 'red');
    return false;
  }
}

// 测试客户管理API
async function testClientAPIs() {
  log('\n=== 测试客户管理API ===', 'blue');

  // 测试获取客户列表
  const listResult = await apiRequest('GET', '/clients');
  if (listResult.success) {
    log('✅ 获取客户列表成功', 'green');
  } else {
    log(`❌ 获取客户列表失败: ${listResult.error}`, 'red');
  }

  // 测试创建客户
  const testClient = {
    name: '测试客户',
    email: 'test@example.com',
    phone: '13800138000',
    address: '测试地址',
    company: '测试公司'
  };

  const createResult = await apiRequest('POST', '/clients', testClient);
  if (createResult.success) {
    log('✅ 创建客户成功', 'green');
    const clientId = createResult.data.id;

    // 测试获取客户详情
    const getResult = await apiRequest('GET', `/clients/${clientId}`);
    if (getResult.success) {
      log('✅ 获取客户详情成功', 'green');
    } else {
      log(`❌ 获取客户详情失败: ${getResult.error}`, 'red');
    }

    // 测试更新客户
    const updateData = { name: '更新后的客户名称' };
    const updateResult = await apiRequest('PUT', `/clients/${clientId}`, updateData);
    if (updateResult.success) {
      log('✅ 更新客户成功', 'green');
    } else {
      log(`❌ 更新客户失败: ${updateResult.error}`, 'red');
    }

    // 测试删除客户
    const deleteResult = await apiRequest('DELETE', `/clients/${clientId}`);
    if (deleteResult.success) {
      log('✅ 删除客户成功', 'green');
    } else {
      log(`❌ 删除客户失败: ${deleteResult.error}`, 'red');
    }
  } else {
    log(`❌ 创建客户失败: ${createResult.error}`, 'red');
  }

  // 测试客户统计
  const statsResult = await apiRequest('GET', '/clients/stats');
  if (statsResult.success) {
    log('✅ 获取客户统计成功', 'green');
  } else {
    log(`❌ 获取客户统计失败: ${statsResult.error}`, 'red');
  }
}

// 测试仪表盘API
async function testDashboardAPIs() {
  log('\n=== 测试仪表盘API ===', 'blue');

  // 测试获取统计数据
  const statsResult = await apiRequest('GET', '/dashboard/statistics');
  if (statsResult.success) {
    log('✅ 获取仪表盘统计成功', 'green');
  } else {
    log(`❌ 获取仪表盘统计失败: ${statsResult.error}`, 'red');
  }

  // 测试获取待办事项
  const todosResult = await apiRequest('GET', '/dashboard/todos');
  if (todosResult.success) {
    log('✅ 获取待办事项成功', 'green');
  } else {
    log(`❌ 获取待办事项失败: ${todosResult.error}`, 'red');
  }

  // 测试获取活动记录
  const activitiesResult = await apiRequest('GET', '/dashboard/activities');
  if (activitiesResult.success) {
    log('✅ 获取活动记录成功', 'green');
  } else {
    log(`❌ 获取活动记录失败: ${activitiesResult.error}`, 'red');
  }
}

// 测试案件管理API
async function testCaseAPIs() {
  log('\n=== 测试案件管理API ===', 'blue');

  // 测试获取案件列表
  const listResult = await apiRequest('GET', '/cases');
  if (listResult.success) {
    log('✅ 获取案件列表成功', 'green');
  } else {
    log(`❌ 获取案件列表失败: ${listResult.error}`, 'red');
  }

  // 测试案件统计
  const statsResult = await apiRequest('GET', '/cases/stats');
  if (statsResult.success) {
    log('✅ 获取案件统计成功', 'green');
  } else {
    log(`❌ 获取案件统计失败: ${statsResult.error}`, 'red');
  }
}

// 主测试函数
async function runTests() {
  log('Law OA Go 系统修复验证测试', 'blue');
  log('=====================================', 'blue');

  // 测试认证
  const authSuccess = await testAuthentication();
  if (!authSuccess) {
    log('\n❌ 认证失败，跳过其他测试', 'red');
    process.exit(1);
  }

  // 测试各个模块
  await testClientAPIs();
  await testDashboardAPIs();
  await testCaseAPIs();

  log('\n=====================================', 'blue');
  log('测试完成！', 'blue');
  log('注意：此测试脚本验证了后端API修复的效果。', 'yellow');
  log('如果仍有问题，请检查：', 'yellow');
  log('1. 后端服务是否正在运行 (端口8080)', 'yellow');
  log('2. 数据库连接是否正常', 'yellow');
  log('3. 测试用户是否存在 (admin@lawfirm.com)', 'yellow');
}

// 运行测试
runTests().catch(error => {
  log(`\n测试脚本执行失败: ${error.message}`, 'red');
  process.exit(1);
});
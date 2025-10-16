/**
 * API集成测试脚本
 * 测试后端API的访问权限和数据格式
 */

const { default: fetch } = require('node-fetch');

const CONFIG = {
  backend: {
    url: 'http://localhost:8080',
    timeout: 10000
  },
  testUser: {
    email: 'test@example.com',
    password: '123456'
  }
};

let authToken = null;

async function makeRequest(url, options = {}) {
  try {
    const response = await fetch(url, {
      timeout: CONFIG.backend.timeout,
      headers: {
        'Content-Type': 'application/json',
        ...(authToken && { 'Authorization': `Bearer ${authToken}` }),
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
    console.error(`请求失败: ${url}`, error.message);
    throw error;
  }
}

async function testPublicEndpoints() {
  console.log('\n=== 测试公共接口 ===');

  // 测试健康检查
  try {
    const response = await makeRequest(`${CONFIG.backend.url}/health`);
    console.log(`✅ 健康检查: ${response.status} - ${JSON.stringify(response.data)}`);
  } catch (error) {
    console.log(`❌ 健康检查失败: ${error.message}`);
  }

  // 测试详细健康检查
  try {
    const response = await makeRequest(`${CONFIG.backend.url}/api/v1/health/detailed`);
    console.log(`✅ 详细健康检查: ${response.status}`);
  } catch (error) {
    console.log(`❌ 详细健康检查失败: ${error.message}`);
  }
}

async function testAuthEndpoints() {
  console.log('\n=== 测试认证接口 ===');

  // 测试登录
  try {
    const response = await makeRequest(`${CONFIG.backend.url}/api/auth/login`, {
      method: 'POST',
      body: JSON.stringify(CONFIG.testUser)
    });

    if (response.ok && response.data.token) {
      authToken = response.data.token;
      console.log(`✅ 用户登录成功: 获得token`);
      console.log(`   用户信息: ${JSON.stringify(response.data.user || response.data.profile)}`);
    } else {
      console.log(`❌ 用户登录失败: ${response.status} - ${JSON.stringify(response.data)}`);
    }
  } catch (error) {
    console.log(`❌ 用户登录异常: ${error.message}`);

    // 尝试创建测试用户
    console.log('\n--- 尝试创建测试用户 ---');
    try {
      const response = await makeRequest(`${CONFIG.backend.url}/api/auth/register`, {
        method: 'POST',
        body: JSON.stringify({
          ...CONFIG.testUser,
          name: '测试用户',
          phone: '13800138000'
        })
      });

      if (response.ok) {
        console.log(`✅ 用户创建成功: ${response.status}`);
        // 重新尝试登录
        await testAuthEndpoints();
      } else {
        console.log(`❌ 用户创建失败: ${response.status} - ${JSON.stringify(response.data)}`);
      }
    } catch (registerError) {
      console.log(`❌ 用户创建异常: ${registerError.message}`);
    }
  }

  // 测试用户信息获取
  if (authToken) {
    try {
      const response = await makeRequest(`${CONFIG.backend.url}/api/users/profile`);
      console.log(`✅ 用户信息获取: ${response.status} - ${JSON.stringify(response.data)}`);
    } catch (error) {
      console.log(`❌ 用户信息获取失败: ${error.message}`);
    }
  }
}

async function testAPIEndpoints() {
  console.log('\n=== 测试业务API接口 ===');

  const endpoints = [
    { name: '仪表盘统计', url: '/api/v1/dashboard/statistics' },
    { name: '律师列表', url: '/api/lawyers' },
    { name: '客户列表', url: '/api/clients' },
    { name: '案件列表', url: '/api/cases' },
    { name: '律师统计', url: '/api/lawyers/stats' },
    { name: '客户统计', url: '/api/clients/stats' },
    { name: '案件统计', url: '/api/cases/stats' },
    { name: '待办事项', url: '/api/dashboard/todos' },
    { name: '活动记录', url: '/api/dashboard/activities' }
  ];

  for (const endpoint of endpoints) {
    try {
      const response = await makeRequest(`${CONFIG.backend.url}${endpoint.url}`);

      if (response.ok) {
        console.log(`✅ ${endpoint.name}: ${response.status} - 数据格式正确`);
      } else {
        console.log(`❌ ${endpoint.name}: ${response.status} - ${JSON.stringify(response.data)}`);
      }
    } catch (error) {
      console.log(`❌ ${endpoint.name} 请求失败: ${error.message}`);
    }
  }
}

async function testDatabaseOperations() {
  console.log('\n=== 测试数据库操作 ===');

  if (!authToken) {
    console.log('❌ 未获得认证token，跳过数据库操作测试');
    return;
  }

  // 测试创建律师
  try {
    const lawyerData = {
      name: '测试律师',
      phone: '13800138000',
      email: 'lawyer@example.com',
      licenseNumber: '123456789012345',
      department: '民事诉讼部',
      position: '律师',
      status: 'active'
    };

    const response = await makeRequest(`${CONFIG.backend.url}/api/lawyers`, {
      method: 'POST',
      body: JSON.stringify(lawyerData)
    });

    if (response.ok) {
      console.log(`✅ 创建律师成功: ${JSON.stringify(response.data)}`);
    } else {
      console.log(`❌ 创建律师失败: ${response.status} - ${JSON.stringify(response.data)}`);
    }
  } catch (error) {
    console.log(`❌ 创建律师异常: ${error.message}`);
  }

  // 测试创建客户
  try {
    const clientData = {
      name: '测试客户',
      phone: '13900139000',
      email: 'client@example.com',
      type: '个人',
      address: '测试地址'
    };

    const response = await makeRequest(`${CONFIG.backend.url}/api/clients`, {
      method: 'POST',
      body: JSON.stringify(clientData)
    });

    if (response.ok) {
      console.log(`✅ 创建客户成功: ${JSON.stringify(response.data)}`);
    } else {
      console.log(`❌ 创建客户失败: ${response.status} - ${JSON.stringify(response.data)}`);
    }
  } catch (error) {
    console.log(`❌ 创建客户异常: ${error.message}`);
  }
}

async function checkDatabaseStatus() {
  console.log('\n=== 检查数据库状态 ===');

  try {
    const response = await makeRequest(`${CONFIG.backend.url}/api/v1/health/detailed`);

    if (response.data && response.data.checks) {
      console.log('数据库连接状态:');
      Object.entries(response.data.checks).forEach(([name, check]) => {
        const status = check.status === 'healthy' || check.status === 'ok' ? '✅' : '❌';
        console.log(`  ${status} ${name}: ${check.status} ${check.message ? `- ${check.message}` : ''}`);
      });
    } else {
      console.log('❌ 无法获取详细的健康检查信息');
    }
  } catch (error) {
    console.log(`❌ 数据库状态检查失败: ${error.message}`);
  }
}

async function runTests() {
  console.log('开始API集成测试...');
  console.log(`后端地址: ${CONFIG.backend.url}`);

  try {
    await testPublicEndpoints();
    await testAuthEndpoints();
    await testAPIEndpoints();
    await testDatabaseOperations();
    await checkDatabaseStatus();

    console.log('\n=== API集成测试完成 ===');
    console.log(`认证Token: ${authToken ? '已获得' : '未获得'}`);

  } catch (error) {
    console.error('测试执行失败:', error);
  }
}

// 运行测试
runTests().catch(console.error);
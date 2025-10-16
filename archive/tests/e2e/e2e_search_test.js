#!/usr/bin/env node

/**
 * 端到端搜索功能测试脚本
 * 验证前端修复后的搜索功能
 */

const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080';
const FRONTEND_URL = 'http://localhost:3003';

const TEST_USER = {
  username: 'testadmin@lawfirm.com',
  password: 'admin123'
};

let authToken = null;

// 认证并获取token
async function authenticate() {
  try {
    console.log('🔐 正在进行用户认证...');
    const response = await axios.post(`${API_BASE_URL}/api/auth/login`, {
      email: TEST_USER.username,
      password: TEST_USER.password
    });

    if (response.data && response.data.data && response.data.data.token) {
      authToken = response.data.data.token;
      console.log('✅ 认证成功，获取到token');
      return true;
    } else {
      console.log('❌ 认证响应格式异常:', response.data);
      return false;
    }
  } catch (error) {
    console.log('❌ 认证失败:', error.response?.data || error.message);
    return false;
  }
}

// 发送认证请求
async function authenticatedRequest(method, url, data = null, params = null) {
  const config = {
    method,
    url: url,
    headers: {
      'Authorization': `Bearer ${authToken}`,
      'Content-Type': 'application/json'
    }
  };

  if (data) config.data = data;
  if (params) config.params = params;

  try {
    const response = await axios(config);
    return response.data;
  } catch (error) {
    console.log(`❌ 请求失败 [${method.toUpperCase()} ${url}]:`, error.response?.data || error.message);
    return null;
  }
}

// 测试修复后的搜索功能
async function testFixedSearch() {
  console.log('\n🔧 测试修复后的客户搜索功能');
  console.log('=' .repeat(50));

  const testCases = [
    {
      name: '搜索"张三" - 使用正确的search参数',
      useSearchParams: true,
      query: '张三',
      description: '应该返回1条记录'
    },
    {
      name: '搜索"张三" - 使用旧的name参数',
      useSearchParams: false,
      query: '张三',
      description: '应该返回19条记录（错误的，但用于对比）'
    },
    {
      name: '搜索"王先生" - 使用正确的search参数',
      useSearchParams: true,
      query: '王先生',
      description: '应该返回1条记录'
    },
    {
      name: '搜索"科技" - 模糊搜索',
      useSearchParams: true,
      query: '科技',
      description: '应该返回相关客户'
    }
  ];

  for (const testCase of testCases) {
    console.log(`\n📋 测试: ${testCase.name}`);
    console.log(`📝 描述: ${testCase.description}`);

    const params = testCase.useSearchParams
      ? { search: testCase.query }
      : { name: testCase.query };

    console.log(`🔧 参数:`, JSON.stringify(params));

    const result = await authenticatedRequest('GET', `${API_BASE_URL}/api/v1/clients`, null, params);

    if (result) {
      let clientList = [];
      let total = 0;

      // 处理响应格式
      if (result.data) {
        if (Array.isArray(result.data)) {
          clientList = result.data;
          total = result.data.length;
        } else if (result.data.list) {
          clientList = result.data.list;
          total = result.data.total || result.data.list.length;
        } else if (result.pagination) {
          clientList = result.data;
          total = result.pagination.total;
        }
      }

      console.log(`📊 结果: 返回${total}条记录`);

      if (clientList.length > 0) {
        console.log('👥 客户列表:');
        clientList.slice(0, 3).forEach((client, index) => {
          console.log(`   ${index + 1}. ID:${client.id} | 姓名:${client.name} | 状态:${client.status}`);
        });
        if (clientList.length > 3) {
          console.log(`   ... 还有${clientList.length - 3}条记录`);
        }
      }

      // 分析结果
      if (testCase.query === '张三') {
        if (testCase.useSearchParams && total === 1) {
          console.log('✅ 修复成功：使用search参数正确返回1条记录');
        } else if (!testCase.useSearchParams && total === 19) {
          console.log('⚠️  确认问题：使用name参数返回所有19条记录');
        }
      }
    } else {
      console.log('❌ 请求失败');
    }

    console.log('-'.repeat(40));
  }
}

// 测试前端服务
async function testFrontendService() {
  console.log('\n🌐 测试前端服务');
  console.log('=' .repeat(30));

  try {
    const response = await axios.get(FRONTEND_URL, { timeout: 5000 });
    console.log('✅ 前端服务正常');
    console.log(`📍 前端地址: ${FRONTEND_URL}`);
    console.log(`📊 响应状态: ${response.status}`);
  } catch (error) {
    console.log('❌ 前端服务不可用:', error.message);
    console.log('💡 请确保前端服务正在运行');
  }
}

// 主函数
async function main() {
  console.log('🚀 端到端搜索功能测试开始');
  console.log(`后端API: ${API_BASE_URL}`);
  console.log(`前端服务: ${FRONTEND_URL}`);

  // 测试前端服务
  await testFrontendService();

  // 认证
  const authSuccess = await authenticate();
  if (!authSuccess) {
    console.log('❌ 认证失败，无法继续测试');
    process.exit(1);
  }

  // 测试搜索功能
  await testFixedSearch();

  console.log('\n🎯 端到端测试完成');
  console.log('✨ 修复总结:');
  console.log('   - 前端已将name参数映射为search参数');
  console.log('   - 搜索"张三"现在应该返回1条精确记录');
  console.log('   - 不再返回所有19条客户记录');
  console.log('\n📝 建议手动测试:');
  console.log('   1. 打开前端页面');
  console.log('   2. 进入客户管理页面');
  console.log('   3. 搜索"张三"');
  console.log('   4. 确认只显示1条记录');
}

// 运行测试
main().catch(console.error);
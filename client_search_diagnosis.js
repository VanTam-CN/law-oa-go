#!/usr/bin/env node

/**
 * 客户搜索功能诊断脚本
 * 用于测试和验证客户搜索API的各种参数组合
 */

const axios = require('axios');

// 配置
const API_BASE_URL = 'http://localhost:8080';
const TEST_USER = {
  username: 'testadmin@lawfirm.com',
  password: 'admin123'
};

// 存储认证token
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
    url: `${API_BASE_URL}${url}`,
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

// 测试客户搜索功能
async function testClientSearch() {
  console.log('\n🔍 开始测试客户搜索功能');
  console.log('=' .repeat(50));

  // 测试场景列表
  const testScenarios = [
    {
      name: '搜索"张三" - 精确匹配期望',
      params: { search: '张三' },
      description: '应该只返回1条记录（张三）'
    },
    {
      name: '使用name参数搜索"张三" - 前端当前方式',
      params: { name: '张三' },
      description: '测试前端当前使用的name参数是否有效'
    },
    {
      name: '搜索"张" - 模糊匹配',
      params: { search: '张' },
      description: '应该返回包含"张"字的客户'
    },
    {
      name: '无参数搜索 - 获取所有客户',
      params: {},
      description: '应该返回所有客户（用于对比）'
    },
    {
      name: '空搜索参数',
      params: { search: '' },
      description: '搜索空字符串的处理'
    },
    {
      name: '不存在的客户搜索',
      params: { search: '不存在的客户' },
      description: '应该返回0条记录'
    }
  ];

  for (const scenario of testScenarios) {
    console.log(`\n📋 测试场景: ${scenario.name}`);
    console.log(`📝 描述: ${scenario.description}`);
    console.log(`🔧 参数:`, JSON.stringify(scenario.params));

    const result = await authenticatedRequest('GET', '/api/v1/clients', null, scenario.params);

    if (result) {
      console.log('📊 响应结果:');
      console.log(`   - 成功状态: ${result.success ? '✅' : '❌'}`);

      let clientList = [];
      let total = 0;

      // 处理不同的响应格式
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
      } else if (result.list) {
        clientList = result.list;
        total = result.total || result.list.length;
      }

      console.log(`   - 返回记录数: ${total}`);
      console.log(`   - 客户列表:`);

      if (clientList.length > 0) {
        clientList.forEach((client, index) => {
          console.log(`     ${index + 1}. ID:${client.id} | 姓名:${client.name} | 类型:${client.type} | 状态:${client.status}`);
        });
      } else {
        console.log('     (无记录)');
      }

      // 分析结果
      if (scenario.params.search === '张三') {
        if (total === 1 && clientList[0]?.name === '张三') {
          console.log('✅ 测试通过：精确匹配"张三"返回1条记录');
        } else {
          console.log('❌ 测试失败：期望1条"张三"记录，实际返回' + total + '条');
        }
      }
    } else {
      console.log('❌ 请求失败');
    }

    console.log('-'.repeat(40));
  }
}

// 测试客户统计功能
async function testClientStats() {
  console.log('\n📊 测试客户统计功能');
  console.log('=' .repeat(30));

  const result = await authenticatedRequest('GET', '/api/v1/clients/stats');

  if (result) {
    console.log('📈 统计结果:');
    console.log(JSON.stringify(result, null, 2));
  } else {
    console.log('❌ 获取统计数据失败');
  }
}

// 主函数
async function main() {
  console.log('🚀 客户搜索功能诊断开始');
  console.log('API地址:', API_BASE_URL);

  // 认证
  const authSuccess = await authenticate();
  if (!authSuccess) {
    console.log('❌ 认证失败，无法继续测试');
    process.exit(1);
  }

  // 测试搜索功能
  await testClientSearch();

  // 测试统计功能
  await testClientStats();

  console.log('\n🎯 诊断完成');
  console.log('请根据以上结果分析问题根源');
}

// 运行诊断
main().catch(console.error);
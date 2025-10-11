const axios = require('axios');

// 基础配置
const BACKEND_URL = 'http://localhost:8080/api';
const FRONTEND_URL = 'http://localhost:3003';

console.log('🔍 开始验证前后端服务连接...\n');

// 验证后端服务
async function verifyBackend() {
  try {
    console.log('📡 测试后端API连接...');
    const response = await axios.post(`${BACKEND_URL}/auth/login`, {
      email: 'admin@example.com',
      password: 'admin123'
    }, { timeout: 5000 });

    if (response.data && response.data.success) {
      console.log('✅ 后端服务连接成功');
      console.log('   - API地址:', BACKEND_URL);
      console.log('   - 登录状态:', response.data.success ? '正常' : '异常');
      return true;
    } else {
      console.log('❌ 后端服务响应异常');
      return false;
    }
  } catch (error) {
    console.log('❌ 后端服务连接失败:', error.message);
    console.log('   - 错误类型:', error.code || '未知');
    console.log('   - 请检查后端服务是否正常运行');
    return false;
  }
}

// 验证前端服务
async function verifyFrontend() {
  try {
    console.log('\n🌐 测试前端服务连接...');
    const response = await axios.get(FRONTEND_URL, { timeout: 5000 });

    if (response.status === 200) {
      console.log('✅ 前端服务连接成功');
      console.log('   - 前端地址:', FRONTEND_URL);
      console.log('   - 服务状态: 正常运行');
      return true;
    } else {
      console.log('❌ 前端服务响应异常');
      return false;
    }
  } catch (error) {
    console.log('❌ 前端服务连接失败:', error.message);
    console.log('   - 错误类型:', error.code || '未知');
    console.log('   - 请检查前端服务是否正常运行');
    return false;
  }
}

// 验证律师管理API
async function verifyLawyerAPI() {
  try {
    console.log('\n⚖️ 测试律师管理API...');

    // 先登录获取token
    const loginResponse = await axios.post(`${BACKEND_URL}/auth/login`, {
      email: 'admin@example.com',
      password: 'admin123'
    });

    if (loginResponse.data && loginResponse.data.success) {
      const token = loginResponse.data.data.token;

      // 测试用户列表API（包含律师信息）
      const usersResponse = await axios.get(`${BACKEND_URL}/users`, {
        headers: {
          'Authorization': `Bearer ${token}`
        },
        timeout: 5000
      });

      if (usersResponse.data && usersResponse.data.data) {
        const users = usersResponse.data.data;
        const lawyers = users.filter(user => user.role === 'lawyer');

        console.log('✅ 律师管理API连接成功');
        console.log('   - 总用户数:', users.length);
        console.log('   - 律师数量:', lawyers.length);

        if (lawyers.length > 0) {
          console.log('   - 律师列表:');
          lawyers.forEach((lawyer, index) => {
            console.log(`     ${index + 1}. ${lawyer.name} (${lawyer.email})`);
          });
        }
        return true;
      }
    }

    console.log('❌ 律师管理API获取数据失败');
    return false;
  } catch (error) {
    console.log('❌ 律师管理API连接失败:', error.message);
    return false;
  }
}

// 验证案件管理API
async function verifyCaseAPI() {
  try {
    console.log('\n📋 测试案件管理API...');

    // 先登录获取token
    const loginResponse = await axios.post(`${BACKEND_URL}/auth/login`, {
      email: 'admin@example.com',
      password: 'admin123'
    });

    if (loginResponse.data && loginResponse.data.success) {
      const token = loginResponse.data.data.token;

      // 测试案件列表API
      const casesResponse = await axios.get(`${BACKEND_URL}/cases`, {
        headers: {
          'Authorization': `Bearer ${token}`
        },
        timeout: 5000
      });

      if (casesResponse.data && casesResponse.data.data) {
        const cases = casesResponse.data.data;

        console.log('✅ 案件管理API连接成功');
        console.log('   - 案件数量:', cases.length);

        if (cases.length > 0) {
          console.log('   - 案件列表:');
          cases.slice(0, 5).forEach((case_item, index) => {
            console.log(`     ${index + 1}. ${case_item.title} (状态: ${case_item.status})`);
          });
        }
        return true;
      }
    }

    console.log('❌ 案件管理API获取数据失败');
    return false;
  } catch (error) {
    console.log('❌ 案件管理API连接失败:', error.message);
    return false;
  }
}

// 主要验证流程
async function runVerification() {
  console.log('🚀 开始前后端服务验证...\n');

  const results = {
    backend: false,
    frontend: false,
    lawyerAPI: false,
    caseAPI: false
  };

  // 验证各项服务
  results.backend = await verifyBackend();
  results.frontend = await verifyFrontend();

  if (results.backend) {
    results.lawyerAPI = await verifyLawyerAPI();
    results.caseAPI = await verifyCaseAPI();
  }

  // 输出验证结果
  console.log('\n📊 验证结果汇总:');
  console.log('─'.repeat(40));
  console.log(`后端服务: ${results.backend ? '✅ 正常' : '❌ 异常'}`);
  console.log(`前端服务: ${results.frontend ? '✅ 正常' : '❌ 异常'}`);
  console.log(`律师API:  ${results.lawyerAPI ? '✅ 正常' : '❌ 异常'}`);
  console.log(`案件API:  ${results.caseAPI ? '✅ 正常' : '❌ 异常'}`);
  console.log('─'.repeat(40));

  // 访问指南
  console.log('\n🌐 访问指南:');
  console.log('─'.repeat(40));
  if (results.frontend) {
    console.log(`前端应用: ${FRONTEND_URL}`);
  }
  if (results.backend) {
    console.log(`后端API: ${BACKEND_URL}`);
  }
  console.log('─'.repeat(40));

  // 功能验证建议
  console.log('\n📋 功能验证建议:');
  console.log('─'.repeat(40));
  if (results.frontend && results.backend) {
    console.log('✅ 前后端服务均正常，可以进行功能验证:');
    console.log('   1. 访问前端应用登录系统');
    console.log('   2. 测试律师管理功能');
    console.log('   3. 测试案件管理功能');
    console.log('   4. 测试数据编辑和保存');
  } else {
    console.log('⚠️ 服务异常，请检查后进行修复');
    if (!results.backend) {
      console.log('   - 启动后端服务: ./law-oa-server');
    }
    if (!results.frontend) {
      console.log('   - 启动前端服务: cd frontend && npm run dev');
    }
  }
  console.log('─'.repeat(40));
}

// 执行验证
runVerification().catch(console.error);
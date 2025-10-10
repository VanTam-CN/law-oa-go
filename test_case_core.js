const axios = require('axios');

// 配置
const BASE_URL = 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api`;

// 存储认证信息
let authToken = '';
let testUserId = null;
let testClientId = null;
let testLawyerId = null;

// 颜色输出
const colors = {
  reset: '\x1b[0m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m'
};

function log(level, message) {
  const timestamp = new Date().toISOString();
  const color = colors[level] || colors.reset;
  console.log(`${color}[${timestamp}] [${level.toUpperCase()}]${colors.reset} ${message}`);
}

function success(message) {
  log('green', message);
}

function error(message) {
  log('red', message);
}

function info(message) {
  log('blue', message);
}

function warn(message) {
  log('yellow', message);
}

// HTTP 客户端
const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// 添加请求拦截器
api.interceptors.request.use((config) => {
  if (authToken) {
    config.headers.Authorization = `Bearer ${authToken}`;
  }
  return config;
});

// 添加响应拦截器
api.interceptors.response.use(
  (response) => response,
  (err) => {
    console.error('API请求失败:', err.response?.data || err.message);
    throw err;
  }
);

// 1. 用户登录
async function login() {
  try {
    info('正在登录...');

    const loginResponse = await api.post('/auth/login', {
      email: 'admin@example.com',
      password: 'admin123'
    });

    if (loginResponse.data && loginResponse.data.data && loginResponse.data.data.token) {
      authToken = loginResponse.data.data.token;
      testUserId = loginResponse.data.data.user.id;
      success(`登录成功，用户ID: ${testUserId}`);
      return true;
    } else {
      error('登录响应格式错误');
      return false;
    }
  } catch (error) {
    error(`登录失败: ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 2. 获取现有客户
async function getTestClient() {
  try {
    info('正在获取测试客户...');

    const clientsResponse = await api.get('/clients');

    if (clientsResponse.data && clientsResponse.data.data && clientsResponse.data.data.length > 0) {
      testClientId = clientsResponse.data.data[0].id;
      success(`使用现有客户，ID: ${testClientId}`);
      return true;
    } else {
      error('没有找到客户');
      return false;
    }
  } catch (error) {
    error(`获取客户失败: ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 3. 获取律师信息
async function getLawyerInfo() {
  try {
    info('正在获取律师信息...');

    const lawyersResponse = await api.get('/lawfirm/lawyers');

    if (lawyersResponse.data && lawyersResponse.data.data && lawyersResponse.data.data.length > 0) {
      testLawyerId = lawyersResponse.data.data[0].id;
      success(`获取律师信息成功，ID: ${testLawyerId}`);
      return true;
    } else {
      error('没有找到律师信息');
      return false;
    }
  } catch (error) {
    error(`获取律师信息失败: ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 4. 创建案件
async function createCase() {
  try {
    info('正在创建案件...');

    const caseData = {
      title: `测试案件 - ${new Date().toISOString().slice(0, 19)}`,
      description: '这是一个测试案件，用于验证新建案件功能是否正常工作。',
      client_id: testClientId,
      lawyer_id: testLawyerId,
      case_type: 'civil',
      priority: 'medium',
      skip_conflict_check: true // 跳过冲突检查以简化测试
    };

    const caseResponse = await api.post('/cases', caseData);

    if (caseResponse.data && caseResponse.data.data) {
      const caseId = caseResponse.data.data.id;
      success(`案件创建成功，ID: ${caseId}`);
      return caseId;
    } else {
      error('案件创建响应格式错误');
      return null;
    }
  } catch (error) {
    error(`创建案件失败: ${error.response?.data?.message || error.message}`);
    return null;
  }
}

// 5. 获取案件详情
async function getCaseDetails(caseId) {
  try {
    info(`正在获取案件详情，ID: ${caseId}...`);

    const caseResponse = await api.get(`/cases/${caseId}`);

    if (caseResponse.data && caseResponse.data.data) {
      const caseData = caseResponse.data.data;
      success('案件详情获取成功');
      info(`案件标题: ${caseData.title}`);
      info(`案件类型: ${caseData.case_type}`);
      info(`案件状态: ${caseData.status}`);
      info(`客户名称: ${caseData.client_name || '未知'}`);
      info(`律师名称: ${caseData.lawyer_name || '未知'}`);
      return true;
    } else {
      error('案件详情响应格式错误');
      return false;
    }
  } catch (error) {
    error(`获取案件详情失败: ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 6. 获取案件列表
async function getCaseList() {
  try {
    info('正在获取案件列表...');

    const casesResponse = await api.get('/cases?limit=5');

    if (casesResponse.data && casesResponse.data.data) {
      const cases = casesResponse.data.data;
      success(`案件列表获取成功，共 ${cases.length} 个案件`);

      cases.forEach((caseItem, index) => {
        info(`案件 ${index + 1}: ${caseItem.title} (${caseItem.status})`);
      });

      return true;
    } else {
      error('案件列表响应格式错误');
      return false;
    }
  } catch (error) {
    error(`获取案件列表失败: ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 7. 更新案件
async function updateCase(caseId) {
  try {
    info('正在更新案件...');

    const updateData = {
      title: `更新后的测试案件标题 - ${new Date().toISOString().slice(0, 19)}`,
      description: '这是更新后的案件描述',
      priority: 'high'
    };

    const updateResponse = await api.put(`/cases/${caseId}`, updateData);

    if (updateResponse.data && updateResponse.data.data) {
      success('案件更新成功');
      return true;
    } else {
      error('案件更新响应格式错误');
      return false;
    }
  } catch (error) {
    error(`更新案件失败: ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 8. 更新案件状态
async function updateCaseStatus(caseId) {
  try {
    info('正在更新案件状态...');

    const statusData = {
      status: 'active'
    };

    const statusResponse = await api.put(`/cases/${caseId}/status`, statusData);

    if (statusResponse.data) {
      success('案件状态更新成功');
      return true;
    } else {
      error('案件状态更新响应格式错误');
      return false;
    }
  } catch (error) {
    error(`更新案件状态失败: ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 9. 删除案件
async function deleteCase(caseId) {
  try {
    info('正在删除案件...');

    const deleteResponse = await api.delete(`/cases/${caseId}`);

    if (deleteResponse.data) {
      success('案件删除成功');
      return true;
    } else {
      error('案件删除响应格式错误');
      return false;
    }
  } catch (error) {
    error(`删除案件失败: ${error.response?.data?.message || error.message}`);
    return false;
  }
}

// 主测试函数
async function runCoreTests() {
  console.log('='.repeat(60));
  console.log('🚀 开始新建案件核心功能测试');
  console.log('='.repeat(60));

  let caseId = null;
  let testResults = {
    login: false,
    getClient: false,
    getLawyer: false,
    createCase: false,
    getCaseDetails: false,
    getCaseList: false,
    updateCase: false,
    updateStatus: false,
    deleteCase: false
  };

  try {
    // 1. 登录
    testResults.login = await login();
    if (!testResults.login) {
      error('登录失败，无法继续测试');
      return;
    }

    // 2. 获取测试客户
    testResults.getClient = await getTestClient();
    if (!testResults.getClient) {
      error('获取客户失败，无法继续测试');
      return;
    }

    // 3. 获取律师信息
    testResults.getLawyer = await getLawyerInfo();
    if (!testResults.getLawyer) {
      error('获取律师信息失败，无法继续测试');
      return;
    }

    // 4. 创建案件
    caseId = await createCase();
    testResults.createCase = caseId !== null;

    if (testResults.createCase) {
      // 5. 获取案件详情
      testResults.getCaseDetails = await getCaseDetails(caseId);

      // 6. 获取案件列表
      testResults.getCaseList = await getCaseList();

      // 7. 更新案件
      testResults.updateCase = await updateCase(caseId);

      // 8. 更新案件状态
      testResults.updateStatus = await updateCaseStatus(caseId);

      // 9. 删除案件
      testResults.deleteCase = await deleteCase(caseId);
    }

  } catch (error) {
    error(`测试过程中发生错误: ${error.message}`);
  }

  // 输出测试结果
  console.log('\n' + '='.repeat(60));
  console.log('📊 核心功能测试结果汇总');
  console.log('='.repeat(60));

  let passedTests = 0;
  let totalTests = Object.keys(testResults).length;

  Object.entries(testResults).forEach(([testName, passed]) => {
    const status = passed ? '✅ 通过' : '❌ 失败';
    const color = passed ? colors.green : colors.red;
    console.log(`${color}${testName}: ${status}${colors.reset}`);
    if (passed) passedTests++;
  });

  console.log('\n' + '='.repeat(60));
  const successRate = ((passedTests / totalTests) * 100).toFixed(1);
  const overallColor = passedTests === totalTests ? colors.green : colors.yellow;
  console.log(`${overallColor}总体结果: ${passedTests}/${totalTests} 测试通过 (${successRate}%)${colors.reset}`);

  if (passedTests === totalTests) {
    success('🎉 所有核心功能测试通过！新建案件功能正常工作');
  } else {
    warn('⚠️  部分核心功能测试失败，请检查相关功能');
  }

  console.log('='.repeat(60));
}

// 运行测试
if (require.main === module) {
  runCoreTests().catch(err => {
    console.error(`测试运行失败: ${err.message}`);
    process.exit(1);
  });
}

module.exports = {
  runCoreTests,
  login,
  getTestClient,
  getLawyerInfo,
  createCase,
  getCaseDetails,
  getCaseList,
  updateCase,
  updateCaseStatus,
  deleteCase
};
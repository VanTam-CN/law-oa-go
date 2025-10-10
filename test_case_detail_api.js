const axios = require('axios');

// 基础配置
const BASE_URL = 'http://localhost:8080/api';
let authToken = '';

// 测试用户登录
async function login() {
  try {
    const response = await axios.post(`${BASE_URL}/auth/login`, {
      username: 'admin',
      password: 'admin123'
    });

    authToken = response.data.data.token;
    console.log('✅ 登录成功，Token已获取');
    return true;
  } catch (error) {
    console.error('❌ 登录失败:', error.response?.data || error.message);
    return false;
  }
}

// 测试获取案件列表
async function testCaseList() {
  try {
    const response = await axios.get(`${BASE_URL}/cases`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    const cases = response.data.data;
    console.log(`✅ 获取案件列表成功，共 ${cases.length} 个案件`);

    if (cases.length > 0) {
      return cases[0].id; // 返回第一个案件ID用于详情测试
    }
    return null;
  } catch (error) {
    console.error('❌ 获取案件列表失败:', error.response?.data || error.message);
    return null;
  }
}

// 测试获取案件详情
async function testCaseDetail(caseId) {
  try {
    const response = await axios.get(`${BASE_URL}/cases/${caseId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    const caseDetail = response.data.data;
    console.log('✅ 获取案件详情成功');
    console.log('案件信息:', {
      id: caseDetail.id,
      title: caseDetail.title,
      case_type: caseDetail.case_type,
      status: caseDetail.status,
      priority: caseDetail.priority,
      client: caseDetail.client,
      lawyer: caseDetail.lawyer
    });
    return caseDetail;
  } catch (error) {
    console.error('❌ 获取案件详情失败:', error.response?.data || error.message);
    return null;
  }
}

// 测试更新案件
async function testCaseUpdate(caseId) {
  try {
    const updateData = {
      title: '测试更新案件标题 ' + new Date().toISOString(),
      description: '测试更新案件描述 ' + new Date().toISOString()
    };

    const response = await axios.put(`${BASE_URL}/cases/${caseId}`, updateData, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    const updatedCase = response.data.data;
    console.log('✅ 更新案件成功');
    console.log('更新后的案件信息:', {
      title: updatedCase.title,
      description: updatedCase.description,
      updated_at: updatedCase.updated_at
    });
    return true;
  } catch (error) {
    console.error('❌ 更新案件失败:', error.response?.data || error.message);
    return false;
  }
}

// 主要测试流程
async function runTests() {
  console.log('🚀 开始案件详情功能测试...\n');

  // 1. 登录测试
  console.log('1. 测试用户登录...');
  const loginSuccess = await login();
  if (!loginSuccess) {
    console.log('❌ 登录失败，无法继续测试');
    return;
  }
  console.log('');

  // 2. 获取案件列表
  console.log('2. 测试获取案件列表...');
  const caseId = await testCaseList();
  if (!caseId) {
    console.log('❌ 无法获取案件列表，无法继续测试');
    return;
  }
  console.log('');

  // 3. 获取案件详情
  console.log('3. 测试获取案件详情...');
  const caseDetail = await testCaseDetail(caseId);
  if (!caseDetail) {
    console.log('❌ 无法获取案件详情，无法继续测试');
    return;
  }
  console.log('');

  // 4. 测试案件更新
  console.log('4. 测试案件更新...');
  const updateSuccess = await testCaseUpdate(caseId);
  if (!updateSuccess) {
    console.log('❌ 案件更新测试失败');
    return;
  }
  console.log('');

  // 5. 再次获取详情验证更新
  console.log('5. 验证案件更新结果...');
  const updatedCaseDetail = await testCaseDetail(caseId);
  if (updatedCaseDetail) {
    console.log('✅ 案件详情功能测试全部通过！');
  } else {
    console.log('❌ 验证更新结果失败');
  }
}

// 执行测试
runTests().catch(console.error);
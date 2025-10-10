const axios = require('axios');

// 基础配置
const BASE_URL = 'http://localhost:8080/api';
let authToken = '';

// 测试用户登录
async function login() {
  try {
    const response = await axios.post(`${BASE_URL}/auth/login`, {
      email: 'admin@example.com',
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

// 测试案件详情的完整响应
async function testCaseDetail() {
  try {
    // 获取第一个案件
    const listResponse = await axios.get(`${BASE_URL}/cases`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    if (listResponse.data.data.length === 0) {
      console.log('❌ 没有找到案件');
      return;
    }

    const caseId = listResponse.data.data[0].id;
    console.log(`\n📋 测试案件ID: ${caseId}`);

    // 获取案件详情
    const detailResponse = await axios.get(`${BASE_URL}/cases/${caseId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    console.log('\n🔍 完整的API响应结构:');
    console.log(JSON.stringify(detailResponse.data, null, 2));

    const caseData = detailResponse.data.data;
    console.log('\n📊 案件数据详细分析:');
    console.log('案件ID:', caseData.id);
    console.log('案件标题:', caseData.title);
    console.log('客户ID:', caseData.client_id);
    console.log('律师ID:', caseData.lawyer_id);
    console.log('客户对象:', caseData.client);
    console.log('律师对象:', caseData.lawyer);
    console.log('客户名称:', caseData.client_name);
    console.log('律师名称:', caseData.lawyer_name);

    // 检查客户信息
    if (caseData.client) {
      console.log('\n👤 客户详细信息:');
      console.log('客户ID:', caseData.client.id);
      console.log('客户姓名:', caseData.client.name);
      console.log('客户公司:', caseData.client.company);
      console.log('客户邮箱:', caseData.client.email);
      console.log('客户电话:', caseData.client.phone);
    } else {
      console.log('\n❌ 客户对象为空');
    }

    // 检查律师信息
    if (caseData.lawyer) {
      console.log('\n👨‍⚖️ 律师详细信息:');
      console.log('律师ID:', caseData.lawyer.id);
      console.log('律师姓名:', caseData.lawyer.name);
      console.log('律师邮箱:', caseData.lawyer.email);
      console.log('律师角色:', caseData.lawyer.role);
    } else {
      console.log('\n❌ 律师对象为空');
    }

  } catch (error) {
    console.error('❌ 测试失败:', error.response?.data || error.message);
  }
}

// 主测试流程
async function runDebugTest() {
  console.log('🚀 开始调试案件详情API响应...\n');

  // 1. 登录
  console.log('1. 登录...');
  const loginSuccess = await login();
  if (!loginSuccess) {
    console.log('❌ 登录失败，无法继续测试');
    return;
  }

  // 2. 测试案件详情
  console.log('\n2. 测试案件详情...');
  await testCaseDetail();
}

// 执行测试
runDebugTest().catch(console.error);
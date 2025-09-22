#!/usr/bin/env node

/**
 * 前端后端集成测试脚本
 * 用于验证frontend2是否正确适配了后端API
 */

const axios = require('axios');

const BASE_URL = 'http://localhost:8080/api/v1';

// 测试用的用户数据
const testUser = {
  email: 'test@example.com',
  password: 'password123',
  name: '测试用户'
};

const api = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  validateStatus: function (status) {
    return status < 500; // 接受所有小于500的状态码
  }
});

async function testAuthAPI() {
  console.log('🔐 测试认证API...');
  
  try {
    // 测试注册
    console.log('  测试注册...');
    const registerResponse = await api.post('/auth/register', testUser);
    console.log('  注册响应:', registerResponse.status, registerResponse.data);
    
    // 测试登录
    console.log('  测试登录...');
    const loginResponse = await api.post('/auth/login', {
      email: testUser.email,
      password: testUser.password
    });
    console.log('  登录响应:', loginResponse.status, loginResponse.data);
    
    if (loginResponse.data.success && loginResponse.data.data.token) {
      const token = loginResponse.data.data.token;
      
      // 设置token进行后续测试
      api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      
      // 测试获取用户信息
      console.log('  测试获取用户信息...');
      const profileResponse = await api.get('/users/profile');
      console.log('  用户信息响应:', profileResponse.status, profileResponse.data);
      
      return token;
    }
  } catch (error) {
    console.error('  认证测试失败:', error.response?.data || error.message);
    return null;
  }
  
  return null;
}

async function testUserAPI(token) {
  if (!token) {
    console.log('⚠️ 跳过用户API测试（无有效token）');
    return;
  }
  
  console.log('👤 测试用户管理API...');
  
  try {
    // 测试获取用户列表
    console.log('  测试获取用户列表...');
    const usersResponse = await api.get('/admin/users');
    console.log('  用户列表响应:', usersResponse.status, usersResponse.data);
    
    // 测试创建用户
    console.log('  测试创建用户...');
    const newUser = {
      name: '新用户',
      email: `newuser${Date.now()}@example.com`,
      password: 'password123',
      role: 'user'
    };
    const createResponse = await api.post('/admin/users', newUser);
    console.log('  创建用户响应:', createResponse.status, createResponse.data);
    
    // 如果创建成功，测试删除用户
    if (createResponse.data.success && createResponse.data.data.id) {
      const userId = createResponse.data.data.id;
      console.log('  测试删除用户...');
      const deleteResponse = await api.delete(`/admin/users/${userId}`);
      console.log('  删除用户响应:', deleteResponse.status, deleteResponse.data);
    }
  } catch (error) {
    console.error('  用户API测试失败:', error.response?.data || error.message);
  }
}

async function testClientAPI(token) {
  if (!token) {
    console.log('⚠️ 跳过客户API测试（无有效token）');
    return;
  }
  
  console.log('👥 测试客户管理API...');
  
  try {
    // 测试获取客户列表
    console.log('  测试获取客户列表...');
    const clientsResponse = await api.get('/clients');
    console.log('  客户列表响应:', clientsResponse.status, clientsResponse.data);
    
    // 测试创建客户
    console.log('  测试创建客户...');
    const newClient = {
      name: '测试客户',
      email: `client${Date.now()}@example.com`,
      phone: '13800138000',
      address: '测试地址',
      company: '测试公司'
    };
    const createResponse = await api.post('/clients', newClient);
    console.log('  创建客户响应:', createResponse.status, createResponse.data);
    
    // 如果创建成功，测试删除客户
    if (createResponse.data.success && createResponse.data.data.id) {
      const clientId = createResponse.data.data.id;
      console.log('  测试删除客户...');
      const deleteResponse = await api.delete(`/clients/${clientId}`);
      console.log('  删除客户响应:', deleteResponse.status, deleteResponse.data);
    }
    
    // 测试获取客户统计
    console.log('  测试获取客户统计...');
    const statsResponse = await api.get('/clients/stats');
    console.log('  客户统计响应:', statsResponse.status, statsResponse.data);
  } catch (error) {
    console.error('  客户API测试失败:', error.response?.data || error.message);
  }
}

async function testCaseAPI(token) {
  if (!token) {
    console.log('⚠️ 跳过案件API测试（无有效token）');
    return;
  }
  
  console.log('⚖️ 测试案件管理API...');
  
  try {
    // 测试获取案件列表
    console.log('  测试获取案件列表...');
    const casesResponse = await api.get('/cases');
    console.log('  案件列表响应:', casesResponse.status, casesResponse.data);
    
    // 测试创建案件
    console.log('  测试创建案件...');
    const newCase = {
      title: '测试案件',
      description: '这是一个测试案件',
      client_id: 1, // 假设存在客户ID为1
      case_type: 'civil',
      priority: 'medium',
      status: 'pending'
    };
    const createResponse = await api.post('/cases', newCase);
    console.log('  创建案件响应:', createResponse.status, createResponse.data);
    
    // 如果创建成功，测试删除案件
    if (createResponse.data.success && createResponse.data.data.id) {
      const caseId = createResponse.data.data.id;
      console.log('  测试删除案件...');
      const deleteResponse = await api.delete(`/cases/${caseId}`);
      console.log('  删除案件响应:', deleteResponse.status, deleteResponse.data);
    }
    
    // 测试获取案件统计
    console.log('  测试获取案件统计...');
    const statsResponse = await api.get('/cases/stats');
    console.log('  案件统计响应:', statsResponse.status, statsResponse.data);
  } catch (error) {
    console.error('  案件API测试失败:', error.response?.data || error.message);
  }
}

async function testAPI() {
  console.log('🚀 开始前端后端集成测试...\n');
  
  // 检查后端服务是否可用
  try {
    console.log('🔍 检查后端服务...');
    const healthResponse = await axios.get('http://localhost:8080/health');
    console.log('  后端服务状态:', healthResponse.status, healthResponse.data);
  } catch (error) {
    console.error('❌ 后端服务不可用，请确保后端服务运行在 http://localhost:8080');
    console.error('   错误:', error.message);
    return;
  }
  
  // 测试认证API
  const token = await testAuthAPI();
  console.log('');
  
  // 测试其他API
  await testUserAPI(token);
  console.log('');
  
  await testClientAPI(token);
  console.log('');
  
  await testCaseAPI(token);
  console.log('');
  
  console.log('✅ 集成测试完成！');
}

// 如果直接运行此脚本
if (require.main === module) {
  testAPI().catch(console.error);
}

module.exports = {
  testAuthAPI,
  testUserAPI,
  testClientAPI,
  testCaseAPI,
  testAPI
};
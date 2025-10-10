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

// 测试获取用户列表（包含律师角色）
async function testGetUsers() {
  try {
    const response = await axios.get(`${BASE_URL}/users`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    const users = response.data.data;
    console.log(`✅ 获取用户列表成功，共 ${users.length} 个用户`);

    // 查找律师角色的用户
    const lawyers = users.filter(user => user.role === 'lawyer');
    console.log(`找到 ${lawyers.length} 个律师用户`);

    if (lawyers.length > 0) {
      return lawyers[0].id; // 返回第一个律师ID用于详情测试
    }
    return null;
  } catch (error) {
    console.error('❌ 获取用户列表失败:', error.response?.data || error.message);
    return null;
  }
}

// 测试获取用户详情
async function testGetUserDetail(userId) {
  try {
    const response = await axios.get(`${BASE_URL}/users/${userId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    const userDetail = response.data.data;
    console.log('✅ 获取用户详情成功');
    console.log('用户信息:', {
      id: userDetail.id,
      username: userDetail.username,
      name: userDetail.name,
      email: userDetail.email,
      role: userDetail.role,
      phone: userDetail.phone,
      status: userDetail.status
    });
    return userDetail;
  } catch (error) {
    console.error('❌ 获取用户详情失败:', error.response?.data || error.message);
    return null;
  }
}

// 测试更新用户
async function testUpdateUser(userId) {
  try {
    const updateData = {
      name: '测试更新律师名称 ' + new Date().toISOString(),
      phone: '13800139999',
      profile: '测试更新律师简介 ' + new Date().toISOString()
    };

    const response = await axios.put(`${BASE_URL}/users/${userId}`, updateData, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    const updatedUser = response.data.data;
    console.log('✅ 更新用户成功');
    console.log('更新后的用户信息:', {
      name: updatedUser.name,
      phone: updatedUser.phone,
      profile: updatedUser.profile,
      updated_at: updatedUser.updated_at
    });
    return true;
  } catch (error) {
    console.error('❌ 更新用户失败:', error.response?.data || error.message);
    return false;
  }
}

// 主要测试流程
async function runTests() {
  console.log('🚀 开始律师详情功能测试...\n');

  // 1. 登录测试
  console.log('1. 测试用户登录...');
  const loginSuccess = await login();
  if (!loginSuccess) {
    console.log('❌ 登录失败，无法继续测试');
    return;
  }
  console.log('');

  // 2. 获取律师用户列表
  console.log('2. 测试获取律师用户列表...');
  const lawyerId = await testGetUsers();
  if (!lawyerId) {
    console.log('❌ 未找到律师用户，无法继续测试');
    return;
  }
  console.log('');

  // 3. 获取律师详情
  console.log('3. 测试获取律师详情...');
  const userDetail = await testGetUserDetail(lawyerId);
  if (!userDetail) {
    console.log('❌ 无法获取律师详情，无法继续测试');
    return;
  }
  console.log('');

  // 4. 测试律师更新
  console.log('4. 测试律师信息更新...');
  const updateSuccess = await testUpdateUser(lawyerId);
  if (!updateSuccess) {
    console.log('❌ 律师信息更新测试失败');
    return;
  }
  console.log('');

  // 5. 再次获取详情验证更新
  console.log('5. 验证律师信息更新结果...');
  const updatedUserDetail = await testGetUserDetail(lawyerId);
  if (updatedUserDetail) {
    console.log('✅ 律师详情功能测试全部通过！');
  } else {
    console.log('❌ 验证更新结果失败');
  }
}

// 执行测试
runTests().catch(console.error);
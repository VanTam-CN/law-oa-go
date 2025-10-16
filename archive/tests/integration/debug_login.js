const axios = require('axios');

async function debugLogin() {
  try {
    console.log('测试1: 简单登录请求');
    const response1 = await axios.post('http://localhost:8080/api/auth/login', {
      email: 'admin@lawfirm.com',
      password: 'Admin123!'
    });
    console.log('✅ 登录成功:', response1.data);
  } catch (error) {
    console.log('❌ 登录失败:', error.response?.status, error.response?.data?.message);
    console.log('详细错误:', error.response?.data);
  }

  try {
    console.log('\n测试2: 注册新用户');
    const response2 = await axios.post('http://localhost:8080/api/auth/register', {
      username: 'testadmin',
      name: 'Test Admin',
      email: 'testadmin@lawfirm.com',
      password: 'TestAdmin123!',
      role: 'admin'
    });
    console.log('✅ 注册成功:', response2.data);
  } catch (error) {
    console.log('❌ 注册失败:', error.response?.status, error.response?.data?.message);
    console.log('详细错误:', error.response?.data);
  }

  try {
    console.log('\n测试3: 使用新注册用户登录');
    const response3 = await axios.post('http://localhost:8080/api/auth/login', {
      email: 'testadmin@lawfirm.com',
      password: 'TestAdmin123!'
    });
    console.log('✅ 登录成功:', response3.data);
  } catch (error) {
    console.log('❌ 登录失败:', error.response?.status, error.response?.data?.message);
    console.log('详细错误:', error.response?.data);
  }
}

debugLogin();
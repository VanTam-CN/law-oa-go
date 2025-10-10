/**
 * 创建测试用户脚本
 */

const { default: fetch } = require('node-fetch');

const CONFIG = {
  backend: {
    url: 'http://localhost:8080'
  },
  testUser: {
    username: 'testuser',
    email: 'testuser2@example.com',
    password: 'Password123',
    name: '测试用户',
    phone: '13800138000',
    role: 'admin'
  }
};

async function createTestUser() {
  console.log('开始创建测试用户...');

  try {
    // 尝试注册用户
    const response = await fetch(`${CONFIG.backend.url}/api/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(CONFIG.testUser)
    });

    const data = await response.json();

    if (response.ok) {
      console.log('✅ 用户创建成功:', JSON.stringify(data, null, 2));
      return true;
    } else {
      console.log('❌ 用户创建失败:', response.status, JSON.stringify(data, null, 2));

      // 如果用户已存在，尝试登录
      if (response.status === 422 && data.error && data.error.message.includes('already exists')) {
        console.log('用户已存在，尝试登录...');
        return await testLogin();
      }

      return false;
    }
  } catch (error) {
    console.error('❌ 创建用户异常:', error.message);
    return false;
  }
}

async function testLogin() {
  try {
    const response = await fetch(`${CONFIG.backend.url}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: CONFIG.testUser.email,
        password: CONFIG.testUser.password
      })
    });

    const data = await response.json();

    if (response.ok) {
      console.log('✅ 用户登录成功:', JSON.stringify(data, null, 2));

      // 保存token到文件
      require('fs').writeFileSync('test-token.txt', data.token);
      console.log('✅ Token已保存到 test-token.txt');

      return true;
    } else {
      console.log('❌ 用户登录失败:', response.status, JSON.stringify(data, null, 2));
      return false;
    }
  } catch (error) {
    console.error('❌ 登录异常:', error.message);
    return false;
  }
}

// 运行脚本
createTestUser().then(success => {
  if (success) {
    console.log('\n测试用户准备完成！');
  } else {
    console.log('\n测试用户准备失败！');
  }
});
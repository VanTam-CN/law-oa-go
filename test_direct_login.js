const jwt = require('jsonwebtoken');

// 直接测试登录API，不使用错误处理中间件
async function testDirectLoginAPI() {
  console.log('=== 直接测试登录API ===');

  const loginData = {
    email: 'admin@law-oa.com',
    password: 'secret'
  };

  console.log('发送登录请求...');
  try {
    const response = await fetch('http://localhost:8080/api/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Debug-Mode': 'true'  // 添加调试头
      },
      body: JSON.stringify(loginData)
    });

    console.log('响应状态:', response.status);
    console.log('响应头:', Object.fromEntries(response.headers.entries()));

    if (response.status === 200) {
      const contentType = response.headers.get('content-type');
      console.log('Content-Type:', contentType);

      if (contentType && contentType.includes('application/json')) {
        const data = await response.json();
        console.log('登录成功，响应数据:', JSON.stringify(data, null, 2));
      } else {
        const text = await response.text();
        console.log('响应不是JSON格式:', text);
      }
    } else {
      const errorText = await response.text();
      console.log('登录失败，错误信息:', errorText);
    }
  } catch (error) {
    console.error('请求异常:', error.message);
  }
}

// 测试不带错误处理中间件的简单端点
async function testSimpleEndpoint() {
  console.log('\n=== 测试简单端点 ===');

  try {
    const response = await fetch('http://localhost:8080/health', {
      method: 'GET'
    });

    console.log('健康检查状态:', response.status);
    const data = await response.text();
    console.log('健康检查响应:', data);
  } catch (error) {
    console.error('健康检查异常:', error.message);
  }
}

// 测试用户服务认证方法
async function testUserServiceAuth() {
  console.log('\n=== 测试用户服务认证 ===');

  // 使用数据库直接查询用户
  const mysql = require('mysql2/promise');

  try {
    const connection = await mysql.createConnection({
      host: 'localhost',
      user: 'root',
      password: 'root',
      database: 'law_oa'
    });

    console.log('数据库连接成功');

    // 查询用户
    const [rows] = await connection.execute('SELECT * FROM users WHERE email = ?', ['admin@law-oa.com']);
    console.log('查询到的用户:', rows[0]);

    if (rows[0]) {
      const user = rows[0];
      console.log('用户ID:', user.id);
      console.log('用户邮箱:', user.email);
      console.log('用户角色:', user.role);
      console.log('密码哈希:', user.password);

      // 测试bcrypt验证
      const bcrypt = require('bcrypt');
      const isValid = await bcrypt.compare('secret', user.password);
      console.log('密码验证结果:', isValid);
    }

    await connection.end();
  } catch (error) {
    console.error('数据库操作异常:', error.message);
  }
}

// 依次执行测试
async function runTests() {
  await testSimpleEndpoint();
  await testDirectLoginAPI();
  await testUserServiceAuth();
}

runTests().catch(console.error);
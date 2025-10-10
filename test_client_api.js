const jwt = require('jsonwebtoken');

// JWT配置（需要与后端一致）
const JWT_SECRET = 'your-very-secure-jwt-secret-key-that-is-at-least-32-characters-long';
const TEST_USER_ID = 4; // admin用户ID

// 生成测试token
function generateTestToken() {
  const payload = {
    user_id: TEST_USER_ID,
    username: 'admin@law-oa.com',
    role: 'admin',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24小时
  };

  return jwt.sign(payload, JWT_SECRET);
}

// 测试客户API
async function testClientAPI() {
  const token = generateTestToken();
  console.log('Generated token:', token);

  // 测试客户列表API
  const clientListResponse = await fetch('http://localhost:8080/api/clients', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('Client list response status:', clientListResponse.status);
  if (clientListResponse.ok) {
    const clientData = await clientListResponse.json();
    console.log('Client list response:', JSON.stringify(clientData, null, 2));
  } else {
    const errorText = await clientListResponse.text();
    console.log('Client list error:', errorText);
  }

  // 测试客户统计API
  const statsResponse = await fetch('http://localhost:8080/api/clients/stats', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('Stats response status:', statsResponse.status);
  if (statsResponse.ok) {
    const statsData = await statsResponse.json();
    console.log('Stats response:', JSON.stringify(statsData, null, 2));
  } else {
    const errorText = await statsResponse.text();
    console.log('Stats error:', errorText);
  }
}

// 执行测试
testClientAPI().catch(console.error);
const jwt = require('jsonwebtoken');

// JWT配置（需要与后端一致）
const JWT_SECRET = 'your-very-secure-jwt-secret-key-that-is-at-least-32-characters-long'; // 需要从后端配置获取
const TEST_USER_ID = 4; // admin用户ID
const TEST_USERNAME = 'admin@law-oa.com';
const TEST_ROLE = 'admin';

// 生成测试token
function generateTestToken() {
  const payload = {
    user_id: TEST_USER_ID,
    username: TEST_USERNAME,
    role: TEST_ROLE,
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24小时
  };

  return jwt.sign(payload, JWT_SECRET);
}

// 测试案件API
async function testCaseAPI() {
  const token = generateTestToken();
  console.log('Generated token:', token);

  // 测试案件列表API
  const caseListResponse = await fetch('http://localhost:8080/api/cases', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('Case list response status:', caseListResponse.status);
  if (caseListResponse.ok) {
    const caseData = await caseListResponse.json();
    console.log('Case list response:', JSON.stringify(caseData, null, 2));
  } else {
    const errorText = await caseListResponse.text();
    console.log('Case list error:', errorText);
  }

  // 测试案件统计API
  const statsResponse = await fetch('http://localhost:8080/api/cases/stats', {
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
testCaseAPI().catch(console.error);
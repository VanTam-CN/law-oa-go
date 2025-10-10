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

// 测试创建案件API
async function testCreateCaseAPI() {
  const token = generateTestToken();
  console.log('Generated token:', token);

  // 测试创建案件
  const newCase = {
    title: '测试劳动合同纠纷案',
    description: '这是一个测试案件，用于验证案件创建功能',
    client_id: 1,
    lawyer_id: 1,
    case_type: 'civil',
    priority: 'high',
    status: 'active',
    start_date: '2025-09-28T10:00:00Z',
    expected_end_date: '2025-12-28T10:00:00Z'
  };

  const createResponse = await fetch('http://localhost:8080/api/v1/cases', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(newCase)
  });

  console.log('Create case response status:', createResponse.status);
  if (createResponse.ok) {
    const caseData = await createResponse.json();
    console.log('Create case response:', JSON.stringify(caseData, null, 2));

    // 获取创建的案件ID
    const caseId = caseData.data?.id || caseData.data?.caseId;
    if (caseId) {
      console.log('Created case ID:', caseId);

      // 测试获取案件详情
      await testGetCaseDetail(token, caseId);

      // 测试更新案件
      await testUpdateCase(token, caseId);

      // 测试删除案件
      await testDeleteCase(token, caseId);
    }
  } else {
    const errorText = await createResponse.text();
    console.log('Create case error:', errorText);
  }
}

// 测试获取案件详情
async function testGetCaseDetail(token, caseId) {
  const detailResponse = await fetch(`http://localhost:8080/api/v1/cases/${caseId}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('Get case detail response status:', detailResponse.status);
  if (detailResponse.ok) {
    const detailData = await detailResponse.json();
    console.log('Case detail response:', JSON.stringify(detailData, null, 2));
  } else {
    const errorText = await detailResponse.text();
    console.log('Get case detail error:', errorText);
  }
}

// 测试更新案件
async function testUpdateCase(token, caseId) {
  const updateData = {
    title: '更新后的测试案件',
    description: '这是一个更新后的测试案件',
    priority: 'medium',
    status: 'pending'
  };

  const updateResponse = await fetch(`http://localhost:8080/api/v1/cases/${caseId}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(updateData)
  });

  console.log('Update case response status:', updateResponse.status);
  if (updateResponse.ok) {
    const updateData = await updateResponse.json();
    console.log('Update case response:', JSON.stringify(updateData, null, 2));
  } else {
    const errorText = await updateResponse.text();
    console.log('Update case error:', errorText);
  }
}

// 测试删除案件
async function testDeleteCase(token, caseId) {
  const deleteResponse = await fetch(`http://localhost:8080/api/v1/cases/${caseId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('Delete case response status:', deleteResponse.status);
  if (deleteResponse.ok) {
    const deleteData = await deleteResponse.json();
    console.log('Delete case response:', JSON.stringify(deleteData, null, 2));
  } else {
    const errorText = await deleteResponse.text();
    console.log('Delete case error:', errorText);
  }
}

// 执行测试
testCreateCaseAPI().catch(console.error);
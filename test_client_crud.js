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

// 测试创建客户API
async function testCreateClientAPI() {
  const token = generateTestToken();
  console.log('Generated token:', token);

  // 测试创建客户
  const newClient = {
    name: '测试客户公司',
    email: 'test@example.com',
    phone: '13800138000',
    address: '北京市朝阳区测试地址',
    company: '测试客户有限公司',
    notes: '这是一个测试客户',
    status: 'active'
  };

  const createResponse = await fetch('http://localhost:8080/api/v1/clients', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(newClient)
  });

  console.log('Create client response status:', createResponse.status);
  if (createResponse.ok) {
    const clientData = await createResponse.json();
    console.log('Create client response:', JSON.stringify(clientData, null, 2));

    // 获取创建的客户ID
    const clientId = clientData.data?.id;
    if (clientId) {
      console.log('Created client ID:', clientId);

      // 测试获取客户详情
      await testGetClientDetail(token, clientId);

      // 测试更新客户
      await testUpdateClient(token, clientId);

      // 测试删除客户
      await testDeleteClient(token, clientId);
    }
  } else {
    const errorText = await createResponse.text();
    console.log('Create client error:', errorText);
  }
}

// 测试获取客户详情
async function testGetClientDetail(token, clientId) {
  const detailResponse = await fetch(`http://localhost:8080/api/v1/clients/${clientId}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('Get client detail response status:', detailResponse.status);
  if (detailResponse.ok) {
    const detailData = await detailResponse.json();
    console.log('Client detail response:', JSON.stringify(detailData, null, 2));
  } else {
    const errorText = await detailResponse.text();
    console.log('Get client detail error:', errorText);
  }
}

// 测试更新客户
async function testUpdateClient(token, clientId) {
  const updateData = {
    name: '更新后的测试客户',
    email: 'updated@example.com',
    phone: '13900139000',
    notes: '这是一个更新后的测试客户'
  };

  const updateResponse = await fetch(`http://localhost:8080/api/v1/clients/${clientId}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(updateData)
  });

  console.log('Update client response status:', updateResponse.status);
  if (updateResponse.ok) {
    const updateData = await updateResponse.json();
    console.log('Update client response:', JSON.stringify(updateData, null, 2));
  } else {
    const errorText = await updateResponse.text();
    console.log('Update client error:', errorText);
  }
}

// 测试删除客户
async function testDeleteClient(token, clientId) {
  const deleteResponse = await fetch(`http://localhost:8080/api/v1/clients/${clientId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('Delete client response status:', deleteResponse.status);
  if (deleteResponse.ok) {
    const deleteData = await deleteResponse.json();
    console.log('Delete client response:', JSON.stringify(deleteData, null, 2));
  } else {
    const errorText = await deleteResponse.text();
    console.log('Delete client error:', errorText);
  }
}

// 执行测试
testCreateClientAPI().catch(console.error);
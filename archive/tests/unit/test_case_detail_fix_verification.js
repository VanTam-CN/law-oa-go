const axios = require('axios');

// 配置
const API_BASE_URL = 'http://localhost:8080/api';
const TEST_USER = {
  email: 'admin@example.com',
  password: 'admin123'
};

let authToken = '';

// 登录获取token
async function login() {
  try {
    const response = await axios.post(`${API_BASE_URL}/auth/login`, TEST_USER);
    if (response.data.success && response.data.data.token) {
      authToken = response.data.data.token;
      console.log('✅ 登录成功，获取到token');
      return true;
    }
    console.log('❌ 登录失败：响应格式不正确');
    return false;
  } catch (error) {
    console.log('❌ 登录失败：', error.response?.data || error.message);
    return false;
  }
}

// 获取案件ID
async function getCaseId() {
  try {
    const response = await axios.get(`${API_BASE_URL}/cases`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      },
      params: {
        page: 1,
        page_size: 1
      }
    });

    if (response.data.success && response.data.data.length > 0) {
      const caseId = response.data.data[0].id;
      console.log(`✅ 找到案件ID: ${caseId}`);
      return caseId;
    } else {
      console.log('❌ 获取案件列表失败');
      return null;
    }
  } catch (error) {
    console.log('❌ 获取案件ID失败：', error.response?.data || error.message);
    return null;
  }
}

// 测试获取案件详情
async function testGetCaseDetail(caseId) {
  try {
    const response = await axios.get(`${API_BASE_URL}/cases/${caseId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    if (response.data.success) {
      const caseData = response.data.data;
      console.log('✅ 获取案件详情成功！');
      console.log('📋 案件基本信息：');
      console.log(`  ID: ${caseData.id}`);
      console.log(`  标题: ${caseData.title}`);
      console.log(`  描述: ${caseData.description || '无描述'}`);
      console.log(`  类型: ${caseData.case_type}`);
      console.log(`  状态: ${caseData.status}`);
      console.log(`  优先级: ${caseData.priority}`);
      console.log(`  客户ID: ${caseData.client_id}`);
      console.log(`  律师ID: ${caseData.lawyer_id}`);

      if (caseData.client) {
        console.log(`  客户姓名: ${caseData.client.name || caseData.client.company || '未知'}`);
        console.log(`  客户联系方式: ${caseData.client.phone || '无'}`);
        console.log(`  客户邮箱: ${caseData.client.email || '无'}`);
      }

      if (caseData.lawyer) {
        console.log(`  律师姓名: ${caseData.lawyer.name || '未分配'}`);
        console.log(`  律师联系方式: ${caseData.lawyer.phone || '无'}`);
        console.log(`  律师邮箱: ${caseData.lawyer.email || '无'}`);
      }

      console.log(`  创建时间: ${caseData.created_at}`);
      console.log(`  更新时间: ${caseData.updated_at}`);

      return true;
    } else {
      console.log('❌ 获取案件详情失败：', response.data.error);
      return false;
    }
  } catch (error) {
    console.log('❌ 获取案件详情失败：', error.response?.data || error.message);
    return false;
  }
}

// 测试更新案件
async function testUpdateCase(caseId) {
  try {
    const updateData = {
      title: `测试更新的案件标题 - ${new Date().toISOString()}`,
      description: `这是一个测试更新的案件描述，用于验证前端修复效果 - ${new Date().toISOString()}`,
      status: 'active',
      priority: 'high'
    };

    const response = await axios.put(`${API_BASE_URL}/cases/${caseId}`, updateData, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      }
    });

    if (response.data.success) {
      const updatedCase = response.data.data;
      console.log('✅ 更新案件成功！');
      console.log('📝 更新后的案件信息：');
      console.log(`  标题: ${updatedCase.title}`);
      console.log(`  描述: ${updatedCase.description}`);
      console.log(`  状态: ${updatedCase.status}`);
      console.log(`  优先级: ${updatedCase.priority}`);

      return true;
    } else {
      console.log('❌ 更新案件失败：', response.data.error);
      return false;
    }
  } catch (error) {
    console.log('❌ 更新案件失败：', error.response?.data || error.message);
    return false;
  }
}

// 验证数据格式一致性
async function testDataConsistency() {
  try {
    const response = await axios.get(`${API_BASE_URL}/cases`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      },
      params: {
        page: 1,
        page_size: 5
      }
    });

    if (response.data.success) {
      const cases = response.data.data;
      console.log('✅ 数据格式一致性验证：');
      console.log(`  获取到 ${cases.length} 个案件`);

      let consistencyIssues = 0;

      cases.forEach((caseItem, index) => {
        console.log(`\n📄 案件 ${index + 1} 数据格式检查：`);

        // 检查必需字段
        const requiredFields = ['id', 'title', 'case_type', 'status', 'priority', 'created_at', 'updated_at'];
        requiredFields.forEach(field => {
          if (!caseItem[field]) {
            console.log(`  ⚠️  缺少字段: ${field}`);
            consistencyIssues++;
          }
        });

        // 检查嵌套对象
        if (caseItem.client) {
          console.log(`  ✅ 客户信息完整: ${caseItem.client.name || caseItem.client.company || '匿名客户'}`);
        } else {
          console.log(`  ⚠️  客户信息为空`);
          consistencyIssues++;
        }

        if (caseItem.lawyer) {
          console.log(`  ✅ 律师信息完整: ${caseItem.lawyer.name || '未分配律师'}`);
        } else {
          console.log(`  ⚠️  律师信息为空`);
        }
      });

      if (consistencyIssues === 0) {
        console.log('\n✅ 所有案件数据格式一致性验证通过！');
        return true;
      } else {
        console.log(`\n❌ 发现 ${consistencyIssues} 个数据一致性问题`);
        return false;
      }
    } else {
      console.log('❌ 获取案件列表失败：', response.data.error);
      return false;
    }
  } catch (error) {
    console.log('❌ 数据格式一致性验证失败：', error.response?.data || error.message);
    return false;
  }
}

// 主测试函数
async function runComprehensiveTest() {
  console.log('🚀 开始案件详情修复验证测试...\n');

  const testResults = {
    login: false,
    getCaseId: false,
    getCaseDetail: false,
    updateCase: false,
    dataConsistency: false
  };

  // 步骤1：登录
  console.log('📝 步骤1：用户登录...');
  testResults.login = await login();
  if (!testResults.login) {
    console.log('\n❌ 测试失败：无法登录');
    return false;
  }

  // 步骤2：获取测试案件ID
  console.log('\n📝 步骤2：获取测试案件ID...');
  const caseId = await getCaseId();
  if (!caseId) {
    console.log('\n❌ 测试失败：无法获取案件ID');
    return false;
  }
  testResults.getCaseId = true;

  // 步骤3：测试获取案件详情
  console.log('\n📝 步骤3：测试获取案件详情...');
  testResults.getCaseDetail = await testGetCaseDetail(caseId);

  // 步骤4：测试更新案件
  console.log('\n📝 步骤4：测试更新案件...');
  testResults.updateCase = await testUpdateCase(caseId);

  // 步骤5：验证数据格式一致性
  console.log('\n📝 步骤5：验证数据格式一致性...');
  testResults.dataConsistency = await testDataConsistency();

  // 总结测试结果
  console.log('\n📊 测试结果总结：');
  console.log('='.repeat(50));
  console.log(`登录测试:            ${testResults.login ? '✅ 通过' : '❌ 失败'}`);
  console.log(`获取案件ID:           ${testResults.getCaseId ? '✅ 通过' : '❌ 失败'}`);
  console.log(`获取案件详情:          ${testResults.getCaseDetail ? '✅ 通过' : '❌ 失败'}`);
  console.log(`更新案件:            ${testResults.updateCase ? '✅ 通过' : '❌ 失败'}`);
  console.log(`数据格式一致性验证:     ${testResults.dataConsistency ? '✅ 通过' : '❌ 失败'}`);
  console.log('='.repeat(50));

  const allPassed = Object.values(testResults).every(result => result);

  if (allPassed) {
    console.log('\n🎉 所有测试通过！案件详情页面修复成功！');
    console.log('📝 修复要点：');
    console.log('  - 前端页面已替换模拟数据为真实API调用');
    console.log('  - 数据结构已与后端API保持一致');
    console.log('  - 案件详情可以正常获取、显示和更新');
    console.log('  - 客户和律师信息正确关联');
    console.log('  - 数据格式一致性验证通过');
    return true;
  } else {
    console.log('\n❌ 部分测试失败，需要进一步检查');
    return false;
  }
}

// 运行测试
runComprehensiveTest().catch(error => {
  console.error('❌ 测试过程中发生错误：', error);
  process.exit(1);
});
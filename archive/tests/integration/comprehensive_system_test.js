const axios = require('axios');

class ComprehensiveSystemTest {
  constructor() {
    this.baseURL = 'http://localhost:8080';
    this.frontendURL = 'http://localhost:3003';
    this.token = null;
    this.results = {
      backend: {},
      frontend: {},
      integration: {},
      fixes: []
    };
  }

  async login() {
    console.log('🔐 登录系统...');
    try {
      const response = await axios.post(`${this.baseURL}/api/auth/login`, {
        email: 'admin@example.com',
        password: 'admin123'
      });
      
      if (response.data.success) {
        this.token = response.data.data.token;
        console.log('✅ 登录成功');
        return true;
      }
    } catch (error) {
      console.log('❌ 登录失败:', error.message);
      return false;
    }
  }

  async testClientAPI() {
    console.log('\n👥 测试客户管理功能...');
    const headers = { Authorization: `Bearer ${this.token}` };

    try {
      // 1. 测试客户列表API
      const listResponse = await axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=5`, { headers });
      const clients = listResponse.data.data || [];
      
      console.log(`📊 客户列表: ${clients.length}个客户`);
      
      // 2. 检查客户数据结构
      let validNameCount = 0;
      let emptyNameCount = 0;
      
      for (const client of clients) {
        if (client.name && client.name.trim() !== '') {
          validNameCount++;
        } else {
          emptyNameCount++;
          console.log(`⚠️ 客户${client.id}缺少name字段`);
        }
      }
      
      console.log(`✅ 客户名称字段: ${validNameCount}个有效，${emptyNameCount}个缺失`);
      
      // 3. 测试客户创建
      const createData = {
        name: '测试新客户',
        email: `test_${Date.now()}@example.com`,
        phone: '13800138999',
        address: '测试地址',
        company: '测试公司',
        status: 'active'
      };
      
      const createResponse = await axios.post(`${this.baseURL}/api/clients`, createData, { headers });
      const newClient = createResponse.data.data;
      
      console.log(`✅ 客户创建成功: ID=${newClient.id}, Name=${newClient.name}`);
      
      // 4. 测试客户更新
      const updateData = { name: '更新后的客户名称' };
      await axios.put(`${this.baseURL}/api/clients/${newClient.id}`, updateData, { headers });
      console.log('✅ 客户更新成功');
      
      // 5. 清理测试数据
      await axios.delete(`${this.baseURL}/api/clients/${newClient.id}`, { headers });
      console.log('✅ 测试数据清理完成');
      
      this.results.backend.clients = {
        status: 'success',
        totalClients: clients.length,
        validNames: validNameCount,
        emptyNames: emptyNameCount,
        crudOperations: 'success'
      };
      
      return true;
      
    } catch (error) {
      console.log('❌ 客户API测试失败:', error.message);
      this.results.backend.clients = { status: 'error', message: error.message };
      return false;
    }
  }

  async testCaseAPI() {
    console.log('\n📋 测试案件管理功能...');
    const headers = { Authorization: `Bearer ${this.token}` };

    try {
      // 1. 测试案件列表API
      const listResponse = await axios.get(`${this.baseURL}/api/cases?page=1&page_size=5`, { headers });
      const cases = listResponse.data.data || [];
      
      console.log(`📊 案件列表: ${cases.length}个案件`);
      
      // 2. 检查案件数据结构和关联
      let validClientNames = 0;
      let validLawyerNames = 0;
      let validClientLinks = 0;
      let validLawyerLinks = 0;
      
      // 获取客户和律师数据用于验证关联
      const [clientsRes, lawyersRes] = await Promise.all([
        axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=100`, { headers }),
        axios.get(`${this.baseURL}/api/lawfirm/lawyers?pageNum=1&pageSize=100`, { headers })
      ]);
      
      const clients = clientsRes.data.data || [];
      const lawyers = lawyersRes.data.data || [];
      
      for (const caseItem of cases) {
        // 检查名称字段
        if (caseItem.client_name && caseItem.client_name.trim() !== '') {
          validClientNames++;
        }
        if (caseItem.lawyer_name && caseItem.lawyer_name.trim() !== '') {
          validLawyerNames++;
        }
        
        // 检查关联关系
        const clientExists = clients.find(c => c.id === caseItem.client_id);
        if (clientExists) validClientLinks++;
        
        const lawyerExists = lawyers.find(l => l.id === caseItem.lawyer_id);
        if (lawyerExists) validLawyerLinks++;
      }
      
      const clientNameRate = cases.length > 0 ? (validClientNames / cases.length * 100).toFixed(1) : 0;
      const lawyerNameRate = cases.length > 0 ? (validLawyerNames / cases.length * 100).toFixed(1) : 0;
      const clientLinkRate = cases.length > 0 ? (validClientLinks / cases.length * 100).toFixed(1) : 0;
      const lawyerLinkRate = cases.length > 0 ? (validLawyerLinks / cases.length * 100).toFixed(1) : 0;
      
      console.log(`✅ 案件数据完整性:`);
      console.log(`   客户名称完整率: ${clientNameRate}% (${validClientNames}/${cases.length})`);
      console.log(`   律师名称完整率: ${lawyerNameRate}% (${validLawyerNames}/${cases.length})`);
      console.log(`   客户关联完整率: ${clientLinkRate}% (${validClientLinks}/${cases.length})`);
      console.log(`   律师关联完整率: ${lawyerLinkRate}% (${validLawyerLinks}/${cases.length})`);
      
      this.results.backend.cases = {
        status: 'success',
        totalCases: cases.length,
        clientNameRate: parseFloat(clientNameRate),
        lawyerNameRate: parseFloat(lawyerNameRate),
        clientLinkRate: parseFloat(clientLinkRate),
        lawyerLinkRate: parseFloat(lawyerLinkRate)
      };
      
      // 记录修复成果
      if (clientNameRate >= 90) {
        this.results.fixes.push('✅ 客户名称字段修复成功');
      }
      if (lawyerNameRate >= 90) {
        this.results.fixes.push('✅ 律师名称字段修复成功');
      }
      
      return true;
      
    } catch (error) {
      console.log('❌ 案件API测试失败:', error.message);
      this.results.backend.cases = { status: 'error', message: error.message };
      return false;
    }
  }

  async testLawyerAPI() {
    console.log('\n⚖️ 测试律师管理功能...');
    const headers = { Authorization: `Bearer ${this.token}` };

    try {
      const response = await axios.get(`${this.baseURL}/api/lawfirm/lawyers?pageNum=1&pageSize=10`, { headers });
      const lawyers = response.data.data || [];
      
      console.log(`📊 律师列表: ${lawyers.length}个律师`);
      
      // 检查律师数据结构
      let validNameCount = 0;
      for (const lawyer of lawyers) {
        if (lawyer.name && lawyer.name.trim() !== '') {
          validNameCount++;
        }
      }
      
      const nameRate = lawyers.length > 0 ? (validNameCount / lawyers.length * 100).toFixed(1) : 0;
      console.log(`✅ 律师名称完整率: ${nameRate}% (${validNameCount}/${lawyers.length})`);
      
      this.results.backend.lawyers = {
        status: 'success',
        totalLawyers: lawyers.length,
        validNames: validNameCount,
        nameRate: parseFloat(nameRate)
      };
      
      return true;
      
    } catch (error) {
      console.log('❌ 律师API测试失败:', error.message);
      this.results.backend.lawyers = { status: 'error', message: error.message };
      return false;
    }
  }

  async testConflictAPI() {
    console.log('\n⚠️ 测试利益冲突检测功能...');
    const headers = { Authorization: `Bearer ${this.token}` };

    try {
      const response = await axios.post(`${this.baseURL}/api/conflict/check`, {
        clientId: "1",
        clientName: "测试客户",
        caseName: "测试案件",
        caseType: "civil",
        clientType: "PERSON",
        otherParties: ["对方当事人"],
        searchYears: 5,
        userId: "1",
        requestTime: new Date().toISOString()
      }, { headers });
      
      const result = response.data.data;
      
      console.log('✅ 利益冲突检测正常');
      console.log(`   检查ID: ${result.checkId}`);
      console.log(`   有冲突: ${result.hasConflict}`);
      console.log(`   风险等级: ${result.riskAssessment?.overallRisk}`);
      console.log(`   检查耗时: ${result.duration}ms`);
      
      this.results.backend.conflict = {
        status: 'success',
        hasConflict: result.hasConflict,
        riskLevel: result.riskAssessment?.overallRisk,
        duration: result.duration
      };
      
      this.results.fixes.push('✅ 利益冲突检测API正常工作');
      
      return true;
      
    } catch (error) {
      console.log('❌ 利益冲突检测失败:', error.message);
      this.results.backend.conflict = { status: 'error', message: error.message };
      return false;
    }
  }

  async testDataConsistency() {
    console.log('\n🔗 测试数据一致性...');
    const headers = { Authorization: `Bearer ${this.token}` };

    try {
      // 获取所有相关数据
      const [casesRes, clientsRes, lawyersRes] = await Promise.all([
        axios.get(`${this.baseURL}/api/cases?page=1&page_size=50`, { headers }),
        axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=100`, { headers }),
        axios.get(`${this.baseURL}/api/lawfirm/lawyers?pageNum=1&pageSize=100`, { headers })
      ]);

      const cases = casesRes.data.data || [];
      const clients = clientsRes.data.data || [];
      const lawyers = lawyersRes.data.data || [];

      console.log(`📊 数据统计:`);
      console.log(`   案件: ${cases.length}个`);
      console.log(`   客户: ${clients.length}个`);
      console.log(`   律师: ${lawyers.length}个`);

      // 检查数据一致性
      let consistentClientNames = 0;
      let consistentLawyerNames = 0;

      for (const caseItem of cases) {
        // 检查客户名称一致性
        const client = clients.find(c => c.id === caseItem.client_id);
        if (client && caseItem.client_name) {
          // 检查名称是否匹配（可能是name或company字段）
          if (caseItem.client_name === client.name || caseItem.client_name === client.company) {
            consistentClientNames++;
          }
        }

        // 检查律师名称一致性
        const lawyer = lawyers.find(l => l.id === caseItem.lawyer_id);
        if (lawyer && caseItem.lawyer_name === lawyer.name) {
          consistentLawyerNames++;
        }
      }

      const clientConsistencyRate = cases.length > 0 ? (consistentClientNames / cases.length * 100).toFixed(1) : 0;
      const lawyerConsistencyRate = cases.length > 0 ? (consistentLawyerNames / cases.length * 100).toFixed(1) : 0;

      console.log(`🔗 数据一致性检查:`);
      console.log(`   客户名称一致率: ${clientConsistencyRate}% (${consistentClientNames}/${cases.length})`);
      console.log(`   律师名称一致率: ${lawyerConsistencyRate}% (${consistentLawyerNames}/${cases.length})`);

      this.results.integration.dataConsistency = {
        status: 'success',
        clientConsistencyRate: parseFloat(clientConsistencyRate),
        lawyerConsistencyRate: parseFloat(lawyerConsistencyRate)
      };

      if (clientConsistencyRate >= 80) {
        this.results.fixes.push('✅ 客户数据一致性良好');
      }
      if (lawyerConsistencyRate >= 80) {
        this.results.fixes.push('✅ 律师数据一致性良好');
      }

      return true;

    } catch (error) {
      console.log('❌ 数据一致性测试失败:', error.message);
      this.results.integration.dataConsistency = { status: 'error', message: error.message };
      return false;
    }
  }

  generateReport() {
    console.log('\n📊 系统测试报告');
    console.log('='.repeat(50));
    
    console.log('\n🔧 修复成果:');
    if (this.results.fixes.length > 0) {
      this.results.fixes.forEach(fix => console.log(`   ${fix}`));
    } else {
      console.log('   暂无修复项目');
    }

    console.log('\n📈 后端API状态:');
    console.log(`   客户管理: ${this.results.backend.clients?.status || '未测试'}`);
    console.log(`   案件管理: ${this.results.backend.cases?.status || '未测试'}`);
    console.log(`   律师管理: ${this.results.backend.lawyers?.status || '未测试'}`);
    console.log(`   冲突检测: ${this.results.backend.conflict?.status || '未测试'}`);

    console.log('\n🔗 数据质量指标:');
    if (this.results.backend.cases) {
      console.log(`   案件客户名称完整率: ${this.results.backend.cases.clientNameRate}%`);
      console.log(`   案件律师名称完整率: ${this.results.backend.cases.lawyerNameRate}%`);
    }
    if (this.results.integration.dataConsistency) {
      console.log(`   客户数据一致率: ${this.results.integration.dataConsistency.clientConsistencyRate}%`);
      console.log(`   律师数据一致率: ${this.results.integration.dataConsistency.lawyerConsistencyRate}%`);
    }

    // 计算总体评分
    let totalScore = 0;
    let maxScore = 0;

    // 后端API评分 (40分)
    const backendTests = ['clients', 'cases', 'lawyers', 'conflict'];
    backendTests.forEach(test => {
      maxScore += 10;
      if (this.results.backend[test]?.status === 'success') {
        totalScore += 10;
      }
    });

    // 数据质量评分 (60分)
    if (this.results.backend.cases) {
      maxScore += 30;
      if (this.results.backend.cases.clientNameRate >= 90) totalScore += 15;
      if (this.results.backend.cases.lawyerNameRate >= 90) totalScore += 15;
    }

    if (this.results.integration.dataConsistency) {
      maxScore += 30;
      if (this.results.integration.dataConsistency.clientConsistencyRate >= 80) totalScore += 15;
      if (this.results.integration.dataConsistency.lawyerConsistencyRate >= 80) totalScore += 15;
    }

    const finalScore = maxScore > 0 ? (totalScore / maxScore * 100).toFixed(1) : 0;

    console.log('\n🎯 系统健康度评分:');
    console.log(`   总分: ${finalScore}/100`);
    
    if (finalScore >= 90) {
      console.log('   状态: 🟢 优秀 - 系统运行良好');
    } else if (finalScore >= 70) {
      console.log('   状态: 🟡 良好 - 有少量问题需要关注');
    } else {
      console.log('   状态: 🔴 需要改进 - 存在重要问题');
    }

    return {
      score: parseFloat(finalScore),
      results: this.results,
      fixes: this.results.fixes
    };
  }

  async run() {
    console.log('🚀 开始全面系统测试...\n');

    const loginSuccess = await this.login();
    if (!loginSuccess) {
      console.log('❌ 登录失败，无法继续测试');
      return false;
    }

    // 执行所有测试
    await this.testClientAPI();
    await this.testCaseAPI();
    await this.testLawyerAPI();
    await this.testConflictAPI();
    await this.testDataConsistency();

    // 生成报告
    const report = this.generateReport();

    // 保存报告
    const fs = require('fs');
    fs.writeFileSync('comprehensive_test_report.json', JSON.stringify(report, null, 2));
    console.log('\n📄 详细报告已保存到 comprehensive_test_report.json');

    return report;
  }
}

// 运行测试
async function main() {
  const test = new ComprehensiveSystemTest();
  try {
    const report = await test.run();
    
    if (report.score >= 90) {
      console.log('\n🎉 系统测试通过！所有主要功能正常工作。');
    } else {
      console.log('\n⚠️ 系统测试完成，但仍有改进空间。');
    }
    
    return report;
  } catch (error) {
    console.error('❌ 测试失败:', error);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = ComprehensiveSystemTest;
const axios = require('axios');

class QuickSystemTest {
  constructor() {
    this.baseURL = 'http://localhost:8080';
    this.frontendURL = 'http://localhost:3003';
    this.token = null;
    this.issues = [];
  }

  async testLogin() {
    console.log('🔐 测试登录功能...');
    try {
      const response = await axios.post(`${this.baseURL}/api/auth/login`, {
        email: 'admin@example.com',
        password: 'admin123'
      });
      
      if (response.data.success) {
        this.token = response.data.data.token;
        console.log('✅ 登录成功，获得token');
        return true;
      } else {
        console.log('❌ 登录失败:', response.data.message);
        return false;
      }
    } catch (error) {
      console.log('❌ 登录API错误:', error.message);
      this.issues.push('登录API不可用');
      return false;
    }
  }

  async testCasesAPI() {
    console.log('📋 测试案件API...');
    if (!this.token) {
      console.log('⚠️ 无token，跳过案件API测试');
      return false;
    }

    try {
      const headers = { Authorization: `Bearer ${this.token}` };
      const response = await axios.get(`${this.baseURL}/api/cases?page=1&page_size=10`, { headers });
      
      const cases = response.data.data || [];
      const total = response.data.pagination?.total || 0;
      
      console.log(`✅ 案件API正常，共${total}个案件，当前页${cases.length}条`);
      
      // 检查数据结构
      if (cases.length > 0) {
        const firstCase = cases[0];
        console.log('📊 案件数据结构:');
        console.log(`   ID: ${firstCase.id}`);
        console.log(`   标题: ${firstCase.title}`);
        console.log(`   客户ID: ${firstCase.client_id}`);
        console.log(`   客户名: ${firstCase.client_name}`);
        console.log(`   律师ID: ${firstCase.lawyer_id}`);
        console.log(`   律师名: ${firstCase.lawyer_name}`);
        console.log(`   状态: ${firstCase.status}`);
        console.log(`   类型: ${firstCase.case_type}`);
      }
      
      return { success: true, total, cases: cases.length };
    } catch (error) {
      console.log('❌ 案件API错误:', error.message);
      this.issues.push('案件API不可用');
      return false;
    }
  }

  async testClientsAPI() {
    console.log('👥 测试客户API...');
    if (!this.token) {
      console.log('⚠️ 无token，跳过客户API测试');
      return false;
    }

    try {
      const headers = { Authorization: `Bearer ${this.token}` };
      const response = await axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=10`, { headers });
      
      const clients = response.data.data?.list || [];
      
      console.log(`✅ 客户API正常，共${clients.length}个客户`);
      
      if (clients.length > 0) {
        const firstClient = clients[0];
        console.log('📊 客户数据结构:');
        console.log(`   ID: ${firstClient.id}`);
        console.log(`   姓名: ${firstClient.name}`);
        console.log(`   公司: ${firstClient.company}`);
        console.log(`   邮箱: ${firstClient.email}`);
        console.log(`   电话: ${firstClient.phone}`);
      }
      
      return { success: true, clients: clients.length };
    } catch (error) {
      console.log('❌ 客户API错误:', error.message);
      this.issues.push('客户API不可用');
      return false;
    }
  }

  async testLawyersAPI() {
    console.log('⚖️ 测试律师API...');
    if (!this.token) {
      console.log('⚠️ 无token，跳过律师API测试');
      return false;
    }

    try {
      const headers = { Authorization: `Bearer ${this.token}` };
      const response = await axios.get(`${this.baseURL}/api/lawfirm/lawyers?pageNum=1&pageSize=10`, { headers });
      
      const lawyers = response.data.data?.list || [];
      
      console.log(`✅ 律师API正常，共${lawyers.length}个律师`);
      
      if (lawyers.length > 0) {
        const firstLawyer = lawyers[0];
        console.log('📊 律师数据结构:');
        console.log(`   ID: ${firstLawyer.id}`);
        console.log(`   姓名: ${firstLawyer.name}`);
        console.log(`   邮箱: ${firstLawyer.email}`);
        console.log(`   角色: ${firstLawyer.role}`);
      }
      
      return { success: true, lawyers: lawyers.length };
    } catch (error) {
      console.log('❌ 律师API错误:', error.message);
      this.issues.push('律师API不可用');
      return false;
    }
  }

  async testConflictAPI() {
    console.log('⚠️ 测试利益冲突检测API...');
    if (!this.token) {
      console.log('⚠️ 无token，跳过冲突检测API测试');
      return false;
    }

    try {
      const headers = { Authorization: `Bearer ${this.token}` };
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
      
      console.log('✅ 利益冲突检测API正常');
      console.log(`   检查ID: ${result.checkId}`);
      console.log(`   有冲突: ${result.hasConflict}`);
      console.log(`   风险等级: ${result.riskAssessment?.overallRisk}`);
      console.log(`   检查耗时: ${result.duration}ms`);
      
      return { success: true, hasConflict: result.hasConflict };
    } catch (error) {
      console.log('❌ 利益冲突检测API错误:', error.message);
      this.issues.push('利益冲突检测API不可用');
      return false;
    }
  }

  async testDataIntegrity() {
    console.log('🔗 测试数据完整性...');
    if (!this.token) {
      console.log('⚠️ 无token，跳过数据完整性测试');
      return false;
    }

    try {
      const headers = { Authorization: `Bearer ${this.token}` };
      
      // 获取数据
      const [casesRes, clientsRes, lawyersRes] = await Promise.all([
        axios.get(`${this.baseURL}/api/cases?page=1&page_size=20`, { headers }),
        axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=100`, { headers }),
        axios.get(`${this.baseURL}/api/lawfirm/lawyers?pageNum=1&pageSize=100`, { headers })
      ]);

      const cases = casesRes.data.data || [];
      const clients = clientsRes.data.data?.list || [];
      const lawyers = lawyersRes.data.data?.list || [];

      console.log(`📊 数据统计:`);
      console.log(`   案件: ${cases.length}个`);
      console.log(`   客户: ${clients.length}个`);
      console.log(`   律师: ${lawyers.length}个`);

      // 检查关联关系
      let validClientLinks = 0;
      let validLawyerLinks = 0;
      let missingClientNames = 0;
      let missingLawyerNames = 0;

      for (const caseItem of cases) {
        // 检查客户关联
        const client = clients.find(c => c.id === caseItem.client_id);
        if (client) {
          validClientLinks++;
        }
        
        // 检查律师关联
        const lawyer = lawyers.find(l => l.id === caseItem.lawyer_id);
        if (lawyer) {
          validLawyerLinks++;
        }

        // 检查名称字段
        if (!caseItem.client_name) missingClientNames++;
        if (!caseItem.lawyer_name) missingLawyerNames++;
      }

      const clientLinkRate = cases.length > 0 ? (validClientLinks / cases.length * 100).toFixed(1) : 0;
      const lawyerLinkRate = cases.length > 0 ? (validLawyerLinks / cases.length * 100).toFixed(1) : 0;

      console.log(`🔗 关联关系检查:`);
      console.log(`   案件-客户关联率: ${clientLinkRate}% (${validClientLinks}/${cases.length})`);
      console.log(`   案件-律师关联率: ${lawyerLinkRate}% (${validLawyerLinks}/${cases.length})`);
      console.log(`   缺失客户名称: ${missingClientNames}个案件`);
      console.log(`   缺失律师名称: ${missingLawyerNames}个案件`);

      // 记录问题
      if (clientLinkRate < 90) {
        this.issues.push(`案件-客户关联率过低: ${clientLinkRate}%`);
      }
      if (lawyerLinkRate < 90) {
        this.issues.push(`案件-律师关联率过低: ${lawyerLinkRate}%`);
      }
      if (missingClientNames > 0) {
        this.issues.push(`${missingClientNames}个案件缺失客户名称`);
      }
      if (missingLawyerNames > 0) {
        this.issues.push(`${missingLawyerNames}个案件缺失律师名称`);
      }

      return {
        success: true,
        clientLinkRate: parseFloat(clientLinkRate),
        lawyerLinkRate: parseFloat(lawyerLinkRate),
        missingClientNames,
        missingLawyerNames
      };

    } catch (error) {
      console.log('❌ 数据完整性测试错误:', error.message);
      this.issues.push('数据完整性测试失败');
      return false;
    }
  }

  async generateFixPlan() {
    console.log('\n🔧 生成修复计划...');
    
    const fixes = [];

    // 基于发现的问题生成修复计划
    this.issues.forEach(issue => {
      if (issue.includes('客户名称')) {
        fixes.push({
          priority: 'high',
          issue: '案件数据缺失客户名称',
          solution: '修复案件API响应，确保包含client_name字段',
          file: 'internal/services/case_service.go'
        });
      }
      
      if (issue.includes('律师名称')) {
        fixes.push({
          priority: 'high',
          issue: '案件数据缺失律师名称',
          solution: '修复案件API响应，确保包含lawyer_name字段',
          file: 'internal/services/case_service.go'
        });
      }

      if (issue.includes('关联率')) {
        fixes.push({
          priority: 'medium',
          issue: '数据关联不完整',
          solution: '检查数据库外键约束和数据完整性',
          file: 'migrations/'
        });
      }

      if (issue.includes('API不可用')) {
        fixes.push({
          priority: 'critical',
          issue: 'API服务异常',
          solution: '检查后端服务状态和路由配置',
          file: 'internal/router/router.go'
        });
      }
    });

    // 前端相关修复
    fixes.push({
      priority: 'high',
      issue: '前端数据显示问题',
      solution: '修复前端API调用和数据映射逻辑',
      file: 'frontend/src/pages/case/CaseManagement.tsx'
    });

    fixes.push({
      priority: 'medium',
      issue: '界面风格不统一',
      solution: '统一各页面的UI组件和样式',
      file: 'frontend/src/pages/'
    });

    console.log('📋 修复计划:');
    fixes.forEach((fix, index) => {
      console.log(`${index + 1}. [${fix.priority.toUpperCase()}] ${fix.issue}`);
      console.log(`   解决方案: ${fix.solution}`);
      console.log(`   文件: ${fix.file}`);
      console.log('');
    });

    return fixes;
  }

  async run() {
    console.log('🚀 开始快速系统测试...\n');

    const results = {};

    // 测试后端API
    results.login = await this.testLogin();
    results.cases = await this.testCasesAPI();
    results.clients = await this.testClientsAPI();
    results.lawyers = await this.testLawyersAPI();
    results.conflict = await this.testConflictAPI();
    results.dataIntegrity = await this.testDataIntegrity();

    // 生成修复计划
    const fixes = await this.generateFixPlan();

    console.log('\n📊 测试结果摘要:');
    console.log(`   登录: ${results.login ? '✅' : '❌'}`);
    console.log(`   案件API: ${results.cases ? '✅' : '❌'}`);
    console.log(`   客户API: ${results.clients ? '✅' : '❌'}`);
    console.log(`   律师API: ${results.lawyers ? '✅' : '❌'}`);
    console.log(`   冲突检测: ${results.conflict ? '✅' : '❌'}`);
    console.log(`   数据完整性: ${results.dataIntegrity ? '✅' : '❌'}`);
    console.log(`\n发现问题: ${this.issues.length}个`);
    console.log(`修复计划: ${fixes.length}项`);

    return { results, issues: this.issues, fixes };
  }
}

// 运行测试
async function main() {
  const test = new QuickSystemTest();
  try {
    const report = await test.run();
    
    // 保存报告
    const fs = require('fs');
    fs.writeFileSync('quick_test_report.json', JSON.stringify(report, null, 2));
    console.log('\n📄 测试报告已保存到 quick_test_report.json');
    
    return report;
  } catch (error) {
    console.error('❌ 测试失败:', error);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = QuickSystemTest;
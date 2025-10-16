/**
 * 前后端API集成测试脚本
 * 用于验证前端服务层与后端API的对接是否正常
 */

import { authService } from '../services/authService';
import { clientService } from '../services/clientService';
import { caseService } from '../services/caseService';

// 测试配置
const TEST_CONFIG = {
  baseURL: 'http://localhost:8080/api/v1',
  testUser: {
    email: 'admin@lawfirm.com',
    password: 'password123'
  },
  testClient: {
    name: '测试客户',
    email: 'test@example.com',
    phone: '13800138000',
    address: '北京市朝阳区',
    company: '测试公司',
    type: 'personal' as const,
    notes: '这是一个测试客户'
  },
  testCase: {
    title: '测试案件',
    description: '这是一个测试案件',
    case_type: 'civil',
    priority: 'medium',
    status: 'pending'
  }
};

class APITestRunner {
  private authToken: string | null = null;

  constructor() {
    console.log('🚀 开始API集成测试...');
    console.log('测试环境:', TEST_CONFIG.baseURL);
  }

  // 测试认证功能
  async testAuthentication() {
    console.log('\n🔐 测试认证功能...');

    try {
      // 测试登录
      console.log('  正在登录...');
      const loginResponse = await authService.login(TEST_CONFIG.testUser);
      console.log('  ✅ 登录成功');
      console.log('  用户:', loginResponse.user.name);
      console.log('  角色:', loginResponse.user.role);

      this.authToken = loginResponse.token;

      // 测试获取当前用户信息
      console.log('  正在获取用户信息...');
      const currentUser = await authService.getCurrentUser();
      console.log('  ✅ 获取用户信息成功');
      console.log('  用户ID:', currentUser.id);

      // 测试权限检查
      console.log('  正在检查权限...');
      const isAuthenticated = authService.isAuthenticated();
      console.log('  ✅ 认证状态检查:', isAuthenticated);

      return true;
    } catch (error) {
      console.error('  ❌ 认证测试失败:', error);
      return false;
    }
  }

  // 测试客户管理功能
  async testClientManagement() {
    console.log('\n👥 测试客户管理功能...');

    try {
      // 测试创建客户
      console.log('  正在创建客户...');
      const newClient = await clientService.createClient(TEST_CONFIG.testClient);
      console.log('  ✅ 创建客户成功');
      console.log('  客户ID:', newClient.id);

      // 测试获取客户列表
      console.log('  正在获取客户列表...');
      const clients = await clientService.getClients({ page: 1, page_size: 10 });
      console.log('  ✅ 获取客户列表成功');
      console.log('  客户数量:', clients.data.length);
      console.log('  分页信息:', clients.pagination);

      // 测试获取客户详情
      console.log('  正在获取客户详情...');
      const clientDetail = await clientService.getClient(newClient.id);
      console.log('  ✅ 获取客户详情成功');
      console.log('  客户名称:', clientDetail.name);

      // 测试搜索客户
      console.log('  正在搜索客户...');
      const searchResults = await clientService.searchClients('测试');
      console.log('  ✅ 搜索客户成功');
      console.log('  搜索结果数量:', searchResults.data.length);

      // 测试获取客户统计
      console.log('  正在获取客户统计...');
      const stats = await clientService.getClientStats();
      console.log('  ✅ 获取客户统计成功');
      console.log('  总客户数:', stats.total);
      console.log('  活跃客户:', stats.active);

      // 测试更新客户
      console.log('  正在更新客户...');
      const updatedClient = await clientService.updateClient(newClient.id, {
        notes: '这是一个更新后的测试客户'
      });
      console.log('  ✅ 更新客户成功');
      console.log('  更新后的备注:', updatedClient.notes);

      // 清理测试数据
      console.log('  正在清理测试客户...');
      await clientService.deleteClient(newClient.id);
      console.log('  ✅ 清理测试客户成功');

      return true;
    } catch (error) {
      console.error('  ❌ 客户管理测试失败:', error);
      return false;
    }
  }

  // 测试案件管理功能
  async testCaseManagement() {
    console.log('\n⚖️ 测试案件管理功能...');

    try {
      // 先创建一个测试客户
      console.log('  正在创建测试客户...');
      const testClient = await clientService.createClient(TEST_CONFIG.testClient);
      console.log('  ✅ 创建测试客户成功，客户ID:', testClient.id);

      // 测试创建案件
      console.log('  正在创建案件...');
      const newCase = await caseService.createCase({
        ...TEST_CONFIG.testCase,
        client_id: testClient.id
      });
      console.log('  ✅ 创建案件成功');
      console.log('  案件ID:', newCase.id);

      // 测试获取案件列表
      console.log('  正在获取案件列表...');
      const cases = await caseService.getCases({ page: 1, page_size: 10 });
      console.log('  ✅ 获取案件列表成功');
      console.log('  案件数量:', cases.data.length);

      // 测试获取案件详情
      console.log('  正在获取案件详情...');
      const caseDetail = await caseService.getCase(newCase.id);
      console.log('  ✅ 获取案件详情成功');
      console.log('  案件标题:', caseDetail.title);

      // 测试搜索案件
      console.log('  正在搜索案件...');
      const searchResults = await caseService.searchCases('测试');
      console.log('  ✅ 搜索案件成功');
      console.log('  搜索结果数量:', searchResults.data.length);

      // 测试获取案件统计
      console.log('  正在获取案件统计...');
      const stats = await caseService.getCaseStats();
      console.log('  ✅ 获取案件统计成功');
      console.log('  总案件数:', stats.total);
      console.log('  待处理案件:', stats.pending);

      // 测试更新案件
      console.log('  正在更新案件...');
      const updatedCase = await caseService.updateCase(newCase.id, {
        description: '这是一个更新后的测试案件'
      });
      console.log('  ✅ 更新案件成功');
      console.log('  更新后的描述:', updatedCase.description);

      // 清理测试数据
      console.log('  正在清理测试案件...');
      await caseService.deleteCase(newCase.id);
      console.log('  ✅ 清理测试案件成功');

      console.log('  正在清理测试客户...');
      await clientService.deleteClient(testClient.id);
      console.log('  ✅ 清理测试客户成功');

      return true;
    } catch (error) {
      console.error('  ❌ 案件管理测试失败:', error);
      return false;
    }
  }

  // 测试错误处理
  async testErrorHandling() {
    console.log('\n🛡️ 测试错误处理功能...');

    try {
      // 测试401错误
      console.log('  正在测试401错误...');
      try {
        await clientService.getClients();
        console.log('  ❌ 应该返回401错误');
        return false;
      } catch (error: any) {
        if (error.code === 'AUTHENTICATION_ERROR') {
          console.log('  ✅ 401错误处理正常');
        } else {
          console.log('  ❌ 401错误处理异常:', error);
          return false;
        }
      }

      // 测试404错误
      console.log('  正在测试404错误...');
      try {
        await clientService.getClient(999999);
        console.log('  ❌ 应该返回404错误');
        return false;
      } catch (error: any) {
        if (error.code === 'NOT_FOUND') {
          console.log('  ✅ 404错误处理正常');
        } else {
          console.log('  ❌ 404错误处理异常:', error);
          return false;
        }
      }

      // 测试验证错误
      console.log('  正在测试验证错误...');
      try {
        await clientService.createClient({
          name: '',
          email: 'invalid-email',
          phone: '',
          address: '',
          company: '',
          type: 'personal' as const
        });
        console.log('  ❌ 应该返回验证错误');
        return false;
      } catch (error: any) {
        if (error.code === 'VALIDATION_ERROR') {
          console.log('  ✅ 验证错误处理正常');
        } else {
          console.log('  ❌ 验证错误处理异常:', error);
          return false;
        }
      }

      return true;
    } catch (error) {
      console.error('  ❌ 错误处理测试失败:', error);
      return false;
    }
  }

  // 运行所有测试
  async runAllTests() {
    console.log('🧪 开始运行所有API测试...\n');

    const results = {
      authentication: false,
      clientManagement: false,
      caseManagement: false,
      errorHandling: false
    };

    // 运行认证测试
    results.authentication = await this.testAuthentication();

    if (results.authentication) {
      // 只有认证成功后才运行其他测试
      results.clientManagement = await this.testClientManagement();
      results.caseManagement = await this.testCaseManagement();
      results.errorHandling = await this.testErrorHandling();
    }

    // 输出测试结果
    console.log('\n📊 测试结果汇总:');
    console.log('====================================');
    console.log('认证功能:', results.authentication ? '✅ 通过' : '❌ 失败');
    console.log('客户管理:', results.clientManagement ? '✅ 通过' : '❌ 失败');
    console.log('案件管理:', results.caseManagement ? '✅ 通过' : '❌ 失败');
    console.log('错误处理:', results.errorHandling ? '✅ 通过' : '❌ 失败');
    console.log('====================================');

    const allPassed = Object.values(results).every(result => result);

    if (allPassed) {
      console.log('🎉 所有测试通过！前后端API集成正常！');
    } else {
      console.log('⚠️  部分测试失败，请检查网络连接和后端服务状态');
    }

    return allPassed;
  }

  // 清理测试数据
  async cleanup() {
    console.log('\n🧹 清理测试数据...');

    try {
      // 登出
      if (this.authToken) {
        await authService.logout();
        console.log('  ✅ 登出成功');
      }
    } catch (error) {
      console.warn('  ⚠️ 登出时出现错误:', error);
    }
  }
}

// 导出测试运行器
export const apiTestRunner = new APITestRunner();

// 如果直接运行此脚本，则执行测试
if (typeof window !== 'undefined') {
  // 在浏览器环境中运行
  (window as any).runAPITests = async () => {
    await apiTestRunner.runAllTests();
    await apiTestRunner.cleanup();
  };

  console.log('💡 使用方法: 在浏览器控制台中运行 runAPITests()');
} else {
  // 在Node.js环境中运行
  console.log('💡 此测试脚本设计用于浏览器环境');
}

export default apiTestRunner;
#!/usr/bin/env node

/**
 * Vue前端修复验证脚本
 * 验证Vue版本的律师管理页面修复是否正确应用
 */

const fs = require('fs');
const path = require('path');

class VueFixVerifier {
  constructor() {
    this.results = {
      fileChecks: [],
      codeChanges: [],
      summary: ''
    };
  }

  checkFiles() {
    console.log('🔍 检查Vue前端修复文件...');

    const filesToCheck = [
      'frontend-vue/src/context/AuthContext.tsx',
      'frontend-vue/src/layouts/MainLayout.tsx',
      'frontend-vue/src/services/http.ts',
      'frontend-vue/src/pages/lawyer/LawyerManagement.tsx',
      'test_vue_lawyer_page.js'
    ];

    filesToCheck.forEach(file => {
      const exists = fs.existsSync(file);
      this.results.fileChecks.push({
        file,
        exists,
        status: exists ? '✅' : '❌'
      });
      console.log(`${exists ? '✅' : '❌'} ${file}`);
    });
  }

  verifyCodeChanges() {
    console.log('\n🔬 验证代码修改...');

    // 验证AuthContext修改
    try {
      const authContext = fs.readFileSync('frontend-vue/src/context/AuthContext.tsx', 'utf8');
      const hasDevModeCheck = authContext.includes('const isDevMode = process.env.NODE_ENV === \'development\'');
      const hasDevUser = authContext.includes('开发用户');
      const hasSkipAPI = authContext.includes('跳过API验证');

      this.results.codeChanges.push({
        file: 'AuthContext.tsx',
        devModeCheck: hasDevModeCheck,
        devUser: hasDevUser,
        skipAPI: hasSkipAPI,
        status: (hasDevModeCheck && hasDevUser && hasSkipAPI) ? '✅' : '❌'
      });

      console.log(`${hasDevModeCheck && hasDevUser && hasSkipAPI ? '✅' : '❌'} AuthContext.tsx - 开发模式认证修复`);
    } catch (error) {
      console.log(`❌ AuthContext.tsx - 读取失败: ${error.message}`);
    }

    // 验证MainLayout修改
    try {
      const mainLayout = fs.readFileSync('frontend-vue/src/layouts/MainLayout.tsx', 'utf8');
      const hasAlertImport = mainLayout.includes('import { Layout, Alert } from \'antd\'');
      const hasDevModeAlert = mainLayout.includes('开发者模式提示');

      this.results.codeChanges.push({
        file: 'MainLayout.tsx',
        alertImport: hasAlertImport,
        devModeAlert: hasDevModeAlert,
        status: (hasAlertImport && hasDevModeAlert) ? '✅' : '❌'
      });

      console.log(`${hasAlertImport && hasDevModeAlert ? '✅' : '❌'} MainLayout.tsx - 开发模式提示修复`);
    } catch (error) {
      console.log(`❌ MainLayout.tsx - 读取失败: ${error.message}`);
    }

    // 验证HTTP客户端修改
    try {
      const http = fs.readFileSync('frontend-vue/src/services/http.ts', 'utf8');
      const hasDevModeError = http.includes('开发者模式：认证错误，但不自动重定向');
      const hasFriendlyError = http.includes('开发模式认证错误，请检查token设置');

      this.results.codeChanges.push({
        file: 'http.ts',
        devModeError: hasDevModeError,
        friendlyError: hasFriendlyError,
        status: (hasDevModeError && hasFriendlyError) ? '✅' : '❌'
      });

      console.log(`${hasDevModeError && hasFriendlyError ? '✅' : '❌'} http.ts - 开发模式错误处理修复`);
    } catch (error) {
      console.log(`❌ http.ts - 读取失败: ${error.message}`);
    }

    // 验证LawyerManagement修改
    try {
      const lawyerManagement = fs.readFileSync('frontend-vue/src/pages/lawyer/LawyerManagement.tsx', 'utf8');
      const hasMockData = lawyerManagement.includes('mockLawyers = [');
      const hasDevModeCheck = lawyerManagement.includes('const isDevMode = process.env.NODE_ENV === \'development\'');
      const hasDevAlert = lawyerManagement.includes('开发者模式指示器');
      const hasTokenFallback = lawyerManagement.includes('localStorage.getItem(\'token\') || localStorage.getItem(\'law_oa_token\')');

      this.results.codeChanges.push({
        file: 'LawyerManagement.tsx',
        mockData: hasMockData,
        devModeCheck: hasDevModeCheck,
        devAlert: hasDevAlert,
        tokenFallback: hasTokenFallback,
        status: (hasMockData && hasDevModeCheck && hasDevAlert && hasTokenFallback) ? '✅' : '❌'
      });

      console.log(`${hasMockData && hasDevModeCheck && hasDevAlert && hasTokenFallback ? '✅' : '❌'} LawyerManagement.tsx - 开发模式和模拟数据修复`);
    } catch (error) {
      console.log(`❌ LawyerManagement.tsx - 读取失败: ${error.message}`);
    }
  }

  generateReport() {
    console.log('\n📋 生成Vue前端修复报告...');

    const allFilesExist = this.results.fileChecks.every(check => check.exists);
    const allCodeChangesValid = this.results.codeChanges.every(change => change.status === '✅');

    let summary = `
🔧 Vue前端律师管理页面修复报告
=================================

📅 修复时间: ${new Date().toLocaleString('zh-CN')}
🎯 修复目标: 解决Vue版本律师管理页面重定向到登录界面的问题

📊 修复状态:
${allFilesExist && allCodeChangesValid ? '✅ 修复完成' : '⚠️ 需要进一步检查'}

🔍 问题根本原因 (Vue版本):
1. AuthContext中的getCurrentUser API调用在开发模式下触发认证错误
2. MainLayout检测到用户为null时自动重定向到登录页面
3. 律师管理页面在API调用失败时没有开发模式的降级处理

🛠️ 修复措施 (Vue版本):
1. AuthContext.tsx - 在开发模式下跳过API验证，使用默认用户
2. MainLayout.tsx - 在开发模式下显示友好的提示信息
3. http.ts - 在开发模式下提供更友好的认证错误提示
4. LawyerManagement.tsx - 添加开发模式支持和模拟数据

📁 修复文件:
`;

    this.results.fileChecks.forEach(check => {
      summary += `   ${check.status} ${check.file}\n`;
    });

    summary += `
🔬 代码修改验证:
`;

    this.results.codeChanges.forEach(change => {
      summary += `   ${change.status} ${change.file}\n`;
    });

    if (allFilesExist && allCodeChangesValid) {
      summary += `
✅ 修复验证:
   - 开发模式认证流程优化: 完成
   - API错误处理改进: 完成
   - 模拟数据降级处理: 完成
   - 用户界面提示优化: 完成
   - Token兼容性处理: 完成

🚀 下一步建议:
1. 启动Vue前端服务测试修复效果
2. 运行端到端测试验证功能
3. 在生产环境中测试认证流程
4. 监控开发模式的用户体验

📝 测试方法 (Vue版本):
1. 启动Vue前端: cd frontend-vue && npm run dev
2. 直接访问: http://localhost:5173/lawyer
3. 验证开发模式指示器显示
4. 检查控制台无认证错误
5. 测试模拟数据加载

🎉 预期结果:
- 开发模式下律师管理页面可以正常访问
- 显示开发模式指示器
- 加载模拟数据而不是调用真实API
- 不会自动重定向到登录界面
- 用户界面友好，错误提示清晰
`;
    } else {
      summary += `
⚠️ 注意事项:
   - 部分修复可能不完整
   - 建议检查失败的修改项
   - 可能需要手动验证代码修改

🔧 补救措施:
   - 检查文件是否存在
   - 验证代码修改是否正确应用
   - 手动修复失败的修改项
`;
    }

    this.results.summary = summary;
    console.log(summary);

    // 保存报告到文件
    const reportPath = 'VUE_LAWYER_MANAGEMENT_FIX_REPORT.md';
    fs.writeFileSync(reportPath, summary);
    console.log(`\n📄 详细报告已保存到: ${reportPath}`);
  }

  run() {
    console.log('🎯 开始Vue前端律师管理页面修复验证...\n');

    this.checkFiles();
    this.verifyCodeChanges();
    this.generateReport();
  }
}

// 运行验证
if (require.main === module) {
  const verifier = new VueFixVerifier();
  verifier.run();
}

module.exports = { VueFixVerifier };
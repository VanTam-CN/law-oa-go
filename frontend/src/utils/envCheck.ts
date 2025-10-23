// 生产环境检查工具

export class EnvironmentChecker {
  /**
   * 检查后端服务连通性
   */
  static async checkBackendConnectivity(): Promise<{ success: boolean, message: string }> {
    try {
      const response = await fetch('/api/conflict-check/history', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        timeout: 5000
      } as RequestInit);
      
      if (response.ok) {
        return {
          success: true,
          message: '后端服务连接正常'
        };
      } else {
        return {
          success: false,
          message: `后端服务响应异常: ${response.status}`
        };
      }
    } catch (error) {
      return {
        success: false,
        message: `后端服务连接失败: ${error instanceof Error ? error.message : '未知错误'}`
      };
    }
  }
  
  /**
   * 检查必要的环境变量和配置
   */
  static checkConfiguration(): { success: boolean, issues: string[] } {
    const issues: string[] = [];
    
    // 检查是否为生产环境
    const isProduction = import.meta.env.PROD;
    if (!isProduction) {
      issues.push('当前不是生产环境');
    }
    
    // 检查API配置
    const apiBase = import.meta.env.VITE_API_BASE_URL;
    if (!apiBase && isProduction) {
      issues.push('生产环境缺少API基础URL配置');
    }
    
    // 检查认证Token
    const token = localStorage.getItem('auth_token');
    if (!token) {
      issues.push('缺少认证Token，可能影响API调用');
    }
    
    return {
      success: issues.length === 0,
      issues
    };
  }
  
  /**
   * 验证冲突检索功能的先决条件
   */
  static validateConflictCheckPrerequisites(): { valid: boolean, messages: string[] } {
    const messages: string[] = [];
    
    // 检查必要的服务
    const requiredServices = [
      'ConflictCheckService',
      'ApiClient'
    ];
    
    // 这里可以添加更多的运行时检查
    
    return {
      valid: messages.length === 0,
      messages
    };
  }
  
  /**
   * 生成环境报告
   */
  static async generateEnvironmentReport(): Promise<string> {
    const configCheck = this.checkConfiguration();
    const backendCheck = await this.checkBackendConnectivity();
    const prereqCheck = this.validateConflictCheckPrerequisites();
    
    const report = [
      '=== 环境检查报告 ===',
      `生成时间: ${new Date().toLocaleString()}`,
      '',
      '1. 配置检查:',
      configCheck.success ? '✅ 配置正常' : '❌ 配置存在问题',
      ...configCheck.issues.map(issue => `   - ${issue}`),
      '',
      '2. 后端连接:',
      backendCheck.success ? '✅ 连接正常' : '❌ 连接异常',
      `   ${backendCheck.message}`,
      '',
      '3. 功能先决条件:',
      prereqCheck.valid ? '✅ 条件满足' : '❌ 条件不满足',
      ...prereqCheck.messages.map(msg => `   - ${msg}`),
      '',
      '=== 报告结束 ==='
    ];
    
    return report.join('\n');
  }
}

// 开发工具：在控制台显示环境信息
if (import.meta.env.DEV) {
  (window as any).envChecker = EnvironmentChecker;
  console.log('开发工具: 使用 envChecker.generateEnvironmentReport() 检查环境状态');
}
import { message } from '@/utils/messageHelper';
import { post } from './api';

// 案件创建服务

// 案件数据验证服务
export class CaseValidationService {
  
  // 验证案件基本信息
  static validateBasicInfo(data: any): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    if (!data.caseName || data.caseName.trim().length === 0) {
      errors.push('案件名称不能为空');
    }
    
    if (!data.clientId) {
      errors.push('必须选择委托人');
    }
    
    if (!data.caseType) {
      errors.push('必须选择案件类型');
    }
    
    if (!data.causeOfAction || data.causeOfAction.trim().length === 0) {
      errors.push('必须选择或输入案由');
    }
    
    if (data.causeOfAction && data.causeOfAction.length > 100) {
      errors.push('案由不能超过100个字符');
    }
    
    if (data.caseName && data.caseName.length > 100) {
      errors.push('案件名称不能超过100个字符');
    }
    
    return {
      isValid: errors.length === 0,
      errors
    };
  }
  
  // 验证内部管理信息
  static validateManagementInfo(data: any): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    if (!data.leadLawyer) {
      errors.push('必须指定主办律师');
    }
    
    if (!data.billingMethod) {
      errors.push('必须选择收费方式');
    }
    
    if (!data.contractAmount || data.contractAmount <= 0) {
      errors.push('合同金额必须大于0');
    }
    
    if (data.contractAmount && data.contractAmount > 100000000) {
      errors.push('合同金额超出合理范围，请核实');
    }
    
    return {
      isValid: errors.length === 0,
      errors
    };
  }
  
  // 验证合规信息
  static validateComplianceInfo(data: any): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    if (!data.conflictCheck) {
      errors.push('必须完成利益冲突检查');
    }
    
    // 如果标记为高风险项目，需要额外验证
    if (data.isHighRisk) {
      if (!data.riskTags || data.riskTags.length === 0) {
        errors.push('高风险项目必须选择至少一个风险标签');
      }
    }
    
    return {
      isValid: errors.length === 0,
      errors
    };
  }
  
  // 验证文档信息
  static validateDocuments(data: any): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    if (!data.contractDocument || data.contractDocument.length === 0) {
      errors.push('必须上传委托代理合同');
    }
    
    return {
      isValid: errors.length === 0,
      errors
    };
  }
  
  // 综合验证
  static validateAll(data: any): { isValid: boolean; errors: string[] } {
    const basicValidation = this.validateBasicInfo(data);
    const managementValidation = this.validateManagementInfo(data);
    const complianceValidation = this.validateComplianceInfo(data);
    const documentValidation = this.validateDocuments(data);
    
    const allErrors = [
      ...basicValidation.errors,
      ...managementValidation.errors,
      ...complianceValidation.errors,
      ...documentValidation.errors
    ];
    
    return {
      isValid: allErrors.length === 0,
      errors: allErrors
    };
  }
}

// 案件工作流服务
export class CaseWorkflowService {
  
  // 根据案件信息确定审批流程
  static determineApprovalWorkflow(caseData: any): {
    requiresApproval: boolean;
    approvalLevel: 'JUNIOR' | 'SENIOR' | 'PARTNER';
    approvers: string[];
    reason: string;
  } {
    let requiresApproval = false;
    let approvalLevel: 'JUNIOR' | 'SENIOR' | 'PARTNER' = 'JUNIOR';
    let reason = '';
    
    // 高风险项目需要合伙人审批
    if (caseData.isHighRisk) {
      requiresApproval = true;
      approvalLevel = 'PARTNER';
      reason = '高风险项目需要合伙人审批';
    }
    
    // 高金额案件需要审批
    if (caseData.contractAmount > 5000000) {
      requiresApproval = true;
      approvalLevel = approvalLevel === 'PARTNER' ? 'PARTNER' : 'SENIOR';
      reason = reason || '高金额案件需要高级审批';
    }
    
    // 特定风险标签需要审批
    const highRiskTags = ['HIGH_VALUE', 'SENSITIVE', 'MEDIA_ATTENTION'];
    if (caseData.riskTags && caseData.riskTags.some((tag: string) => highRiskTags.includes(tag))) {
      requiresApproval = true;
      approvalLevel = 'PARTNER';
      reason = reason || '敏感案件需要合伙人审批';
    }
    
    // 模拟获取审批人列表
    const approvers = this.getApproversByLevel(approvalLevel);
    
    return {
      requiresApproval,
      approvalLevel,
      approvers,
      reason
    };
  }
  
  // 根据审批级别获取审批人
  private static getApproversByLevel(level: string): string[] {
    switch (level) {
      case 'PARTNER':
        return ['partner1@lawfirm.com', 'partner2@lawfirm.com'];
      case 'SENIOR':
        return ['senior1@lawfirm.com', 'senior2@lawfirm.com'];
      default:
        return ['junior1@lawfirm.com'];
    }
  }
  
  // 触发审批流程
  static async triggerApprovalWorkflow(caseId: string, caseData: any): Promise<boolean> {
    try {
      const workflow = this.determineApprovalWorkflow(caseData);
      
      if (workflow.requiresApproval) {
        // 这里应该调用实际的工作流API
        console.log('触发审批流程:', {
          caseId,
          level: workflow.approvalLevel,
          approvers: workflow.approvers,
          reason: workflow.reason
        });
        
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        message.info(`案件已提交${workflow.approvalLevel === 'PARTNER' ? '合伙人' : '高级'}审批`);
        return true;
      }
      
      return true;
    } catch (error) {
      console.error('工作流触发失败:', error);
      message.error('审批流程启动失败');
      return false;
    }
  }
  
  // 案件状态变更通知
  static async notifyStatusChange(caseId: string, oldStatus: string, newStatus: string, participants: string[]): Promise<void> {
    try {
      // 模拟发送通知
      console.log('发送状态变更通知:', {
        caseId,
        oldStatus,
        newStatus,
        participants
      });
      
      // 这里应该调用通知服务API
      await new Promise(resolve => setTimeout(resolve, 500));
      
    } catch (error) {
      console.error('通知发送失败:', error);
    }
  }
  
  // 生成案件编号
  static generateCaseNumber(caseType: string, clientId: string): string {
    const typePrefix = {
      'CIVIL': 'CIV',
      'COMMERCIAL': 'COM',
      'CRIMINAL': 'CRI',
      'ADMINISTRATIVE': 'ADM'
    }[caseType] || 'OTH';
    
    const year = new Date().getFullYear();
    const month = String(new Date().getMonth() + 1).padStart(2, '0');
    const random = Math.floor(Math.random() * 10000).toString().padStart(4, '0');
    
    return `${typePrefix}${year}${month}${random}`;
  }
  
  // 计算案件优先级
  static calculatePriority(caseData: any): 'HIGH' | 'MEDIUM' | 'LOW' {
    let score = 0;
    
    // 金额因素
    if (caseData.contractAmount > 10000000) score += 3;
    else if (caseData.contractAmount > 1000000) score += 2;
    else if (caseData.contractAmount > 100000) score += 1;
    
    // 风险因素
    if (caseData.isHighRisk) score += 3;
    if (caseData.riskTags && caseData.riskTags.length > 0) score += 1;
    
    // 案件类型因素
    if (caseData.caseType === 'CRIMINAL') score += 2;
    if (caseData.caseType === 'COMMERCIAL') score += 1;
    
    if (score >= 5) return 'HIGH';
    if (score >= 2) return 'MEDIUM';
    return 'LOW';
  }
}

// 案件创建服务
export class CaseCreationService {
  
  // 创建案件的完整流程
  static async createCase(formData: any): Promise<{ success: boolean; caseId?: string; message: string }> {
    try {
      // 1. 数据验证
      const validation = CaseValidationService.validateAll(formData);
      if (!validation.isValid) {
        return {
          success: false,
          message: `数据验证失败: ${validation.errors.join(', ')}`
        };
      }
      
      // 2. 生成案件编号
      const caseNumber = CaseWorkflowService.generateCaseNumber(formData.caseType, formData.clientId);
      
      // 3. 计算优先级
      const priority = CaseWorkflowService.calculatePriority(formData);
      
      // 4. 准备案件数据
      const caseData = {
        ...formData,
        caseNo: caseNumber,
        priority,
        status: 'PENDING_APPROVAL',
        createTime: new Date().toISOString(),
        updateTime: new Date().toISOString()
      };
      
      // 5. 保存案件到数据库（模拟）
      const caseId = await this.saveCaseToDatabase(caseData);
      
      // 6. 触发工作流
      const workflowSuccess = await CaseWorkflowService.triggerApprovalWorkflow(caseId, caseData);
      if (!workflowSuccess) {
        return {
          success: false,
          message: '工作流启动失败'
        };
      }
      
      // 7. 发送通知
      await CaseWorkflowService.notifyStatusChange(
        caseId,
        '',
        'PENDING_APPROVAL',
        [formData.leadLawyer, ...formData.assistingLawyers || []]
      );
      
      return {
        success: true,
        caseId,
        message: '案件创建成功'
      };
      
    } catch (error) {
      console.error('案件创建失败:', error);
      return {
        success: false,
        message: '案件创建失败，请稍后重试'
      };
    }
  }
  
  // 保存案件到数据库
  private static async saveCaseToDatabase(caseData: any): Promise<string> {
    try {
      // 准备API所需的数据格式
      const apiData = {
        title: caseData.caseName || caseData.title,
        description: caseData.description || caseData.caseDescription,
        client_id: parseInt(caseData.clientId),
        lawyer_id: parseInt(caseData.leadLawyer),
        case_type: caseData.caseType?.toLowerCase() || 'civil',
        priority: caseData.priority?.toLowerCase() || 'medium'
      };

      console.log('正在保存案件到数据库:', apiData);

      // 调用真实的后端API
      const response = await post('/cases', apiData);

      if (response && response.data) {
        const caseId = response.data.id.toString();
        console.log('案件保存成功:', { caseId, ...response.data });
        message.success('案件创建成功');
        return caseId;
      } else {
        console.error('API响应格式错误:', response);
        throw new Error('API响应格式错误');
      }

    } catch (error) {
      console.error('案件保存失败:', error);
      const errorMessage = error instanceof Error ? error.message : '案件保存失败，请稍后重试';
      message.error(errorMessage);
      throw new Error(errorMessage);
    }
  }
}

export default {
  CaseValidationService,
  CaseWorkflowService,
  CaseCreationService
};
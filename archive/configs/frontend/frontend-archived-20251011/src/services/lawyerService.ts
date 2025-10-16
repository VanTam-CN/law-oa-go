import apiClient from "./api";

// 律师接口定义
export interface Lawyer {
  id?: number;
  name?: string;
  username?: string;
  phone?: string;
  email?: string;
  licenseNumber?: string;
  department?: string;
  position?: string;
  status?: string;
  specialty?: string[];
  experience?: number;
  gender?: string;
  joinDate?: string;
  profile?: string;
  avatar?: string;
  address?: string;
  education?: string;
  achievements?: string[];
  hourlyRate?: number;
  consultationHours?: string;
  caseCount?: number;
  successRate?: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface LawyerListRequest {
  page?: number;
  pageSize?: number;
  name?: string;
  department?: string;
  status?: string;
  specialty?: string;
}

export interface LawyerListResponse {
  data: Lawyer[];
  total: number;
  page: number;
  pageSize: number;
}

export interface LawyerStats {
  total: number;
  active: number;
  inactive: number;
  onLeave: number;
}

export interface CreateLawyerRequest {
  name: string;
  username: string;
  phone: string;
  email: string;
  licenseNumber: string;
  department?: string;
  position?: string;
  status?: string;
  specialty?: string[];
  experience?: number;
  gender?: string;
  profile?: string;
  address?: string;
  education?: string;
  achievements?: string[];
  hourlyRate?: number;
  consultationHours?: string;
}

export interface UpdateLawyerRequest {
  name?: string;
  phone?: string;
  email?: string;
  licenseNumber?: string;
  department?: string;
  position?: string;
  status?: string;
  specialty?: string[];
  experience?: number;
  profile?: string;
  address?: string;
  education?: string;
  achievements?: string[];
  hourlyRate?: number;
  consultationHours?: string;
}

class LawyerService {
  // 获取律师列表
  async getLawyers(params: LawyerListRequest): Promise<LawyerListResponse> {
    try {
      // 由于目前使用模拟数据，这里返回模拟的响应
      console.log('🛠️ 开发模式：调用律师列表API', params);

      // 模拟网络延迟
      await new Promise(resolve => setTimeout(resolve, 500));

      // 模拟数据
      const mockLawyers: Lawyer[] = [
        {
          id: 1,
          name: '张律师',
          username: 'zhang_lawyer',
          phone: '13800138001',
          email: 'zhang.lawyer@example.com',
          licenseNumber: '123456789012345',
          department: '民事诉讼部',
          position: '合伙人',
          status: 'active',
          specialty: ['合同纠纷', '侵权责任'],
          experience: 15,
          gender: 'male',
          joinDate: '2010-01-15',
          profile: '资深律师，专注于民事诉讼领域',
          address: '北京市朝阳区建国门外大街1号',
          education: '中国政法大学 法学硕士',
          achievements: ['2023年度优秀律师', '处理案件成功率95%'],
          hourlyRate: 800,
          consultationHours: '周一至周五 9:00-18:00',
          caseCount: 156,
          successRate: 95,
          createdAt: '2024-01-15T10:30:00Z',
          updatedAt: '2025-10-09T14:20:00Z'
        },
        {
          id: 2,
          name: '李律师',
          username: 'li_lawyer',
          phone: '13800138002',
          email: 'li.lawyer@example.com',
          licenseNumber: '123456789012346',
          department: '刑事辩护部',
          position: '合伙人',
          status: 'active',
          specialty: ['刑事辩护', '知识产权'],
          experience: 12,
          gender: 'male',
          joinDate: '2012-03-20',
          profile: '专业刑事辩护律师',
          address: '北京市海淀区中关村大街2号',
          education: '北京大学 法学博士',
          achievements: ['2022年度刑事辩护奖', '无罪辩护成功率90%'],
          hourlyRate: 1000,
          consultationHours: '周一至周五 10:00-19:00',
          caseCount: 128,
          successRate: 90,
          createdAt: '2024-01-15T10:30:00Z',
          updatedAt: '2025-10-09T14:20:00Z'
        },
        {
          id: 3,
          name: '王律师',
          username: 'wang_lawyer',
          phone: '13800138003',
          email: 'wang.lawyer@example.com',
          licenseNumber: '123456789012347',
          department: '公司法务部',
          position: '资深律师',
          status: 'active',
          specialty: ['公司法务', '劳动争议'],
          experience: 8,
          gender: 'female',
          joinDate: '2016-06-10',
          profile: '公司法务专家',
          address: '北京市东城区王府井大街3号',
          education: '清华大学 法学硕士',
          achievements: ['2023年度公司法务专家'],
          hourlyRate: 600,
          consultationHours: '周一至周五 9:00-17:00',
          caseCount: 89,
          successRate: 92,
          createdAt: '2024-01-15T10:30:00Z',
          updatedAt: '2025-10-09T14:20:00Z'
        },
        {
          id: 4,
          name: '赵律师',
          username: 'zhao_lawyer',
          phone: '13800138004',
          email: 'zhao.lawyer@example.com',
          licenseNumber: '123456789012348',
          department: '行政诉讼部',
          position: '律师',
          status: 'on_leave',
          specialty: ['行政诉讼', '行政复议'],
          experience: 5,
          gender: 'male',
          joinDate: '2019-09-01',
          profile: '行政诉讼专业律师',
          address: '北京市西城区金融大街4号',
          education: '中国人民大学 法学硕士',
          achievements: ['2021年度新锐律师'],
          hourlyRate: 500,
          consultationHours: '暂时休假',
          caseCount: 45,
          successRate: 88,
          createdAt: '2024-01-15T10:30:00Z',
          updatedAt: '2025-10-09T14:20:00Z'
        }
      ];

      // 过滤数据
      let filteredLawyers = mockLawyers;

      if (params.name) {
        filteredLawyers = filteredLawyers.filter(lawyer =>
          lawyer.name?.toLowerCase().includes(params.name!.toLowerCase())
        );
      }

      if (params.department) {
        filteredLawyers = filteredLawyers.filter(lawyer =>
          lawyer.department === params.department
        );
      }

      if (params.status) {
        filteredLawyers = filteredLawyers.filter(lawyer =>
          lawyer.status === params.status
        );
      }

      if (params.specialty) {
        filteredLawyers = filteredLawyers.filter(lawyer =>
          lawyer.specialty?.some(s => s.toLowerCase().includes(params.specialty!.toLowerCase()))
        );
      }

      // 分页
      const page = params.page || 1;
      const pageSize = params.pageSize || 10;
      const startIndex = (page - 1) * pageSize;
      const endIndex = startIndex + pageSize;
      const paginatedLawyers = filteredLawyers.slice(startIndex, endIndex);

      return {
        data: paginatedLawyers,
        total: filteredLawyers.length,
        page: page,
        pageSize: pageSize
      };
    } catch (error: any) {
      console.error("获取律师列表失败:", error);
      throw error;
    }
  }

  // 获取律师详情
  async getLawyerById(id: number): Promise<Lawyer> {
    try {
      console.log(`🛠️ 开发模式：调用律师详情API (ID: ${id})`);

      // 模拟网络延迟
      await new Promise(resolve => setTimeout(resolve, 300));

      // 查找模拟数据
      const lawyers = await this.getLawyers({ pageSize: 100 });
      const lawyer = lawyers.data.find(l => l.id === id);

      if (!lawyer) {
        throw new Error('律师不存在');
      }

      return lawyer;
    } catch (error: any) {
      console.error("获取律师详情失败:", error);
      throw error;
    }
  }

  // 创建律师
  async createLawyer(lawyer: CreateLawyerRequest): Promise<Lawyer> {
    try {
      console.log('🛠️ 开发模式：调用创建律师API', lawyer);

      // 模拟网络延迟
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 返回模拟的创建结果
      const newLawyer: Lawyer = {
        id: Math.floor(Math.random() * 1000) + 100, // 模拟新ID
        ...lawyer,
        caseCount: 0,
        successRate: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      return newLawyer;
    } catch (error: any) {
      console.error("创建律师失败:", error);
      throw error;
    }
  }

  // 更新律师
  async updateLawyer(id: number, lawyer: UpdateLawyerRequest): Promise<Lawyer> {
    try {
      console.log(`🛠️ 开发模式：调用更新律师API (ID: ${id})`, lawyer);

      // 模拟网络延迟
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 获取原有数据并合并
      const existingLawyer = await this.getLawyerById(id);
      const updatedLawyer: Lawyer = {
        ...existingLawyer,
        ...lawyer,
        updatedAt: new Date().toISOString()
      };

      return updatedLawyer;
    } catch (error: any) {
      console.error("更新律师失败:", error);
      throw error;
    }
  }

  // 删除律师
  async deleteLawyer(id: number): Promise<void> {
    try {
      console.log(`🛠️ 开发模式：调用删除律师API (ID: ${id})`);

      // 模拟网络延迟
      await new Promise(resolve => setTimeout(resolve, 500));

      // 模拟删除操作
      console.log(`律师 ${id} 删除成功`);
    } catch (error: any) {
      console.error("删除律师失败:", error);
      throw error;
    }
  }

  // 获取律师统计
  async getLawyerStats(): Promise<LawyerStats> {
    try {
      console.log('🛠️ 开发模式：调用律师统计API');

      // 模拟网络延迟
      await new Promise(resolve => setTimeout(resolve, 300));

      // 获取律师数据并计算统计
      const lawyers = await this.getLawyers({ pageSize: 100 });
      const stats: LawyerStats = {
        total: lawyers.total,
        active: lawyers.data.filter(l => l.status === 'active').length,
        inactive: lawyers.data.filter(l => l.status === 'inactive').length,
        onLeave: lawyers.data.filter(l => l.status === 'on_leave').length
      };

      return stats;
    } catch (error: any) {
      console.error("获取律师统计失败:", error);
      throw error;
    }
  }
}

export default new LawyerService();
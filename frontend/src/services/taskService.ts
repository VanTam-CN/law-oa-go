import apiClient from "./api";
import { AppError } from "../types/errors";

interface Task {
  id: number;
  title: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assignee_id?: number;
  assignee_name?: string;
  creator_id: number;
  creator_name: string;
  due_date?: string;
  created_at: string;
  updated_at: string;
  case_id?: number;
  client_id?: number;
}

interface TaskListRequest {
  page?: number;
  page_size?: number;
  status?: string;
  priority?: string;
  assignee_id?: number;
  creator_id?: number;
  case_id?: number;
  client_id?: number;
  search?: string;
  due_soon?: boolean;
  overdue?: boolean;
}

interface CreateTaskRequest {
  title: string;
  description: string;
  status?: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  assignee_id?: number;
  due_date?: string;
  case_id?: number;
  client_id?: number;
}

interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  assignee_id?: number;
  due_date?: string;
  case_id?: number;
  client_id?: number;
}

interface TaskStats {
  total: number;
  pending: number;
  in_progress: number;
  completed: number;
  cancelled: number;
  overdue: number;
  due_soon: number;
  by_priority: {
    low: number;
    medium: number;
    high: number;
    urgent: number;
  };
  by_assignee: Array<{
    assignee_id: number;
    assignee_name: string;
    count: number;
  }>;
}

class TaskService {
  // 获取任务列表
  async getTasks(params?: TaskListRequest): Promise<{
    data: Task[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Task>("/tasks", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取任务列表失败:", error);
      throw new AppError(
        error.message || "获取任务列表失败",
        error.code || "GET_TASKS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取任务详情
  async getTask(id: number): Promise<Task> {
    try {
      return await apiClient.get<Task>(`/tasks/${id}`, {
        
      });
    } catch (error: any) {
      console.error("获取任务详情失败:", error);
      throw new AppError(
        error.message || "获取任务详情失败",
        error.code || "GET_TASK_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 创建任务
  async createTask(data: CreateTaskRequest): Promise<Task> {
    try {
      return await apiClient.post<Task>("/tasks", data, {
      });
    } catch (error: any) {
      console.error("创建任务失败:", error);
      throw new AppError(
        error.message || "创建任务失败",
        error.code || "CREATE_TASK_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新任务
  async updateTask(id: number, data: UpdateTaskRequest): Promise<Task> {
    try {
      return await apiClient.put<Task>(`/tasks/${id}`, data, {
      });
    } catch (error: any) {
      console.error("更新任务失败:", error);
      throw new AppError(
        error.message || "更新任务失败",
        error.code || "UPDATE_TASK_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 删除任务
  async deleteTask(id: number): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/tasks/${id}`, {
      });
    } catch (error: any) {
      console.error("删除任务失败:", error);
      throw new AppError(
        error.message || "删除任务失败",
        error.code || "DELETE_TASK_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取任务统计信息
  async getTaskStats(): Promise<TaskStats> {
    try {
      return await apiClient.get<TaskStats>("/tasks/stats", {
        
      });
    } catch (error: any) {
      console.error("获取任务统计信息失败:", error);
      throw new AppError(
        error.message || "获取任务统计信息失败",
        error.code || "GET_TASK_STATS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 分配任务
  async assignTask(
    taskId: number,
    assigneeId: number,
  ): Promise<{ message: string }> {
    try {
      return await apiClient.post<{ message: string }>(
        `/tasks/${taskId}/assign`,
        { assignee_id: assigneeId },
      );
    } catch (error: any) {
      console.error("分配任务失败:", error);
      throw new AppError(
        error.message || "分配任务失败",
        error.code || "ASSIGN_TASK_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新任务状态
  async updateTaskStatus(
    taskId: number,
    status: string,
  ): Promise<{ message: string }> {
    try {
      return await apiClient.post<{ message: string }>(
        `/tasks/${taskId}/status`,
        { status },
      );
    } catch (error: any) {
      console.error("更新任务状态失败:", error);
      throw new AppError(
        error.message || "更新任务状态失败",
        error.code || "UPDATE_TASK_STATUS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取我的任务
  async getMyTasks(params?: Omit<TaskListRequest, "assignee_id">): Promise<{
    data: Task[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Task>("/tasks/my", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取我的任务失败:", error);
      throw new AppError(
        error.message || "获取我的任务失败",
        error.code || "GET_MY_TASKS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取案件任务
  async getCaseTasks(
    caseId: number,
    params?: Omit<TaskListRequest, "case_id">,
  ): Promise<{
    data: Task[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Task>(`/cases/${caseId}/tasks`, {
        params,
        
      });
    } catch (error: any) {
      console.error("获取案件任务失败:", error);
      throw new AppError(
        error.message || "获取案件任务失败",
        error.code || "GET_CASE_TASKS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取客户任务
  async getClientTasks(
    client_id: number,
    params?: Omit<TaskListRequest, "client_id">,
  ): Promise<{
    data: Task[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Task>(`/clients/${client_id}/tasks`, {
        params,
        
      });
    } catch (error: any) {
      console.error("获取客户任务失败:", error);
      throw new AppError(
        error.message || "获取客户任务失败",
        error.code || "GET_CLIENT_TASKS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索任务
  async searchTasks(
    query: string,
    params?: Omit<TaskListRequest, "search">,
  ): Promise<{
    data: Task[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      const searchParams = {
        ...params,
        search: query,
      };
      return await apiClient.getPaginated<Task>("/tasks/search", {
        params: searchParams,
        
      });
    } catch (error: any) {
      console.error("搜索任务失败:", error);
      throw new AppError(
        error.message || "搜索任务失败",
        error.code || "SEARCH_TASKS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 批量更新任务状态
  async batchUpdateStatus(
    taskIds: number[],
    status: string,
  ): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/tasks/batch/status",
        { task_ids: taskIds, status },
      );
    } catch (error: any) {
      console.error("批量更新任务状态失败:", error);
      throw new AppError(
        error.message || "批量更新任务状态失败",
        error.code || "BATCH_UPDATE_STATUS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量分配任务
  async batchAssignTasks(
    taskIds: number[],
    assigneeId: number,
  ): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/tasks/batch/assign",
        { task_ids: taskIds, assignee_id: assigneeId },
      );
    } catch (error: any) {
      console.error("批量分配任务失败:", error);
      throw new AppError(
        error.message || "批量分配任务失败",
        error.code || "BATCH_ASSIGN_TASKS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量删除任务
  async batchDeleteTasks(taskIds: number[]): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/tasks/batch/delete",
        { task_ids: taskIds },
      );
    } catch (error: any) {
      console.error("批量删除任务失败:", error);
      throw new AppError(
        error.message || "批量删除任务失败",
        error.code || "BATCH_DELETE_TASKS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取任务类型列表
  async getTaskTypes(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
        }>
      >("/tasks/types", {
        
      });
    } catch (error: any) {
      console.error("获取任务类型列表失败:", error);
      throw new AppError(
        error.message || "获取任务类型列表失败",
        error.code || "GET_TASK_TYPES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取任务优先级列表
  async getTaskPriorities(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
      color: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
          color: string;
        }>
      >("/tasks/priorities", {
        
      });
    } catch (error: any) {
      console.error("获取任务优先级列表失败:", error);
      throw new AppError(
        error.message || "获取任务优先级列表失败",
        error.code || "GET_TASK_PRIORITIES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取任务状态列表
  async getTaskStatuses(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
      color: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
          color: string;
        }>
      >("/tasks/statuses", {
        
      });
    } catch (error: any) {
      console.error("获取任务状态列表失败:", error);
      throw new AppError(
        error.message || "获取任务状态列表失败",
        error.code || "GET_TASK_STATUSES_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const taskService = new TaskService();

// 为了向后兼容，也导出独立的函数
export const getTasks = (params?: TaskListRequest) => taskService.getTasks(params);
export const getTask = (id: number) => taskService.getTask(id);
export const createTask = (data: CreateTaskRequest) => taskService.createTask(data);
export const updateTask = (id: number, data: UpdateTaskRequest) => taskService.updateTask(id, data);
export const deleteTask = (id: number) => taskService.deleteTask(id);
export const getTaskStats = () => taskService.getTaskStats();
export const assignTask = (taskId: number, assigneeId: number) => taskService.assignTask(taskId, assigneeId);
export const updateTaskStatus = (taskId: number, status: string) => taskService.updateTaskStatus(taskId, status);
export const batchDeleteTasks = (taskIds: number[]) => taskService.batchDeleteTasks(taskIds);
export const batchUpdateStatus = (taskIds: number[], status: string) => taskService.batchUpdateStatus(taskIds, status);
export const batchAssignTasks = (taskIds: number[], assigneeId: number) => taskService.batchAssignTasks(taskIds, assigneeId);
export const getMyTasks = (params?: Omit<TaskListRequest, "assignee_id">) => taskService.getMyTasks(params);
export const getCaseTasks = (caseId: number, params?: Omit<TaskListRequest, "case_id">) => taskService.getCaseTasks(caseId, params);
export const getClientTasks = (client_id: number, params?: Omit<TaskListRequest, "client_id">) => taskService.getClientTasks(client_id, params);
export const searchTasks = (query: string, params?: Omit<TaskListRequest, "search">) => taskService.searchTasks(query, params);
export const getTaskTypes = () => taskService.getTaskTypes();
export const getTaskPriorities = () => taskService.getTaskPriorities();
export const getTaskStatuses = () => taskService.getTaskStatuses();

export default taskService;

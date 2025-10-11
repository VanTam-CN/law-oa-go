import { AppError } from './errors';

// API统一响应格式
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: string;
    suggestions?: string[];
  };
  meta?: {
    timestamp: string;
    request_id: string;
    version: string;
  };
}

// 分页响应格式
export interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  meta?: {
    timestamp: string;
    request_id: string;
    version: string;
  };
}

// 认证相关类型
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  phone?: string;
  role?: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: UserProfile;
}

export interface RefreshTokenRequest {
  token: string;
}

export interface RefreshTokenResponse {
  token: string;
  expires_at: string;
}

// 用户相关类型
export interface UserProfile {
  id: number;
  name: string;
  email: string;
  role: string;
  status: string;
  phone?: string;
  avatar?: string;
  created_at: string;
  updated_at: string;
}

export interface UserListRequest {
  page?: number;
  page_size?: number;
  role?: string;
  status?: string;
  search?: string;
}

export interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
  phone?: string;
  role: string;
  status?: string;
}

export interface UpdateUserRequest {
  name?: string;
  email?: string;
  phone?: string;
  role?: string;
  status?: string;
}

export interface UserListResponse {
  data: UserProfile[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// 客户相关类型
export interface Client {
  id: number;
  name: string;
  email: string;
  phone: string;
  address: string;
  company: string;
  type: "personal" | "company";
  id_card?: string;
  industry?: string;
  contact_person?: string;
  contact_phone?: string;
  source?: string;
  notes?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface ClientListRequest {
  page?: number;
  page_size?: number;
  status?: string;
  search?: string;
  type?: "personal" | "company";
  source?: string;
  industry?: string;
}

export interface CreateClientRequest {
  name: string;
  email: string;
  phone: string;
  address: string;
  company: string;
  type: "personal" | "company";
  id_card?: string;
  industry?: string;
  contact_person?: string;
  contact_phone?: string;
  source?: string;
  notes?: string;
}

export interface UpdateClientRequest {
  name?: string;
  email?: string;
  phone?: string;
  address?: string;
  company?: string;
  type?: "personal" | "company";
  id_card?: string;
  industry?: string;
  contact_person?: string;
  contact_phone?: string;
  source?: string;
  notes?: string;
  status?: string;
}

export interface ClientStats {
  total: number;
  active: number;
  inactive: number;
  by_status: {
    active: number;
    inactive: number;
  };
  created_this_month: number;
  updated_this_month: number;
}

// 文件相关类型
export interface FileInfo {
  id: string;
  name: string;
  originalName?: string;
  size: number;
  contentType: string;
  category: string;
  categoryName?: string;
  description: string;
  uploadTime: string;
  uploadPath: string;
  type?: string;
  path?: string;
  url?: string;
  lastModified?: number;
  status?: string;
  created_at?: string;
}

export interface FileListRequest {
  page?: number;
  page_size?: number;
  category?: string;
  search?: string;
  file_type?: string;
}

export interface FileListResponse {
  data: FileInfo[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export interface FileStats {
  total_files: number;
  total_size: number;
  by_category: {
    document: number;
    image: number;
    spreadsheet: number;
    other: number;
  };
  by_type: Record<string, number>;
}

export interface UploadFileRequest {
  file: File;
  category: string;
  description?: string;
}

export interface FileUploadResponse {
  id: string;
  name: string;
  originalName: string;
  size: number;
  contentType: string;
  category: string;
  description: string;
  uploadTime: string;
  uploadPath: string;
  url: string;
}

// 文件类型定义
export type FileCategory = "document" | "image" | "spreadsheet" | "other";
export type FileType =
  | "pdf"
  | "word"
  | "excel"
  | "powerpoint"
  | "image"
  | "text"
  | "other";

// 通知相关类型
export interface Notification {
  id: number;
  title: string;
  message: string;
  type: "info" | "warning" | "error" | "success" | "case_update" | "client_update" | "system" | "reminder" | "deadline" | "document" | "message";
  read: boolean;
  created_at: string;
  updated_at: string;
  relatedEntity?: {
    id: number;
    type: "case" | "client" | "document";
    name: string;
  };
}

export interface NotificationListResponse {
  data: Notification[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// 消息相关类型
export interface Message {
  id: number;
  sender_id: number;
  sender_name: string;
  recipient_id: number;
  recipient_name: string;
  subject: string;
  content: string;
  type: "inbox" | "sent" | "draft";
  read: boolean;
  created_at: string;
  updated_at: string;
}

export interface MessageListRequest {
  page?: number;
  page_size?: number;
  type?: "inbox" | "sent" | "draft";
  search?: string;
  status?: "read" | "unread";
}

export interface SendMessageRequest {
  recipient_id: number;
  subject: string;
  content: string;
  attachments?: File[];
}

export interface UpdateMessageRequest {
  subject?: string;
  content?: string;
  read?: boolean;
}

export interface MessageListResponse {
  data: Message[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// 报告相关类型
export interface Report {
  id: number;
  title: string;
  description: string;
  type: string;
  parameters: Record<string, any>;
  format: "pdf" | "excel" | "csv";
  status: "pending" | "generating" | "completed" | "failed" | "generated";
  created_at: string;
  updated_at: string;
  file_url?: string;
  startDate?: string;
  endDate?: string;
  size?: number;
}

export interface ReportListResponse {
  data: Report[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export interface CreateReportRequest {
  title: string;
  description: string;
  type: string;
  parameters: Record<string, any>;
  format: "pdf" | "excel" | "csv";
}

export interface UpdateReportRequest {
  title?: string;
  description?: string;
  type?: string;
  parameters?: Record<string, any>;
  format?: "pdf" | "excel" | "csv";
}

export interface ReportListRequest {
  page?: number;
  page_size?: number;
  search?: string;
  type?: string;
  status?: string;
  created_by?: number;
  start_date?: string;
  end_date?: string;
}

// 案件相关类型
export interface Case {
  id: number;
  case_number?: string;
  title: string;
  description: string;
  client_id: number;
  client_name?: string;
  lawyer_id: number;
  lawyer_name?: string;
  case_type: string;
  priority: string;
  status: string;
  start_date?: string;
  end_date?: string | null;
  created_at: string;
  updated_at: string;
  case_amount?: number;
  expected_end_date?: string;
  principal_info?: string;
  opponent_info?: string;
  client?: {
    id: number;
    name: string;
    email?: string;
    phone?: string;
    address?: string;
    company?: string;
    status?: string;
  };
  lawyer?: {
    id: number;
    name: string;
    email?: string;
    role?: string;
    phone?: string;
    status?: string;
  };
  conflict_check_result?: {
    has_conflict: boolean;
    risk_level?: string;
    details?: any;
  };
}

export interface CaseListRequest {
  page?: number;
  page_size?: number;
  status?: string;
  case_type?: string;
  priority?: string;
  search?: string;
  client_id?: number;
  lawyer_id?: number;
  // 后端不支持的字段暂时保留，但不在API调用中使用
  start_date?: string;
  end_date?: string;
  urgent_only?: boolean;
  my_cases?: boolean;
  recent_only?: boolean;
  high_priority_only?: boolean;
}

export interface CreateCaseRequest {
  title: string;
  description: string;
  client_id: number;
  lawyer_id?: number;
  case_type: string;
  priority: string;
  status?: string;
  // 后端暂时不支持的字段，保留但不用于API调用
  start_date?: string;
  expected_end_date?: string;
  case_amount?: number;
  principal_info?: string;
  opponent_info?: string;
  skip_conflict_check?: boolean;
}

export interface UpdateCaseRequest {
  title?: string;
  description?: string;
  lawyer_id?: number;
  case_type?: string;
  priority?: string;
  status?: string;
  // 后端暂时不支持的字段，保留但不用于API调用
  client_id?: number;
  start_date?: string;
  end_date?: string;
  expected_end_date?: string;
  case_amount?: number;
  principal_info?: string;
  opponent_info?: string;
}

export interface AssignLawyerRequest {
  lawyer_id: number;
}

export interface UpdateCaseStatusRequest {
  status: string;
}

// 统计数据类型
export interface CaseStats {
  total: number;
  pending: number;
  active: number;
  closed: number;
  by_type: {
    civil: number;
    commercial: number;
    criminal: number;
    administrative: number;
  };
  by_priority: {
    low: number;
    medium: number;
    high: number;
    urgent: number;
  };
  by_status: {
    pending: number;
    active: number;
    closed: number;
  };
}

// 错误类型
export interface ApiError {
  code: string;
  message: string;
  details?: string;
  suggestions?: string[];
}

// 通用请求参数
export interface PaginationParams {
  page?: number;
  page_size?: number;
}

export interface SearchParams {
  search?: string;
}

export interface FilterParams {
  status?: string;
  role?: string;
  case_type?: string;
  priority?: string;
}

// 用户设置相关类型
export interface UserSettings {
  id: number;
  user_id: number;
  language: 'zh-CN' | 'en-US' | 'zh-TW';
  theme: 'light' | 'dark' | 'auto';
  timezone: string;
  date_format: 'YYYY-MM-DD' | 'MM/DD/YYYY' | 'DD/MM/YYYY';
  time_format: '24h' | '12h';
  notifications: {
    email: boolean;
    push: boolean;
    sms: boolean;
    case_updates: boolean;
    client_updates: boolean;
    system_announcements: boolean;
  };
  privacy: {
    profile_visibility: 'public' | 'private' | 'contacts';
    activity_tracking: boolean;
    data_sharing: boolean;
  };
  preferences: {
    items_per_page: number;
    default_view: 'list' | 'grid' | 'table';
    auto_save: boolean;
    confirm_actions: boolean;
    keyboard_shortcuts: boolean;
  };
  created_at: string;
  updated_at: string;
}

export interface UpdateUserSettingsRequest {
  language?: 'zh-CN' | 'en-US' | 'zh-TW';
  theme?: 'light' | 'dark' | 'auto';
  timezone?: string;
  date_format?: 'YYYY-MM-DD' | 'MM/DD/YYYY' | 'DD/MM/YYYY';
  time_format?: '24h' | '12h';
  notifications?: Partial<UserSettings['notifications']>;
  privacy?: Partial<UserSettings['privacy']>;
  preferences?: Partial<UserSettings['preferences']>;
}

// 任务相关类型
export interface Task {
  id: number;
  title: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assigned_to?: number;
  assigned_to_name?: string;
  assigned_by?: number;
  due_date?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  case_id?: number;
  client_id?: number;
  tags?: string[];
}

export interface TaskListRequest {
  page?: number;
  page_size?: number;
  status?: string;
  priority?: string;
  assigned_to?: number;
  case_id?: number;
  client_id?: number;
  search?: string;
  due_soon?: boolean;
  overdue?: boolean;
}

export interface CreateTaskRequest {
  title: string;
  description: string;
  status?: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  assigned_to?: number;
  due_date?: string;
  case_id?: number;
  client_id?: number;
  tags?: string[];
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  assigned_to?: number;
  due_date?: string;
  case_id?: number;
  client_id?: number;
  tags?: string[];
}

export interface TaskListResponse {
  data: Task[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// 导出AppError类
export { AppError } from './errors';

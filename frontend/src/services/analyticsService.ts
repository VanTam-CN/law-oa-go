import { apiClient } from './api';

// 用户行为分析相关的API服务

// 用户会话相关接口
export interface UserSession {
  id: string;
  user_id: string;
  ip_address?: string;
  user_agent?: string;
  start_time: string;
  end_time?: string;
  duration?: number;
  is_active: boolean;
  page_views: number;
  last_active: string;
  referrer?: string;
  source?: string;
  campaign?: string;
  device_type?: string;
  platform?: string;
  browser?: string;
  location?: GeoLocation;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface GeoLocation {
  country?: string;
  country_code?: string;
  region?: string;
  city?: string;
  latitude?: number;
  longitude?: number;
  timezone?: string;
}

// 页面浏览相关接口
export interface PageView {
  id: string;
  session_id: string;
  user_id: string;
  url: string;
  path: string;
  title?: string;
  referrer?: string;
  timestamp: string;
  duration?: number;
  scroll_depth?: number;
  viewport_size?: string;
  screen_size?: string;
  interaction?: string;
  is_bounce?: boolean;
  exit_page?: boolean;
  entry_page?: boolean;
  properties?: Record<string, any>;
  created_at: string;
}

// 用户事件相关接口
export interface UserEvent {
  id: string;
  session_id: string;
  user_id: string;
  event_type: string;
  event_category: string;
  event_action: string;
  event_label?: string;
  event_value?: number;
  url?: string;
  element?: string;
  properties?: Record<string, any>;
  timestamp: string;
  created_at: string;
}

// 用户旅程相关接口
export interface JourneyStep {
  step_name: string;
  step_order: number;
  step_type: string;
  url?: string;
  completed: boolean;
  completed_at?: string;
  time_spent?: number;
}

export interface UserJourney {
  id: string;
  user_id: string;
  journey_type: string;
  start_time: string;
  end_time?: string;
  steps?: JourneyStep[];
  current_step: number;
  is_completed: boolean;
  properties?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// 行为模式相关接口
export interface BehaviorPattern {
  id: string;
  user_id: string;
  pattern_type: string;
  pattern_name: string;
  description?: string;
  confidence: number;
  frequency: number;
  pattern_data?: Record<string, any>;
  last_detected: string;
  created_at: string;
}

// 用户行为分析结果接口
export interface UserSessionSummary {
  total_sessions: number;
  active_sessions: number;
  avg_duration: number;
  max_duration: number;
  first_session: string;
  last_session: string;
  total_page_views: number;
  unique_pages: number;
  avg_page_duration: number;
  total_events: number;
  click_events: number;
  form_events: number;
}

export interface AnalysisPeriod {
  start_date: string;
  end_date: string;
}

export interface UserBehaviorAnalysis {
  user_id: string;
  analysis_period: AnalysisPeriod;
  session_summary?: UserSessionSummary;
  behavior_patterns?: BehaviorPattern[];
  page_view_stats?: Array<Record<string, any>>;
  event_stats?: Array<Record<string, any>>;
  generated_at: string;
}

// 实时统计接口
export interface RealTimeDashboard {
  active_users: number;
  page_views: number;
  events: number;
  last_updated: string;
  time_range: string;
}

// 请求参数接口
export interface CreateSessionRequest {
  ip_address: string;
  user_agent: string;
  referrer?: string;
  metadata?: Record<string, any>;
}

export interface UpdateSessionRequest {
  end_time?: string;
  metadata?: Record<string, any>;
}

export interface TrackPageViewRequest {
  session_id: string;
  url: string;
  title?: string;
  referrer?: string;
  duration?: number;
  scroll_depth?: number;
  viewport_size?: string;
  screen_size?: string;
  interaction?: string;
  is_bounce?: boolean;
  exit_page?: boolean;
  entry_page?: boolean;
  properties?: Record<string, any>;
}

export interface TrackEventRequest {
  session_id: string;
  event_type: string;
  event_category: string;
  event_action: string;
  event_label?: string;
  event_value?: number;
  url?: string;
  element?: string;
  properties?: Record<string, any>;
}

export interface CreateJourneyRequest {
  user_id: string;
  journey_type: string;
  end_time?: string;
  steps?: JourneyStep[];
  current_step?: number;
  is_completed?: boolean;
  properties?: Record<string, any>;
}

export interface BatchTrackEventsRequest {
  events: TrackEventRequest[];
}

// API响应接口
export interface ApiResponse<T = any> {
  error?: {
    message: string;
    code: string;
  };
  data?: T;
}

export interface PaginatedResponse<T = any> {
  stats?: T[];
  page?: number;
  page_size?: number;
  total?: number;
}

// AnalyticsService 类
class AnalyticsService {
  // 用户会话相关方法
  async createSession(request: CreateSessionRequest): Promise<ApiResponse<UserSession>> {
    return apiClient.post('/api/v1/analytics/sessions', request);
  }

  async getSession(sessionId: string): Promise<ApiResponse<UserSession>> {
    return apiClient.get(`/api/v1/analytics/sessions/${sessionId}`);
  }

  async updateSession(sessionId: string, request: UpdateSessionRequest): Promise<ApiResponse<Record<string, any>>> {
    return apiClient.put(`/api/v1/analytics/sessions/${sessionId}`, request);
  }

  // 页面浏览相关方法
  async trackPageView(request: TrackPageViewRequest): Promise<ApiResponse<Record<string, any>>> {
    return apiClient.post('/api/v1/analytics/page-views', request);
  }

  // 事件追踪相关方法
  async trackEvent(request: TrackEventRequest): Promise<ApiResponse<Record<string, any>>> {
    return apiClient.post('/api/v1/analytics/events', request);
  }

  async batchTrackEvents(request: BatchTrackEventsRequest): Promise<ApiResponse<Record<string, any>>> {
    return apiClient.post('/api/v1/analytics/events/batch', request);
  }

  // 用户旅程相关方法
  async createJourney(request: CreateJourneyRequest): Promise<ApiResponse<Record<string, any>>> {
    return apiClient.post('/api/v1/analytics/journeys', request);
  }

  // 行为分析相关方法
  async getUserBehaviorAnalysis(
    userId: string,
    startDate?: string,
    endDate?: string
  ): Promise<ApiResponse<UserBehaviorAnalysis>> {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);

    const url = `/api/v1/analytics/analysis/users/${userId}/behavior${params.toString() ? `?${params.toString()}` : ''}`;
    return apiClient.get(url);
  }

  async detectBehaviorPatterns(userId: string): Promise<ApiResponse<Record<string, any>>> {
    return apiClient.post(`/api/v1/analytics/analysis/users/${userId}/patterns`);
  }

  // 实时统计相关方法
  async getRealTimeDashboard(): Promise<ApiResponse<RealTimeDashboard>> {
    return apiClient.get('/api/v1/analytics/stats/realtime/dashboard');
  }

  async updateRealTimeStats(): Promise<ApiResponse<Record<string, any>>> {
    return apiClient.post('/api/v1/analytics/stats/realtime/update');
  }

  async getDailyActiveUsers(startDate?: string, endDate?: string): Promise<ApiResponse<Array<Record<string, any>>>> {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);

    const url = `/api/v1/analytics/stats/users/daily-active${params.toString() ? `?${params.toString()}` : ''}`;
    return apiClient.get(url);
  }

  async getPageViewStats(
    startDate?: string,
    endDate?: string,
    page?: number,
    pageSize?: number
  ): Promise<ApiResponse<PaginatedResponse>> {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('page_size', pageSize.toString());

    const url = `/api/v1/analytics/stats/page-views${params.toString() ? `?${params.toString()}` : ''}`;
    return apiClient.get(url);
  }

  async getEventStats(
    startDate?: string,
    endDate?: string,
    page?: number,
    pageSize?: number
  ): Promise<ApiResponse<PaginatedResponse>> {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('page_size', pageSize.toString());

    const url = `/api/v1/analytics/stats/events${params.toString() ? `?${params.toString()}` : ''}`;
    return apiClient.get(url);
  }

  // 健康检查
  async healthCheck(): Promise<ApiResponse<Record<string, any>>> {
    return apiClient.get('/api/v1/analytics/public/health');
  }

  // 辅助方法：格式化日期为API需要的格式
  formatDateForAPI(date: Date): string {
    return date.toISOString().split('T')[0]; // YYYY-MM-DD格式
  }

  // 辅助方法：获取默认的时间范围（最近7天）
  getDefaultDateRange(): { startDate: string; endDate: string } {
    const endDate = new Date();
    const startDate = new Date();
    startDate.setDate(endDate.getDate() - 7);

    return {
      startDate: this.formatDateForAPI(startDate),
      endDate: this.formatDateForAPI(endDate)
    };
  }

  // 辅助方法：获取当前页面的URL信息
  getCurrentPageInfo(): { url: string; path: string; title?: string } {
    if (typeof window === 'undefined') {
      return { url: '', path: '' };
    }

    return {
      url: window.location.href,
      path: window.location.pathname,
      title: document.title
    };
  }

  // 辅助方法：获取设备信息
  getDeviceInfo(): { userAgent: string; deviceType: string; platform: string; browser: string } {
    if (typeof navigator === 'undefined') {
      return { userAgent: '', deviceType: '', platform: '', browser: '' };
    }

    const userAgent = navigator.userAgent;
    let deviceType = 'desktop';
    let platform = 'unknown';
    let browser = 'unknown';

    // 检测设备类型
    if (/mobile|android|iphone/i.test(userAgent)) {
      deviceType = 'mobile';
    } else if (/tablet|ipad/i.test(userAgent)) {
      deviceType = 'tablet';
    }

    // 检测操作系统
    if (/windows/i.test(userAgent)) {
      platform = 'windows';
    } else if (/mac|os x/i.test(userAgent)) {
      platform = 'mac';
    } else if (/linux/i.test(userAgent)) {
      platform = 'linux';
    } else if (/android/i.test(userAgent)) {
      platform = 'android';
    } else if (/ios|iphone|ipad/i.test(userAgent)) {
      platform = 'ios';
    }

    // 检测浏览器
    if (/chrome/.test(userAgent) && !/edg/.test(userAgent)) {
      browser = 'chrome';
    } else if (/firefox/.test(userAgent)) {
      browser = 'firefox';
    } else if (/safari/.test(userAgent) && !/chrome/.test(userAgent)) {
      browser = 'safari';
    } else if (/edg/.test(userAgent)) {
      browser = 'edge';
    } else if (/opera/.test(userAgent)) {
      browser = 'opera';
    }

    return {
      userAgent,
      deviceType,
      platform,
      browser
    };
  }

  // 辅助方法：获取视口和屏幕信息
  getScreenInfo(): { viewportSize?: string; screenSize?: string } {
    if (typeof window === 'undefined') {
      return {};
    }

    const viewportSize = `${window.innerWidth}x${window.innerHeight}`;
    const screenSize = `${screen.width}x${screen.height}`;

    return {
      viewportSize,
      screenSize
    };
  }
}

// 创建单例实例
const analyticsService = new AnalyticsService();

export default analyticsService;
export { AnalyticsService };
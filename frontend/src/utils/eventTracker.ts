import analyticsService from '../services/analyticsService';

// 事件追踪器配置接口
export interface EventTrackerConfig {
  enableAutoTracking?: boolean;
  enablePageViewTracking?: boolean;
  enableClickTracking?: boolean;
  enableFormTracking?: boolean;
  enableScrollTracking?: boolean;
  enablePerformanceTracking?: boolean;
  apiEndpoint?: string;
  batchSize?: number;
  flushInterval?: number;
  debugMode?: boolean;
}

// 事件数据接口
export interface TrackedEvent {
  id: string;
  type: string;
  category: string;
  action: string;
  label?: string;
  value?: number;
  timestamp: number;
  url: string;
  element?: string;
  properties?: Record<string, any>;
}

// 页面浏览数据接口
export interface PageViewData {
  url: string;
  title?: string;
  referrer?: string;
  timestamp: number;
  duration?: number;
  scrollDepth?: number;
  viewportSize?: string;
  screenSize?: string;
  interaction?: string;
  isBounce?: boolean;
  exitPage?: boolean;
  entryPage?: boolean;
  properties?: Record<string, any>;
}

// 默认配置
const DEFAULT_CONFIG: EventTrackerConfig = {
  enableAutoTracking: true,
  enablePageViewTracking: true,
  enableClickTracking: true,
  enableFormTracking: true,
  enableScrollTracking: false, // 默认关闭滚动追踪，性能开销较大
  enablePerformanceTracking: true,
  batchSize: 10,
  flushInterval: 30000, // 30秒
  debugMode: false
};

// EventTracker 类
class EventTracker {
  private config: EventTrackerConfig;
  private sessionId: string | null = null;
  private eventQueue: TrackedEvent[] = [];
  private pageViewStartTime: number = 0;
  private lastScrollDepth: number = 0;
  private flushTimer: NodeJS.Timeout | null = null;
  private isInitialized: boolean = false;
  private pageViewData: PageViewData | null = null;

  constructor(config: Partial<EventTrackerConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.init();
  }

  // 初始化追踪器
  private async init(): Promise<void> {
    if (this.isInitialized) return;

    try {
      // 创建用户会话
      await this.createSession();

      // 设置自动追踪
      if (this.config.enableAutoTracking) {
        this.setupAutoTracking();
      }

      // 设置定期刷新
      this.setupPeriodicFlush();

      // 页面卸载时发送剩余事件
      this.setupPageUnloadHandler();

      this.isInitialized = true;
      this.log('Event tracker initialized successfully');
    } catch (error) {
      this.log('Failed to initialize event tracker:', error);
    }
  }

  // 创建用户会话
  private async createSession(): Promise<void> {
    try {
      const deviceInfo = analyticsService.getDeviceInfo();
      const pageInfo = analyticsService.getCurrentPageInfo();

      const sessionRequest = {
        ip_address: '', // 将在后端获取
        user_agent: deviceInfo.userAgent,
        referrer: document.referrer || pageInfo.url,
        metadata: {
          device_type: deviceInfo.deviceType,
          platform: deviceInfo.platform,
          browser: deviceInfo.browser,
          screen_info: analyticsService.getScreenInfo()
        }
      };

      const response = await analyticsService.createSession(sessionRequest);
      if (response.data) {
        this.sessionId = response.data.id;
        this.log('Session created:', this.sessionId);
      }
    } catch (error) {
      this.log('Failed to create session:', error);
    }
  }

  // 设置自动追踪
  private setupAutoTracking(): void {
    // 页面浏览追踪
    if (this.config.enablePageViewTracking) {
      this.trackPageView();
    }

    // 点击事件追踪
    if (this.config.enableClickTracking) {
      this.setupClickTracking();
    }

    // 表单事件追踪
    if (this.config.enableFormTracking) {
      this.setupFormTracking();
    }

    // 滚动事件追踪
    if (this.config.enableScrollTracking) {
      this.setupScrollTracking();
    }

    // 性能指标追踪
    if (this.config.enablePerformanceTracking) {
      this.setupPerformanceTracking();
    }
  }

  // 追踪页面浏览
  public trackPageView(customData?: Partial<PageViewData>): void {
    if (!this.sessionId) {
      this.log('Cannot track page view: no session ID');
      return;
    }

    const pageInfo = analyticsService.getCurrentPageInfo();
    const screenInfo = analyticsService.getScreenInfo();
    const now = Date.now();

    this.pageViewData = {
      url: pageInfo.url,
      title: pageInfo.title,
      referrer: document.referrer,
      timestamp: now,
      entryPage: true, // 新页面默认为入口页
      ...screenInfo,
      ...customData
    };

    this.pageViewStartTime = now;
    this.lastScrollDepth = 0;

    // 发送页面浏览事件
    this.sendPageView();
  }

  // 发送页面浏览数据
  private async sendPageView(): Promise<void> {
    if (!this.sessionId || !this.pageViewData) return;

    try {
      const pageViewRequest = {
        session_id: this.sessionId,
        url: this.pageViewData.url,
        title: this.pageViewData.title,
        referrer: this.pageViewData.referrer,
        viewport_size: this.pageViewData.viewportSize,
        screen_size: this.pageViewData.screenSize,
        properties: this.pageViewData.properties
      };

      await analyticsService.trackPageView(pageViewRequest);
      this.log('Page view tracked:', this.pageViewData.url);
    } catch (error) {
      this.log('Failed to track page view:', error);
    }
  }

  // 追踪事件
  public trackEvent(
    category: string,
    action: string,
    label?: string,
    value?: number,
    properties?: Record<string, any>
  ): void {
    if (!this.sessionId) {
      this.log('Cannot track event: no session ID');
      return;
    }

    const pageInfo = analyticsService.getCurrentPageInfo();
    const event: TrackedEvent = {
      id: this.generateId(),
      type: 'user_event',
      category,
      action,
      label,
      value,
      timestamp: Date.now(),
      url: pageInfo.url,
      properties
    };

    this.eventQueue.push(event);
    this.log('Event tracked:', { category, action, label, value });

    // 如果队列达到批量大小，立即发送
    if (this.eventQueue.length >= (this.config.batchSize || 10)) {
      this.flushEvents();
    }
  }

  // 设置点击追踪
  private setupClickTracking(): void {
    document.addEventListener('click', (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      const elementInfo = this.getElementInfo(target);

      if (elementInfo.trackable) {
        this.trackEvent(
          'click',
          elementInfo.category || 'general',
          elementInfo.action || target.tagName.toLowerCase(),
          elementInfo.label,
          undefined,
          {
            element_type: target.tagName.toLowerCase(),
            element_id: target.id,
            element_class: target.className,
            element_text: target.textContent?.slice(0, 100),
            coordinates: {
              x: event.clientX,
              y: event.clientY
            },
            ...elementInfo.properties
          }
        );
      }
    });
  }

  // 设置表单追踪
  private setupFormTracking(): void {
    // 表单提交追踪
    document.addEventListener('submit', (event: Event) => {
      const form = event.target as HTMLFormElement;
      const elementInfo = this.getElementInfo(form);

      this.trackEvent(
        'form',
        'submit',
        elementInfo.label || form.id || form.className || 'unknown_form',
        undefined,
        {
          form_id: form.id,
          form_class: form.className,
          form_action: form.action,
          form_method: form.method,
          field_count: form.elements.length,
          ...elementInfo.properties
        }
      );
    });

    // 表单字段交互追踪
    document.addEventListener('focus', (event: FocusEvent) => {
      const target = event.target as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement;
      if (['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName)) {
        const elementInfo = this.getElementInfo(target);

        this.trackEvent(
          'form',
          'field_focus',
          elementInfo.label || target.name || target.id || 'unknown_field',
          undefined,
          {
            field_type: target.type || target.tagName.toLowerCase(),
            field_name: target.name,
            field_id: target.id,
            field_class: target.className,
            ...elementInfo.properties
          }
        );
      }
    }, true);
  }

  // 设置滚动追踪
  private setupScrollTracking(): void {
    let scrollTimer: NodeJS.Timeout | null = null;

    const handleScroll = () => {
      if (scrollTimer) return;

      scrollTimer = setTimeout(() => {
        const scrollDepth = this.calculateScrollDepth();

        if (scrollDepth > this.lastScrollDepth) {
          this.lastScrollDepth = scrollDepth;

          this.trackEvent(
            'scroll',
            'depth',
            `scroll_${Math.floor(scrollDepth / 10) * 10}%`,
            scrollDepth
          );
        }

        scrollTimer = null;
      }, 100);
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
  }

  // 设置性能追踪
  private setupPerformanceTracking(): void {
    // 等待页面加载完成
    if (document.readyState === 'complete') {
      this.trackPerformanceMetrics();
    } else {
      window.addEventListener('load', () => {
        setTimeout(this.trackPerformanceMetrics.bind(this), 0);
      });
    }
  }

  // 追踪性能指标
  private trackPerformanceMetrics(): void {
    if (!window.performance || !window.performance.timing) return;

    const timing = window.performance.timing;
    const navigation = window.performance.navigation;

    const metrics = {
      // 页面加载时间
      dom_content_loaded: timing.domContentLoadedEventEnd - timing.navigationStart,
      load_complete: timing.loadEventEnd - timing.navigationStart,

      // 网络时间
      dns_lookup: timing.domainLookupEnd - timing.domainLookupStart,
      tcp_connect: timing.connectEnd - timing.connectStart,
      server_response: timing.responseEnd - timing.requestStart,

      // 前端渲染时间
      dom_processing: timing.domComplete - timing.domLoading,

      // 导航类型
      navigation_type: navigation.type,
      redirect_count: navigation.redirectCount
    };

    this.trackEvent('performance', 'page_load', 'load_complete', metrics.load_complete, metrics);
  }

  // 设置定期刷新
  private setupPeriodicFlush(): void {
    if (this.config.flushInterval && this.config.flushInterval > 0) {
      this.flushTimer = setInterval(() => {
        this.flushEvents();
      }, this.config.flushInterval);
    }
  }

  // 设置页面卸载处理器
  private setupPageUnloadHandler(): void {
    const flushBeforeUnload = () => {
      if (this.eventQueue.length > 0) {
        // 使用 sendBeacon API 在页面卸载时发送数据
        this.flushEvents(true);
      }
    };

    window.addEventListener('beforeunload', flushBeforeUnload);
    window.addEventListener('pagehide', flushBeforeUnload);
  }

  // 刷新事件队列
  private async flushEvents(isBeacon: boolean = false): Promise<void> {
    if (this.eventQueue.length === 0) return;

    const eventsToSend = [...this.eventQueue];
    this.eventQueue = [];

    try {
      if (isBeacon && navigator.sendBeacon) {
        // 使用 sendBeacon API 发送数据
        const data = JSON.stringify({
          events: eventsToSend.map(event => ({
            session_id: this.sessionId,
            event_type: event.type,
            event_category: event.category,
            event_action: event.action,
            event_label: event.label,
            event_value: event.value,
            url: event.url,
            element: event.element,
            properties: event.properties
          }))
        });

        navigator.sendBeacon('/api/v1/analytics/events/batch', data);
        this.log(`Sent ${eventsToSend.length} events via beacon`);
      } else {
        // 使用常规 API 发送
        const batchRequests = eventsToSend.map(event => ({
          session_id: this.sessionId!,
          event_type: event.type,
          event_category: event.category,
          event_action: event.action,
          event_label: event.label,
          event_value: event.value,
          url: event.url,
          element: event.element,
          properties: event.properties
        }));

        await analyticsService.batchTrackEvents({ events: batchRequests });
        this.log(`Sent ${eventsToSend.length} events via API`);
      }
    } catch (error) {
      this.log('Failed to flush events:', error);
      // 如果发送失败，将事件重新加入队列（限制重试次数）
      eventsToSend.forEach(event => {
        if ((event.properties?.retryCount || 0) < 3) {
          event.properties = { ...event.properties, retryCount: (event.properties?.retryCount || 0) + 1 };
          this.eventQueue.push(event);
        }
      });
    }
  }

  // 结束当前页面浏览
  public endPageView(): void {
    if (!this.pageViewData || !this.pageViewStartTime) return;

    const duration = Date.now() - this.pageViewStartTime;
    const scrollDepth = this.calculateScrollDepth();

    // 更新页面浏览数据
    this.pageViewData.duration = duration;
    this.pageViewData.scrollDepth = scrollDepth;
    this.pageViewData.exitPage = true;
    this.pageViewData.isBounce = duration < 5000; // 停留少于5秒视为跳出

    // 发送更新后的页面浏览数据
    this.sendPageView();
    this.pageViewData = null;
  }

  // 获取元素信息
  private getElementInfo(element: HTMLElement): {
    trackable: boolean;
    category?: string;
    action?: string;
    label?: string;
    properties?: Record<string, any>;
  } {
    const result = {
      trackable: false,
      category: undefined as string | undefined,
      action: undefined as string | undefined,
      label: undefined as string | undefined,
      properties: {} as Record<string, any>
    };

    // 检查是否有追踪属性
    const trackAttr = element.getAttribute('data-track');
    const categoryAttr = element.getAttribute('data-category');
    const actionAttr = element.getAttribute('data-action');
    const labelAttr = element.getAttribute('data-label');

    if (trackAttr === 'false') {
      return result;
    }

    if (trackAttr === 'true' || categoryAttr || actionAttr || labelAttr) {
      result.trackable = true;
      result.category = categoryAttr || 'general';
      result.action = actionAttr || element.textContent?.slice(0, 50) || 'unknown';
      result.label = labelAttr || undefined;
    } else {
      // 自动检测可追踪的元素
      const tagName = element.tagName.toLowerCase();
      const trackableTags = ['a', 'button', 'input[type="button"]', 'input[type="submit"]'];

      if (trackableTags.includes(tagName) || element.onclick) {
        result.trackable = true;
        result.category = 'interaction';
        result.action = tagName;

        if (tagName === 'a') {
          result.label = (element as HTMLAnchorElement).href;
        } else if (element.textContent) {
          result.label = element.textContent.slice(0, 50);
        }
      }
    }

    // 添加额外的属性
    if (element.id) {
      result.properties.element_id = element.id;
    }
    if (element.className) {
      result.properties.element_class = element.className;
    }

    return result;
  }

  // 计算滚动深度
  private calculateScrollDepth(): number {
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
    const documentHeight = document.documentElement.scrollHeight;
    const windowHeight = window.innerHeight;
    const scrollableHeight = documentHeight - windowHeight;

    if (scrollableHeight <= 0) return 100;

    return Math.min(100, Math.round((scrollTop / scrollableHeight) * 100));
  }

  // 生成唯一ID
  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // 日志输出
  private log(...args: any[]): void {
    if (this.config.debugMode) {
      console.log('[EventTracker]', ...args);
    }
  }

  // 销毁追踪器
  public destroy(): void {
    // 结束当前页面浏览
    this.endPageView();

    // 刷新剩余事件
    this.flushEvents();

    // 清理定时器
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }

    this.isInitialized = false;
    this.log('Event tracker destroyed');
  }

  // 获取当前会话ID
  public getSessionId(): string | null {
    return this.sessionId;
  }

  // 更新配置
  public updateConfig(newConfig: Partial<EventTrackerConfig>): void {
    this.config = { ...this.config, ...newConfig };
    this.log('Config updated:', this.config);
  }
}

// 创建默认实例
export const eventTracker = new EventTracker();

// 便捷方法
export const trackEvent = (
  category: string,
  action: string,
  label?: string,
  value?: number,
  properties?: Record<string, any>
): void => {
  eventTracker.trackEvent(category, action, label, value, properties);
};

export const trackPageView = (customData?: Partial<PageViewData>): void => {
  eventTracker.trackPageView(customData);
};

export const endPageView = (): void => {
  eventTracker.endPageView();
};

// React Hook
export const useEventTracker = () => {
  return {
    trackEvent,
    trackPageView,
    endPageView,
    getSessionId: () => eventTracker.getSessionId(),
    updateConfig: (config: Partial<EventTrackerConfig>) => eventTracker.updateConfig(config)
  };
};

export default EventTracker;
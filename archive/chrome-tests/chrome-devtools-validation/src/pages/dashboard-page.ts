import { PageObject } from '../mcp/page-object';
import { ChromeDevToolsService } from '../mcp/devtools-service';
import { Logger } from '../core/logger';

/**
 * 仪表板页面Page Object
 */
export class DashboardPage extends PageObject {
  protected override selectors = {
    userProfile: '.user-profile',
    userName: '.user-name',
    welcomeMessage: '.welcome-message',
    quickActions: '.quick-actions',
    caseList: '.case-list',
    caseItem: '.case-item',
    notificationBell: '.notification-bell',
    notificationBadge: '.notification-badge',
    searchBox: '.search-box',
    sidebar: '.sidebar',
    menuItems: '.menu-item',
    statsCards: '.stats-card',
    activityFeed: '.activity-feed',
    recentCases: '.recent-cases',
    upcomingDeadlines: '.upcoming-deadlines',
    logoutButton: '.logout-button',
  };

  constructor(service: ChromeDevToolsService, logger?: Logger) {
    super(service, logger);
    this.url = '/dashboard';
  }

  /**
   * 验证仪表板页面加载
   */
  override async validatePage(): Promise<void> {
    this.logger.info('验证仪表板页面');

    await this.waitForPageLoad();
    await this.expectVisible(this.selectors.userProfile);
    await this.expectVisible(this.selectors.welcomeMessage);
    await this.expectVisible(this.selectors.quickActions);
    await this.expectVisible(this.selectors.sidebar);

    this.logger.info('仪表板页面验证通过');
  }

  /**
   * 获取用户名
   */
  override async getUserName(): Promise<string> {
    await this.waitForElement(this.selectors.userName);
    return await this.getText(this.selectors.userName);
  }

  /**
   * 获取欢迎消息
   */
  override async getWelcomeMessage(): Promise<string> {
    await this.waitForElement(this.selectors.welcomeMessage);
    return await this.getText(this.selectors.welcomeMessage);
  }

  /**
   * 获取案件列表
   */
  override async getCaseList(): Promise<Array<{ title: string; status: string; id: string }>> {
    const cases: Array<{ title: string; status: string; id: string }> = [];

    const caseElements = await this.executeScript<Element[]>(this.selectors.caseItem);

    for (const caseElement of caseElements) {
      const title = caseElement.querySelector('.case-title')?.textContent || '';
      const status = caseElement.querySelector('.case-status')?.textContent || '';
      const id = caseElement.getAttribute('data-case-id') || '';

      if (title) {
        cases.push({ title, status, id });
      }
    }

    this.logger.debug('获取案件列表', { count: cases.length });
    return cases;
  }

  /**
   * 点击特定案件
   */
  override async clickCase(caseId: string): Promise<void> {
    const caseSelector = `${this.selectors.caseItem}[data-case-id="${caseId}"]`;
    await this.waitForElement(caseSelector);
    await this.click(caseSelector, true);
    this.logger.debug('点击案件', { caseId });
  }

  /**
   * 获取通知数量
   */
  override async getNotificationCount(): Promise<number> {
    if (!(await this.exists(this.selectors.notificationBadge))) {
      return 0;
    }

    const badgeText = await this.getText(this.selectors.notificationBadge);
    return parseInt(badgeText) || 0;
  }

  /**
   * 点击通知铃铛
   */
  override async clickNotifications(): Promise<void> {
    await this.click(this.selectors.notificationBell);
    this.logger.debug('点击通知铃铛');
  }

  /**
   * 搜索功能
   */
  override async search(query: string): Promise<void> {
    await this.fill(this.selectors.searchBox, query);
    await this.delay(1000); // 等待搜索结果
    this.logger.debug('执行搜索', { query });
  }

  /**
   * 获取统计数据
   */
  override async getStatistics(): Promise<{
    totalCases: number;
    activeCases: number;
    completedCases: number;
    pendingTasks: number;
  }> {
    const stats = {
      totalCases: 0,
      activeCases: 0,
      completedCases: 0,
      pendingTasks: 0,
    };

    // 从统计卡片中提取数据
    const statCards = await this.executeScript<Element[]>(this.selectors.statsCards);

    for (const card of statCards) {
      const title = card.querySelector('.stat-title')?.textContent || '';
      const value = card.querySelector('.stat-value')?.textContent || '';

      if (title.includes('总案件')) {
        stats.totalCases = parseInt(value.replace(/[^0-9]/g, '')) || 0;
      } else if (title.includes('进行中')) {
        stats.activeCases = parseInt(value.replace(/[^0-9]/g, '')) || 0;
      } else if (title.includes('已完成')) {
        stats.completedCases = parseInt(value.replace(/[^0-9]/g, '')) || 0;
      } else if (title.includes('待办')) {
        stats.pendingTasks = parseInt(value.replace(/[^0-9]/g, '')) || 0;
      }
    }

    this.logger.debug('获取统计数据', stats);
    return stats;
  }

  /**
   * 获取活动流
   */
  override async getActivityFeed(): Promise<Array<{ action: string; time: string; user: string }>> {
    const activities: Array<{ action: string; time: string; user: string }> = [];

    const activityItems = await this.executeScript<Element[]>( `${this.selectors.activityFeed} .activity-item`);

    for (const item of activityItems) {
      const action = item.querySelector('.activity-action')?.textContent || '';
      const time = item.querySelector('.activity-time')?.textContent || '';
      const user = item.querySelector('.activity-user')?.textContent || '';

      if (action) {
        activities.push({ action, time, user });
      }
    }

    this.logger.debug('获取活动流', { count: activities.length });
    return activities;
  }

  /**
   * 获取即将到期的截止日期
   */
  override async getUpcomingDeadlines(): Promise<Array<{ title: string; dueDate: string; daysLeft: number }>> {
    const deadlines: Array<{ title: string; dueDate: string; daysLeft: number }> = [];

    const deadlineItems = await this.executeScript<Element[]>( `${this.selectors.upcomingDeadlines} .deadline-item`);

    for (const item of deadlineItems) {
      const title = item.querySelector('.deadline-title')?.textContent || '';
      const dueDate = item.querySelector('.deadline-date')?.textContent || '';
      const daysLeftText = item.querySelector('.days-left')?.textContent || '';
      const daysLeft = parseInt(daysLeftText.replace(/[^0-9]/g, '')) || 0;

      if (title) {
        deadlines.push({ title, dueDate, daysLeft });
      }
    }

    this.logger.debug('获取即将到期的截止日期', { count: deadlines.length });
    return deadlines;
  }

  /**
   * 导航到案件管理
   */
  override async navigateToCases(): Promise<void> {
    const casesMenuItem = `${this.selectors.menuItems}[data-menu="cases"]`;
    await this.click(casesMenuItem, true);
    this.logger.debug('导航到案件管理');
  }

  /**
   * 导航到客户管理
   */
  override async navigateToClients(): Promise<void> {
    const clientsMenuItem = `${this.selectors.menuItems}[data-menu="clients"]`;
    await this.click(clientsMenuItem, true);
    this.logger.debug('导航到客户管理');
  }

  /**
   * 导航到文档管理
   */
  override async navigateToDocuments(): Promise<void> {
    const documentsMenuItem = `${this.selectors.menuItems}[data-menu="documents"]`;
    await this.click(documentsMenuItem, true);
    this.logger.debug('导航到文档管理');
  }

  /**
   * 导航到财务管理
   */
  override async navigateToFinance(): Promise<void> {
    const financeMenuItem = `${this.selectors.menuItems}[data-menu="finance"]`;
    await this.click(financeMenuItem, true);
    this.logger.debug('导航到财务管理');
  }

  /**
   * 点击用户资料
   */
  override async clickUserProfile(): Promise<void> {
    await this.click(this.selectors.userProfile);
    this.logger.debug('点击用户资料');
  }

  /**
   * 执行登出
   */
  override async logout(): Promise<void> {
    this.logger.info('执行登出操作');

    // 点击用户资料打开菜单
    await this.clickUserProfile();
    await this.delay(500);

    // 点击登出按钮
    if (await this.exists(this.selectors.logoutButton)) {
      await this.click(this.selectors.logoutButton, true);
    } else {
      // 如果没有找到登出按钮，尝试JavaScript方式
      await this.executeScript(`
        const logoutButtons = Array.from(document.querySelectorAll('button'))
          .filter(btn => btn.textContent.includes('登出') || btn.textContent.includes('退出') || btn.textContent.includes('Logout'));
        if (logoutButtons.length > 0) {
          logoutButtons[0].click();
        }
      `);
    }

    this.logger.info('登出操作完成');
  }

  /**
   * 检查是否有未读通知
   */
  override async hasUnreadNotifications(): Promise<boolean> {
    return await this.exists(this.selectors.notificationBadge);
  }

  /**
   * 获取快速操作列表
   */
  override async getQuickActions(): Promise<Array<{ name: string; icon: string; action: string }>> {
    const actions: Array<{ name: string; icon: string; action: string }> = [];

    const actionItems = await this.executeScript<Element[]>( `${this.selectors.quickActions} .quick-action`);

    for (const item of actionItems) {
      const name = item.querySelector('.action-name')?.textContent || '';
      const icon = item.querySelector('.action-icon')?.getAttribute('class') || '';
      const action = item.getAttribute('data-action') || '';

      if (name) {
        actions.push({ name, icon, action });
      }
    }

    this.logger.debug('获取快速操作列表', { count: actions.length });
    return actions;
  }

  /**
   * 执行快速操作
   */
  override async executeQuickAction(actionName: string): Promise<void> {
    const actionSelector = `${this.selectors.quickActions} .quick-action[data-action="${actionName}"]`;
    await this.waitForElement(actionSelector);
    await this.click(actionSelector, true);
    this.logger.debug('执行快速操作', { actionName });
  }

  /**
   * 截图仪表板
   */
  override async captureDashboard(): Promise<string> {
    this.logger.info('截图仪表板');
    return await this.screenshot('dashboard');
  }
}
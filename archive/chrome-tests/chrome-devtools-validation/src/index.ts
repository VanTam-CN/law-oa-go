/**
 * Chrome DevTools Validation Framework
 *
 * A comprehensive testing framework for Law OA System validation
 * using Chrome DevTools MCP services.
 */

// Core framework exports
export * from './types';
export * from './core';
export * from './mcp';
export * from './pages';

// Main framework entry point
export class ChromeDevToolsValidator {
  private initialized = false;
  private chromeService?: import('./mcp').ChromeDevToolsService;

  override async initialize(): Promise<void> {
    if (this.initialized) {
      return;
    }

    try {
      // Initialize Chrome DevTools MCP service
      const { ChromeDevToolsService } = await import('./mcp');
      this.chromeService = new ChromeDevToolsService();

      // Initialize the service
      await this.chromeService.initialize();

      this.initialized = true;
      console.log('Chrome DevTools Validator initialized successfully');
    } catch (error) {
      console.error('Failed to initialize Chrome DevTools Validator:', error);
      throw error;
    }
  }

  override async cleanup(): Promise<void> {
    if (!this.initialized || !this.chromeService) {
      return;
    }

    try {
      // Clean up resources
      await this.chromeService.close();
      this.chromeService = undefined as any;
      this.initialized = false;
      console.log('Chrome DevTools Validator cleaned up successfully');
    } catch (error) {
      console.error('Error during cleanup:', error);
      throw error;
    }
  }

  isInitialized(): boolean {
    return this.initialized;
  }

  /**
   * Get the Chrome DevTools service instance
   */
  getChromeService(): import('./mcp').ChromeDevToolsService | undefined {
    return this.chromeService;
  }

  /**
   * Create a new login page instance
   */
  override async createLoginPage(): Promise<import('./pages').LoginPage> {
    if (!this.chromeService) {
      throw new Error('Validator not initialized');
    }

    const { LoginPage } = await import('./pages');
    return new LoginPage(this.chromeService);
  }

  /**
   * Create a new dashboard page instance
   */
  override async createDashboardPage(): Promise<import('./pages').DashboardPage> {
    if (!this.chromeService) {
      throw new Error('Validator not initialized');
    }

    const { DashboardPage } = await import('./pages');
    return new DashboardPage(this.chromeService);
  }
}

// Default export for convenience
export default ChromeDevToolsValidator;
import { TestConfig, MCPConfig, ChromeDevToolsConfig } from '../types';
import * as fs from 'fs';
import * as path from 'path';

export class ConfigManager {
  private static instance: ConfigManager;
  private config: TestConfig;

  private constructor() {
    this.config = this.loadConfig();
  }

  public static getInstance(): ConfigManager {
    if (!ConfigManager.instance) {
      ConfigManager.instance = new ConfigManager();
    }
    return ConfigManager.instance;
  }

  private loadConfig(): TestConfig {
    // Default configuration
    const defaultConfig: TestConfig = {
      baseUrl: 'http://localhost:3000',
      apiUrl: 'http://localhost:8080/api',
      defaultTimeout: 30000,
      retryAttempts: 3,
      retryDelay: 1000,
      concurrencyLevel: 1,
      screenshotOnFailure: true,
      headless: false,
      slowMo: 100,
      outputDir: './test-results',
      reporting: {
        formats: ['html', 'json'],
        includeScreenshots: true,
        includeLogs: true,
      },
    };

    try {
      // Try to load configuration file
      const configPath = path.resolve(process.cwd(), 'test.config.json');
      if (fs.existsSync(configPath)) {
        const fileConfig = JSON.parse(fs.readFileSync(configPath, 'utf8'));
        return { ...defaultConfig, ...fileConfig };
      }
    } catch (error) {
      console.warn('Failed to load configuration file, using defaults:', error);
    }

    // Override with environment variables
    const envConfig = this.loadEnvironmentConfig();
    return { ...defaultConfig, ...envConfig };
  }

  private loadEnvironmentConfig(): Partial<TestConfig> {
    const envConfig: Partial<TestConfig> = {};

    if (process.env['TEST_BASE_URL']) {
      envConfig.baseUrl = process.env['TEST_BASE_URL'];
    }
    if (process.env['TEST_API_URL']) {
      envConfig.apiUrl = process.env['TEST_API_URL'];
    }
    if (process.env['TEST_TIMEOUT']) {
      envConfig.defaultTimeout = parseInt(process.env['TEST_TIMEOUT'], 10);
    }
    if (process.env['TEST_RETRY_ATTEMPTS']) {
      envConfig.retryAttempts = parseInt(process.env['TEST_RETRY_ATTEMPTS'], 10);
    }
    if (process.env['TEST_HEADLESS']) {
      envConfig.headless = process.env['TEST_HEADLESS'].toLowerCase() === 'true';
    }
    if (process.env['TEST_OUTPUT_DIR']) {
      envConfig.outputDir = process.env['TEST_OUTPUT_DIR'];
    }

    return envConfig;
  }

  public getConfig(): TestConfig {
    return { ...this.config };
  }

  public getMCPConfig(): MCPConfig {
    return {
      chromeDevTools: {
        enabled: true,
        config: {
          headless: this.config.headless,
          slowMo: this.config.slowMo,
          defaultTimeout: this.config.defaultTimeout,
          viewport: {
            width: 1280,
            height: 720,
          },
        },
      },
    };
  }

  public getChromeDevToolsConfig(): ChromeDevToolsConfig {
    return {
      enabled: true,
      headless: this.config.headless,
      slowMo: this.config.slowMo,
      defaultTimeout: this.config.defaultTimeout,
      viewport: {
        width: 1280,
        height: 720,
      },
      userAgent: 'Chrome DevTools MCP Test Framework',
      ignoreHTTPSErrors: true,
    };
  }

  public updateConfig(updates: Partial<TestConfig>): void {
    this.config = { ...this.config, ...updates };
  }

  public saveConfig(filePath?: string): void {
    const configPath = filePath || path.resolve(process.cwd(), 'test.config.json');
    try {
      fs.writeFileSync(configPath, JSON.stringify(this.config, null, 2));
      console.log(`Configuration saved to: ${configPath}`);
    } catch (error) {
      console.error('Failed to save configuration:', error);
    }
  }

  public validateConfig(): { valid: boolean; errors: string[] } {
    let errors: string[] = [];

    if (!this.config.baseUrl) {
      errors.push('baseUrl is required');
    }
    if (!this.config.apiUrl) {
      errors.push('apiUrl is required');
    }
    if (this.config.defaultTimeout <= 0) {
      errors.push('defaultTimeout must be positive');
    }
    if (this.config.retryAttempts < 0) {
      errors.push('retryAttempts must be non-negative');
    }
    if (this.config.concurrencyLevel <= 0) {
      errors.push('concurrencyLevel must be positive');
    }

    return {
      valid: (errors || []).length === 0,
      errors,
    };
  }
}
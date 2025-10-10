/**
 * Chrome DevTools MCP service type definitions
 */

export interface ChromeDevToolsConfig {
  enabled: boolean;
  headless: boolean;
  slowMo: number;
  defaultTimeout: number;
  viewport: {
    width: number;
    height: number;
  };
  userAgent?: string;
  ignoreHTTPSErrors: boolean;
}

export interface DevToolsPage {
  id: string;
  url: string;
  title: string;
  timestamp: Date;
  viewport: {
    width: number;
    height: number;
  };
  networkRequests: DevToolsNetworkRequest[];
  consoleMessages: DevToolsConsoleMessage[];
}

export interface DevToolsElement {
  uid: string;
  tagName: string;
  attributes: Record<string, string>;
  textContent?: string;
  visible: boolean;
  enabled: boolean;
  x: number;
  y: number;
  width: number;
  height: number;
  children?: DevToolsElement[];
}

export interface DevToolsNetworkRequest {
  id: string;
  url: string;
  method: string;
  status: number;
  statusText: string;
  headers: Record<string, string>;
  contentType?: string;
  timestamp: Date;
  duration: number;
  size: number;
  response?: any;
}

export interface DevToolsConsoleMessage {
  type: 'log' | 'warn' | 'error' | 'info' | 'debug';
  text: string;
  timestamp: Date;
  location?: {
    url: string;
    lineNumber: number;
    columnNumber: number;
  };
  stack?: string;
}

export interface MCPServiceConfig {
  serviceName: string;
  endpoint?: string;
  timeout: number;
  retryAttempts: number;
  retryDelay: number;
  headers?: Record<string, string>;
}

export interface MCPRequest {
  method: string;
  params: Record<string, any>;
  timeout?: number;
  headers?: Record<string, string>;
}

export interface MCPResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  metadata?: {
    requestId: string;
    timestamp: Date;
    duration: number;
  };
}

export interface MCPSession {
  id: string;
  createdAt: Date;
  lastActivity: Date;
  pages: string[];
  currentPage: number;
  capabilities: {
    browserName: string;
    browserVersion: string;
    platform: string;
  };
}

// MCP specific operation types
export interface NavigateOperation {
  url: string;
  waitUntil?: 'load' | 'domcontentloaded' | 'networkidle';
  timeout?: number;
}

export interface ClickOperation {
  element: DevToolsElement;
  button?: 'left' | 'right' | 'middle';
  clickCount?: number;
  delay?: number;
}

export interface FillOperation {
  element: DevToolsElement;
  value: string;
  clear?: boolean;
  delay?: number;
}

export interface SelectOperation {
  element: DevToolsElement;
  values: string[];
  force?: boolean;
}

export interface WaitOperation {
  condition: 'element' | 'navigation' | 'timeout' | 'network';
  selector?: string;
  timeout?: number;
  state?: 'visible' | 'hidden' | 'enabled' | 'disabled';
}

export interface ScreenshotOperation {
  fullPage?: boolean;
  selector?: string;
  format?: 'png' | 'jpeg';
  quality?: number;
  filename?: string;
}

export type MCPOperation =
  | NavigateOperation
  | ClickOperation
  | FillOperation
  | SelectOperation
  | WaitOperation
  | ScreenshotOperation;
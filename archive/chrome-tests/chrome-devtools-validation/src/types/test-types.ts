/**
 * Test framework core type definitions
 */

export interface TestSuite {
  id: string;
  name: string;
  description: string;
  testCases: TestCase[];
  tags: string[];
  timeout: number;
  setup?: () => Promise<void>;
  teardown?: () => Promise<void>;
}

export interface TestCase {
  id: string;
  name: string;
  description: string;
  steps: TestStep[];
  assertions: Assertion[];
  tags: string[];
  timeout: number;
  priority: 'P0' | 'P1' | 'P2';
  dependencies?: string[];
  setup?: () => Promise<void>;
  teardown?: () => Promise<void>;
}

export interface TestStep {
  id: string;
  name: string;
  type: 'navigate' | 'click' | 'fill' | 'select' | 'wait' | 'verify' | 'screenshot' | 'executeScript';
  selector?: string;
  url?: string;
  value?: string | FormData;
  timeout?: number;
  expectedState?: ElementState;
  retry?: number;
  script?: string;
}

export interface Assertion {
  id: string;
  type: 'element-exists' | 'element-visible' | 'element-enabled' | 'text-contains' | 'value-equals' | 'url-contains';
  selector: string;
  expected: any;
  actual?: any;
  message?: string;
}

export interface ElementState {
  visible?: boolean;
  enabled?: boolean;
  text?: string;
  value?: string;
  attribute?: string;
}

export interface FormData {
  [selector: string]: string;
}

export interface TestResult {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'passed' | 'failed' | 'skipped';
  startTime?: Date;
  endTime?: Date;
  duration?: number;
  error?: TestError;
  screenshots?: string[];
  logs?: TestLog[];
  metadata?: Record<string, any>;
}

export interface TestError {
  message: string;
  stack?: string;
  type: string;
  code?: string;
}

export interface TestLog {
  timestamp: Date;
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
  metadata?: Record<string, any>;
}

export interface TestExecutionResult {
  id: string;
  timestamp: Date;
  duration: number;
  totalTests: number;
  passedTests: number;
  failedTests: number;
  skippedTests: number;
  testSuites: TestSuiteResult[];
  config: TestConfig;
}

export interface TestSuiteResult {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'passed' | 'failed' | 'skipped';
  startTime?: Date;
  endTime?: Date;
  duration?: number;
  testCases: TestResult[];
  error?: TestError;
}

export interface TestConfig {
  baseUrl: string;
  apiUrl: string;
  defaultTimeout: number;
  retryAttempts: number;
  retryDelay: number;
  concurrencyLevel: number;
  screenshotOnFailure: boolean;
  headless: boolean;
  slowMo: number;
  outputDir: string;
  reporting: {
    formats: ('html' | 'json' | 'junit')[];
    includeScreenshots: boolean;
    includeLogs: boolean;
  };
}

export interface TestExecutionContext {
  id: string;
  testSuite: TestSuite;
  config: TestConfig;
  startTime: Date;
  status: 'running' | 'completed' | 'failed';
  testCases: TestCaseContext[];
  logs: TestLog[];
  metadata: Record<string, any>;
}

export interface TestCaseContext {
  id: string;
  testCase: TestCase;
  status: 'pending' | 'running' | 'passed' | 'failed' | 'skipped';
  startTime?: Date;
  endTime?: Date;
  duration?: number;
  error?: TestError;
  screenshots?: string[];
  steps: TestStepResult[];
}

export interface TestStepResult {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'passed' | 'failed' | 'skipped';
  startTime?: Date;
  endTime?: Date;
  duration?: number;
  error?: TestError;
  screenshot?: string;
}

export interface BrowserSession {
  sessionId: string;
  pages: string[];
  currentPage: number;
  capabilities: BrowserCapabilities;
}

export interface BrowserCapabilities {
  browserName: string;
  browserVersion: string;
  platform: string;
  headless: boolean;
  viewport: {
    width: number;
    height: number;
  };
}

export interface ElementInfo {
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
}

export interface PageSnapshot {
  url: string;
  title: string;
  elements: ElementInfo[];
  timestamp: Date;
  viewport: {
    width: number;
    height: number;
  };
}

export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export interface MCPConfig {
  chromeDevTools: {
    enabled: boolean;
    config: {
      headless: boolean;
      slowMo: number;
      defaultTimeout: number;
      viewport: {
        width: number;
        height: number;
      };
    };
  };
}

export interface TestReporter {
  generateReport(results: TestExecutionResult): Promise<void>;
  generateSummary(results: TestExecutionResult): Promise<string>;
}

export interface TestDataManager {
  loadTestData(suiteId: string): Promise<DataSet>;
  generateTestData(template: DataTemplate): Promise<DataSet>;
  cleanupTestData(dataSetId: string): Promise<void>;
  saveDataState(state: DataState): Promise<void>;
  restoreDataState(stateId: string): Promise<void>;
  verifyDataConsistency(): Promise<ConsistencyReport>;
}

export interface DataSet {
  id: string;
  name: string;
  users: UserData[];
  clients: ClientData[];
  cases: CaseData[];
  documents: DocumentData[];
  metadata: Record<string, any>;
}

export interface DataTemplate {
  id: string;
  name: string;
  description: string;
  template: any;
  generation: {
    count: number;
    indexPattern: 'sequential' | 'random';
  };
}

export interface DataState {
  id: string;
  timestamp: Date;
  database: DatabaseState;
  fileSystem: FileSystemState;
  cache: CacheState;
}

export interface ConsistencyReport {
  timestamp: Date;
  database: ConsistencyCheck[];
  fileSystem: ConsistencyCheck[];
  cache: ConsistencyCheck[];
  summary: {
    totalChecks: number;
    passedChecks: number;
    failedChecks: number;
    warnings: number;
  };
}

export interface ConsistencyCheck {
  id: string;
  type: 'database' | 'file-system' | 'cache';
  status: 'passed' | 'failed' | 'warning';
  message: string;
  details?: Record<string, any>;
}

// Domain-specific types
export interface UserData {
  id: string;
  username: string;
  email: string;
  password: string;
  role: 'admin' | 'attorney' | 'paralegal' | 'client';
  department?: string;
  status: 'active' | 'inactive' | 'pending';
  createdAt: Date;
  updatedAt: Date;
}

export interface ClientData {
  id: string;
  name: string;
  type: 'individual' | 'corporate';
  registrationNumber?: string;
  industry?: string;
  contactPerson?: string;
  email?: string;
  phone?: string;
  address?: string;
  status: 'active' | 'inactive' | 'prospect';
  createdAt: Date;
  updatedAt: Date;
}

export interface CaseData {
  id: string;
  title: string;
  caseNumber: string;
  type: 'litigation' | 'corporate' | 'family' | 'criminal' | 'other';
  status: 'active' | 'closed' | 'pending' | 'archived';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assignedTo: string;
  client: string;
  description?: string;
  estimatedValue?: string;
  createdDate: Date;
  updatedDate: Date;
  dueDate?: Date;
}

export interface DocumentData {
  id: string;
  name: string;
  type: 'contract' | 'evidence' | 'correspondence' | 'court_filing' | 'other';
  caseId: string;
  clientId?: string;
  uploadedBy: string;
  fileSize: number;
  mimeType: string;
  path: string;
  tags: string[];
  status: 'active' | 'archived' | 'deleted';
  uploadedAt: Date;
  updatedAt: Date;
}

// Database state types
export interface DatabaseState {
  tables: TableState[];
  version: string;
  checksum: string;
}

export interface TableState {
  name: string;
  rowCount: number;
  checksum: string;
  lastModified: Date;
}

// File system state types
export interface FileSystemState {
  files: FileState[];
  directories: DirectoryState[];
  totalSize: number;
}

export interface FileState {
  path: string;
  size: number;
  checksum: string;
  lastModified: Date;
  permissions: string;
}

export interface DirectoryState {
  path: string;
  fileCount: number;
  totalSize: number;
  lastModified: Date;
  permissions: string;
}

// Cache state types
export interface CacheState {
  keys: CacheKeyState[];
  totalSize: number;
  hitRate: number;
}

export interface CacheKeyState {
  key: string;
  size: number;
  ttl: number;
  created: Date;
  accessed: Date;
}

// Test execution types
export interface ExecutionConfig {
  concurrency: number;
  timeout: number;
  retryAttempts: number;
  retryDelay: number;
  bailMode: boolean;
  verbose: boolean;
}

export interface TestExecutorConfig {
  mcpConfig: MCPConfig;
  reporterConfig: ReporterConfig;
  dataConfig: DataConfig;
}

export interface ReporterConfig {
  formats: ('html' | 'json' | 'junit')[];
  outputDir: string;
  includeScreenshots: boolean;
  includeLogs: boolean;
}

export interface DataConfig {
  testDataPath: string;
  templatesPath: string;
  cleanupAfterTest: boolean;
}
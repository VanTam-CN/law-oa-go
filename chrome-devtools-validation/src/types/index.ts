export * from './test-types';
export * from './mcp-types';
export * from './domain-types';

// Re-export commonly used types for convenience
export type {
  TestSuite,
  TestCase,
  TestStep,
  Assertion,
  TestResult,
  TestError,
  TestLog,
  TestExecutionResult,
  ElementInfo,
  PageSnapshot,
  TestConfig,
  BrowserSession,
  FormData,
  ElementState
} from './test-types';

export type {
  ChromeDevToolsConfig,
  DevToolsPage,
  DevToolsElement,
  DevToolsNetworkRequest,
  DevToolsConsoleMessage
} from './mcp-types';

export type {
  UserData,
  ClientData,
  CaseData,
  DocumentData,
  DataSet,
  DataTemplate
} from './domain-types';
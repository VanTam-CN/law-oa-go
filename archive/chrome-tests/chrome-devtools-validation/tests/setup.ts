/**
 * Jest test setup file
 */

// Set test environment variables
process.env['NODE_ENV'] = 'test';

// Set longer timeout for integration tests
jest.setTimeout(30000);

// Mock console methods in tests to reduce noise
global.console = {
  ...console,
  // Uncomment to ignore a specific log level
  // log: jest.fn(),
  // debug: jest.fn(),
  // info: jest.fn(),
  // warn: jest.fn(),
  // error: jest.fn(),
};

// Global test utilities
(global as any).createTestLogger = (context: string) => {
  return {
    debug: (message: string, meta?: any) => console.log(`[DEBUG] [${context}] ${message}`, meta || ''),
    info: (message: string, meta?: any) => console.log(`[INFO] [${context}] ${message}`, meta || ''),
    warn: (message: string, meta?: any) => console.warn(`[WARN] [${context}] ${message}`, meta || ''),
    error: (message: string, meta?: any) => console.error(`[ERROR] [${context}] ${message}`, meta || ''),
  };
};

// Global test helpers
(global as any).delay = (ms: number): Promise<void> => {
  return new Promise(resolve => setTimeout(resolve, ms));
};

(global as any).generateTestId = (): string => {
  return `test_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
};

// Setup global mocks
beforeAll(() => {
  // Global setup before all tests
  console.log('Test suite starting...');
});

afterAll(() => {
  // Global cleanup after all tests
  console.log('Test suite completed.');
});

beforeEach(() => {
  // Setup before each test
});

afterEach(() => {
  // Cleanup after each test
  jest.clearAllMocks();
});
import { Logger } from '../../src/core/logger';

describe('Logger', () => {
  let logger: Logger;

  beforeEach(() => {
    logger = new Logger('test-context');
    // Clear console spies
    jest.spyOn(console, 'debug').mockImplementation(() => {});
    jest.spyOn(console, 'info').mockImplementation(() => {});
    jest.spyOn(console, 'warn').mockImplementation(() => {});
    jest.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('constructor', () => {
    it('should create a logger with context and correlation ID', () => {
      const customLogger = new Logger('custom-context', 'custom-correlation-id');
      expect(customLogger.getCorrelationId()).toBe('custom-correlation-id');
    });

    it('should generate correlation ID if not provided', () => {
      expect(logger.getCorrelationId()).toMatch(/^ctx_\d+_[a-zA-Z0-9]+$/);
    });
  });

  describe('logging methods', () => {
    it('should log debug messages', () => {
      const message = 'Debug message';
      const metadata = { key: 'value' };

      logger.debug(message, metadata);

      expect(console.debug).toHaveBeenCalledWith(
        expect.stringContaining('[DEBUG] [test-context]'),
        expect.stringContaining(message),
        metadata
      );
    });

    it('should log info messages', () => {
      const message = 'Info message';

      logger.info(message);

      expect(console.info).toHaveBeenCalledWith(
        expect.stringContaining('[INFO] [test-context]'),
        expect.stringContaining(message),
        ''
      );
    });

    it('should log warning messages', () => {
      const message = 'Warning message';

      logger.warn(message);

      expect(console.warn).toHaveBeenCalledWith(
        expect.stringContaining('[WARN] [test-context]'),
        expect.stringContaining(message),
        ''
      );
    });

    it('should log error messages', () => {
      const message = 'Error message';
      const metadata = { error: 'details' };

      logger.error(message, metadata);

      expect(console.error).toHaveBeenCalledWith(
        expect.stringContaining('[ERROR] [test-context]'),
        expect.stringContaining(message),
        metadata
      );
    });
  });

  describe('log management', () => {
    it('should store all logs internally', () => {
      logger.debug('Debug message');
      logger.info('Info message');
      logger.warn('Warning message');
      logger.error('Error message');

      const logs = logger.getLogs();
      expect(logs).toHaveLength(4);
    });

    it('should filter logs by level', () => {
      logger.debug('Debug 1');
      logger.info('Info 1');
      logger.debug('Debug 2');
      logger.error('Error 1');

      const debugLogs = logger.getLogsByLevel('debug');
      expect(debugLogs).toHaveLength(2);

      const infoLogs = logger.getLogsByLevel('info');
      expect(infoLogs).toHaveLength(1);

      const errorLogs = logger.getErrorLogs();
      expect(errorLogs).toHaveLength(1);
    });

    it('should clear all logs', () => {
      logger.debug('Message 1');
      logger.info('Message 2');

      expect(logger.getLogs()).toHaveLength(2);

      logger.clearLogs();

      expect(logger.getLogs()).toHaveLength(0);
    });
  });

  describe('log structure', () => {
    it('should create properly structured log entries', () => {
      const metadata = { userId: '123', action: 'login' };
      logger.info('User logged in', metadata);

      const logs = logger.getLogs();
      const log = logs[0];

      expect(log).toMatchObject({
        level: 'info',
        message: 'User logged in',
        metadata: {
          correlationId: expect.any(String),
          context: 'test-context',
          userId: '123',
          action: 'login',
        },
      });
      expect(log?.timestamp).toBeInstanceOf(Date);
    });
  });

  describe('export functionality', () => {
    it('should export logs as JSON string', () => {
      logger.info('Test message', { key: 'value' });

      const exported = logger.exportLogs();

      expect(() => JSON.parse(exported)).not.toThrow();
      const parsed = JSON.parse(exported);
      expect(parsed).toHaveLength(1);
      expect(parsed[0].message).toBe('Test message');
    });
  });

  describe('child logger', () => {
    it('should create child logger with same correlation ID', () => {
      const childLogger = Logger.createChildLogger(logger, 'child-context');

      expect(childLogger.getCorrelationId()).toBe(logger.getCorrelationId());
    });

    it('should child logger should work independently', () => {
      const childLogger = Logger.createChildLogger(logger, 'child-context');

      logger.info('Parent message');
      childLogger.info('Child message');

      expect(logger.getLogs()).toHaveLength(1);
      expect(childLogger.getLogs()).toHaveLength(1);
    });
  });
});
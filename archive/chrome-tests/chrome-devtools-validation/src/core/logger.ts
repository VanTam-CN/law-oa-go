import { TestLog, LogLevel } from '../types';

export class Logger {
  private logs: TestLog[] | undefined = undefined;
  private correlationId: string;
  private context: string;

  constructor(context: string, correlationId?: string) {
    this.context = context;
    this.correlationId = correlationId || this.generateCorrelationId();
  }

  private generateCorrelationId(): string {
    return `ctx_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private createLog(level: LogLevel, message: string, metadata?: Record<string, any>): TestLog {
    return {
      timestamp: new Date(),
      level,
      message,
      metadata: {
        ...metadata,
        correlationId: this.correlationId,
        context: this.context,
      },
    };
  }

  private log(level: LogLevel, message: string, metadata?: Record<string, any>): void {
    const log = this.createLog(level, message, metadata);
    this.logs.push(log);

    // Also output to console for development
    const timestamp = log.timestamp.toISOString();
    const prefix = `[${timestamp}] [${level.toUpperCase()}] [${this.context}]`;
    const messageWithId = `${message} (cid: ${this.correlationId})`;

    switch (level) {
      case 'debug':
        console.debug(prefix, messageWithId, metadata || '');
        break;
      case 'info':
        console.info(prefix, messageWithId, metadata || '');
        break;
      case 'warn':
        console.warn(prefix, messageWithId, metadata || '');
        break;
      case 'error':
        console.error(prefix, messageWithId, metadata || '');
        break;
    }
  }

  trace(message: string, metadata?: Record<string, any>): void {
    this.log('debug', message, metadata);
  }

  debug(message: string, metadata?: Record<string, any>): void {
    this.log('debug', message, metadata);
  }

  info(message: string, metadata?: Record<string, any>): void {
    this.log('info', message, metadata);
  }

  warn(message: string, metadata?: Record<string, any>): void {
    this.log('warn', message, metadata);
  }

  error(message: string, metadata?: Record<string, any>): void {
    this.log('error', message, metadata);
  }

  getLogs(): TestLog[] {
    return [...this.logs];
  }

  getLogsByLevel(level: LogLevel): TestLog[] {
    return this.logs.filter(log => log.level === level);
  }

  getErrorLogs(): TestLog[] {
    return this.getLogsByLevel('error');
  }

  getWarningLogs(): TestLog[] {
    return this.getLogsByLevel('warn');
  }

  clearLogs(): void {
    this.logs = [];
  }

  exportLogs(): string {
    return JSON.stringify(this.logs, null, 2);
  }

  getCorrelationId(): string {
    return this.correlationId;
  }

  static createChildLogger(parent: Logger, context: string): Logger {
    return new Logger(context, parent.getCorrelationId());
  }
}
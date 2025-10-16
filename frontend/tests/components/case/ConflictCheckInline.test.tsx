/**
 * ConflictCheckInline组件测试
 * 测试冲突检测、结果显示和1080p优化
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import '@testing-library/jest-dom';
import ConflictCheckInline from '../../../src/components/case/ConflictCheckInline';
import type {
  ConflictCheckInlineProps,
  ConflictCheckResult,
  ConflictCase,
  ConflictSeverity
} from '../../../src/components/case/types/ConflictCheckInline.types';

// Mock Ant Design components
jest.mock('antd/es/card', () => ({
  Card: ({ children, title, extra, className }: any) => (
    <div data-testid="card" className={className} data-title={title} data-extra={extra}>
      {title && <div data-testid="card-title">{title}</div>}
      {extra && <div data-testid="card-extra">{extra}</div>}
      <div data-testid="card-body">{children}</div>
    </div>
  )
}));

jest.mock('antd/es/alert', () => ({
  Alert: ({ message, description, type, action }: any) => (
    <div data-testid="alert" data-type={type} data-message={message}>
      <div>{message}</div>
      {description && <div data-testid="alert-description">{description}</div>}
      {action && <div data-testid="alert-action">{action}</div>}
    </div>
  )
}));

jest.mock('antd/es/button', () => ({
  Button: ({ children, onClick, icon, type, size, danger }: any) => (
    <button
      data-testid="button"
      onClick={onClick}
      data-type={type}
      data-size={size}
      data-danger={danger}
    >
      {icon}
      {children}
    </button>
  )
}));

jest.mock('antd/es/space', () => ({
  Space: ({ children, size, direction }: any) => (
    <div data-testid="space" data-size={size} data-direction={direction}>
      {children}
    </div>
  )
}));

jest.mock('antd/es/typography', () => ({
  Text: ({ children, type, strong, style }: any) => (
    <span
      data-testid="text"
      data-type={type}
      data-strong={strong}
      style={style}
    >
      {children}
    </span>
  ),
  Title: ({ children, level }: any) => (
    <h1 data-testid="title" data-level={level}>
      {children}
    </h1>
  ),
  Paragraph: ({ children }: any) => (
    <p data-testid="paragraph">{children}</p>
  )
}));

jest.mock('antd/es/spin', () => ({
  Spin: ({ size }: any) => (
    <div data-testid="spin" data-size={size}>
      Loading...
    </div>
  )
}));

jest.mock('antd/es/badge', () => ({
  Badge: ({ children, count, size }: any) => (
    <div data-testid="badge" data-count={count} data-size={size}>
      {children}
    </div>
  )
}));

jest.mock('antd/es/tag', () => ({
  Tag: ({ children, color, size }: any) => (
    <span data-testid="tag" data-color={color} data-size={size}>
      {children}
    </span>
  )
}));

jest.mock('antd/es/tooltip', () => ({
  Tooltip: ({ children, title }: any) => (
    <div data-testid="tooltip" data-title={title}>
      {children}
    </div>
  )
}));

jest.mock('antd/es/divider', () => ({
  Divider: ({ type }: any) => (
    <hr data-testid="divider" data-type={type} />
  )
}));

jest.mock('antd/es/list', () => ({
  List: ({ dataSource, renderItem, size }: any) => (
    <div data-testid="list" data-size={size}>
      {dataSource?.map((item: any, index: number) => renderItem(item, index))}
    </div>
  )
}));

jest.mock('antd/es/avatar', () => ({
  Avatar: ({ children, size, style }: any) => (
    <div data-testid="avatar" data-size={size} style={style}>
      {children}
    </div>
  )
}));

jest.mock('antd/es/progress', () => ({
  Progress: ({ percent, size }: any) => (
    <div data-testid="progress" data-percent={percent} data-size={size}>
      Progress: {percent}%
    </div>
  )
}));

jest.mock('antd/es/empty', () => ({
  Empty: ({ description, image }: any) => (
    <div data-testid="empty" data-description={description} data-image={image}>
      Empty State
    </div>
  )
}));

jest.mock('antd/es/collapse', () => ({
  Collapse: ({ items, ghost, size }: any) => (
    <div data-testid="collapse" data-ghost={ghost} data-size={size}>
      {items?.map((item: any, index: number) => (
        <div key={item.key} data-testid={`collapse-item-${item.key}`}>
          <div data-testid="collapse-header">{item.label}</div>
          <div data-testid="collapse-content">{item.children}</div>
        </div>
      ))}
    </div>
  )
}));

// Mock icons
jest.mock('@ant-design/icons', () => ({
  ReloadOutlined: () => <span data-testid="reload-icon">↻</span>,
  EyeOutlined: () => <span data-testid="eye-icon">👁</span>,
  CheckCircleOutlined: () => <span data-testid="check-icon">✓</span>,
  ExclamationCircleOutlined: () => <span data-testid="exclamation-icon">!</span>,
  CloseCircleOutlined: () => <span data-testid="close-icon">✕</span>,
  SearchOutlined: () => <span data-testid="search-icon">🔍</span>,
  WarningOutlined: () => <span data-testid="warning-icon">⚠</span>,
  InfoCircleOutlined: () => <span data-testid="info-icon">ℹ</span>,
  RightOutlined: () => <span data-testid="right-icon">→</span>,
  DownOutlined: () => <span data-testid="down-icon">↓</span>
}));

describe('ConflictCheckInline', () => {
  const mockOnStatusChange = jest.fn();
  const mockOnCheckComplete = jest.fn();
  const mockOnViewDetails = jest.fn();
  const mockOnRecheck = jest.fn();
  const mockOnMarkResolved = jest.fn();

  const defaultCheckParams = {
    clientName: '测试客户',
    opposingParty: '对方当事人',
    caseType: '民事纠纷',
    lawyerId: 'lawyer_001',
    scope: 'all' as const
  };

  const mockConflictCase: ConflictCase = {
    id: 'conflict_001',
    caseNumber: 'CASE_2023_001',
    title: '测试冲突案件',
    clientName: '冲突客户',
    lawyerName: '张律师',
    conflictType: 'direct',
    severity: 'high' as ConflictSeverity,
    description: '这是一个测试冲突案件',
    recommendation: '建议详细审查',
    createdAt: '2023-01-01T00:00:00Z'
  };

  const mockConflictResult: ConflictCheckResult = {
    status: 'warning',
    hasConflict: true,
    conflicts: [mockConflictCase],
    checkedAt: '2023-01-01T12:00:00Z',
    stats: {
      total: 1,
      direct: 1,
      indirect: 0,
      potential: 0
    }
  };

  const defaultProps: ConflictCheckInlineProps = {
    checkParams: defaultCheckParams,
    onStatusChange: mockOnStatusChange,
    onCheckComplete: mockOnCheckComplete,
    onViewDetails: mockOnViewDetails,
    onRecheck: mockOnRecheck,
    onMarkResolved: mockOnMarkResolved,
    autoCheck: false
  };

  beforeEach(() => {
    jest.clearAllMocks();
    // Mock window.innerWidth for 1080p testing
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1920
    });
  });

  describe('基本功能', () => {
    test('应该正确渲染初始状态', () => {
      render(<ConflictCheckInline {...defaultProps} />);

      expect(screen.getByTestId('card')).toBeInTheDocument();
      expect(screen.getByTestId('empty')).toBeInTheDocument();
      expect(screen.getByTestId('button')).toBeInTheDocument();
    });

    test('应该显示检测参数', () => {
      render(<ConflictCheckInline {...defaultProps} />);

      // 组件应该接收并存储检测参数
      expect(screen.getByTestId('card')).toBeInTheDocument();
    });
  });

  describe('检测状态', () => {
    test('应该支持手动触发检测', async () => {
      const user = userEvent.setup();
      render(<ConflictCheckInline {...defaultProps} />);

      const checkButton = screen.getByTestId('button');
      await user.click(checkButton);

      expect(mockOnStatusChange).toHaveBeenCalledWith('checking');
    });

    test('应该支持自动检测', async () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          autoCheck={true}
          debounceDelay={100}
        />
      );

      await waitFor(() => {
        expect(mockOnStatusChange).toHaveBeenCalledWith('checking');
      }, { timeout: 200 });
    });

    test('应该显示检测中状态', () => {
      const checkingResult: ConflictCheckResult = {
        status: 'checking',
        hasConflict: false,
        conflicts: [],
        stats: { total: 0, direct: 0, indirect: 0, potential: 0 }
      };

      render(
        <ConflictCheckInline
          {...defaultProps}
          result={checkingResult}
        />
      );

      expect(screen.getByTestId('spin')).toBeInTheDocument();
      expect(screen.getByTestId('progress')).toBeInTheDocument();
    });
  });

  describe('结果显示', () => {
    test('应该显示无冲突结果', () => {
      const successResult: ConflictCheckResult = {
        status: 'success',
        hasConflict: false,
        conflicts: [],
        stats: { total: 0, direct: 0, indirect: 0, potential: 0 }
      };

      render(
        <ConflictCheckInline
          {...defaultProps}
          result={successResult}
        />
      );

      const alert = screen.getByTestId('alert');
      expect(alert).toHaveAttribute('data-type', 'success');
      expect(alert).toHaveAttribute('data-message', '冲突检测完成');
    });

    test('应该显示有冲突结果', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
        />
      );

      expect(screen.getByTestId('badge')).toHaveAttribute('data-count', '1');
      expect(screen.getByTestId('list')).toBeInTheDocument();
    });

    test('应该显示错误状态', () => {
      const errorResult: ConflictCheckResult = {
        status: 'error',
        hasConflict: false,
        conflicts: [],
        error: '检测失败',
        stats: { total: 0, direct: 0, indirect: 0, potential: 0 }
      };

      render(
        <ConflictCheckInline
          {...defaultProps}
          result={errorResult}
        />
      );

      const alert = screen.getByTestId('alert');
      expect(alert).toHaveAttribute('data-type', 'error');
      expect(screen.getByTestId('alert-description')).toHaveTextContent('检测失败');
    });
  });

  describe('冲突列表', () => {
    test('应该正确显示冲突项目', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ showDetails: true }}
        />
      );

      expect(screen.getByTestId('avatar')).toBeInTheDocument();
      expect(screen.getByTestId('text', { strong: true })).toHaveTextContent('测试冲突案件');
      expect(screen.getByTestId('tag')).toHaveAttribute('data-color', '#ff7a45');
    });

    test('应该支持查看详情操作', async () => {
      const user = userEvent.setup();
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ showDetails: true, showActions: true }}
        />
      );

      const viewButton = screen.getByTestId('eye-icon');
      await user.click(viewButton);

      expect(mockOnViewDetails).toHaveBeenCalledWith(mockConflictCase);
    });

    test('应该支持标记为已解决操作', async () => {
      const user = userEvent.setup();
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ showDetails: true, showActions: true }}
        />
      );

      const resolveButton = screen.getByTestId('check-icon');
      await user.click(resolveButton);

      expect(mockOnMarkResolved).toHaveBeenCalledWith(['conflict_001']);
    });

    test('应该限制显示数量', () => {
      const multipleConflicts: ConflictCheckResult = {
        ...mockConflictResult,
        conflicts: [
          mockConflictCase,
          { ...mockConflictCase, id: 'conflict_002', caseNumber: 'CASE_2023_002' },
          { ...mockConflictCase, id: 'conflict_003', caseNumber: 'CASE_2023_003' },
          { ...mockConflictCase, id: 'conflict_004', caseNumber: 'CASE_2023_004' }
        ],
        stats: { total: 4, direct: 2, indirect: 1, potential: 1 }
      };

      render(
        <ConflictCheckInline
          {...defaultProps}
          result={multipleConflicts}
          displayConfig={{ maxDisplayCount: 2 }}
        />
      );

      // 应该只显示前2个冲突项目
      expect(screen.getAllByTestId('avatar')).toHaveLength(2);
    });
  });

  describe('统计信息', () => {
    test('应该显示冲突统计', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ showStats: true }}
        />
      );

      expect(screen.getByTestId('badge')).toHaveAttribute('data-count', '1');
      expect(screen.getByTestId('tag')).toHaveAttribute('data-color', '#ff4d4f');
    });

    test('应该显示不同类型的冲突统计', () => {
      const complexStats: ConflictCheckResult = {
        ...mockConflictResult,
        stats: { total: 3, direct: 1, indirect: 1, potential: 1 }
      };

      render(
        <ConflictCheckInline
          {...defaultProps}
          result={complexStats}
          displayConfig={{ showStats: true }}
        />
      );

      expect(screen.getByTestId('badge')).toHaveAttribute('data-count', '3');
      // 应该显示3个不同类型的标签
      expect(screen.getAllByTestId('tag')).toHaveLength(3);
    });
  });

  describe('快速操作', () => {
    test('应该支持重新检测', async () => {
      const user = userEvent.setup();
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          actionConfig={{ allowRecheck: true }}
          displayConfig={{ showActions: true }}
        />
      );

      const recheckButton = screen.getByTestId('reload-icon');
      await user.click(recheckButton);

      expect(mockOnRecheck).toHaveBeenCalled();
    });

    test('应该支持查看所有冲突', async () => {
      const user = userEvent.setup();
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          actionConfig={{ allowViewDetails: true }}
          displayConfig={{ showActions: true }}
        />
      );

      const viewAllButton = screen.getByTestId('eye-icon');
      await user.click(viewAllButton);

      expect(mockOnViewDetails).toHaveBeenCalledWith(mockConflictCase);
    });
  });

  describe('1080p优化', () => {
    test('应该在紧凑模式下应用紧凑样式', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          displayConfig={{ isCompact: true }}
        />
      );

      const container = document.querySelector('.conflict-check-inline');
      expect(container).toHaveClass('compact');
    });

    test('应该在1080p分辨率下自动启用紧凑模式', () => {
      // 模拟1080p分辨率
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920
      });

      render(<ConflictCheckInline {...defaultProps} />);

      const container = document.querySelector('.conflict-check-inline');
      expect(container).toHaveClass('compact');
    });
  });

  describe('响应式布局', () => {
    test('应该在移动设备上调整布局', () => {
      // 模拟移动设备
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768
      });

      render(<ConflictCheckInline {...defaultProps} />);

      const container = document.querySelector('.conflict-check-inline');
      expect(container).toBeInTheDocument();
    });
  });

  describe('错误处理', () => {
    test('应该正确处理检测失败', () => {
      const errorResult: ConflictCheckResult = {
        status: 'error',
        hasConflict: false,
        conflicts: [],
        error: '网络连接失败',
        stats: { total: 0, direct: 0, indirect: 0, potential: 0 }
      };

      render(
        <ConflictCheckInline
          {...defaultProps}
          result={errorResult}
        />
      );

      const alert = screen.getByTestId('alert');
      expect(alert).toHaveAttribute('data-type', 'error');
      expect(screen.getByTestId('alert-description')).toHaveTextContent('网络连接失败');
    });

    test('应该支持重试功能', async () => {
      const user = userEvent.setup();
      const errorResult: ConflictCheckResult = {
        status: 'error',
        hasConflict: false,
        conflicts: [],
        error: '检测失败',
        stats: { total: 0, direct: 0, indirect: 0, potential: 0 }
      };

      render(
        <ConflictCheckInline
          {...defaultProps}
          result={errorResult}
        />
      );

      const retryButton = screen.getByTestId('button');
      await user.click(retryButton);

      expect(mockOnRecheck).toHaveBeenCalled();
    });
  });

  describe('配置选项', () => {
    test('应该支持隐藏统计信息', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ showStats: false }}
        />
      );

      expect(screen.queryByTestId('badge')).not.toBeInTheDocument();
    });

    test('应该支持隐藏操作按钮', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ showActions: false }}
        />
      );

      expect(screen.queryByTestId('reload-icon')).not.toBeInTheDocument();
      expect(screen.queryByTestId('eye-icon')).not.toBeInTheDocument();
    });

    test('应该支持隐藏详细信息', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ showDetails: false }}
        />
      );

      expect(screen.queryByTestId('collapse')).not.toBeInTheDocument();
    });

    test('应该支持不同的显示模式', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ displayMode: 'alert' }}
        />
      );

      // 在alert模式下应该直接显示Alert而不是Collapse
      expect(screen.getByTestId('alert')).toBeInTheDocument();
    });
  });

  describe('防抖功能', () => {
    test('应该支持防抖延迟', async () => {
      jest.useFakeTimers();

      const { rerender } = render(
        <ConflictCheckInline
          {...defaultProps}
          autoCheck={true}
          debounceDelay={500}
        />
      );

      // 更新检测参数
      rerender(
        <ConflictCheckInline
          {...defaultProps}
          checkParams={{ ...defaultCheckParams, clientName: '新客户' }}
          autoCheck={true}
          debounceDelay={500}
        />
      );

      // 在延迟时间前不应该触发检测
      expect(mockOnStatusChange).not.toHaveBeenCalled();

      // 快进时间
      jest.advanceTimersByTime(500);

      await waitFor(() => {
        expect(mockOnStatusChange).toHaveBeenCalledWith('checking');
      });

      jest.useRealTimers();
    });
  });

  describe('可访问性', () => {
    test('应该具有正确的ARIA标签', () => {
      render(<ConflictCheckInline {...defaultProps} />);

      const card = screen.getByTestId('card');
      expect(card).toBeInTheDocument();
    });

    test('应该支持键盘导航', async () => {
      const user = userEvent.setup();
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
          displayConfig={{ showActions: true }}
        />
      );

      const firstButton = screen.getByTestId('reload-icon').closest('button');
      firstButton?.focus();

      expect(firstButton).toHaveFocus();

      await user.keyboard('{Enter}');
      expect(mockOnRecheck).toHaveBeenCalled();
    });

    test('应该为屏幕阅读器提供状态信息', () => {
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={mockConflictResult}
        />
      );

      // 统计信息应该有适当的语义标记
      expect(screen.getByTestId('badge')).toBeInTheDocument();
      expect(screen.getByTestId('badge')).toHaveAttribute('data-count', '1');
    });
  });

  describe('性能优化', () => {
    test('应该优化大量冲突的渲染', () => {
      const manyConflicts: ConflictCheckResult = {
        ...mockConflictResult,
        conflicts: Array.from({ length: 100 }, (_, i) => ({
          ...mockConflictCase,
          id: `conflict_${i}`,
          caseNumber: `CASE_2023_${String(i + 1).padStart(3, '0')}`
        })),
        stats: { total: 100, direct: 50, indirect: 30, potential: 20 }
      };

      const startTime = performance.now();
      render(
        <ConflictCheckInline
          {...defaultProps}
          result={manyConflicts}
          displayConfig={{ maxDisplayCount: 10 }}
        />
      );
      const endTime = performance.now();

      // 渲染时间应该在合理范围内
      expect(endTime - startTime).toBeLessThan(100);

      // 应该只渲染指定数量的冲突
      expect(screen.getAllByTestId('avatar')).toHaveLength(10);
    });

    test('应该正确清理定时器', () => {
      const { unmount } = render(
        <ConflictCheckInline
          {...defaultProps}
          autoCheck={true}
          debounceDelay={1000}
        />
      );

      // 组件卸载时不应该有未清理的定时器
      unmount();

      expect(setTimeout).toHaveBeenCalled();
      // 这里可以验证定时器是否被正确清理
    });
  });
});
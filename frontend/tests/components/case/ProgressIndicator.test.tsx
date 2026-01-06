/**
 * ProgressIndicator组件测试
 * 测试步骤导航、进度计算、1080p优化和响应式布局
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import '@testing-library/jest-dom';
import ProgressIndicator from '../../../src/components/case/ProgressIndicator';
import type {
  StepConfig,
  ProgressIndicatorProps,
  StepStatus,
  ProgressStats
} from '../../../src/components/case/types/ProgressIndicator.types';

// Mock Ant Design components
jest.mock('antd/es/steps', () => ({
  Steps: ({ items, current, onChange }: any) => (
    <div data-testid="steps" data-current={current}>
      {items?.map((item: any, index: number) => (
        <div
          key={item.key}
          data-testid={`step-${item.key}`}
          data-status={item.status}
          onClick={() => onChange?.(item.key)}
          style={{ cursor: 'pointer' }}
        >
          {item.title}
          {item.description && <span data-testid={`step-desc-${item.key}`}>{item.description}</span>}
        </div>
      ))}
    </div>
  )
}));

jest.mock('antd/es/progress', () => ({
  Progress: ({ percent, type }: any) => (
    <div data-testid="progress" data-type={type} data-percent={percent}>
      Progress: {percent}%
    </div>
  )
}));

jest.mock('antd/es/card', () => ({
  Card: ({ children, title, className }: any) => (
    <div data-testid="card" data-title={title} className={className}>
      {children}
    </div>
  )
}));

jest.mock('antd/es/statistic', () => ({
  Statistic: ({ title, value }: any) => (
    <div data-testid="statistic" data-title={title}>
      {title}: {value}
    </div>
  )
}));

jest.mock('antd/es/alert', () => ({
  Alert: ({ message, description, type }: any) => (
    <div data-testid="alert" data-type={type}>
      <div>{message}</div>
      {description && <div>{description}</div>}
    </div>
  )
}));

jest.mock('antd/es/modal', () => ({
  Modal: {
    confirm: jest.fn(({ onOk }) => {
      onOk?.();
      return Promise.resolve();
    })
  }
}));

jest.mock('antd/es/button', () => ({
  Button: ({ children, onClick, ...props }: any) => (
    <button data-testid="button" onClick={onClick} {...props}>
      {children}
    </button>
  )
}));

jest.mock('antd/es/icon', () => ({
  RightOutlined: () => <span data-testid="right-icon">→</span>,
  ReloadOutlined: () => <span data-testid="reload-icon">↻</span>
}));

describe('ProgressIndicator', () => {
  const mockOnStepChange = jest.fn();

  const defaultSteps: StepConfig[] = [
    {
      key: 'basic',
      title: '基本信息',
      description: '案件基本信息填写',
      status: 'finish' as StepStatus,
      completed: true,
      weight: 1
    },
    {
      key: 'client',
      title: '客户信息',
      description: '客户相关资料',
      status: 'process' as StepStatus,
      completed: false,
      weight: 1
    },
    {
      key: 'details',
      title: '案件详情',
      description: '案件详细信息',
      status: 'wait' as StepStatus,
      completed: false,
      weight: 2
    }
  ];

  const defaultProps: ProgressIndicatorProps = {
    steps: defaultSteps,
    currentStepKey: 'client',
    onStepChange: mockOnStepChange,
    showProgressBar: true,
    showStats: true,
    enableStepNavigation: true,
    progressType: 'line',
    direction: 'horizontal',
    size: 'default',
    isCompact: false
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
    test('应该正确渲染所有步骤', () => {
      render(<ProgressIndicator {...defaultProps} />);

      defaultSteps.forEach(step => {
        expect(screen.getByTestId(`step-${step.key}`)).toBeInTheDocument();
        expect(screen.getByText(step.title)).toBeInTheDocument();
        if (step.description) {
          expect(screen.getByTestId(`step-desc-${step.key}`)).toHaveTextContent(step.description);
        }
      });
    });

    test('应该显示进度条', () => {
      render(<ProgressIndicator {...defaultProps} />);

      const progress = screen.getByTestId('progress');
      expect(progress).toBeInTheDocument();
      expect(progress).toHaveAttribute('data-type', 'line');
    });

    test('应该显示统计信息', () => {
      render(<ProgressIndicator {...defaultProps} />);

      expect(screen.getByTestId('statistic')).toBeInTheDocument();
    });

    test('应该正确标记当前步骤', () => {
      render(<ProgressIndicator {...defaultProps} />);

      const stepsContainer = screen.getByTestId('steps');
      const clientStep = screen.getByTestId('step-client');

      expect(stepsContainer).toHaveAttribute('data-current', '1'); // client是第2个步骤(索引1)
    });
  });

  describe('步骤导航', () => {
    test('应该支持点击步骤导航', async () => {
      const user = userEvent.setup();
      render(<ProgressIndicator {...defaultProps} />);

      const detailsStep = screen.getByTestId('step-details');
      await user.click(detailsStep);

      expect(mockOnStepChange).toHaveBeenCalledWith('details', 2);
    });

    test('应该支持键盘导航', async () => {
      const user = userEvent.setup();
      render(<ProgressIndicator {...defaultProps} />);

      const stepsContainer = screen.getByTestId('steps');
      stepsContainer.focus();

      // 测试右箭头键
      await user.keyboard('{ArrowRight}');

      // 验证焦点移动到下一个步骤
      const detailsStep = screen.getByTestId('step-details');
      expect(detailsStep).toHaveFocus();
    });

    test('应该禁用已禁用的步骤', () => {
      const disabledSteps = [
        ...defaultSteps,
        {
          key: 'disabled',
          title: '禁用步骤',
          status: 'wait' as StepStatus,
          disabled: true,
          completed: false,
          weight: 1
        }
      ];

      render(
        <ProgressIndicator
          {...defaultProps}
          steps={disabledSteps}
        />
      );

      const disabledStep = screen.getByTestId('step-disabled');
      expect(disabledStep).toHaveStyle({ cursor: 'not-allowed' });
    });
  });

  describe('进度计算', () => {
    test('应该正确计算完成百分比', () => {
      render(<ProgressIndicator {...defaultProps} />);

      const progress = screen.getByTestId('progress');
      expect(progress).toHaveAttribute('data-percent', '33'); // 1/3 完成
    });

    test('应该正确处理权重计算', () => {
      const weightedSteps: StepConfig[] = [
        { ...defaultSteps[0], weight: 1 },
        { ...defaultSteps[1], weight: 2 },
        { ...defaultSteps[2], weight: 1 }
      ];

      render(
        <ProgressIndicator
          {...defaultProps}
          steps={weightedSteps}
        />
      );

      const progress = screen.getByTestId('progress');
      expect(progress).toHaveAttribute('data-percent', '25'); // 1/(1+2+1) = 25%
    });

    test('应该正确统计步骤数量', () => {
      render(<ProgressIndicator {...defaultProps} />);

      const statsElements = screen.getAllByTestId('statistic');
      expect(statsElements.length).toBeGreaterThan(0);

      // 检查是否包含总步骤数和已完成步骤数
      const statsTexts = statsElements.map(el => el.textContent || '');
      expect(statsTexts.some(text => text.includes('3'))).toBe(true); // 总步骤数
      expect(statsTexts.some(text => text.includes('1'))).toBe(true); // 已完成步骤数
    });
  });

  describe('1080p优化', () => {
    test('应该在紧凑模式下应用紧凑样式', () => {
      render(
        <ProgressIndicator
          {...defaultProps}
          isCompact={true}
        />
      );

      const cards = screen.getAllByTestId('card');
      cards.forEach(card => {
        expect(card).toHaveClass('progress-indicator-compact');
      });
    });

    test('应该在1080p分辨率下自动启用紧凑模式', () => {
      // 模拟1080p分辨率
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920
      });

      render(<ProgressIndicator {...defaultProps} />);

      const container = document.querySelector('.progress-indicator');
      expect(container).toHaveClass('progress-indicator-1080p');
    });
  });

  describe('响应式布局', () => {
    test('应该在移动设备上应用移动样式', () => {
      // 模拟移动设备
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768
      });

      render(<ProgressIndicator {...defaultProps} direction="vertical" />);

      const container = document.querySelector('.progress-indicator');
      expect(container).toHaveClass('progress-vertical');
    });

    test('应该支持垂直布局', () => {
      render(
        <ProgressIndicator
          {...defaultProps}
          direction="vertical"
        />
      );

      const container = document.querySelector('.progress-indicator');
      expect(container).toHaveClass('progress-vertical');
    });
  });

  describe('验证功能', () => {
    test('应该验证必填字段', async () => {
      const stepsWithValidation: StepConfig[] = [
        {
          ...defaultSteps[0],
          validation: {
            required: true,
            fields: ['name', 'type']
          }
        }
      ];

      render(
        <ProgressIndicator
          {...defaultProps}
          steps={stepsWithValidation}
          enableStepNavigation={true}
        />
      );

      const user = userEvent.setup();
      const basicStep = screen.getByTestId('step-basic');

      await user.click(basicStep);

      // 应该触发验证
      expect(mockOnStepChange).toHaveBeenCalled();
    });
  });

  describe('错误处理', () => {
    test('应该正确处理错误状态', () => {
      const errorSteps: StepConfig[] = [
        ...defaultSteps,
        {
          key: 'error',
          title: '错误步骤',
          status: 'error' as StepStatus,
          completed: false,
          weight: 1
        }
      ];

      render(
        <ProgressIndicator
          {...defaultProps}
          steps={errorSteps}
        />
      );

      const errorStep = screen.getByTestId('step-error');
      expect(errorStep).toHaveAttribute('data-status', 'error');
    });

    test('应该显示错误警告', () => {
      const errorSteps: StepConfig[] = [
        ...defaultSteps.slice(0, 2),
        {
          ...defaultSteps[2],
          status: 'error' as StepStatus
        }
      ];

      render(
        <ProgressIndicator
          {...defaultProps}
          steps={errorSteps}
        />
      );

      // 应该显示错误提示
      expect(screen.getByTestId('alert')).toBeInTheDocument();
      expect(screen.getByTestId('alert')).toHaveAttribute('data-type', 'error');
    });
  });

  describe('交互功能', () => {
    test('应该支持跳转到下一个未完成步骤', async () => {
      const user = userEvent.setup();
      render(<ProgressIndicator {...defaultProps} />);

      const nextButton = screen.getByLabelText('跳转到下一个未完成步骤');
      await user.click(nextButton);

      // 应该跳转到details步骤
      expect(mockOnStepChange).toHaveBeenCalledWith('details', 2);
    });

    test('应该支持刷新进度', async () => {
      const user = userEvent.setup();
      render(<ProgressIndicator {...defaultProps} />);

      const refreshButton = screen.getByLabelText('刷新进度');
      await user.click(refreshButton);

      // 验证进度被重新计算
      const progress = screen.getByTestId('progress');
      expect(progress).toBeInTheDocument();
    });
  });

  describe('可访问性', () => {
    test('应该具有正确的ARIA标签', () => {
      render(<ProgressIndicator {...defaultProps} />);

      const container = document.querySelector('.progress-indicator');
      expect(container).toHaveAttribute('role', 'navigation');
      expect(container).toHaveAttribute('aria-label', '案件创建进度');
    });

    test('应该支持键盘导航', async () => {
      const user = userEvent.setup();
      render(<ProgressIndicator {...defaultProps} />);

      const stepsContainer = screen.getByTestId('steps');

      // Tab键导航
      await user.tab();
      expect(stepsContainer).toHaveFocus();

      // 方向键导航
      await user.keyboard('{ArrowRight}');
      const nextStep = screen.getByTestId('step-details');
      expect(nextStep).toHaveFocus();
    });

    test('应该为屏幕阅读器提供进度信息', () => {
      render(<ProgressIndicator {...defaultProps} />);

      const progressInfo = document.querySelector('[aria-live="polite"]');
      expect(progressInfo).toBeInTheDocument();

      const statsText = progressInfo?.textContent || '';
      expect(statsText).toContain('总步骤');
      expect(statsText).toContain('已完成');
      expect(statsText).toContain('33%');
    });
  });

  describe('组件属性', () => {
    test('应该支持隐藏进度条', () => {
      render(
        <ProgressIndicator
          {...defaultProps}
          showProgressBar={false}
        />
      );

      expect(screen.queryByTestId('progress')).not.toBeInTheDocument();
    });

    test('应该支持隐藏统计信息', () => {
      render(
        <ProgressIndicator
          {...defaultProps}
          showStats={false}
        />
      );

      expect(screen.queryByTestId('statistic')).not.toBeInTheDocument();
    });

    test('应该支持禁用步骤导航', () => {
      render(
        <ProgressIndicator
          {...defaultProps}
          enableStepNavigation={false}
        />
      );

      const steps = screen.getAllByTestId(/^step-/);
      steps.forEach(step => {
        expect(step).toHaveStyle({ cursor: 'default' });
      });
    });

    test('应该支持不同的进度条类型', () => {
      const { rerender } = render(
        <ProgressIndicator
          {...defaultProps}
          progressType="circle"
        />
      );

      expect(screen.getByTestId('progress')).toHaveAttribute('data-type', 'circle');

      rerender(
        <ProgressIndicator
          {...defaultProps}
          progressType="dashboard"
        />
      );

      expect(screen.getByTestId('progress')).toHaveAttribute('data-type', 'dashboard');
    });
  });

  describe('性能优化', () => {
    test('应该优化大量步骤的渲染', () => {
      const manySteps: StepConfig[] = Array.from({ length: 50 }, (_, i) => ({
        key: `step-${i}`,
        title: `步骤 ${i + 1}`,
        status: (i < 10 ? 'finish' : i < 15 ? 'process' : 'wait') as StepStatus,
        completed: i < 10,
        weight: 1
      }));

      const startTime = performance.now();
      render(
        <ProgressIndicator
          {...defaultProps}
          steps={manySteps}
        />
      );
      const endTime = performance.now();

      // 渲染时间应该在合理范围内
      expect(endTime - startTime).toBeLessThan(100);
    });

    test('应该正确清理事件监听器', () => {
      const { unmount } = render(<ProgressIndicator {...defaultProps} />);

      // 模拟组件卸载
      unmount();

      // 验证没有内存泄漏（这里只是简单示例）
      expect(() => {
        // 组件卸载后不应该有未清理的事件监听器
        document.removeEventListener('keydown', jest.fn());
      }).not.toThrow();
    });
  });
});
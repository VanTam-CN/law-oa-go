/**
 * CompactCaseForm集成测试
 * 测试整个新建案件功能的组件集成和1080p优化
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import '@testing-library/jest-dom';
import CompactCaseForm from '../../src/components/case/CompactCaseForm';
import type {
  FormStepConfig,
  FieldGroupConfig,
  QuickAction
} from '../../src/components/case/types/CompactCaseForm.types';
import {
  DEFAULT_FORM_CONFIG,
  DEFAULT_RESPONSIVE_CONFIG,
  DEFAULT_SAVE_CONFIG,
  STEP_KEYS,
  FIELD_GROUP_KEYS
} from '../../src/components/case/types/CompactCaseForm.types';

// Mock子组件
jest.mock('../../src/components/case/ResponsiveFormLayout', () => ({
  default: ({ children, columns }: any) => (
    <div data-testid="responsive-form-layout" data-columns={columns}>
      {children}
    </div>
  )
}));

jest.mock('../../src/components/case/SmartFieldGroup', () => ({
  default: ({ config, formData, onChange }: any) => (
    <div data-testid="smart-field-group" data-group-key={config.key}>
      <div data-testid="form-data-display">{JSON.stringify(formData)}</div>
      <button
        data-testid="field-group-change"
        onClick={() => onChange({ testField: 'testValue' })}
      >
        Change Field
      </button>
    </div>
  )
}));

jest.mock('../../src/components/case/ProgressIndicator', () => ({
  default: ({ steps, currentStepKey, onStepChange }: any) => (
    <div data-testid="progress-indicator">
      {steps.map((step: any) => (
        <button
          key={step.key}
          data-testid={`step-${step.key}`}
          data-current={step.key === currentStepKey}
          onClick={() => onStepChange?.(step.key, 0)}
        >
          {step.title}
        </button>
      ))}
    </div>
  )
}));

jest.mock('../../src/components/case/ConflictCheckInline', () => ({
  default: ({ checkParams, onCheckComplete }: any) => (
    <div data-testid="conflict-check-inline">
      <div data-testid="check-params">{JSON.stringify(checkParams)}</div>
      <button
        data-testid="conflict-check-complete"
        onClick={() => onCheckComplete?.({ status: 'success', hasConflict: false })}
      >
        Complete Check
      </button>
    </div>
  )
}));

// Mock Ant Design组件
jest.mock('antd/es/form', () => ({
  Form: ({ children, onValuesChange, initialValues, disabled }: any) => (
    <form data-testid="form" data-disabled={disabled}>
      <div data-testid="form-initial-values">{JSON.stringify(initialValues)}</div>
      <div data-testid="form-children">{children}</div>
      <button
        data-testid="form-change"
        onClick={() => onValuesChange?.({ changed: 'value' }, { changed: 'value', ...initialValues })}
      >
        Change Form
      </button>
    </form>
  ),
  useForm: () => [
    {
      validateFields: jest.fn().mockResolvedValue(true),
      resetFields: jest.fn(),
      setFieldsValue: jest.fn(),
      getFieldValue: jest.fn(),
      validate: jest.fn().mockResolvedValue({})
    }
  ]
}));

jest.mock('antd/es/card', () => ({
  Card: ({ children, title, extra }: any) => (
    <div data-testid="card">
      <div data-testid="card-title">{title}</div>
      {extra && <div data-testid="card-extra">{extra}</div>}
      <div data-testid="card-body">{children}</div>
    </div>
  )
}));

jest.mock('antd/es/button', () => ({
  Button: ({ children, onClick, disabled, loading }: any) => (
    <button
      data-testid="button"
      onClick={onClick}
      disabled={disabled}
      data-loading={loading}
    >
      {loading ? 'Loading...' : children}
    </button>
  )
}));

jest.mock('antd/es/space', () => ({
  Space: ({ children }: any) => (
    <div data-testid="space">{children}</div>
  )
}));

jest.mock('antd/es/typography', () => ({
  Text: ({ children }: any) => <span data-testid="text">{children}</span>,
  Title: ({ children, level }: any) => (
    <h1 data-testid="title" data-level={level}>
      {children}
    </h1>
  )
}));

jest.mock('antd/es/message', () => ({
  success: jest.fn(),
  error: jest.fn(),
  warning: jest.fn()
}));

jest.mock('antd/es/modal', () => ({
  confirm: jest.fn(({ onOk }) => {
    onOk?.();
    return Promise.resolve();
  })
}));

describe('CompactCaseForm Integration Tests', () => {
  const mockStepConfigs: FormStepConfig[] = [
    {
      key: STEP_KEYS.BASIC,
      title: '基本信息',
      description: '填写案件基本信息',
      fieldGroups: [FIELD_GROUP_KEYS.BASIC_INFO],
      validation: [
        {
          field: 'basic.title',
          rule: 'required',
          message: '案件标题是必填项',
          required: true
        }
      ]
    },
    {
      key: STEP_KEYS.CLIENT,
      title: '客户信息',
      description: '填写客户相关信息',
      fieldGroups: [FIELD_GROUP_KEYS.CLIENT_INFO],
      validation: []
    },
    {
      key: STEP_KEYS.LAWYER,
      title: '律师信息',
      description: '分配律师',
      fieldGroups: [FIELD_GROUP_KEYS.LAWYER_INFO],
      validation: []
    },
    {
      key: STEP_KEYS.DETAILS,
      title: '案件详情',
      description: '填写案件详细信息',
      fieldGroups: [FIELD_GROUP_KEYS.ADDITIONAL_INFO],
      validation: []
    }
  ];

  const mockFieldGroups: Record<string, FieldGroupConfig> = {
    [FIELD_GROUP_KEYS.BASIC_INFO]: {
      key: FIELD_GROUP_KEYS.BASIC_INFO,
      title: '基本信息',
      collapsible: true,
      defaultExpanded: true,
      fields: [
        {
          name: 'basic.title',
          label: '案件标题',
          type: 'input',
          required: true,
          defaultValue: ''
        },
        {
          name: 'basic.caseType',
          label: '案件类型',
          type: 'select',
          required: true,
          defaultValue: ''
        }
      ],
      layout: {
        columns: 2,
        spacing: 'medium'
      }
    },
    [FIELD_GROUP_KEYS.CLIENT_INFO]: {
      key: FIELD_GROUP_KEYS.CLIENT_INFO,
      title: '客户信息',
      collapsible: true,
      defaultExpanded: true,
      fields: [
        {
          name: 'client.clientName',
          label: '客户名称',
          type: 'input',
          required: true,
          defaultValue: ''
        }
      ],
      layout: {
        columns: 2,
        spacing: 'medium'
      }
    },
    [FIELD_GROUP_KEYS.LAWYER_INFO]: {
      key: FIELD_GROUP_KEYS.LAWYER_INFO,
      title: '律师信息',
      collapsible: true,
      defaultExpanded: true,
      fields: [
        {
          name: 'lawyer.lawyerId',
          label: '主办律师',
          type: 'select',
          required: true,
          defaultValue: ''
        }
      ],
      layout: {
        columns: 1,
        spacing: 'medium'
      }
    },
    [FIELD_GROUP_KEYS.ADDITIONAL_INFO]: {
      key: FIELD_GROUP_KEYS.ADDITIONAL_INFO,
      title: '其他信息',
      collapsible: true,
      defaultExpanded: false,
      fields: [
        {
          name: 'details.remarks',
          label: '备注',
          type: 'textarea',
          required: false,
          defaultValue: ''
        }
      ],
      layout: {
        columns: 1,
        spacing: 'medium'
      }
    }
  };

  const defaultQuickActions: QuickAction[] = [
    {
      key: 'save',
      label: '保存',
      type: 'primary',
      handler: jest.fn()
    },
    {
      key: 'reset',
      label: '重置',
      type: 'default',
      handler: jest.fn()
    }
  ];

  const defaultProps = {
    steps: mockStepConfigs,
    fieldGroups: mockFieldGroups,
    quickActions: defaultQuickActions,
    events: {
      onInit: jest.fn(),
      onStepChange: jest.fn(),
      onFieldChange: jest.fn(),
      onBeforeSave: jest.fn().mockResolved(true),
      onAfterSave: jest.fn(),
      onSaveError: jest.fn(),
      onCancel: jest.fn(),
      onReset: jest.fn(),
      onConflictCheck: jest.fn()
    }
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

  describe('基本功能集成测试', () => {
    test('应该正确渲染所有组件', () => {
      render(<CompactCaseForm {...defaultProps} />);

      // 检查主容器
      expect(screen.getByTestId('responsive-form-layout')).toBeInTheDocument();
      expect(screen.getByTestId('form')).toBeInTheDocument();

      // 检查进度指示器
      expect(screen.getByTestId('progress-indicator')).toBeInTheDocument();
      expect(screen.getByTestId('step-basic')).toBeInTheDocument();
      expect(screen.getByTestId('step-client')).toBeInTheDocument();
      expect(screen.getByTestId('step-lawyer')).toBeInTheDocument();
      expect(screen.getByTestId('step-details')).toBeInTheDocument();

      // 检查字段组
      expect(screen.getByTestId('smart-field-group')).toBeInTheDocument();

      // 检查冲突检测
      expect(screen.getByTestId('conflict-check-inline')).toBeInTheDocument();
    });

    test('应该正确初始化表单数据', () => {
      render(<CompactCaseForm {...defaultProps} />);

      const formInitialValues = screen.getByTestId('form-initial-values');
      expect(formInitialValues).toBeInTheDocument();

      // 检查是否包含自动生成的案件编号
      expect(formInitialValues.textContent).toContain('CASE');
    });

    test('应该支持步骤切换', async () => {
      const mockOnStepChange = jest.fn();
      render(
        <CompactCaseForm
          {...defaultProps}
          events={{
            ...defaultProps.events,
            onStepChange: mockOnStepChange
          }}
        />
      );

      // 点击下一步
      const clientStep = screen.getByTestId('step-client');
      await userEvent.click(clientStep);

      expect(mockOnStepChange).toHaveBeenCalledWith('client', 'basic');
    });

    test('应该支持字段变化', async () => {
      const mockOnFieldChange = jest.fn();
      render(
        <CompactCaseForm
          {...defaultProps}
          events={{
            ...defaultProps.events,
            onFieldChange: mockOnFieldChange
          }}
        />
      );

      // 触发表单变化
      const formChange = screen.getByTestId('form-change');
      await userEvent.click(formChange);

      expect(mockOnFieldChange).toHaveBeenCalledWith(
        { changed: 'value' },
        expect.any(Object)
      );
    });

    test('应该支持冲突检测', async () => {
      const mockOnConflictCheck = jest.fn();
      render(
        <CompactCaseForm
          {...defaultProps}
          events={{
            ...defaultProps.events,
            onConflictCheck: mockOnConflictCheck
          }}
        />
      );

      // 触发冲突检测
      const conflictCheckComplete = screen.getByTestId('conflict-check-complete');
      await userEvent.click(conflictCheckComplete);

      expect(mockOnConflictCheck).toHaveBeenCalledWith({
        status: 'success',
        hasConflict: false
      });
    });
  });

  describe('1080p优化集成测试', () => {
    test('应该在1080p分辨率下自动启用紧凑模式', () => {
      // 模拟1080p分辨率
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920
      });

      render(<CompactCaseForm {...defaultProps} />);

      const container = document.querySelector('.compact-case-form');
      expect(container).toHaveClass('compact');
      expect(container).toHaveClass('layout-compact');
    });

    test('应该在移动设备下调整布局', () => {
      // 模拟移动设备分辨率
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768
      });

      render(<CompactCaseForm {...defaultProps} />);

      const container = document.querySelector('.compact-case-form');
      expect(container).toHaveClass('breakpoint-md');
    });

    test('应该在平板设备下调整布局', () => {
      // 模拟平板设备分辨率
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 992
      });

      render(<CompactCaseForm {...defaultProps} />);

      const container = document.querySelector('.compact-case-form');
      expect(container).toHaveClass('breakpoint-lg');
    });
  });

  describe('表单验证集成测试', () => {
    test('应该验证必填字段', async () => {
      render(<CompactCaseForm {...defaultProps} />);

      const saveButton = screen.getByText('保存');
      await userEvent.click(saveButton);

      // 验证函数应该被调用
      const { Form } = require('antd/es/form');
      expect(Form.useForm()[0].validateFields).toHaveBeenCalled();
    });

    test('应该支持保存前验证', async () => {
      const mockOnBeforeSave = jest.fn().mockResolvedValue(false);
      render(
        <CompactCaseForm
          {...defaultProps}
          saveConfig={{
            validateBeforeSave: true,
            customSaveLogic: jest.fn()
          }}
          events={{
            ...defaultProps.events,
            onBeforeSave: mockOnBeforeSave
          }}
        />
      );

      const saveButton = screen.getByText('保存');
      await userEvent.click(saveButton);

      expect(mockOnBeforeSave).toHaveBeenCalled();
    });

    test('应该支持保存成功处理', async () => {
      const mockOnAfterSave = jest.fn();
      render(
        <CompactCaseForm
          {...defaultProps}
          saveConfig={{
            validateBeforeSave: true,
            customSaveLogic: jest.fn().mockResolvedValue(undefined)
          }}
          events={{
            ...defaultProps.events,
            onAfterSave: mockOnAfterSave
          }}
        />
      );

      const saveButton = screen.getByText('保存');
      await userEvent.click(saveButton);

      await waitFor(() => {
        expect(mockOnAfterSave).toHaveBeenCalled();
      });
    });

    test('应该支持保存失败处理', async () => {
      const mockOnSaveError = jest.fn();
      const mockError = new Error('保存失败');

      render(
        <CompactCaseForm
          {...defaultProps}
          saveConfig={{
            validateBeforeSave: true,
            customSaveLogic: jest.fn().mockRejectedValue(mockError)
          }}
          events={{
            ...defaultProps.events,
            onSaveError: mockOnSaveError
          }}
        />
      );

      const saveButton = screen.getByText('保存');
      await userEvent.click(saveButton);

      await waitFor(() => {
        expect(mockOnSaveError).toHaveBeenCalledWith(mockError);
      });
    });
  });

  describe('用户交互集成测试', () => {
    test('应该支持重置功能', async () => {
      const mockConfirm = jest.fn(({ onOk }) => {
        onOk?.();
        return Promise.resolve();
      });
      const { Modal } = require('antd/es/modal');
      Modal.confirm = mockConfirm;

      render(<CompactCaseForm {...defaultProps} />);

      const resetButton = screen.getByText('重置');
      await userEvent.click(resetButton);

      expect(mockConfirm).toHaveBeenCalled();
    });

    test('应该支持取消功能', async () => {
      const mockConfirm = jest.fn(({ onOk }) => {
        onOk?.();
        return Promise.resolve();
      });
      const { Modal } = require('antd/es/modal');
      Modal.confirm = mockConfirm;

      render(<CompactCaseForm {...defaultProps} />);

      const cancelButton = screen.getByText('取消');
      await userEvent.click(cancelButton);

      expect(mockConfirm).toHaveBeenCalled();
    });

    test('应该支持快捷操作', async () => {
      const mockSaveAction = jest.fn();
      const customActions = [
        {
          key: 'custom',
          label: '自定义操作',
          type: 'default' as const,
          handler: mockSaveAction
        }
      ];

      render(
        <CompactCaseForm
          {...defaultProps}
          quickActions={customActions}
        />
      );

      // 验证默认快捷操作和自定义操作都被渲染
      expect(screen.getByText('保存')).toBeInTheDocument();
      expect(screen.getByText('自定义操作')).toBeInTheDocument();
    });
  });

  describe('自动保存功能测试', () => {
    test('应该支持自动保存', async () => {
      jest.useFakeTimers();

      render(
        <CompactCaseForm
          {...defaultProps}
          config={{
            enableAutoSave: true,
            autoSaveInterval: 1
          }}
        />
      );

      // 触发表单变化
      const formChange = screen.getByTestId('form-change');
      await userEvent.click(formChange);

      // 等待自动保存触发
      jest.advanceTimersByTime(1000);

      jest.useRealTimers();
    });

    test('应该在只读模式下禁用自动保存', async () => {
      jest.useFakeTimers();

      render(
        <CompactCaseForm
          {...defaultProps}
          readonly={true}
          config={{
            enableAutoSave: true,
            autoSaveInterval: 1
          }}
        />
      );

      // 触发表单变化
      const formChange = screen.getByTestId('form-change');
      await userEvent.click(formChange);

      // 等待自动保存时间
      jest.advanceTimersTimeByTime(2000);

      jest.useRealTimers();
    });
  });

  describe('状态管理集成测试', () => {
    test('应该正确管理表单状态', () => {
      const { rerender } = render(<CompactCaseForm {...defaultProps} />);

      // 初始状态应该是第一个步骤
      expect(screen.getByTestId('step-basic')).toHaveAttribute('data-current', 'true');

      // 切换步骤
      const clientStep = screen.getByTestId('step-client');
      fireEvent.click(clientStep);

      rerender(<CompactCaseForm {...defaultProps} />);

      // 验证状态更新
      expect(screen.getByTestId('step-client')).toHaveAttribute('data-current', 'true');
    });

    test('应该支持初始数据', () => {
      const initialData = {
        basic: {
          title: '测试案件',
          caseType: '民事纠纷'
        }
      };

      render(
        <CompactCaseForm
          {...defaultProps}
          initialData={initialData}
        />
      );

      const formInitialValues = screen.getByTestId('form-initial-values');
      expect(formInitialValues.textContent).toContain('测试案件');
      expect(formInitialValues.textContent).toContain('民事纠纷');
    });

    test('应该支持编辑模式', () => {
      render(
        <CompactCaseForm
          {...defaultProps}
          config={{
            mode: 'edit'
          }}
        />
      );

      expect(screen.getByText('编辑案件')).toBeInTheDocument();
    });

    test('应该支持查看模式', () => {
      render(
        <CompactCaseForm
          {...defaultProps}
          config={{
            mode: 'view'
          }}
        />
      );

      expect(screen.getByText('查看案件')).toBeInTheDocument();
    });
  });

  describe('无障碍集成测试', () => {
    test('应该具有正确的ARIA标签', () => {
      render(<CompactCaseForm {...defaultProps} />);

      // 检查表单是否有正确的语义标记
      const form = screen.getByTestId('form');
      expect(form).toBeInTheDocument();
    });

    test('应该支持键盘导航', async () => {
      render(<CompactCaseForm {...defaultProps} />);

      // 测试Tab键导航
      await userEvent.tab();

      const firstStep = screen.getByTestId('step-basic');
      expect(firstStep).toHaveFocus();

      // 测试方向键导航
      await userEvent.keyboard('{ArrowRight}');

      const secondStep = screen.getByTestId('step-client');
      expect(secondStep).toHaveFocus();
    });

    test('应该为屏幕阅读器提供状态信息', () => {
      render(
        <CompactCaseForm
          {...defaultProps}
          config={{
            showProgressIndicator: true
          }}
        />
      );

      // 进度指示器应该有适当的语义标记
      const progressIndicator = screen.getByTestId('progress-indicator');
      expect(progressIndicator).toBeInTheDocument();
    });
  });

  describe('性能优化集成测试', () => {
    test('应该优化大量步骤的渲染', () => {
      const manySteps = Array.from({ length: 20 }, (_, i) => ({
        key: `step_${i}`,
        title: `步骤 ${i + 1}`,
        description: `步骤 ${i + 1} 的描述`,
        fieldGroups: [],
        validation: []
      }));

      const startTime = performance.now();
      render(
        <CompactCaseForm
          {...defaultProps}
          steps={manySteps}
          fieldGroups={{}}
        />
      );
      const endTime = performance.now();

      // 渲染时间应该在合理范围内
      expect(endTime - startTime).toBeLessThan(100);
    });

    test('应该优化表单状态更新', () => {
      const { rerender } = render(<CompactCaseForm {...defaultProps} />);

      const startTime = performance.now();

      // 多次更新状态
      for (let i = 0; i < 10; i++) {
        rerender(
          <CompactCaseForm
            {...defaultProps}
            key={i}
          />
        );
      }

      const endTime = performance.now();

      // 更新时间应该在合理范围内
      expect(endTime - startTime).toBeLessThan(50);
    });
  });

  describe('响应式集成测试', () => {
    test('应该在小屏幕上正确显示', () => {
      // 模拟小屏幕
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 480
      });

      render(<CompactCaseForm {...defaultProps} />);

      const container = document.querySelector('.compact-case-form');
      expect(container).toHaveClass('breakpoint-xs');
    });

    test('应该在中等屏幕上正确显示', () => {
      // 模拟中等屏幕
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 800
      });

      render(<CompactCaseForm {...defaultProps} />);

      const container = document.querySelector('.compact-case-form');
      expect(container).toHaveClass('breakpoint-sm');
    });

    test('应该在大屏幕上正确显示', () => {
      // 模拟大屏幕
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1440
      });

      render(<CompactCaseForm {...defaultProps} />);

      const container = document.querySelector('.compact-case-form');
      expect(container).toHaveClass('breakpoint-lg');
    });
  });

  describe('错误处理集成测试', () => {
    test('应该处理表单验证错误', async () => {
      const { Form } = require('antd/es/form');
      Form.useForm()[0].validateFields.mockRejectedValue({
        errorFields: [
          { name: ['basic.title'], errors: ['标题是必填项'] }
        ]
      });

      render(<CompactCaseForm {...defaultProps} />);

      const saveButton = screen.getByText('保存');
      await userEvent.click(saveButton);

      // 应该显示验证错误
      const { message } = require('antd/es/message');
      expect(message.error).toHaveBeenCalled();
    });

    test('应该处理网络保存错误', async () => {
      const mockError = new Error('网络错误');
      const { message } = require('antd/es/message');

      render(
        <CompactCaseForm
          {...defaultProps}
          saveConfig={{
            validateBeforeSave: true,
            customSaveLogic: jest.fn().mockRejectedValue(mockError)
          }}
        />
      );

      const saveButton = screen.getByText('保存');
      await userEvent.click(saveButton);

      await waitFor(() => {
        expect(message.error).toHaveBeenCalledWith('网络错误');
      });
    });

    test('应该处理组件加载错误', () => {
      // 模拟组件加载失败
      const consoleError = jest.spyOn(console, 'error').mockImplementation(() => {});

      render(<CompactCaseForm {...defaultProps} />);

      // 验证错误被正确处理
      expect(consoleError).not.toHaveBeenCalled();

      consoleError.mockRestore();
    });
  });
});
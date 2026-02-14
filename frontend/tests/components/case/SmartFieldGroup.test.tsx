/**
 * SmartFieldGroup组件测试
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import SmartFieldGroup from '../../../src/components/case/SmartFieldGroup';
import { Input, Select } from 'antd';
import type { FieldGroupConfig } from '../../../src/components/case/types/SmartFieldGroup.types';

// Mock数据
const mockGroups: FieldGroupConfig[] = [
  {
    key: 'basic-info',
    title: '基本信息',
    description: '案件的基本信息，包括案件标题、类型等',
    icon: '📋',
    priority: 1,
    fields: [
      {
        name: 'caseTitle',
        label: '案件标题',
        component: <Input placeholder="请输入案件标题" />,
        required: true,
        priority: 'high'
      },
      {
        name: 'caseType',
        label: '案件类型',
        component: (
          <Select placeholder="请选择案件类型">
            <Select.Option value="civil">民事案件</Select.Option>
            <Select.Option value="criminal">刑事案件</Select.Option>
            <Select.Option value="administrative">行政案件</Select.Option>
          </Select>
        ),
        required: true,
        priority: 'high'
      },
      {
        name: 'clientName',
        label: '客户姓名',
        component: <Input placeholder="请输入客户姓名" />,
        priority: 'medium'
      }
    ],
    defaultExpanded: true
  },
  {
    key: 'details',
    title: '详细信息',
    description: '案件的详细信息，包括金额、律师等',
    icon: '📝',
    priority: 2,
    fields: [
      {
        name: 'amount',
        label: '涉案金额',
        component: <Input placeholder="请输入涉案金额" />,
        priority: 'medium',
        condition: [
          { field: 'caseType', operator: 'equals', value: 'civil' }
        ]
      },
      {
        name: 'lawyer',
        label: '负责律师',
        component: (
          <Select placeholder="请选择负责律师">
            <Select.Option value="lawyer1">张律师</Select.Option>
            <Select.Option value="lawyer2">李律师</Select.Option>
          </Select>
        ),
        priority: 'low'
      }
    ],
    defaultExpanded: false
  }
];

describe('SmartFieldGroup', () => {
  // 基础渲染测试
  test('应该正确渲染分组', () => {
    render(
      <SmartFieldGroup
        groups={mockGroups}
        formData={{}}
        defaultActiveGroups={['basic-info']}
      />
    );

    // 检查分组标题
    expect(screen.getByText('基本信息')).toBeInTheDocument();
    expect(screen.getByText('详细信息')).toBeInTheDocument();

    // 检查字段
    expect(screen.getByText('案件标题')).toBeInTheDocument();
    expect(screen.getByText('案件类型')).toBeInTheDocument();
    expect(screen.getByText('客户姓名')).toBeInTheDocument();
  });

  // 折叠功能测试
  test('应该支持分组折叠和展开', async () => {
    // formData must have caseType='civil' for the conditional field '涉案金额' to be visible
    render(
      <SmartFieldGroup
        groups={mockGroups}
        formData={{ caseType: 'civil' }}
        defaultActiveGroups={['basic-info']}
      />
    );

    // 基本信息应该展开
    expect(screen.getByText('案件标题')).toBeInTheDocument();

    // 详细信息应该折叠（字段不可见）
    expect(screen.queryByText('涉案金额')).not.toBeInTheDocument();

    // 点击详细信息标题展开
    const detailsHeader = screen.getByText('详细信息');
    fireEvent.click(detailsHeader);

    await waitFor(() => {
      expect(screen.getByText('涉案金额')).toBeInTheDocument();
    });
  });

  // 条件显示测试
  test('应该根据条件显示字段', () => {
    const formData1 = { caseType: 'civil' };
    const { rerender } = render(
      <SmartFieldGroup
        groups={mockGroups}
        formData={formData1}
        defaultActiveGroups={['basic-info', 'details']}
      />
    );

    // 民事案件应该显示涉案金额
    expect(screen.getByText('涉案金额')).toBeInTheDocument();

    // 更新为刑事案件
    const formData2 = { caseType: 'criminal' };
    rerender(
      <SmartFieldGroup
        groups={mockGroups}
        formData={formData2}
        defaultActiveGroups={['basic-info', 'details']}
      />
    );

    // 刑事案件不应该显示涉案金额
    expect(screen.queryByText('涉案金额')).not.toBeInTheDocument();
  });

  // 1080p紧凑模式测试
  test('应该在紧凑模式下正确显示', () => {
    render(
      <SmartFieldGroup
        groups={mockGroups}
        formData={{}}
        isCompact={true}
      />
    );

    // 检查紧凑模式样式类
    const container = screen.getByTestId('smart-field-group');
    expect(container).toHaveClass('smart-field-group-compact');

    // 检查优化提示
    expect(screen.getByText(/1080p显示器优化/)).toBeInTheDocument();
  });

  // 自动排序测试
  test('应该按优先级自动排序字段', () => {
    const customGroups: FieldGroupConfig[] = [
      {
        key: 'test',
        title: '测试分组',
        fields: [
          {
            name: 'low-priority',
            label: '低优先级字段',
            component: <Input />,
            priority: 'low'
          },
          {
            name: 'high-priority',
            label: '高优先级字段',
            component: <Input />,
            priority: 'high'
          },
          {
            name: 'medium-priority',
            label: '中优先级字段',
            component: <Input />,
            priority: 'medium'
          }
        ]
      }
    ];

    render(
      <SmartFieldGroup
        groups={customGroups}
        formData={{}}
        enableAutoSort={true}
        defaultActiveGroups={['test']}
      />
    );

    // 获取所有字段标签
    const fieldLabels = screen.getAllByText(/优先级字段/);

    // 检查排序：高优先级 -> 中优先级 -> 低优先级
    expect(fieldLabels[0]).toHaveTextContent('高优先级字段');
    expect(fieldLabels[1]).toHaveTextContent('中优先级字段');
    expect(fieldLabels[2]).toHaveTextContent('低优先级字段');
  });

  // 分组变化回调测试
  test('应该正确触发分组变化回调', () => {
    const mockOnGroupChange = jest.fn();

    render(
      <SmartFieldGroup
        groups={mockGroups}
        formData={{}}
        onGroupChange={mockOnGroupChange}
      />
    );

    // 点击详细信息标题
    const detailsHeader = screen.getByText('详细信息');
    fireEvent.click(detailsHeader);

    expect(mockOnGroupChange).toHaveBeenCalledWith(['details']);
  });

  // 空状态测试
  test('应该在没有可见字段时显示空状态', () => {
    const emptyGroups: FieldGroupConfig[] = [
      {
        key: 'empty',
        title: '空分组',
        fields: [
          {
            name: 'hidden',
            label: '隐藏字段',
            component: <Input />,
            condition: [
              { field: 'nonexistent', operator: 'exists' }
            ]
          }
        ]
      }
    ];

    render(
      <SmartFieldGroup
        groups={emptyGroups}
        formData={{}}
      />
    );

    expect(screen.getByText('没有可显示的字段')).toBeInTheDocument();
  });

  // 可见字段计数测试
  test('应该正确显示可见字段数量', () => {
    render(
      <SmartFieldGroup
        groups={mockGroups}
        formData={{}}
        defaultActiveGroups={['basic-info']}
      />
    );

    // 基本信息分组应该显示字段数量徽章
    const badge = screen.getByText('3'); // 3个字段可见
    expect(badge).toBeInTheDocument();
  });

  // 1080p自动折叠测试
  test('应该在1080p模式下自动折叠指定分组', () => {
    const autoCollapseGroups: FieldGroupConfig[] = [
      {
        key: 'auto-collapse',
        title: '自动折叠分组',
        fields: [
          {
            name: 'field1',
            label: '字段1',
            component: <Input />,
            priority: 'medium'
          }
        ],
        autoCollapseInCompact: true,
        defaultExpanded: true
      }
    ];

    render(
      <SmartFieldGroup
        groups={autoCollapseGroups}
        formData={{}}
        isCompact={true}
        defaultActiveGroups={['auto-collapse']}
      />
    );

    // 在紧凑模式下，应该显示优化提示
    expect(screen.getByText(/已优化/)).toBeInTheDocument();
  });
});

// 辅助函数：为组件添加data-testid (use unique testid to avoid conflict with component's own testid)
const SmartFieldGroupWithTestId = (props: any) => (
  <div data-testid="smart-field-group-wrapper">
    <SmartFieldGroup {...props} />
  </div>
);

// 更新测试以使用带data-testid的组件
describe('SmartFieldGroup with data-testid', () => {
  test('应该正确设置data-testid', () => {
    render(
      <SmartFieldGroupWithTestId
        groups={mockGroups}
        formData={{}}
      />
    );

    // Check for the wrapper's unique testid
    const wrapper = screen.getByTestId('smart-field-group-wrapper');
    expect(wrapper).toBeInTheDocument();
    // Check the actual component's testid is inside
    const container = screen.getByTestId('smart-field-group');
    expect(container).toBeInTheDocument();
  });
});
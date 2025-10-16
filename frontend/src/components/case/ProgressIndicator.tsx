import React, { useState, useEffect, useMemo, useCallback } from 'react';
import {
  Steps,
  Progress,
  Card,
  Typography,
  Space,
  Alert,
  Tooltip,
  Modal,
  message
} from 'antd';
import {
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  LoadingOutlined,
  ClockCircleOutlined,
  InfoCircleOutlined,
  WarningOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  FastForwardOutlined
} from '@ant-design/icons';
import type { StepsProps } from 'antd/es/steps';
import type {
  ProgressIndicatorProps,
  StepConfig,
  ProgressStats,
  StepStatus,
  StepNavigationEvent,
  ValidationResult,
  ProgressCalculationConfig,
  NavigationConfig
} from './types/ProgressIndicator.types';
import './ProgressIndicator.less';

const { Text, Title, Paragraph } = Typography;

/**
 * ProgressIndicator - 进度指示器组件
 * 显示表单完成进度和当前步骤，支持步骤点击导航
 * 专为案件创建表单设计，提供清晰的用户进度反馈
 */
const ProgressIndicator: React.FC<ProgressIndicatorProps> = ({
  steps,
  currentStepKey,
  onStepChange,
  showProgressBar = true,
  showStats = true,
  enableStepNavigation = true,
  progressType = 'line',
  direction = 'horizontal',
  size = 'default',
  isCompact = false,
  className = '',
  style,
  stepsProps = {}
}) => {
  const [currentStepIndex, setCurrentStepIndex] = useState<number>(0);
  const [navigationConfig, setNavigationConfig] = useState<NavigationConfig>({
    allowSkipIncomplete: true,
    validateOnClick: false,
    confirmDialog: {
      title: '切换步骤确认',
      content: '您确定要切换到这个步骤吗？当前步骤的未保存数据可能会丢失。',
      okText: '确定',
      cancelText: '取消'
    }
  });
  const [validating, setValidating] = useState(false);

  // 计算当前步骤索引
  useEffect(() => {
    if (currentStepKey) {
      const index = steps.findIndex(step => step.key === currentStepKey);
      if (index !== -1) {
        setCurrentStepIndex(index);
      }
    }
  }, [currentStepKey, steps]);

  // 计算进度统计
  const calculateProgressStats = useCallback((
    stepConfigs: StepConfig[],
    currentIdx: number,
    config: ProgressCalculationConfig = {}
  ): ProgressStats => {
    const {
      includeDisabled = true,
      useWeighting = false,
      weights = {},
      minProgress = 0
    } = config;

    let totalWeight = 0;
    let completedWeight = 0;
    let errorCount = 0;
    let warningCount = 0;

    stepConfigs.forEach((step, index) => {
      const stepWeight = useWeighting ? (weights[step.key] || step.weight || 1) : 1;
      const isDisabled = step.disabled && !includeDisabled;

      if (!isDisabled) {
        totalWeight += stepWeight;

        if (step.completed) {
          completedWeight += stepWeight;
        }

        if (step.status === 'error') {
          errorCount++;
        } else if (step.status === 'wait' && index < currentIdx) {
          warningCount++;
        }
      }
    });

    const percentage = totalWeight > 0
      ? Math.max(minProgress, Math.round((completedWeight / totalWeight) * 100))
      : 0;

    return {
      totalSteps: stepConfigs.length,
      completedSteps: stepConfigs.filter(s => s.completed).length,
      currentStep: currentIdx,
      percentage,
      errorSteps: errorCount,
      warningSteps: warningCount
    };
  }, []);

  // 计算进度统计
  const stats = useMemo(() => {
    return calculateProgressStats(steps, currentStepIndex);
  }, [steps, currentStepIndex, calculateProgressStats]);

  // 生成Ant Design Steps items
  const stepsItems = useMemo(() => {
    return steps.map((step, index) => {
      let status: StepStatus = 'wait';

      if (step.status) {
        status = step.status;
      } else if (step.completed) {
        status = 'finish';
      } else if (index < currentStepIndex) {
        status = 'finish';
      } else if (index === currentStepIndex) {
        status = 'process';
      }

      // 根据验证状态调整
      if (step.validation && !step.completed && index <= currentStepIndex) {
        if (step.validation.required && !step.completed) {
          status = 'error';
        } else if (!step.completed) {
          status = 'wait';
        }
      }

      // 获取图标
      let icon = step.icon;
      if (!icon) {
        switch (status) {
          case 'finish':
            icon = <CheckCircleOutlined />;
            break;
          case 'process':
            icon = <LoadingOutlined />;
            break;
          case 'error':
            icon = <ExclamationCircleOutlined />;
            break;
          case 'wait':
            icon = <ClockCircleOutlined />;
            break;
          default:
            icon = <InfoCircleOutlined />;
        }
      }

      return {
        key: step.key,
        title: (
          <div className="progress-step-title">
            <Space size="small" align="center">
              {icon}
              <Text strong={status === 'process' || status === 'finish'}>{step.title}</Text>
              {status === 'error' && (
                <Tooltip title="此步骤有错误需要解决">
                  <WarningOutlined style={{ color: '#ff4d4f' }} />
                </Tooltip>
              )}
            </Space>
          </div>
        ),
        description: step.description && (
          <div className="progress-step-description">
            <Text type="secondary">{step.description}</Text>
          </div>
        ),
        status,
        disabled: step.disabled,
        onClick: enableStepNavigation && !step.disabled ? () => handleStepClick(step, index) : undefined
      };
    });
  }, [steps, currentStepIndex, enableStepNavigation]);

  // 处理步骤点击
  const handleStepClick = useCallback(async (step: StepConfig, index: number) => {
    if (!enableStepNavigation || step.disabled) return;

    // 验证逻辑
    if (navigationConfig.validateOnClick && index > currentStepIndex) {
      setValidating(true);

      try {
        // 这里可以添加验证逻辑
        const isValid = step.validation?.validator ? step.validation.validator() : true;

        if (!isValid) {
          message.error('当前步骤未完成，无法继续');
          return;
        }
      } finally {
        setValidating(false);
      }
    }

    // 确认对话框逻辑
    if (navigationConfig.allowSkipIncomplete && index > currentStepIndex) {
      Modal.confirm({
        ...navigationConfig.confirmDialog,
        content: (
          <div>
            <p>{navigationConfig.confirmDialog?.content}</p>
            <p>
              当前进度: {stats.percentage}% ({stats.completedSteps}/{stats.totalSteps} 步骤完成)
            </p>
          </div>
        ),
        onOk: () => {
          if (step.onClick) {
            step.onClick();
          } else if (onStepChange) {
            onStepChange(step.key, index);
          }
          setCurrentStepIndex(index);
        },
        okText: navigationConfig.confirmDialog?.okText,
        cancelText: navigationConfig.confirmDialog?.cancelText
      });
    } else {
      if (step.onClick) {
        step.onClick();
      } else if (onStepChange) {
        onStepChange(step.key, index);
      }
      setCurrentStepIndex(index);
    }
  }, [enableStepNavigation, currentStepIndex, navigationConfig, stats, onStepChange, validating]);

  // 键盘导航支持
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!enableStepNavigation) return;

      if (e.key === 'ArrowLeft' && currentStepIndex > 0) {
        e.preventDefault();
        const prevStep = steps[currentStepIndex - 1];
        if (prevStep && !prevStep.disabled) {
          handleStepClick(prevStep, currentStepIndex - 1);
        }
      } else if (e.key === 'ArrowRight' && currentStepIndex < steps.length - 1) {
        e.preventDefault();
        const nextStep = steps[currentStepIndex + 1];
        if (nextStep && !nextStep.disabled) {
          handleStepClick(nextStep, currentStepIndex + 1);
        }
      } else if (e.key === 'Enter' && steps[currentStepIndex]) {
        e.preventDefault();
        const currentStep = steps[currentStepIndex];
        if (currentStep && !currentStep.completed) {
          // 标记当前步骤为完成
          currentStep.completed = true;
          setCurrentStepIndex(currentStepIndex); // 触发重新渲染
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [enableStepNavigation, currentStepIndex, steps, handleStepClick]);

  // 格式化统计信息
  const formatStats = useCallback((stats: ProgressStats) => {
    return {
      percentage: `${stats.percentage}%`,
      completed: `${stats.completedSteps}`,
      total: `${stats.totalSteps}`,
      errors: stats.errorSteps > 0 ? `${stats.errorSteps} 个错误` : '',
      warnings: stats.warningSteps > 0 ? `${stats.warningSteps} 个警告` : ''
    };
  }, []);

  return (
    <div
      className={`progress-indicator ${className} ${isCompact ? 'progress-indicator-compact' : ''} progress-${direction}`}
      style={style}
    >
      {/* 步骤导航 */}
      <Card
        className="progress-steps-card"
        bordered={false}
        size={isCompact ? 'small' : 'default'}
      >
        <Steps
          current={currentStepIndex}
          items={stepsItems}
          size={size}
          direction={direction}
          {...stepsProps}
          onChange={(current) => {
            if (enableStepNavigation && typeof current === 'number') {
              const step = steps[current];
              if (step && !step.disabled) {
                handleStepClick(step, current);
              }
            }
          }}
        />
      </Card>

      {/* 进度条 */}
      {showProgressBar && (
        <Card
          className="progress-bar-card"
          bordered={false}
          size={isCompact ? 'small' : 'default'}
        >
          <Progress
            type={progressType}
            percent={stats.percentage}
            status={stats.errorSteps > 0 ? 'exception' : stats.percentage === 100 ? 'success' : 'active'}
            strokeColor={stats.errorSteps > 0 ? '#ff4d4f' : '#1890ff'}
            trailColor="#f0f0f0"
            strokeWidth={isCompact ? 4 : 6}
            size={size === 'small' ? 'default' : size}
            format={(percent) => (
              <Text strong>{percent}%</Text>
            )}
          />
        </Card>
      )}

      {/* 统计信息 */}
      {showStats && (
        <Card
          className="progress-stats-card"
          bordered={false}
          size={isCompact ? 'small' : 'default'}
        >
          <div className="progress-stats-content">
            <Space direction="vertical" size="small" className="stats-main">
              <Title level={isCompact ? 5 : 4} className="stats-title">
                表单进度
              </Title>
              <div className="stats-overview">
                <Space wrap>
                  <div className="stat-item">
                    <Text type="secondary">完成进度</Text>
                    <Text strong className="stat-value">
                      {formatStats(stats).percentage}
                    </Text>
                  </div>
                  <div className="stat-item">
                    <Text type="secondary">已步骤</Text>
                    <Text strong className="stat-value">
                      {formatStats(stats).completed} / {formatStats(stats).total}
                    </Text>
                  </div>
                </Space>
              </div>

              {/* 错误和警告信息 */}
              {(stats.errorSteps > 0 || stats.warningSteps > 0) && (
                <div className="stats-alerts">
                  {stats.errorSteps > 0 && (
                    <Alert
                      message={`发现 ${formatStats(stats).errors}`}
                      type="error"
                      showIcon
                      closable
                      className="stats-alert"
                    />
                  )}
                  {stats.warningSteps > 0 && (
                    <Alert
                      message={`发现 ${formatStats(stats).warnings}`}
                      type="warning"
                      showIcon
                      closable
                      className="stats-alert"
                    />
                  )}
                </div>
              )}

              {/* 快捷操作 */}
              <div className="stats-actions">
                <Space wrap>
                  {stats.completedSteps < stats.totalSteps && (
                    <Tooltip title="跳转到下一个未完成步骤">
                      <FastForwardOutlined
                        className="action-icon"
                        onClick={() => {
                          const nextIncompleteIndex = steps.findIndex(
                            (step, index) => index > currentStepIndex && !step.completed && !step.disabled
                          );
                          if (nextIncompleteIndex !== -1) {
                            handleStepClick(steps[nextIncompleteIndex], nextIncompleteIndex);
                          }
                        }}
                      />
                    </Tooltip>
                  )}
                  <Tooltip title="刷新进度">
                    <PlayCircleOutlined
                      className="action-icon"
                      onClick={() => {
                        // 重新计算进度
                        setCurrentStepIndex(currentStepIndex);
                      }}
                    />
                  </Tooltip>
                </Space>
              </div>
            </Space>
          </div>
        </Card>
      )}

      {/* 1080p优化提示 */}
      {isCompact && (
        <div className="progress-optimization-hint">
          <Alert
            message="表单进度指示器已为1080p显示器优化"
            description="使用键盘方向键快速导航，按Enter键标记当前步骤为完成"
            type="info"
            showIcon
            closable
          />
        </div>
      )}
    </div>
  );
};

export default ProgressIndicator;
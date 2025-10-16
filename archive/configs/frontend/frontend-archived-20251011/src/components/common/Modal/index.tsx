import React, { useEffect } from "react";
import { Modal as BootstrapModal, Button, Spinner } from "react-bootstrap";
import { useDispatch, useSelector } from 'react-redux';

// 直接定义类型避免导入问题
interface RootState {
  ui: {
    modals: Array<{
      id: string;
      type: "confirm" | "info" | "warning" | "error" | "custom";
      title: string;
      message?: string;
      content?: React.ReactNode;
      onConfirm?: () => void;
      onCancel?: () => void;
      confirmText?: string;
      cancelText?: string;
      size?: "sm" | "lg" | "xl";
      fullscreen?: true | "sm-down" | "md-down" | "lg-down" | "xl-down" | "xxl-down";
      icon?: string;
      loading?: boolean;
      backdrop?: "static" | boolean;
      centered?: boolean;
      scrollable?: boolean;
      footer?: React.ReactNode;
      className?: string;
      showCloseButton?: boolean;
    }>;
  };
}

type AppDispatch = any;

// 使用动态导入避免模块解析问题
const selectModals = (state: RootState) => state.ui.modals;
const hideModal = (modalId: string) => ({
  type: 'ui/hideModal',
  payload: modalId
});

export interface ModalConfig {
  id: string;
  type: "confirm" | "info" | "warning" | "error" | "custom";
  title: string;
  message?: string;
  content?: React.ReactNode;
  onConfirm?: () => void;
  onCancel?: () => void;
  confirmText?: string;
  cancelText?: string;
  size?: "sm" | "lg" | "xl";
  fullscreen?:
    | boolean
    | "sm-down"
    | "md-down"
    | "lg-down"
    | "xl-down"
    | "xxl-down";
  icon?: string;
  loading?: boolean;
  backdrop?: "static" | boolean;
  centered?: boolean;
  scrollable?: boolean;
  footer?: React.ReactNode;
  className?: string;
  showCloseButton?: boolean;
}

interface ModalProps extends ModalConfig {
  show?: boolean;
}

const ModalComponent: React.FC<ModalProps> = ({
  show = false,
  id,
  type,
  title,
  message,
  content,
  onConfirm,
  onCancel,
  confirmText,
  cancelText,
  size = "lg",
  icon,
  loading = false,
  backdrop = true,
  centered = true,
  scrollable = false,
  fullscreen = false,
  footer,
  className = "",
  showCloseButton = true,
}) => {
  const dispatch = useDispatch<AppDispatch>();
  const modals = useSelector(selectModals) as any;
  const modalConfig = modals.find((m: ModalConfig) => m.id === id);

  const modalShow = show || !!modalConfig;
  const currentConfig = modalConfig || {
    id,
    type,
    title,
    message,
    content,
    onConfirm,
    onCancel,
    confirmText,
    cancelText,
    size,
    icon,
    loading,
    backdrop,
    centered,
    scrollable,
    fullscreen,
    footer,
    className,
    showCloseButton,
  };

  const handleHide = () => {
    if (currentConfig.onCancel) {
      currentConfig.onCancel();
    }
    dispatch(hideModal(id));
  };

  const handleConfirm = () => {
    if (currentConfig.onConfirm) {
      currentConfig.onConfirm();
    }
  };

  // 获取默认的确认文本
  const getDefaultConfirmText = () => {
    switch (currentConfig.type) {
      case "confirm":
        return "确认";
      case "warning":
        return "继续";
      case "error":
        return "确定";
      case "info":
        return "知道了";
      default:
        return "确认";
    }
  };

  // 获取默认的取消文本
  const getDefaultCancelText = () => {
    return currentConfig.type === "info" ? "关闭" : "取消";
  };

  // 获取默认图标
  const getDefaultIcon = () => {
    switch (currentConfig.type) {
      case "confirm":
        return "fa-question-circle text-primary";
      case "warning":
        return "fa-exclamation-triangle text-warning";
      case "error":
        return "fa-times-circle text-danger";
      case "info":
        return "fa-info-circle text-info";
      default:
        return currentConfig.icon;
    }
  };

  const finalIcon = currentConfig.icon || getDefaultIcon();
  const finalConfirmText = currentConfig.confirmText || getDefaultConfirmText();
  const finalCancelText = currentConfig.cancelText || getDefaultCancelText();

  // 渲染内容
  const renderContent = () => {
    if (currentConfig.content) {
      return currentConfig.content;
    }

    if (currentConfig.message) {
      return (
        <div className="d-flex align-items-center">
          {finalIcon && (
            <div className="me-3">
              <i className={`fas ${finalIcon} fa-2x`}></i>
            </div>
          )}
          <div className="flex-grow-1">
            <p className="mb-0">{currentConfig.message}</p>
          </div>
        </div>
      );
    }

    return null;
  };

  // 渲染Footer
  const renderFooter = () => {
    if (currentConfig.footer) {
      return currentConfig.footer;
    }

    // 确认类型的Modal显示确认和取消按钮
    if (currentConfig.type === "confirm" || currentConfig.onConfirm) {
      return (
        <>
          <Button
            variant="outline-secondary"
            onClick={handleHide}
            disabled={currentConfig.loading}
          >
            <i className="fas fa-times me-2"></i>
            {finalCancelText}
          </Button>
          <Button variant="primary" onClick={handleConfirm} disabled={currentConfig.loading}>
            {loading ? (
              <>
                <Spinner
                  as="span"
                  animation="border"
                  size="sm"
                  role="status"
                  aria-hidden="true"
                  className="me-2"
                />
                处理中...
              </>
            ) : (
              <>
                <i className="fas fa-check me-2"></i>
                {finalConfirmText}
              </>
            )}
          </Button>
        </>
      );
    }

    // 其他类型只显示确定按钮
    return (
      <Button variant="outline-secondary" onClick={handleHide}>
        <i className="fas fa-times me-2"></i>
        {finalCancelText}
      </Button>
    );
  };

  return (
    <BootstrapModal
      show={modalShow}
      onHide={handleHide}
      size={size}
      backdrop={backdrop}
      centered={centered}
      scrollable={scrollable}
      fullscreen={fullscreen || undefined}
      className={className}
      keyboard={!loading}
    >
      <BootstrapModal.Header closeButton={showCloseButton && !loading}>
        <BootstrapModal.Title className="d-flex align-items-center">
          {finalIcon && <i className={`fas ${finalIcon} me-2`}></i>}
          {currentConfig.title}
        </BootstrapModal.Title>
      </BootstrapModal.Header>
      <BootstrapModal.Body>{renderContent()}</BootstrapModal.Body>
      <BootstrapModal.Footer>{renderFooter()}</BootstrapModal.Footer>
    </BootstrapModal>
  );
};

// Modal容器组件 - 用于渲染所有激活的Modal
export const ModalContainer: React.FC = () => {
  const modals = useSelector(selectModals);

  return (
    <>
      {modals.map((modal: ModalConfig) => (
        <ModalComponent key={modal.id} {...modal} />
      ))}
    </>
  );
};

// 便捷的Modal显示Hook
export const useModal = () => {
  const dispatch = useDispatch<AppDispatch>();

  const showModal = (config: Omit<ModalConfig, "id">) => {
    const id = `modal_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    dispatch({
      type: "ui/showModal",
      payload: { ...config, id },
    });
    return id;
  };

  const showConfirmModal = (config: Omit<ModalConfig, "id" | "type">) => {
    return showModal({ ...config, type: "confirm" });
  };

  const showInfoModal = (config: Omit<ModalConfig, "id" | "type">) => {
    return showModal({ ...config, type: "info" });
  };

  const showWarningModal = (config: Omit<ModalConfig, "id" | "type">) => {
    return showModal({ ...config, type: "warning" });
  };

  const showErrorModal = (config: Omit<ModalConfig, "id" | "type">) => {
    return showModal({ ...config, type: "error" });
  };

  const hideModal = (id: string) => {
    dispatch({
      type: "ui/hideModal",
      payload: id,
    });
  };

  const hideAllModals = () => {
    dispatch({ type: "ui/hideAllModals" });
  };

  return {
    showModal,
    showConfirmModal,
    showInfoModal,
    showWarningModal,
    showErrorModal,
    hideModal,
    hideAllModals,
  };
};

// 预设的Modal配置
export const ModalPresets = {
  deleteConfirm: (itemName: string, onConfirm: () => void) => ({
    type: "confirm" as const,
    title: "确认删除",
    message: `确定要删除 "${itemName}" 吗？此操作不可恢复。`,
    confirmText: "删除",
    cancelText: "取消",
    icon: "fa-trash-alt text-danger",
    onConfirm,
  }),

  saveConfirm: (onConfirm: () => void) => ({
    type: "confirm" as const,
    title: "确认保存",
    message: "确定要保存更改吗？",
    confirmText: "保存",
    cancelText: "取消",
    icon: "fa-save text-primary",
    onConfirm,
  }),

  unsavedChanges: (onConfirm: () => void, onCancel?: () => void) => ({
    type: "warning" as const,
    title: "未保存的更改",
    message: "您有未保存的更改，确定要离开吗？",
    confirmText: "离开",
    cancelText: "取消",
    icon: "fa-exclamation-triangle text-warning",
    onConfirm,
    onCancel,
  }),

  success: (title: string, message: string) => ({
    type: "info" as const,
    title,
    message,
    confirmText: "确定",
    icon: "fa-check-circle text-success",
  }),

  error: (title: string, message: string) => ({
    type: "error" as const,
    title,
    message,
    confirmText: "确定",
    icon: "fa-times-circle text-danger",
  }),

  networkError: () => ({
    type: "error" as const,
    title: "网络错误",
    message: "网络连接失败，请检查网络设置后重试。",
    confirmText: "确定",
    icon: "fa-wifi text-danger",
  }),
};

export default ModalComponent;

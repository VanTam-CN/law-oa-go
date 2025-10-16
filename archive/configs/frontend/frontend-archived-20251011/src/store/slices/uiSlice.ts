import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import React from "react";

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

export interface NotificationConfig {
  id: string;
  type: "success" | "info" | "warning" | "error";
  title: string;
  message: string;
  duration?: number;
}

interface UIState {
  loading: boolean;
  modals: ModalConfig[];
  notifications: NotificationConfig[];
  theme: "light" | "dark" | "auto";
  sidebarCollapsed: boolean;
  searchHistory: string[];
}

const initialState: UIState = {
  loading: false,
  modals: [],
  notifications: [],
  theme: "light",
  sidebarCollapsed: false,
  searchHistory: [],
};

const uiSlice = createSlice({
  name: "ui",
  initialState,
  reducers: {
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    showModal: (state, action: PayloadAction<ModalConfig>) => {
      state.modals.push(action.payload);
    },
    hideModal: (state, action: PayloadAction<string>) => {
      state.modals = state.modals.filter(
        (modal) => modal.id !== action.payload,
      );
    },
    clearModals: (state) => {
      state.modals = [];
    },
    showNotification: (state, action: PayloadAction<NotificationConfig>) => {
      state.notifications.push(action.payload);
    },
    hideNotification: (state, action: PayloadAction<string>) => {
      state.notifications = state.notifications.filter(
        (notification) => notification.id !== action.payload,
      );
    },
    clearNotifications: (state) => {
      state.notifications = [];
    },
    setTheme: (state, action: PayloadAction<"light" | "dark" | "auto">) => {
      state.theme = action.payload;
    },
    toggleSidebar: (state) => {
      state.sidebarCollapsed = !state.sidebarCollapsed;
    },
    setSidebarCollapsed: (state, action: PayloadAction<boolean>) => {
      state.sidebarCollapsed = action.payload;
    },
    addToSearchHistory: (state, action: PayloadAction<string>) => {
      if (!action.payload.trim()) return;

      // 移除重复项
      const filtered = state.searchHistory.filter(item => item !== action.payload);

      // 添加到开头
      state.searchHistory = [action.payload, ...filtered].slice(0, 10); // 保留最多10个
    },
    clearSearchHistory: (state) => {
      state.searchHistory = [];
    },
  },
});

export const {
  setLoading,
  showModal,
  hideModal,
  clearModals,
  showNotification,
  hideNotification,
  clearNotifications,
  setTheme,
  toggleSidebar,
  setSidebarCollapsed,
  addToSearchHistory,
  clearSearchHistory,
} = uiSlice.actions;

export const selectModals = (state: { ui: UIState }) => state.ui.modals;
export const selectNotifications = (state: { ui: UIState }) =>
  state.ui.notifications;
export const selectTheme = (state: { ui: UIState }) => state.ui.theme;
export const selectSidebarCollapsed = (state: { ui: UIState }) =>
  state.ui.sidebarCollapsed;
export const selectSearchHistory = (state: { ui: UIState }) =>
  state.ui.searchHistory;

// Loading state selectors
export const setGlobalLoading = (loading: boolean) => ({
  type: "ui/setGlobalLoading",
  payload: loading,
});

export const setActionLoading = (key: string, loading: boolean) => ({
  type: "ui/setActionLoading",
  payload: { key, loading },
});

export const clearAllLoading = () => ({
  type: "ui/clearAllLoading",
});

export const selectGlobalLoading = (state: { ui: UIState }) => state.ui.loading;
export const selectActionLoading = (key: string) => (state: { ui: UIState }) =>
  state.ui.loading; // Simplified for now

export default uiSlice.reducer;

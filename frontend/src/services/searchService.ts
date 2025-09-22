import apiClient from "./api";
import { AppError, Client, Case, UserProfile } from "../types";

interface SearchParams {
  query: string;
  page?: number;
  page_size?: number;
  type?: "client" | "case" | "user" | "all";
  filters?: Record<string, any>;
}

interface SearchResult<T> {
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  facets?: {
    type: Record<string, number>;
    status: Record<string, number>;
    date_range: {
      start: string;
      end: string;
    };
  };
}

class SearchService {
  // 全局搜索
  async search(params: SearchParams): Promise<{
    clients: SearchResult<Client>;
    cases: SearchResult<Case>;
    users: SearchResult<UserProfile>;
  }> {
    try {
      return await apiClient.get<{
        clients: SearchResult<Client>;
        cases: SearchResult<Case>;
        users: SearchResult<UserProfile>;
      }>("/search", {
        params,
        useCache: true,
        cacheTTL: 30 * 1000, // 30秒缓存
      });
    } catch (error: any) {
      console.error("全局搜索失败:", error);
      throw new AppError(
        error.message || "搜索失败",
        error.code || "SEARCH_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索客户
  async searchClients(params: Omit<SearchParams, "type">): Promise<SearchResult<Client>> {
    try {
      const searchParams = {
        ...params,
        type: "client",
      };
      return await apiClient.get<SearchResult<Client>>("/search/clients", {
        params: searchParams,
        useCache: true,
        cacheTTL: 30 * 1000, // 30秒缓存
      });
    } catch (error: any) {
      console.error("搜索客户失败:", error);
      throw new AppError(
        error.message || "搜索客户失败",
        error.code || "SEARCH_CLIENTS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索案件
  async searchCases(params: Omit<SearchParams, "type">): Promise<SearchResult<Case>> {
    try {
      const searchParams = {
        ...params,
        type: "case",
      };
      return await apiClient.get<SearchResult<Case>>("/search/cases", {
        params: searchParams,
        useCache: true,
        cacheTTL: 30 * 1000, // 30秒缓存
      });
    } catch (error: any) {
      console.error("搜索案件失败:", error);
      throw new AppError(
        error.message || "搜索案件失败",
        error.code || "SEARCH_CASES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索用户
  async searchUsers(params: Omit<SearchParams, "type">): Promise<SearchResult<UserProfile>> {
    try {
      const searchParams = {
        ...params,
        type: "user",
      };
      return await apiClient.get<SearchResult<UserProfile>>("/search/users", {
        params: searchParams,
        useCache: true,
        cacheTTL: 30 * 1000, // 30秒缓存
      });
    } catch (error: any) {
      console.error("搜索用户失败:", error);
      throw new AppError(
        error.message || "搜索用户失败",
        error.code || "SEARCH_USERS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取搜索建议
  async getSearchSuggestions(query: string, limit: number = 5): Promise<{
    clients: Array<{ id: number; name: string; type: "client" }>;
    cases: Array<{ id: number; title: string; type: "case" }>;
    users: Array<{ id: number; name: string; type: "user" }>;
  }> {
    try {
      return await apiClient.get<{
        clients: Array<{ id: number; name: string; type: "client" }>;
        cases: Array<{ id: number; title: string; type: "case" }>;
        users: Array<{ id: number; name: string; type: "user" }>;
      }>("/search/suggestions", {
        params: { query, limit },
        useCache: true,
        cacheTTL: 10 * 1000, // 10秒缓存
      });
    } catch (error: any) {
      console.error("获取搜索建议失败:", error);
      throw new AppError(
        error.message || "获取搜索建议失败",
        error.code || "GET_SUGGESTIONS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取搜索历史
  async getSearchHistory(limit: number = 10): Promise<Array<{
    id: number;
    query: string;
    type: string;
    created_at: string;
  }>> {
    try {
      return await apiClient.get<Array<{
        id: number;
        query: string;
        type: string;
        created_at: string;
      }>>("/search/history", {
        params: { limit },
        useCache: true,
        cacheTTL: 5 * 60 * 1000, // 5分钟缓存
      });
    } catch (error: any) {
      console.error("获取搜索历史失败:", error);
      throw new AppError(
        error.message || "获取搜索历史失败",
        error.code || "GET_SEARCH_HISTORY_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 添加搜索历史
  async addToSearchHistory(query: string, type: string): Promise<void> {
    try {
      await apiClient.post("/search/history", { query, type }, {
      });
    } catch (error: any) {
      console.error("添加搜索历史失败:", error);
      // 静默失败，不影响主要搜索功能
    }
  }

  // 清除搜索历史
  async clearSearchHistory(): Promise<void> {
    try {
      await apiClient.delete("/search/history", {
      });
    } catch (error: any) {
      console.error("清除搜索历史失败:", error);
      throw new AppError(
        error.message || "清除搜索历史失败",
        error.code || "CLEAR_SEARCH_HISTORY_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取热门搜索
  async getPopularSearches(limit: number = 10): Promise<Array<{
    query: string;
    count: number;
    type: string;
  }>> {
    try {
      return await apiClient.get<Array<{
        query: string;
        count: number;
        type: string;
      }>>("/search/popular", {
        params: { limit },
        useCache: true,
        cacheTTL: 10 * 60 * 1000, // 10分钟缓存
      });
    } catch (error: any) {
      console.error("获取热门搜索失败:", error);
      throw new AppError(
        error.message || "获取热门搜索失败",
        error.code || "GET_POPULAR_SEARCHES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 高级搜索
  async advancedSearch(params: {
    query?: string;
    filters: Record<string, any>;
    page?: number;
    page_size?: number;
    sort_by?: string;
    sort_order?: "asc" | "desc";
  }): Promise<{
    clients: SearchResult<Client>;
    cases: SearchResult<Case>;
    users: SearchResult<UserProfile>;
  }> {
    try {
      return await apiClient.post<{
        clients: SearchResult<Client>;
        cases: SearchResult<Case>;
        users: SearchResult<UserProfile>;
      }>("/search/advanced", params, {});
    } catch (error: any) {
      console.error("高级搜索失败:", error);
      throw new AppError(
        error.message || "高级搜索失败",
        error.code || "ADVANCED_SEARCH_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const searchService = new SearchService();

// 为了向后兼容，也导出独立的函数
export const searchClients = (params: Omit<SearchParams, "type">) =>
  searchService.searchClients(params);
export const searchCases = (params: Omit<SearchParams, "type">) =>
  searchService.searchCases(params);
export const searchUsers = (params: Omit<SearchParams, "type">) =>
  searchService.searchUsers(params);
export const getSearchSuggestions = (query: string, limit?: number) =>
  searchService.getSearchSuggestions(query, limit);
export const getSearchHistory = (limit?: number) =>
  searchService.getSearchHistory(limit);
export const addToSearchHistory = (query: string, type: string) =>
  searchService.addToSearchHistory(query, type);
export const clearSearchHistory = () => searchService.clearSearchHistory();
export const getPopularSearches = (limit?: number) =>
  searchService.getPopularSearches(limit);
export const advancedSearch = (params: {
  query?: string;
  filters: Record<string, any>;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: "asc" | "desc";
}) => searchService.advancedSearch(params);

export default searchService;
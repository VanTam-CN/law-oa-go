package services

// Common pagination and summary types used across services

// Pagination 分页信息
type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// PaginationWithTotalPage 带总页数的分页信息（用于通知队列等）
type PaginationWithTotalPage struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int64 `json:"total_page"`
}

// PaginationInfo 分页信息（通用别名）
type PaginationInfo = Pagination

// ClientSummary 客户摘要
type ClientSummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// CaseSummary 案件摘要
type CaseSummary struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	CaseNumber string `json:"case_number"`
}

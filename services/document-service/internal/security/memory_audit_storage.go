package security

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

// MemoryAuditStorage 内存审计存储实现
type MemoryAuditStorage struct {
	events  map[string]*AuditEvent
	reports map[string]*AuditReport
	mutex   sync.RWMutex
	logger  *slog.Logger
}

// NewMemoryAuditStorage 创建内存审计存储
func NewMemoryAuditStorage(logger *slog.Logger) *MemoryAuditStorage {
	return &MemoryAuditStorage{
		events:  make(map[string]*AuditEvent),
		reports: make(map[string]*AuditReport),
		logger:  logger,
	}
}

// StoreEvent 存储事件
func (mas *MemoryAuditStorage) StoreEvent(event *AuditEvent) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	mas.events[event.ID] = event
	return nil
}

// StoreReport 存储报告
func (mas *MemoryAuditStorage) StoreReport(report *AuditReport) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	mas.reports[report.ID] = report
	return nil
}

// GetEvents 获取事件
func (mas *MemoryAuditStorage) GetEvents(filter *EventFilter) ([]*AuditEvent, error) {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	var events []*AuditEvent

	for _, event := range mas.events {
		if mas.matchesEventFilter(event, filter) {
			events = append(events, event)
		}
	}

	// 按时间排序（最新的在前）
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	// 应用分页
	if filter.Limit != nil || filter.Offset != nil {
		limit := 100
		offset := 0

		if filter.Limit != nil {
			limit = *filter.Limit
		}
		if filter.Offset != nil {
			offset = *filter.Offset
		}

		if offset >= len(events) {
			return []*AuditEvent{}, nil
		}

		end := offset + limit
		if end > len(events) {
			end = len(events)
		}

		events = events[offset:end]
	}

	return events, nil
}

// GetReports 获取报告
func (mas *MemoryAuditStorage) GetReports(filter *ReportFilter) ([]*AuditReport, error) {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	var reports []*AuditReport

	for _, report := range mas.reports {
		if mas.matchesReportFilter(report, filter) {
			reports = append(reports, report)
		}
	}

	// 按生成时间排序（最新的在前）
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].GeneratedAt.After(reports[j].GeneratedAt)
	})

	// 应用分页
	if filter.Limit != nil || filter.Offset != nil {
		limit := 50
		offset := 0

		if filter.Limit != nil {
			limit = *filter.Limit
		}
		if filter.Offset != nil {
			offset = *filter.Offset
		}

		if offset >= len(reports) {
			return []*AuditReport{}, nil
		}

		end := offset + limit
		if end > len(reports) {
			end = len(reports)
		}

		reports = reports[offset:end]
	}

	return reports, nil
}

// DeleteEventsBefore 删除指定时间之前的事件
func (mas *MemoryAuditStorage) DeleteEventsBefore(t time.Time) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	deletedCount := 0
	for id, event := range mas.events {
		if event.Timestamp.Before(t) {
			delete(mas.events, id)
			deletedCount++
		}
	}

	mas.logger.Info("删除过期事件完成",
		"deleted_count", deletedCount,
		"expired_before", t.Format(time.RFC3339),
	)

	return nil
}

// GetEventCount 获取事件数量
func (mas *MemoryAuditStorage) GetEventCount(filter *EventFilter) (int64, error) {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	var count int64 = 0

	for _, event := range mas.events {
		if mas.matchesEventFilter(event, filter) {
			count++
		}
	}

	return count, nil
}

// matchesEventFilter 检查事件是否匹配过滤器
func (mas *MemoryAuditStorage) matchesEventFilter(event *AuditEvent, filter *EventFilter) bool {
	if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
		return false
	}

	if filter.Level != nil && event.Level != *filter.Level {
		return false
	}

	if filter.Category != nil && event.Category != *filter.Category {
		return false
	}

	if filter.UserID != nil && event.UserID != *filter.UserID {
		return false
	}

	if filter.Action != nil && event.Action != *filter.Action {
		return false
	}

	if filter.Resource != nil && event.Resource != *filter.Resource {
		return false
	}

	if filter.SessionID != nil && event.SessionID != *filter.SessionID {
		return false
	}

	if filter.RequestID != nil && event.RequestID != *filter.RequestID {
		return false
	}

	if filter.IPAddress != nil && event.IPAddress != *filter.IPAddress {
		return false
	}

	return true
}

// matchesReportFilter 检查报告是否匹配过滤器
func (mas *MemoryAuditStorage) matchesReportFilter(report *AuditReport, filter *ReportFilter) bool {
	if filter.StartTime != nil && report.Period.StartTime.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && report.Period.EndTime.After(*filter.EndTime) {
		return false
	}

	if filter.Type != nil && report.Period.Type != *filter.Type {
		return false
	}

	return true
}

// GetEvent 获取单个事件
func (mas *MemoryAuditStorage) GetEvent(id string) (*AuditEvent, error) {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	if event, exists := mas.events[id]; exists {
		return event, nil
	}

	return nil, fmt.Errorf("事件不存在: %s", id)
}

// GetReport 获取单个报告
func (mas *MemoryAuditStorage) GetReport(id string) (*AuditReport, error) {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	if report, exists := mas.reports[id]; exists {
		return report, nil
	}

	return nil, fmt.Errorf("报告不存在: %s", id)
}

// UpdateEvent 更新事件
func (mas *MemoryAuditStorage) UpdateEvent(event *AuditEvent) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	if _, exists := mas.events[event.ID]; !exists {
		return fmt.Errorf("事件不存在: %s", event.ID)
	}

	mas.events[event.ID] = event
	return nil
}

// UpdateReport 更新报告
func (mas *MemoryAuditStorage) UpdateReport(report *AuditReport) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	if _, exists := mas.reports[report.ID]; !exists {
		return fmt.Errorf("报告不存在: %s", report.ID)
	}

	mas.reports[report.ID] = report
	return nil
}

// DeleteEvent 删除事件
func (mas *MemoryAuditStorage) DeleteEvent(id string) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	if _, exists := mas.events[id]; !exists {
		return fmt.Errorf("事件不存在: %s", id)
	}

	delete(mas.events, id)
	return nil
}

// DeleteReport 删除报告
func (mas *MemoryAuditStorage) DeleteReport(id string) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	if _, exists := mas.reports[id]; !exists {
		return fmt.Errorf("报告不存在: %s", id)
	}

	delete(mas.reports, id)
	return nil
}

// GetStorageStats 获取存储统计信息
func (mas *MemoryAuditStorage) GetStorageStats() map[string]interface{} {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	return map[string]interface{}{
		"total_events":  len(mas.events),
		"total_reports": len(mas.reports),
		"storage_type": "memory",
		"last_updated": time.Now(),
	}
}
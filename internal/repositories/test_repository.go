package repositories

import (
	"context"
	"errors"
	"time"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// Testing Repository Errors
var (
	ErrTestSuiteNotFound       = errors.New("test suite not found")
	ErrTestSuiteAlreadyExists = errors.New("test suite already exists")
	ErrTestSuiteInvalid       = errors.New("invalid test suite data")

	ErrTestExecutionNotFound  = errors.New("test execution not found")
	ErrTestExecutionInvalid   = errors.New("invalid test execution data")

	ErrTestResultNotFound     = errors.New("test result not found")
	ErrTestResultInvalid      = errors.New("invalid test result data")

	ErrUserSessionNotFound    = errors.New("user session not found")
	ErrUserEventNotFound      = errors.New("user event not found")

	ErrAnalyticsReportNotFound = errors.New("analytics report not found")
	ErrSystemMetricNotFound   = errors.New("system metric not found")
	ErrAlertNotFound          = errors.New("alert not found")
)

// TestRepository 测试数据仓库接口
type TestRepository interface {
	// Test Suite operations
	CreateTestSuite(ctx context.Context, suite *models.TestSuite) error
	GetTestSuiteByID(ctx context.Context, id string) (*models.TestSuite, error)
	GetTestSuites(ctx context.Context, params *TestSuiteListParams) ([]*models.TestSuite, int64, error)
	UpdateTestSuite(ctx context.Context, suite *models.TestSuite) error
	DeleteTestSuite(ctx context.Context, id string) error
	GetActiveTestSuites(ctx context.Context) ([]*models.TestSuite, error)

	// Test Execution operations
	CreateTestExecution(ctx context.Context, execution *models.TestExecution) error
	GetTestExecutionByID(ctx context.Context, id string) (*models.TestExecution, error)
	GetTestExecutions(ctx context.Context, params *TestExecutionListParams) ([]*models.TestExecution, int64, error)
	UpdateTestExecution(ctx context.Context, execution *models.TestExecution) error
	UpdateTestExecutionStatus(ctx context.Context, id string, status models.TestExecutionStatus, errorMessage string) error
	GetTestExecutionsBySuiteID(ctx context.Context, suiteID string, limit int) ([]*models.TestExecution, error)

	// Test Result operations
	CreateTestResult(ctx context.Context, result *models.TestResult) error
	CreateTestResults(ctx context.Context, results []*models.TestResult) error
	GetTestResultsByExecutionID(ctx context.Context, executionID string) ([]*models.TestResult, error)
	GetTestResults(ctx context.Context, params *TestResultListParams) ([]*models.TestResult, int64, error)

	// User Event operations
	CreateUserEvent(ctx context.Context, event *models.UserEvent) error
	CreateUserEvents(ctx context.Context, events []*models.UserEvent) error
	GetUserEvents(ctx context.Context, params *UserEventListParams) ([]*models.UserEvent, int64, error)
	GetUserEventsBySessionID(ctx context.Context, sessionID string, limit int) ([]*models.UserEvent, error)
	DeleteUserEvents(ctx context.Context, olderThan time.Time) error

	// User Session operations
	CreateUserSession(ctx context.Context, session *models.UserSession) error
	GetUserSessionByID(ctx context.Context, id string) (*models.UserSession, error)
	GetUserSessions(ctx context.Context, params *UserSessionListParams) ([]*models.UserSession, int64, error)
	UpdateUserSession(ctx context.Context, session *models.UserSession) error
	GetUserSessionsByUserID(ctx context.Context, userID string, limit int) ([]*models.UserSession, error)

	// Analytics Report operations
	CreateAnalyticsReport(ctx context.Context, report *models.AnalyticsReport) error
	GetAnalyticsReportByID(ctx context.Context, id string) (*models.AnalyticsReport, error)
	GetAnalyticsReports(ctx context.Context, params *AnalyticsReportListParams) ([]*models.AnalyticsReport, int64, error)
	UpdateAnalyticsReport(ctx context.Context, report *models.AnalyticsReport) error
	DeleteAnalyticsReport(ctx context.Context, id string) error

	// System Metric operations
	CreateSystemMetric(ctx context.Context, metric *models.SystemMetric) error
	CreateSystemMetrics(ctx context.Context, metrics []*models.SystemMetric) error
	GetSystemMetrics(ctx context.Context, params *SystemMetricListParams) ([]*models.SystemMetric, int64, error)
	DeleteSystemMetrics(ctx context.Context, olderThan time.Time) error

	// Alert operations
	CreateAlert(ctx context.Context, alert *models.Alert) error
	GetAlertByID(ctx context.Context, id string) (*models.Alert, error)
	GetAlerts(ctx context.Context, params *AlertListParams) ([]*models.Alert, int64, error)
	UpdateAlert(ctx context.Context, alert *models.Alert) error
	UpdateAlertStatus(ctx context.Context, id string, status models.AlertStatus, acknowledgedBy, resolvedBy string) error
	GetActiveAlerts(ctx context.Context) ([]*models.Alert, error)

	// Alert Notification operations
	CreateAlertNotification(ctx context.Context, notification *models.AlertNotification) error
	GetAlertNotificationsByAlertID(ctx context.Context, alertID string) ([]*models.AlertNotification, error)
	UpdateAlertNotificationStatus(ctx context.Context, id string, status models.NotificationStatus, errorMessage string) error
}

// testRepository 测试数据仓库实现
type testRepository struct {
	db *gorm.DB
}

// NewTestRepository 创建测试数据仓库
func NewTestRepository(db *gorm.DB) TestRepository {
	return &testRepository{db: db}
}

// Test Suite operations
func (r *testRepository) CreateTestSuite(ctx context.Context, suite *models.TestSuite) error {
	if err := r.db.WithContext(ctx).Create(suite).Error; err != nil {
		return NewRepositoryError("create test suite", "TestSuite", err)
	}
	return nil
}

func (r *testRepository) GetTestSuiteByID(ctx context.Context, id string) (*models.TestSuite, error) {
	var suite models.TestSuite
	err := r.db.WithContext(ctx).First(&suite, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("get test suite by id", "TestSuite", id, ErrTestSuiteNotFound)
		}
		return nil, NewRepositoryError("get test suite by id", "TestSuite", err)
	}
	return &suite, nil
}

func (r *testRepository) GetTestSuites(ctx context.Context, params *TestSuiteListParams) ([]*models.TestSuite, int64, error) {
	var suites []*models.TestSuite
	var total int64

	query := r.db.WithContext(ctx).Model(&models.TestSuite{})

	// 应用过滤条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}
	if params.CreatedBy != "" {
		query = query.Where("created_by = ?", params.CreatedBy)
	}
	if params.Search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count test suites", "TestSuite", err)
	}

	// 应用排序和分页
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("created_at DESC")
	}

	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	if err := query.Find(&suites).Error; err != nil {
		return nil, 0, NewRepositoryError("list test suites", "TestSuite", err)
	}

	return suites, total, nil
}

func (r *testRepository) UpdateTestSuite(ctx context.Context, suite *models.TestSuite) error {
	result := r.db.WithContext(ctx).Model(suite).Where("id = ?", suite.ID).Updates(suite)
	if result.Error != nil {
		return NewRepositoryError("update test suite", "TestSuite", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update test suite", "TestSuite", suite.ID, ErrTestSuiteNotFound)
	}
	return nil
}

func (r *testRepository) DeleteTestSuite(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.TestSuite{}, "id = ?", id)
	if result.Error != nil {
		return NewRepositoryError("delete test suite", "TestSuite", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("delete test suite", "TestSuite", id, ErrTestSuiteNotFound)
	}
	return nil
}

func (r *testRepository) GetActiveTestSuites(ctx context.Context) ([]*models.TestSuite, error) {
	var suites []*models.TestSuite
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&suites).Error
	if err != nil {
		return nil, NewRepositoryError("get active test suites", "TestSuite", err)
	}
	return suites, nil
}

// Test Execution operations
func (r *testRepository) CreateTestExecution(ctx context.Context, execution *models.TestExecution) error {
	if err := r.db.WithContext(ctx).Create(execution).Error; err != nil {
		return NewRepositoryError("create test execution", "TestExecution", err)
	}
	return nil
}

func (r *testRepository) GetTestExecutionByID(ctx context.Context, id string) (*models.TestExecution, error) {
	var execution models.TestExecution
	err := r.db.WithContext(ctx).
		Preload("TestSuite").
		Preload("TestResults").
		First(&execution, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("get test execution by id", "TestExecution", id, ErrTestExecutionNotFound)
		}
		return nil, NewRepositoryError("get test execution by id", "TestExecution", err)
	}
	return &execution, nil
}

func (r *testRepository) GetTestExecutions(ctx context.Context, params *TestExecutionListParams) ([]*models.TestExecution, int64, error) {
	var executions []*models.TestExecution
	var total int64

	query := r.db.WithContext(ctx).Model(&models.TestExecution{})

	// 应用过滤条件
	if params.SuiteID != "" {
		query = query.Where("suite_id = ?", params.SuiteID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Environment != "" {
		query = query.Where("environment = ?", params.Environment)
	}
	if params.TriggerType != "" {
		query = query.Where("trigger_type = ?", params.TriggerType)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("started_at >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("started_at <= ?", params.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count test executions", "TestExecution", err)
	}

	// 应用排序和分页
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("started_at DESC")
	}

	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	if err := query.Preload("TestSuite").Find(&executions).Error; err != nil {
		return nil, 0, NewRepositoryError("list test executions", "TestExecution", err)
	}

	return executions, total, nil
}

func (r *testRepository) UpdateTestExecution(ctx context.Context, execution *models.TestExecution) error {
	result := r.db.WithContext(ctx).Model(execution).Where("id = ?", execution.ID).Updates(execution)
	if result.Error != nil {
		return NewRepositoryError("update test execution", "TestExecution", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update test execution", "TestExecution", execution.ID, ErrTestExecutionNotFound)
	}
	return nil
}

func (r *testRepository) UpdateTestExecutionStatus(ctx context.Context, id string, status models.TestExecutionStatus, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}

	if status == models.TestStatusCompleted || status == models.TestStatusFailed || status == models.TestStatusCancelled {
		updates["completed_at"] = time.Now()
	}

	result := r.db.WithContext(ctx).Model(&models.TestExecution{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return NewRepositoryError("update test execution status", "TestExecution", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update test execution status", "TestExecution", id, ErrTestExecutionNotFound)
	}
	return nil
}

func (r *testRepository) GetTestExecutionsBySuiteID(ctx context.Context, suiteID string, limit int) ([]*models.TestExecution, error) {
	var executions []*models.TestExecution
	err := r.db.WithContext(ctx).
		Where("suite_id = ?", suiteID).
		Order("started_at DESC").
		Limit(limit).
		Find(&executions).Error
	if err != nil {
		return nil, NewRepositoryError("get test executions by suite id", "TestExecution", err)
	}
	return executions, nil
}

// Test Result operations
func (r *testRepository) CreateTestResult(ctx context.Context, result *models.TestResult) error {
	if err := r.db.WithContext(ctx).Create(result).Error; err != nil {
		return NewRepositoryError("create test result", "TestResult", err)
	}
	return nil
}

func (r *testRepository) CreateTestResults(ctx context.Context, results []*models.TestResult) error {
	if len(results) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(results, 100).Error; err != nil {
		return NewRepositoryError("create test results", "TestResult", err)
	}
	return nil
}

func (r *testRepository) GetTestResultsByExecutionID(ctx context.Context, executionID string) ([]*models.TestResult, error) {
	var results []*models.TestResult
	err := r.db.WithContext(ctx).
		Where("execution_id = ?", executionID).
		Order("created_at ASC").
		Find(&results).Error
	if err != nil {
		return nil, NewRepositoryError("get test results by execution id", "TestResult", err)
	}
	return results, nil
}

func (r *testRepository) GetTestResults(ctx context.Context, params *TestResultListParams) ([]*models.TestResult, int64, error) {
	var results []*models.TestResult
	var total int64

	query := r.db.WithContext(ctx).Model(&models.TestResult{})

	// 应用过滤条件
	if params.ExecutionID != "" {
		query = query.Where("execution_id = ?", params.ExecutionID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.TestType != "" {
		query = query.Where("test_type = ?", params.TestType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count test results", "TestResult", err)
	}

	// 应用排序和分页
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("created_at ASC")
	}

	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, 0, NewRepositoryError("list test results", "TestResult", err)
	}

	return results, total, nil
}

// User Event operations
func (r *testRepository) CreateUserEvent(ctx context.Context, event *models.UserEvent) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return NewRepositoryError("create user event", "UserEvent", err)
	}
	return nil
}

func (r *testRepository) CreateUserEvents(ctx context.Context, events []*models.UserEvent) error {
	if len(events) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(events, 200).Error; err != nil {
		return NewRepositoryError("create user events", "UserEvent", err)
	}
	return nil
}

func (r *testRepository) GetUserEvents(ctx context.Context, params *UserEventListParams) ([]*models.UserEvent, int64, error) {
	var events []*models.UserEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&models.UserEvent{})

	// 应用过滤条件
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.SessionID != "" {
		query = query.Where("session_id = ?", params.SessionID)
	}
	if params.EventType != "" {
		query = query.Where("event_type = ?", params.EventType)
	}
	if params.PageURL != "" {
		query = query.Where("page_url LIKE ?", "%"+params.PageURL+"%")
	}
	if !params.StartTime.IsZero() {
		query = query.Where("timestamp >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("timestamp <= ?", params.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count user events", "UserEvent", err)
	}

	// 应用排序和分页
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("timestamp DESC")
	}

	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, 0, NewRepositoryError("list user events", "UserEvent", err)
	}

	return events, total, nil
}

func (r *testRepository) GetUserEventsBySessionID(ctx context.Context, sessionID string, limit int) ([]*models.UserEvent, error) {
	var events []*models.UserEvent
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("timestamp ASC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, NewRepositoryError("get user events by session id", "UserEvent", err)
	}
	return events, nil
}

func (r *testRepository) DeleteUserEvents(ctx context.Context, olderThan time.Time) error {
	result := r.db.WithContext(ctx).Where("created_at < ?", olderThan).Delete(&models.UserEvent{})
	if result.Error != nil {
		return NewRepositoryError("delete old user events", "UserEvent", result.Error)
	}
	return nil
}

// User Session operations
func (r *testRepository) CreateUserSession(ctx context.Context, session *models.UserSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return NewRepositoryError("create user session", "UserSession", err)
	}
	return nil
}

func (r *testRepository) GetUserSessionByID(ctx context.Context, id string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("get user session by id", "UserSession", id, ErrUserSessionNotFound)
		}
		return nil, NewRepositoryError("get user session by id", "UserSession", err)
	}
	return &session, nil
}

func (r *testRepository) GetUserSessions(ctx context.Context, params *UserSessionListParams) ([]*models.UserSession, int64, error) {
	var sessions []*models.UserSession
	var total int64

	query := r.db.WithContext(ctx).Model(&models.UserSession{})

	// 应用过滤条件
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.DeviceType != "" {
		query = query.Where("device_type = ?", params.DeviceType)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("started_at >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("started_at <= ?", params.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count user sessions", "UserSession", err)
	}

	// 应用排序和分页
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("started_at DESC")
	}

	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	if err := query.Find(&sessions).Error; err != nil {
		return nil, 0, NewRepositoryError("list user sessions", "UserSession", err)
	}

	return sessions, total, nil
}

func (r *testRepository) UpdateUserSession(ctx context.Context, session *models.UserSession) error {
	result := r.db.WithContext(ctx).Model(session).Where("id = ?", session.ID).Updates(session)
	if result.Error != nil {
		return NewRepositoryError("update user session", "UserSession", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update user session", "UserSession", session.ID, ErrUserSessionNotFound)
	}
	return nil
}

func (r *testRepository) GetUserSessionsByUserID(ctx context.Context, userID string, limit int) ([]*models.UserSession, error) {
	var sessions []*models.UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("started_at DESC").
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, NewRepositoryError("get user sessions by user id", "UserSession", err)
	}
	return sessions, nil
}

// Analytics Report operations
func (r *testRepository) CreateAnalyticsReport(ctx context.Context, report *models.AnalyticsReport) error {
	if err := r.db.WithContext(ctx).Create(report).Error; err != nil {
		return NewRepositoryError("create analytics report", "AnalyticsReport", err)
	}
	return nil
}

func (r *testRepository) GetAnalyticsReportByID(ctx context.Context, id string) (*models.AnalyticsReport, error) {
	var report models.AnalyticsReport
	err := r.db.WithContext(ctx).First(&report, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("get analytics report by id", "AnalyticsReport", id, ErrAnalyticsReportNotFound)
		}
		return nil, NewRepositoryError("get analytics report by id", "AnalyticsReport", err)
	}
	return &report, nil
}

func (r *testRepository) GetAnalyticsReports(ctx context.Context, params *AnalyticsReportListParams) ([]*models.AnalyticsReport, int64, error) {
	var reports []*models.AnalyticsReport
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AnalyticsReport{})

	// 应用过滤条件
	if params.ReportType != "" {
		query = query.Where("report_type = ?", params.ReportType)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.GeneratedBy != "" {
		query = query.Where("generated_by = ?", params.GeneratedBy)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("created_at >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("created_at <= ?", params.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count analytics reports", "AnalyticsReport", err)
	}

	// 应用排序和分页
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("created_at DESC")
	}

	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	if err := query.Find(&reports).Error; err != nil {
		return nil, 0, NewRepositoryError("list analytics reports", "AnalyticsReport", err)
	}

	return reports, total, nil
}

func (r *testRepository) UpdateAnalyticsReport(ctx context.Context, report *models.AnalyticsReport) error {
	result := r.db.WithContext(ctx).Model(report).Where("id = ?", report.ID).Updates(report)
	if result.Error != nil {
		return NewRepositoryError("update analytics report", "AnalyticsReport", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update analytics report", "AnalyticsReport", report.ID, ErrAnalyticsReportNotFound)
	}
	return nil
}

func (r *testRepository) DeleteAnalyticsReport(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.AnalyticsReport{}, "id = ?", id)
	if result.Error != nil {
		return NewRepositoryError("delete analytics report", "AnalyticsReport", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("delete analytics report", "AnalyticsReport", id, ErrAnalyticsReportNotFound)
	}
	return nil
}

// System Metric operations
func (r *testRepository) CreateSystemMetric(ctx context.Context, metric *models.SystemMetric) error {
	if err := r.db.WithContext(ctx).Create(metric).Error; err != nil {
		return NewRepositoryError("create system metric", "SystemMetric", err)
	}
	return nil
}

func (r *testRepository) CreateSystemMetrics(ctx context.Context, metrics []*models.SystemMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(metrics, 500).Error; err != nil {
		return NewRepositoryError("create system metrics", "SystemMetric", err)
	}
	return nil
}

func (r *testRepository) GetSystemMetrics(ctx context.Context, params *SystemMetricListParams) ([]*models.SystemMetric, int64, error) {
	var metrics []*models.SystemMetric
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SystemMetric{})

	// 应用过滤条件
	if params.MetricName != "" {
		query = query.Where("metric_name = ?", params.MetricName)
	}
	if params.MetricType != "" {
		query = query.Where("metric_type = ?", params.MetricType)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("timestamp >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("timestamp <= ?", params.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count system metrics", "SystemMetric", err)
	}

	// 应用排序和分页
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("timestamp DESC")
	}

	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	if err := query.Find(&metrics).Error; err != nil {
		return nil, 0, NewRepositoryError("list system metrics", "SystemMetric", err)
	}

	return metrics, total, nil
}

func (r *testRepository) DeleteSystemMetrics(ctx context.Context, olderThan time.Time) error {
	result := r.db.WithContext(ctx).Where("created_at < ?", olderThan).Delete(&models.SystemMetric{})
	if result.Error != nil {
		return NewRepositoryError("delete old system metrics", "SystemMetric", result.Error)
	}
	return nil
}

// Alert operations
func (r *testRepository) CreateAlert(ctx context.Context, alert *models.Alert) error {
	if err := r.db.WithContext(ctx).Create(alert).Error; err != nil {
		return NewRepositoryError("create alert", "Alert", err)
	}
	return nil
}

func (r *testRepository) GetAlertByID(ctx context.Context, id string) (*models.Alert, error) {
	var alert models.Alert
	err := r.db.WithContext(ctx).
		Preload("Notifications").
		First(&alert, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("get alert by id", "Alert", id, ErrAlertNotFound)
		}
		return nil, NewRepositoryError("get alert by id", "Alert", err)
	}
	return &alert, nil
}

func (r *testRepository) GetAlerts(ctx context.Context, params *AlertListParams) ([]*models.Alert, int64, error) {
	var alerts []*models.Alert
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Alert{})

	// 应用过滤条件
	if params.Severity != "" {
		query = query.Where("severity = ?", params.Severity)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Source != "" {
		query = query.Where("source = ?", params.Source)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("triggered_at >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("triggered_at <= ?", params.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count alerts", "Alert", err)
	}

	// 应用排序和分页
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("triggered_at DESC")
	}

	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	if err := query.Find(&alerts).Error; err != nil {
		return nil, 0, NewRepositoryError("list alerts", "Alert", err)
	}

	return alerts, total, nil
}

func (r *testRepository) UpdateAlert(ctx context.Context, alert *models.Alert) error {
	result := r.db.WithContext(ctx).Model(alert).Where("id = ?", alert.ID).Updates(alert)
	if result.Error != nil {
		return NewRepositoryError("update alert", "Alert", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update alert", "Alert", alert.ID, ErrAlertNotFound)
	}
	return nil
}

func (r *testRepository) UpdateAlertStatus(ctx context.Context, id string, status models.AlertStatus, acknowledgedBy, resolvedBy string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	if status == models.AlertStatusAcknowledged {
		updates["acknowledged_at"] = now
		updates["acknowledged_by"] = acknowledgedBy
	}
	if status == models.AlertStatusResolved {
		updates["resolved_at"] = now
		updates["resolved_by"] = resolvedBy
	}

	result := r.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return NewRepositoryError("update alert status", "Alert", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update alert status", "Alert", id, ErrAlertNotFound)
	}
	return nil
}

func (r *testRepository) GetActiveAlerts(ctx context.Context) ([]*models.Alert, error) {
	var alerts []*models.Alert
	err := r.db.WithContext(ctx).
		Where("status = ?", models.AlertStatusActive).
		Order("triggered_at DESC").
		Find(&alerts).Error
	if err != nil {
		return nil, NewRepositoryError("get active alerts", "Alert", err)
	}
	return alerts, nil
}

// Alert Notification operations
func (r *testRepository) CreateAlertNotification(ctx context.Context, notification *models.AlertNotification) error {
	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
		return NewRepositoryError("create alert notification", "AlertNotification", err)
	}
	return nil
}

func (r *testRepository) GetAlertNotificationsByAlertID(ctx context.Context, alertID string) ([]*models.AlertNotification, error) {
	var notifications []*models.AlertNotification
	err := r.db.WithContext(ctx).
		Where("alert_id = ?", alertID).
		Order("created_at DESC").
		Find(&notifications).Error
	if err != nil {
		return nil, NewRepositoryError("get alert notifications by alert id", "AlertNotification", err)
	}
	return notifications, nil
}

func (r *testRepository) UpdateAlertNotificationStatus(ctx context.Context, id string, status models.NotificationStatus, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.NotificationStatusSent {
		updates["sent_at"] = time.Now()
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
		updates["retry_count"] = gorm.Expr("retry_count + 1")
	}

	result := r.db.WithContext(ctx).Model(&models.AlertNotification{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return NewRepositoryError("update alert notification status", "AlertNotification", result.Error)
	}
	return nil
}

// List Parameters
type TestSuiteListParams struct {
	Page      int
	PageSize  int
	Type      models.TestType
	IsActive  *bool
	CreatedBy string
	Search    string
	OrderBy   string
}

type TestExecutionListParams struct {
	Page        int
	PageSize    int
	SuiteID     string
	Status      models.TestExecutionStatus
	Environment string
	TriggerType models.TriggerType
	StartTime   time.Time
	EndTime     time.Time
	OrderBy     string
}

type TestResultListParams struct {
	Page         int
	PageSize     int
	ExecutionID  string
	Status       models.TestResultStatus
	TestType     string
	OrderBy      string
}

type UserEventListParams struct {
	Page      int
	PageSize  int
	UserID    string
	SessionID string
	EventType string
	PageURL   string
	StartTime time.Time
	EndTime   time.Time
	OrderBy   string
}

type UserSessionListParams struct {
	Page       int
	PageSize   int
	UserID     string
	DeviceType string
	StartTime  time.Time
	EndTime    time.Time
	OrderBy    string
}

type AnalyticsReportListParams struct {
	Page       int
	PageSize   int
	ReportType models.ReportType
	Status     models.ReportStatus
	GeneratedBy string
	StartTime  time.Time
	EndTime    time.Time
	OrderBy    string
}

type SystemMetricListParams struct {
	Page       int
	PageSize   int
	MetricName string
	MetricType models.MetricType
	StartTime  time.Time
	EndTime    time.Time
	OrderBy    string
}

type AlertListParams struct {
	Page      int
	PageSize  int
	Severity  models.AlertSeverity
	Status    models.AlertStatus
	Source    string
	StartTime time.Time
	EndTime   time.Time
	OrderBy   string
}
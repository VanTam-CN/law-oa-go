package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// Inbox Repository Sentinel Errors
var (
	ErrInboxItemNotFound            = errors.New("inbox item not found")
	ErrInboxItemInvalid             = errors.New("invalid inbox item data")
	ErrInboxItemDeletionUnavailable = errors.New("inbox item deletion is disabled for audit retention")
	ErrReminderRuleNotFound         = errors.New("reminder rule not found")
)

// InboxRepository 待办事项数据仓库接口
type InboxRepository interface {
	// Create 创建待办事项
	Create(ctx context.Context, item *models.InboxItem) error
	// FindByID 根据ID查找待办事项
	FindByID(ctx context.Context, id uint) (*models.InboxItem, error)
	// FindByIDAndUserID 根据ID和所属用户查找待办事项
	FindByIDAndUserID(ctx context.Context, id, userID uint) (*models.InboxItem, error)
	// Update 更新待办事项
	Update(ctx context.Context, item *models.InboxItem) error
	// UpdateByUserID 仅更新指定用户的待办事项
	UpdateByUserID(ctx context.Context, item *models.InboxItem, userID uint) error
	// Delete 删除待办事项
	Delete(ctx context.Context, id uint) error
	// DeleteByUserID 仅删除指定用户的待办事项
	DeleteByUserID(ctx context.Context, id, userID uint) error
	// List 查询待办事项列表
	List(ctx context.Context, params *InboxListParams) ([]*models.InboxItem, int64, error)
	// FindByUserID 根据用户ID查询待办事项
	FindByUserID(ctx context.Context, userID uint) ([]*models.InboxItem, error)
	// MarkAsRead 标记为已读
	MarkAsRead(ctx context.Context, id uint) error
	// MarkAsReadByUserID 仅标记指定用户的待办事项为已读
	MarkAsReadByUserID(ctx context.Context, id, userID uint) error
	// MarkAsCompleted 标记为已完成
	MarkAsCompleted(ctx context.Context, id uint) error
	// MarkAsCompletedByUserID 仅标记指定用户的待办事项为已完成
	MarkAsCompletedByUserID(ctx context.Context, id, userID uint) error
	// Snooze 延后待办事项
	Snooze(ctx context.Context, id uint, until time.Time) error
	// SnoozeByUserID 仅延后指定用户的待办事项
	SnoozeByUserID(ctx context.Context, id, userID uint, until time.Time) error
	// Escalate 升级待办事项
	Escalate(ctx context.Context, id uint) error
	// GetUnreadCount 获取未读数量
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	// GetDueItems 获取到期待办事项
	GetDueItems(ctx context.Context, before time.Time) ([]*models.InboxItem, error)
	// GetOverdueCriticalItems 获取超时的critical待办
	GetOverdueCriticalItems(ctx context.Context, before time.Time) ([]*models.InboxItem, error)
	// DB 暴露底层句柄，仅供服务层只读投影查询（N5 动作生成）。
	DB() *gorm.DB

	// ConflictReviewAuthorizedItemTitles resolves, for one inbox owner, the
	// conflict-review todos whose linked matter name the actor may still
	// see. The map is inbox item id -> conflict_check_records.case_name.
	// When search is non-empty it filters by the same SQL LIKE semantics as
	// the stored title/content search, so a non-matching term can never
	// surface an authorized todo through its id alone. Fail closed: a
	// stale/revoked/reassigned/expired appointment, an inactive or
	// non-reviewer account, or a missing/forbidden intake binding hides the
	// title; the row stays searchable only by its stored title/content.
	ConflictReviewAuthorizedItemTitles(userID uint, search string) map[uint]string

	// Reminder Rules
	CreateReminderRule(ctx context.Context, rule *models.InboxReminderRule) error
	GetReminderRules(ctx context.Context, isActive bool) ([]*models.InboxReminderRule, error)
	GetReminderRuleByTypeAndPriority(ctx context.Context, dueDateType, priority string) (*models.InboxReminderRule, error)
	UpdateReminderRule(ctx context.Context, rule *models.InboxReminderRule) error
	DeleteReminderRule(ctx context.Context, id uint) error
}

// InboxReminderRuleRepository 提醒规则数据仓库接口
type InboxReminderRuleRepository interface {
	// Create 创建提醒规则
	Create(ctx context.Context, rule *models.InboxReminderRule) error
	// FindByID 根据ID查找规则
	FindByID(ctx context.Context, id uint) (*models.InboxReminderRule, error)
	// Update 更新规则
	Update(ctx context.Context, rule *models.InboxReminderRule) error
	// Delete 删除规则
	Delete(ctx context.Context, id uint) error
	// FindAll 查询所有规则
	FindAll(ctx context.Context) ([]*models.InboxReminderRule, error)
	// FindActive 查询启用的规则
	FindActive(ctx context.Context) ([]*models.InboxReminderRule, error)
	// FindByDueDateType 根据日期类型查询规则
	FindByDueDateType(ctx context.Context, dueDateType string) ([]*models.InboxReminderRule, error)
}

// InboxListParams 待办事项列表查询参数
type InboxListParams struct {
	Page        int
	PageSize    int
	UserID      uint
	IsRead      *bool
	IsCompleted *bool
	Priority    string
	SourceType  string
	DueBefore   *time.Time
	DueAfter    *time.Time
	Search      string
	OrderBy     string // due_date, priority, created_at
}

// InboxStats 待办事项统计信息
type InboxStats struct {
	Total       int64
	Unread      int64
	Pending     int64
	Completed   int64
	Critical    int64
	High        int64
	Overdue     int64
	DueToday    int64
	DueThisWeek int64
}

// InboxRepositoryImpl 待办事项数据仓库的GORM实现
type InboxRepositoryImpl struct {
	*BaseRepository[models.InboxItem]
	db *gorm.DB
}

// NewInboxRepository 创建待办事项数据仓库实例
func NewInboxRepository(db *gorm.DB) InboxRepository {
	return &InboxRepositoryImpl{
		BaseRepository: NewBaseRepository[models.InboxItem](db),
		db:             db,
	}
}

// DB 返回底层 *gorm.DB。只读投影使用；业务写入仍走仓储方法。
func (r *InboxRepositoryImpl) DB() *gorm.DB {
	return r.db
}

// Create 创建待办事项
func (r *InboxRepositoryImpl) Create(ctx context.Context, item *models.InboxItem) error {
	return r.BaseRepository.Create(ctx, item)
}

// FindByID 根据ID查找待办事项
func (r *InboxRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.InboxItem, error) {
	item, err := r.BaseRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("find", "inbox_item", id, ErrInboxItemNotFound)
		}
		return nil, NewRepositoryErrorWithID("find", "inbox_item", id, err)
	}
	return item, nil
}

func (r *InboxRepositoryImpl) FindByIDAndUserID(ctx context.Context, id, userID uint) (*models.InboxItem, error) {
	var item models.InboxItem
	if err := r.ownedItemQuery(ctx, id, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("find", "inbox_item", id, ErrInboxItemNotFound)
		}
		return nil, NewRepositoryErrorWithID("find", "inbox_item", id, err)
	}
	return &item, nil
}

// Update 更新待办事项
func (r *InboxRepositoryImpl) Update(ctx context.Context, item *models.InboxItem) error {
	return r.BaseRepository.Update(ctx, item.ID, inboxItemUpdates(item))
}

func (r *InboxRepositoryImpl) UpdateByUserID(ctx context.Context, item *models.InboxItem, userID uint) error {
	result := r.ownedItemQuery(ctx, item.ID, userID).
		Model(&models.InboxItem{}).
		Updates(inboxItemUpdates(item))
	if result.Error != nil {
		return NewRepositoryErrorWithID("update", "inbox_item", item.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		// MySQL may report zero affected rows when every submitted value is unchanged.
		if _, err := r.FindByIDAndUserID(ctx, item.ID, userID); err != nil {
			return err
		}
	}
	return nil
}

func inboxItemUpdates(item *models.InboxItem) map[string]interface{} {
	return map[string]interface{}{
		"title":          item.Title,
		"content":        item.Content,
		"priority":       item.Priority,
		"due_date":       item.DueDate,
		"due_date_type":  item.DueDateType,
		"is_read":        item.IsRead,
		"read_at":        item.ReadAt,
		"is_completed":   item.IsCompleted,
		"completed_at":   item.CompletedAt,
		"reminder_sent":  item.ReminderSent,
		"reminder_count": item.ReminderCount,
		"escalated":      item.Escalated,
		"escalated_at":   item.EscalatedAt,
		"snoozed_until":  item.SnoozedUntil,
		"snoozed_count":  item.SnoozedCount,
	}
}

// Delete 删除待办事项
func (r *InboxRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return NewRepositoryErrorWithID("delete", "inbox_item", id, ErrInboxItemDeletionUnavailable)
}

func (r *InboxRepositoryImpl) DeleteByUserID(ctx context.Context, id, userID uint) error {
	return NewRepositoryErrorWithID("delete", "inbox_item", id, ErrInboxItemDeletionUnavailable)
}

// List 查询待办事项列表
func (r *InboxRepositoryImpl) List(ctx context.Context, params *InboxListParams) ([]*models.InboxItem, int64, error) {
	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	queryBuilder := NewQueryBuilder[models.InboxItem](r.db)

	// 添加用户ID过滤
	if params.UserID > 0 {
		queryBuilder = queryBuilder.Where("user_id = ?", params.UserID)
	}

	// 添加已读状态过滤
	if params.IsRead != nil {
		queryBuilder = queryBuilder.Where("is_read = ?", *params.IsRead)
	}

	// 添加完成状态过滤
	if params.IsCompleted != nil {
		queryBuilder = queryBuilder.Where("is_completed = ?", *params.IsCompleted)
	}

	// 添加优先级过滤
	if params.Priority != "" {
		queryBuilder = queryBuilder.Where("priority = ?", params.Priority)
	}

	// 添加来源类型过滤
	if params.SourceType != "" {
		queryBuilder = queryBuilder.Where("source_type = ?", params.SourceType)
	}

	// 添加到期时间范围过滤
	if params.DueBefore != nil {
		queryBuilder = queryBuilder.Where("due_date <= ?", *params.DueBefore)
	}
	if params.DueAfter != nil {
		queryBuilder = queryBuilder.Where("due_date >= ?", *params.DueAfter)
	}

	// 添加搜索条件
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		// Case-name search is layered on top of the stored title/content
		// search BEFORE pagination: only authorized conflict-review todos
		// participate, so both the page and the total stay consistent.
		authorizedTitles := r.ConflictReviewAuthorizedItemTitles(params.UserID, params.Search)
		if len(authorizedTitles) > 0 {
			authorizedIDs := make([]uint, 0, len(authorizedTitles))
			for itemID := range authorizedTitles {
				authorizedIDs = append(authorizedIDs, itemID)
			}
			// The grouped OR keeps the original title/content search and the
			// authorized case-name branch as one condition inside the same
			// WHERE conjunction.
			queryBuilder = queryBuilder.Where(
				"(title LIKE ? OR content LIKE ?) OR id IN ?",
				searchTerm, searchTerm, authorizedIDs,
			)
		} else {
			queryBuilder = queryBuilder.Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm)
		}
	}

	// 排除延后的待办（如果延后时间未到）
	queryBuilder = queryBuilder.Where("(snoozed_until IS NULL OR snoozed_until <= NOW())")

	// 获取总数
	total, err := queryBuilder.Count(ctx)
	if err != nil {
		return nil, 0, NewRepositoryError("list", "inbox_item", err)
	}

	// 排序
	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "due_date"
	}
	switch orderBy {
	case "due_date":
		queryBuilder = queryBuilder.Order("due_date ASC").OrderDesc("priority")
	case "priority":
		// 优先级排序: critical > high > medium > low
		queryBuilder = queryBuilder.Raw(`
		 FIELD(priority, 'critical', 'high', 'medium', 'low') ASC,
		 due_date ASC
		`)
	case "created_at":
		queryBuilder = queryBuilder.OrderDesc("created_at")
	default:
		queryBuilder = queryBuilder.Order(orderBy)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	items, err := queryBuilder.Offset(offset).Limit(pageSize).Find(ctx)
	if err != nil {
		return nil, 0, NewRepositoryError("list", "inbox_item", err)
	}

	return items, total, nil
}

// FindByUserID 根据用户ID查询待办事项
func (r *InboxRepositoryImpl) FindByUserID(ctx context.Context, userID uint) ([]*models.InboxItem, error) {
	var items []*models.InboxItem
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("is_completed = ?", false).
		Where("(snoozed_until IS NULL OR snoozed_until <= NOW())").
		Order("priority DESC, due_date ASC").
		Find(&items).Error

	if err != nil {
		return nil, NewRepositoryError("find_by_user_id", "inbox_item", err)
	}

	return items, nil
}

// inboxConflictReviewRoleAllowed mirrors the reviewer role gate used by the
// conflict reviewer service; it lives here because repositories cannot import
// the services package.
func inboxConflictReviewRoleAllowed(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "director", "partner", "compliance", "risk", "risk_control", "management", "conflict_officer":
		return true
	default:
		return false
	}
}

// ConflictReviewAuthorizedItemTitles resolves the conflict-review handoff
// titles for one inbox owner. Authorization mirrors the reviewer gate:
// the newest active appointment must still belong to this user, the account
// must be an active reviewer role, and the recorded intake binding must
// still point at this user's client context behind the ethical wall.
// Any read failure hides that branch (fail closed) instead of failing the
// whole inbox page.
func (r *InboxRepositoryImpl) ConflictReviewAuthorizedItemTitles(userID uint, search string) map[uint]string {
	out := make(map[uint]string)
	if userID == 0 || r.db == nil {
		return out
	}
	db := r.db.WithContext(context.Background())

	var rows []struct {
		ID              uint
		SourceRef       string
		WorkflowTaskKey string
	}
	if err := db.Table("inbox_items").
		Select("id, source_ref, workflow_task_key").
		Where("user_id = ? AND is_completed = ? AND completion_mode = ? AND workflow_task_key LIKE ? AND deleted_at IS NULL",
			userID, false, "workflow", "conflict_review:%").
		Scan(&rows).Error; err != nil {
		return out
	}
	checkIDs := make([]string, 0, len(rows))
	itemIDsByCheck := make(map[string][]uint, len(rows))
	for _, row := range rows {
		checkID := strings.TrimSpace(row.SourceRef)
		// The deterministic handoff key must agree with the stored target;
		// a tampered or stale key can never unlock another check's title.
		if checkID == "" || row.WorkflowTaskKey != "conflict_review:"+checkID+":"+strconv.FormatUint(uint64(userID), 10) {
			continue
		}
		checkIDs = append(checkIDs, checkID)
		itemIDsByCheck[checkID] = append(itemIDsByCheck[checkID], row.ID)
	}
	if len(checkIDs) == 0 {
		return out
	}

	var assignments []struct {
		CheckID    string
		ReviewerID uint
	}
	if err := db.Table("conflict_reviewer_assignments").
		Select("check_id, reviewer_id").
		Where("check_id IN ?", checkIDs).
		Where("status = ?", models.ConflictReviewerAssignmentActive).
		Where("effective_from IS NULL OR effective_from <= ?", time.Now()).
		Where("effective_to IS NULL OR effective_to > ?", time.Now()).
		Order("created_at DESC, id DESC").
		Scan(&assignments).Error; err != nil {
		return out
	}
	activeReviewerByCheck := make(map[string]uint, len(assignments))
	for _, assignment := range assignments {
		if _, seen := activeReviewerByCheck[assignment.CheckID]; !seen {
			activeReviewerByCheck[assignment.CheckID] = assignment.ReviewerID
		}
	}
	boundCheckIDs := make([]string, 0, len(activeReviewerByCheck))
	for checkID, reviewerID := range activeReviewerByCheck {
		if reviewerID == userID {
			boundCheckIDs = append(boundCheckIDs, checkID)
		}
	}
	if len(boundCheckIDs) == 0 {
		return out
	}

	var account struct {
		Role   string
		Status string
	}
	if err := db.Table("users").
		Select("role, status").
		Where("id = ? AND deleted_at IS NULL", userID).
		Take(&account).Error; err != nil {
		return out
	}
	if !strings.EqualFold(strings.TrimSpace(account.Status), "active") || !inboxConflictReviewRoleAllowed(account.Role) {
		return out
	}

	var records []models.ConflictCheckRecord
	if err := db.Where("check_id IN ?", boundCheckIDs).Find(&records).Error; err != nil {
		return out
	}
	// A search term filters the frozen case name with the same SQL LIKE
	// semantics as the stored title/content search: an unmatched term never
	// contributes ids, so authorization can never turn into a ride-along.
	titleFilter := strings.TrimSpace(search)
	casesAvailable := db.Migrator().HasTable("cases")
	for _, record := range records {
		subjectCaseID := conflictSubjectCaseID(record)
		intakeID, _ := conflictSubjectBinding(record)
		// The explicit case reference is an OBJECT binding: the review
		// endpoint (CanReadConflictContext via ResolveConflictSubjectContext)
		// authorizes the bound case itself, so the inbox title must use the
		// same case and the same isolation rules. A visible sibling case on
		// the same client can never ride along, and the same-client count
		// branch below stays the fallback ONLY when no case binding exists.
		if subjectCaseID > 0 {
			if !casesAvailable {
				continue
			}
			authorized := false
			var caseRow models.Case
			if err := db.Where("id = ? AND deleted_at IS NULL", subjectCaseID).
				Take(&caseRow).Error; err == nil && inboxReviewerCanReadCase(db, userID, account.Role, &caseRow) {
				authorized = true
			} else if err != nil {
				continue
			}
			if !authorized {
				continue
			}
			title := strings.TrimSpace(record.CaseName)
			if title == "" {
				continue
			}
			if titleFilter != "" {
				var matching int64
				if err := db.Table("conflict_check_records").
					Where("check_id = ? AND case_name LIKE ?", record.CheckID, "%"+titleFilter+"%").
					Count(&matching).Error; err != nil {
					continue
				}
				if matching == 0 {
					continue
				}
			}
			for _, itemID := range itemIDsByCheck[record.CheckID] {
				out[itemID] = title
			}
			continue
		}
		// Without an explicit frozen case id, the reverse cases.conflict_check_id
		// binding resolves the subject case exactly like the review entry point
		// (exactly one live row; ambiguous rows fail closed).
		if casesAvailable && strings.TrimSpace(record.CheckID) != "" {
			var boundIDs []uint
			if err := db.Model(&models.Case{}).
				Where("conflict_check_id = ? AND deleted_at IS NULL", record.CheckID).
				Limit(2).
				Pluck("id", &boundIDs).Error; err != nil {
				continue
			}
			if len(boundIDs) == 1 {
				var caseRow models.Case
				if err := db.Where("id = ?", boundIDs[0]).Take(&caseRow).Error; err != nil {
					continue
				}
				if !inboxReviewerCanReadCase(db, userID, account.Role, &caseRow) {
					continue
				}
				title := strings.TrimSpace(record.CaseName)
				if title == "" {
					continue
				}
				if titleFilter != "" {
					var matching int64
					if err := db.Table("conflict_check_records").
						Where("check_id = ? AND case_name LIKE ?", record.CheckID, "%"+titleFilter+"%").
						Count(&matching).Error; err != nil {
						continue
					}
					if matching == 0 {
						continue
					}
				}
				for _, itemID := range itemIDsByCheck[record.CheckID] {
					out[itemID] = title
				}
				continue
			}
			if len(boundIDs) > 1 {
				// Ambiguous reverse binding cannot identify ONE subject case;
				// the review endpoint resolves only exactly-one, so the title
				// must not pick a winner here either (fail closed).
				continue
			}
		}
		if intakeID == "" {
			// A check bound only to a formal case has no intake todo
			// contract; keep its title hidden until the todo contract
			// defines that display path.
			continue
		}
		var intake struct {
			CreatedBy string
			ClientID  uint
		}
		if err := db.Table("case_intakes").
			Select("created_by, client_id").
			Where("id = ?", intakeID).
			Take(&intake).Error; err != nil {
			continue
		}
		recordClientID := strings.TrimSpace(record.ClientID)
		if recordClientID != "" && recordClientID != "0" &&
			recordClientID != strconv.FormatUint(uint64(intake.ClientID), 10) {
			continue
		}
		ownerID, _ := strconv.ParseUint(strings.TrimSpace(intake.CreatedBy), 10, 32)
		// The frozen intake binding must still agree with the check's
		// applicant; the reviewer is authorized by the active appointment
		// and reviewer-role gate, not by owning the intake.
		if ownerID == 0 || ownerID != uint64(record.UserID) || intake.ClientID == 0 {
			continue
		}
		if casesAvailable {
			var visible int64
			if err := db.Model(&models.Case{}).
				Where("client_id = ?", intake.ClientID).
				Where("ethical_wall_enabled = ? OR EXISTS (SELECT 1 FROM case_ethical_wall_whitelist wall_access WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?)", false, userID).
				Count(&visible).Error; err != nil {
				continue
			}
			var total int64
			if err := db.Model(&models.Case{}).
				Where("client_id = ?", intake.ClientID).
				Count(&total).Error; err != nil {
				continue
			}
			if total > 0 && visible == 0 {
				continue
			}
		}
		title := strings.TrimSpace(record.CaseName)
		if title == "" {
			continue
		}
		if titleFilter != "" {
			var matching int64
			if err := db.Table("conflict_check_records").
				Where("check_id = ? AND case_name LIKE ?", record.CheckID, "%"+titleFilter+"%").
				Count(&matching).Error; err != nil {
				continue
			}
			if matching == 0 {
				continue
			}
		}
		for _, itemID := range itemIDsByCheck[record.CheckID] {
			out[itemID] = title
		}
	}
	return out
}

// conflictSubjectCaseID mirrors the review endpoint's explicit-case
// resolution exactly like services.subjectCaseIDFromConflictParameters:
// firstNonEmpty selects ONE candidate (subjectCaseId, subject_case_id,
// caseId, case_id) and only that selected value is parsed, so an invalid
// selected value stays unbound and falls back to the reverse binding
// instead of silently authorizing a later candidate's case.
func conflictSubjectCaseID(record models.ConflictCheckRecord) uint {
	if len(record.SearchParameters) == 0 {
		return 0
	}
	values := struct {
		SubjectCaseID      string `json:"subjectCaseId"`
		SubjectCaseIDSnake string `json:"subject_case_id"`
		CaseID             string `json:"caseId"`
		CaseIDSnake        string `json:"case_id"`
	}{}
	raw, err := json.Marshal(record.SearchParameters)
	if err != nil || json.Unmarshal(raw, &values) != nil {
		return 0
	}
	selected := firstNonEmptyConflictParameter(values.SubjectCaseID, values.SubjectCaseIDSnake, values.CaseID, values.CaseIDSnake)
	if selected == "" {
		return 0
	}
	caseID, err := strconv.ParseUint(selected, 10, 32)
	if err != nil {
		return 0
	}
	return uint(caseID)
}

// firstNonEmptyConflictParameter mirrors services.firstNonEmptyContextValue:
// trim, then take the first candidate that is neither blank nor a zero
// placeholder. Selection happens before parsing, so an unparsable selected
// value never retries later candidates.
func firstNonEmptyConflictParameter(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && trimmed != "<nil>" && trimmed != "0" {
			return trimmed
		}
	}
	return ""
}

// inboxReviewerCanReadCase inlines the review endpoint's case read rule
// (services.CanReadConflictContext) inside the repositories package: the
// technical admin is denied, an ethical wall requires the explicit
// whitelist row, then a conflict-review/management role or case ownership
// passes. Query failures deny (fail closed).
func inboxReviewerCanReadCase(db *gorm.DB, userID uint, role string, caseRow *models.Case) bool {
	if caseRow == nil || userID == 0 {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(role), "admin") || strings.EqualFold(strings.TrimSpace(role), "super_admin") {
		return false
	}
	if db == nil {
		return false
	}
	if caseRow.EthicalWallEnabled {
		var granted int64
		if err := db.Model(&models.CaseEthicalWallWhitelist{}).
			Where("case_id = ? AND user_id = ?", caseRow.ID, userID).
			Count(&granted).Error; err != nil || granted == 0 {
			return false
		}
	}
	if inboxConflictReviewRoleAllowed(role) {
		return true
	}
	return caseRow.LawyerID == userID || caseRow.CreatedBy == strconv.FormatUint(uint64(userID), 10)
}

// conflictSubjectBinding reads the immutable intake association and any
// subject case id from frozen conflict search parameters.
func conflictSubjectBinding(record models.ConflictCheckRecord) (string, uint) {
	if len(record.SearchParameters) == 0 {
		return "", 0
	}
	values := struct {
		IntakeID      string `json:"intakeId"`
		IntakeIDSnake string `json:"intake_id"`
		CaseID        string `json:"subjectCaseId"`
		CaseIDSnake   string `json:"subject_case_id"`
	}{}
	raw, err := json.Marshal(record.SearchParameters)
	if err != nil || json.Unmarshal(raw, &values) != nil {
		return "", 0
	}
	intakeID := strings.TrimSpace(values.IntakeID)
	if intakeID == "" {
		intakeID = strings.TrimSpace(values.IntakeIDSnake)
	}
	caseIDText := strings.TrimSpace(values.CaseID)
	if caseIDText == "" {
		caseIDText = strings.TrimSpace(values.CaseIDSnake)
	}
	caseID, _ := strconv.ParseUint(caseIDText, 10, 32)
	return intakeID, uint(caseID)
}

// MarkAsRead 标记为已读
func (r *InboxRepositoryImpl) MarkAsRead(ctx context.Context, id uint) error {
	return r.markAsRead(ctx, r.db.WithContext(ctx).Where("id = ?", id), id)
}

func (r *InboxRepositoryImpl) MarkAsReadByUserID(ctx context.Context, id, userID uint) error {
	return r.markAsRead(ctx, r.ownedItemQuery(ctx, id, userID), id)
}

func (r *InboxRepositoryImpl) markAsRead(ctx context.Context, query *gorm.DB, id uint) error {
	now := time.Now()
	result := query.
		Model(&models.InboxItem{}).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return NewRepositoryErrorWithID("mark_as_read", "inbox_item", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("mark_as_read", "inbox_item", id, ErrInboxItemNotFound)
	}

	return nil
}

// MarkAsCompleted 标记为已完成
func (r *InboxRepositoryImpl) MarkAsCompleted(ctx context.Context, id uint) error {
	return r.markAsCompleted(ctx, r.db.WithContext(ctx).Where("id = ?", id), id)
}

func (r *InboxRepositoryImpl) MarkAsCompletedByUserID(ctx context.Context, id, userID uint) error {
	return r.markAsCompleted(ctx, r.ownedItemQuery(ctx, id, userID), id)
}

func (r *InboxRepositoryImpl) markAsCompleted(ctx context.Context, query *gorm.DB, id uint) error {
	now := time.Now()
	result := query.
		Model(&models.InboxItem{}).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": now,
		})

	if result.Error != nil {
		return NewRepositoryErrorWithID("mark_as_completed", "inbox_item", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("mark_as_completed", "inbox_item", id, ErrInboxItemNotFound)
	}

	return nil
}

func (r *InboxRepositoryImpl) Snooze(ctx context.Context, id uint, until time.Time) error {
	return r.snooze(ctx, r.db.WithContext(ctx).Where("id = ?", id), id, until)
}

// SnoozeByUserID 仅延后指定用户的待办：只推迟提醒可见性，不改期限、
// 状态或责任人（§5.4）。
func (r *InboxRepositoryImpl) SnoozeByUserID(ctx context.Context, id, userID uint, until time.Time) error {
	return r.snooze(ctx, r.ownedItemQuery(ctx, id, userID), id, until)
}

func (r *InboxRepositoryImpl) snooze(ctx context.Context, query *gorm.DB, id uint, until time.Time) error {
	result := query.
		Model(&models.InboxItem{}).
		Updates(map[string]interface{}{
			"snoozed_until": until,
			"snoozed_count": gorm.Expr("snoozed_count + 1"),
		})

	if result.Error != nil {
		return NewRepositoryErrorWithID("snooze", "inbox_item", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("snooze", "inbox_item", id, ErrInboxItemNotFound)
	}

	return nil
}

func (r *InboxRepositoryImpl) ownedItemQuery(ctx context.Context, id, userID uint) *gorm.DB {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID)
}

// Escalate 升级待办事项
func (r *InboxRepositoryImpl) Escalate(ctx context.Context, id uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.InboxItem{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"escalated":    true,
			"escalated_at": now,
		})

	if result.Error != nil {
		return NewRepositoryErrorWithID("escalate", "inbox_item", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("escalate", "inbox_item", id, ErrInboxItemNotFound)
	}

	return nil
}

// GetUnreadCount 获取未读数量
func (r *InboxRepositoryImpl) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.InboxItem{}).
		Where("user_id = ?", userID).
		Where("is_read = ?", false).
		Where("(snoozed_until IS NULL OR snoozed_until <= NOW())").
		Count(&count).Error

	if err != nil {
		return 0, NewRepositoryError("get_unread_count", "inbox_item", err)
	}

	return count, nil
}

// GetDueItems 获取到期待办事项
func (r *InboxRepositoryImpl) GetDueItems(ctx context.Context, before time.Time) ([]*models.InboxItem, error) {
	var items []*models.InboxItem
	err := r.db.WithContext(ctx).
		Where("is_completed = ?", false).
		Where("due_date <= ?", before).
		Where("reminder_sent = ?", false).
		Where("(snoozed_until IS NULL OR snoozed_until <= NOW())").
		Find(&items).Error

	if err != nil {
		return nil, NewRepositoryError("get_due_items", "inbox_item", err)
	}

	return items, nil
}

// GetOverdueCriticalItems 获取超时的critical待办
func (r *InboxRepositoryImpl) GetOverdueCriticalItems(ctx context.Context, before time.Time) ([]*models.InboxItem, error) {
	var items []*models.InboxItem
	err := r.db.WithContext(ctx).
		Where("is_completed = ?", false).
		Where("priority = ?", "critical").
		Where("due_date < ?", before).
		Where("escalated = ?", false).
		Where("(snoozed_until IS NULL OR snoozed_until <= NOW())").
		Find(&items).Error

	if err != nil {
		return nil, NewRepositoryError("get_overdue_critical_items", "inbox_item", err)
	}

	return items, nil
}

// CreateReminderRule 创建提醒规则
func (r *InboxRepositoryImpl) CreateReminderRule(ctx context.Context, rule *models.InboxReminderRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetReminderRules 获取提醒规则
func (r *InboxRepositoryImpl) GetReminderRules(ctx context.Context, isActive bool) ([]*models.InboxReminderRule, error) {
	var rules []*models.InboxReminderRule
	query := r.db.WithContext(ctx)
	if isActive {
		query = query.Where("is_active = ?", true)
	}
	err := query.Find(&rules).Error
	if err != nil {
		return nil, NewRepositoryError("get_reminder_rules", "inbox_reminder_rule", err)
	}
	return rules, nil
}

// GetReminderRuleByTypeAndPriority 根据日期类型和优先级获取规则
func (r *InboxRepositoryImpl) GetReminderRuleByTypeAndPriority(ctx context.Context, dueDateType, priority string) (*models.InboxReminderRule, error) {
	var rule models.InboxReminderRule
	err := r.db.WithContext(ctx).
		Where("due_date_type = ?", dueDateType).
		Where("priority = ?", priority).
		Where("is_active = ?", true).
		First(&rule).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, NewRepositoryError("get_reminder_rule", "inbox_reminder_rule", err)
	}

	return &rule, nil
}

// UpdateReminderRule 更新提醒规则
func (r *InboxRepositoryImpl) UpdateReminderRule(ctx context.Context, rule *models.InboxReminderRule) error {
	result := r.db.WithContext(ctx).
		Model(&models.InboxReminderRule{}).
		Where("id = ?", rule.ID).
		Updates(rule)

	if result.Error != nil {
		return NewRepositoryErrorWithID("update_reminder_rule", "inbox_reminder_rule", rule.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update_reminder_rule", "inbox_reminder_rule", rule.ID, ErrReminderRuleNotFound)
	}

	return nil
}

// DeleteReminderRule 删除提醒规则
func (r *InboxRepositoryImpl) DeleteReminderRule(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Delete(&models.InboxReminderRule{}, id)

	if result.Error != nil {
		return NewRepositoryErrorWithID("delete_reminder_rule", "inbox_reminder_rule", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("delete_reminder_rule", "inbox_reminder_rule", id, ErrReminderRuleNotFound)
	}

	return nil
}

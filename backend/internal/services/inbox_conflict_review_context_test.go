package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// Use only the existing isolated SQLite fixture. Register NOW on its sole
// connection so these tests exercise the real inbox repository search/count.
func conflictReviewInboxDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := reviewerTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	conn, err := sqlDB.Conn(context.Background())
	require.NoError(t, err)
	require.NoError(t, conn.Raw(func(driverConn interface{}) error {
		return driverConn.(*sqlite3.SQLiteConn).RegisterFunc("NOW", func() string {
			return time.Now().UTC().Format("2006-01-02 15:04:05")
		}, false)
	}))
	require.NoError(t, conn.Close())
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&models.Client{}, &models.CaseEthicalWallWhitelist{}))
	require.NoError(t, db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, created_by INTEGER, client_id INTEGER)").Error)
	seedReviewerCheck(t, db)
	return db
}

func appointInboxReviewer(t *testing.T, db *gorm.DB, checkID, caseName string) models.InboxItem {
	t.Helper()
	intakeID := "intake-" + checkID
	require.NoError(t, db.Exec("INSERT INTO case_intakes (id, created_by, client_id) VALUES (?, 10, 1)", intakeID).Error)
	parameters := models.JSON{"intakeId": intakeID}
	if checkID == "check-reviewer" {
		require.NoError(t, db.Model(&models.ConflictCheckRecord{}).Where("check_id = ?", checkID).
			Updates(map[string]interface{}{"case_name": caseName, "client_id": "1", "search_parameters": parameters}).Error)
	} else {
		require.NoError(t, db.Create(&models.ConflictCheckRecord{
			CheckID: checkID, CaseName: caseName, ClientID: "1", UserID: 10,
			CheckStatus: "COMPLETED", SearchParameters: parameters,
		}).Error)
	}
	_, err := AssignConflictReviewer(context.Background(), db, AuthActor{UserID: 20, Role: "conflict_officer"}, checkID,
		ConflictReviewerAssignmentInput{ReviewerID: 20, RecusalDeclared: true})
	require.NoError(t, err)
	var item models.InboxItem
	require.NoError(t, db.Where("workflow_task_key = ?", "conflict_review:"+checkID+":20").Take(&item).Error)
	return item
}

// The inbox title parser must never drift from the formal review entry
// point: firstNonEmpty selects ONE candidate (camelCase, snake_case,
// caseId, case_id; blanks and zero are skipped) and only that selected
// value is parsed. An invalid selected value stays unbound - the parser
// must not silently fall through to a later candidate and authorize a
// different case than the review endpoint would resolve.
func TestSubjectCaseIDFromConflictParametersCandidateSemantics(t *testing.T) {
	cases := []struct {
		name       string
		parameters models.JSON
		want       uint
	}{
		{"camelcase_wins", models.JSON{"subjectCaseId": "101", "case_id": "102"}, 101},
		{"snake_case_fallback", models.JSON{"subject_case_id": "103"}, 103},
		{"case_id_fallback", models.JSON{"case_id": "105"}, 105},
		{"blank_and_zero_candidates_skipped", models.JSON{"subjectCaseId": "", "subject_case_id": "0", "caseId": "  ", "case_id": "106"}, 106},
		{"selected_value_is_trimmed", models.JSON{"subjectCaseId": " 108 "}, 108},
		{"invalid_selected_value_does_not_retry_later_candidates", models.JSON{"subjectCaseId": "abc", "caseId": "107"}, 0},
		{"missing_binding_stays_zero", models.JSON{"intakeId": "intake-x"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, subjectCaseIDFromConflictParameters(tc.parameters))
		})
	}
}

func TestConflictReviewInbox_AssignmentExposesCaseContext(t *testing.T) {
	db := conflictReviewInboxDB(t)
	caseName := "新版浏览器练习甲（虚构）"
	item := appointInboxReviewer(t, db, "check-reviewer", caseName)
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	resp, err := svc.GetInboxItemByID(context.Background(), item.ID, 20)
	require.NoError(t, err)
	require.NotNil(t, resp.WorkflowAction)
	require.Equal(t, "check-reviewer", resp.WorkflowAction.ResourceID, "the review target is the CCT, not an approval UUID")
	require.Equal(t, "review_conflict", resp.WorkflowAction.ActionCode)
	require.Equal(t, caseName, resp.WorkflowAction.RelatedTitle)
	require.Contains(t, resp.WorkflowAction.AssignedTo, "独立核查人")
	require.NotEmpty(t, resp.WorkflowAction.NextStep)
	var stored models.InboxItem
	require.NoError(t, db.First(&stored, item.ID).Error)
	require.Equal(t, "利益冲突独立复核待处理", stored.Title, "display hydration must not rewrite the audit-retained todo")
}

func TestConflictReviewInbox_SearchFindsGenericAssociatedTodoBeforePagination(t *testing.T) {
	db := conflictReviewInboxDB(t)
	// Only synthetic rows in the test's temporary database: never the QA inbox.
	for i := 0; i < 48; i++ {
		require.NoError(t, db.Create(&models.InboxItem{
			UserID: 20, SourceType: "approval", SourceRef: fmt.Sprintf("other-%d", i),
			Title: "利益冲突独立复核待处理", Content: "其他演练事项", CompletionMode: "manual", Priority: "high",
		}).Error)
	}
	caseName := "新版浏览器练习甲（虚构）"
	item := appointInboxReviewer(t, db, "check-reviewer", caseName)
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	for _, query := range []string{caseName, "新版浏览器练习甲"} {
		resp, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{
			Page: 1, PageSize: 1, Search: query, OrderBy: "created_at",
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), resp.Pagination.Total, "case-name search must participate in the count before pagination")
		require.Len(t, resp.Items, 1)
		require.Equal(t, item.ID, resp.Items[0].ID)
		require.Equal(t, caseName, resp.Items[0].WorkflowAction.RelatedTitle)
	}
	var stored models.InboxItem
	require.NoError(t, db.First(&stored, item.ID).Error)
	require.Equal(t, item.Title, stored.Title)
	require.Equal(t, item.Content, stored.Content)
	require.Equal(t, item.SourceRef, stored.SourceRef)
}

func TestConflictReviewInbox_SearchPreservesFiltersAndExistingTextSearch(t *testing.T) {
	db := conflictReviewInboxDB(t)
	first := appointInboxReviewer(t, db, "check-reviewer", "分页练习甲（虚构）")
	second := appointInboxReviewer(t, db, "check-second", "分页练习乙（虚构）")
	require.NoError(t, db.Model(&models.InboxItem{}).Where("id = ?", first.ID).Update("is_read", true).Error)
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	resp, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{
		Page: 2, PageSize: 1, Search: "分页练习", OrderBy: "created_at",
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), resp.Pagination.Total)
	require.Equal(t, int64(2), resp.Pagination.TotalPage)
	require.Len(t, resp.Items, 1)
	require.Equal(t, first.ID, resp.Items[0].ID)
	unread := false
	resp, err = svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{
		Page: 1, PageSize: 20, Search: "分页练习", IsRead: &unread, SourceType: second.SourceType,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Pagination.Total)
	require.Len(t, resp.Items, 1)
	require.Equal(t, second.ID, resp.Items[0].ID)
	resp, err = svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{Search: "独立复核待处理"})
	require.NoError(t, err)
	require.Equal(t, int64(2), resp.Pagination.Total, "existing title/content search is retained")
	resp, err = svc.ListInboxItems(context.Background(), 10, &ListInboxItemsRequest{Search: "分页练习"})
	require.NoError(t, err)
	require.Zero(t, resp.Pagination.Total, "case-name search cannot cross inbox ownership")
}

func TestConflictReviewInbox_UnassignedCheckDoesNotCreateOrFindTodo(t *testing.T) {
	db := conflictReviewInboxDB(t)
	caseName := "尚未指定核查人（虚构）"
	require.NoError(t, db.Model(&models.ConflictCheckRecord{}).Where("check_id = ?", "check-reviewer").Update("case_name", caseName).Error)
	var count int64
	require.NoError(t, db.Model(&models.InboxItem{}).Count(&count).Error)
	require.Zero(t, count)
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	resp, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{Search: caseName})
	require.NoError(t, err)
	require.Empty(t, resp.Items)
	require.Zero(t, resp.Pagination.Total)
}

func TestConflictReviewInbox_UnmatchedSearchCannotRideAuthorizedIDs(t *testing.T) {
	db := conflictReviewInboxDB(t)
	caseName := "随机关键词过滤练习甲（虚构）"
	item := appointInboxReviewer(t, db, "check-reviewer", caseName)
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	// Any search term that does not match the authorized case name must
	// return zero, even though this user owns an authorized conflict todo.
	for _, query := range []string{"随机关键词完全不同词", "zzz-no-such-case"} {
		resp, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{
			Page: 1, PageSize: 20, Search: query, OrderBy: "created_at",
		})
		require.NoError(t, err)
		require.Empty(t, resp.Items, "unmatched search cannot surface authorized todos via ID ride-along")
		require.Zero(t, resp.Pagination.Total)
	}
	// The authorized todos only participate when the SQL LIKE on the
	// frozen case name itself matches; the same database collation rules
	// as title/content apply, including the LIKE wildcard suffix match.
	resp, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{
		Page: 1, PageSize: 20, Search: "练习甲", OrderBy: "created_at",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Pagination.Total)
	require.Len(t, resp.Items, 1)
	require.Equal(t, item.ID, resp.Items[0].ID)
}

func TestConflictReviewInbox_DoesNotLeakUnavailableOrRestrictedCaseTitle(t *testing.T) {
	for _, scenario := range []string{"wall", "wrong-owner", "wrong-client", "missing-intake", "revoked", "reassigned", "expired", "admin", "inactive", "wrong-key"} {
		t.Run(scenario, func(t *testing.T) {
			db := conflictReviewInboxDB(t)
			caseName := "受限练习标题（虚构）"
			item := appointInboxReviewer(t, db, "check-reviewer", caseName)
			switch scenario {
			case "wall":
				require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).
					Updates(map[string]interface{}{"client_id": 1, "ethical_wall_enabled": true}).Error)
			case "wrong-owner":
				require.NoError(t, db.Table("case_intakes").Where("id = ?", "intake-check-reviewer").Update("created_by", 11).Error)
			case "wrong-client":
				require.NoError(t, db.Table("case_intakes").Where("id = ?", "intake-check-reviewer").Update("client_id", 2).Error)
			case "missing-intake":
				require.NoError(t, db.Table("case_intakes").Where("id = ?", "intake-check-reviewer").Delete(nil).Error)
			case "revoked":
				require.NoError(t, db.Model(&models.ConflictReviewerAssignment{}).Where("check_id = ?", "check-reviewer").Update("status", models.ConflictReviewerAssignmentRevoked).Error)
			case "reassigned":
				require.NoError(t, db.Model(&models.ConflictReviewerAssignment{}).Where("check_id = ?", "check-reviewer").Update("reviewer_id", 21).Error)
			case "expired":
				require.NoError(t, db.Model(&models.ConflictReviewerAssignment{}).Where("check_id = ?", "check-reviewer").Update("effective_to", time.Now().Add(-time.Hour)).Error)
			case "admin", "inactive":
				column, value := "role", "admin"
				if scenario == "inactive" {
					column, value = "status", "inactive"
				}
				require.NoError(t, db.Model(&models.User{}).Where("id = ?", 20).Update(column, value).Error)
			case "wrong-key":
				require.NoError(t, db.Model(&models.InboxItem{}).Where("id = ?", item.ID).Update("workflow_task_key", "conflict_review:another-check:20").Error)
			}
			svc := NewInboxService(repositories.NewInboxRepository(db), nil)
			resp, err := svc.GetInboxItemByID(context.Background(), item.ID, 20)
			require.NoError(t, err)
			require.NotNil(t, resp.WorkflowAction)
			require.Empty(t, resp.WorkflowAction.RelatedTitle, "a stale/unauthorized binding cannot disclose a title")
			list, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{Search: caseName})
			require.NoError(t, err)
			require.Empty(t, list.Items)
			require.Zero(t, list.Pagination.Total, "even a title-search count must not leak restricted context")
		})
	}
}

// A formal case reference is an object binding, not a display fact. When the
// frozen check names an explicit subject case (or the intake later produced
// exactly one bound case), title authorization must use that same case and
// the same isolation rules as the real review endpoint; a visible case on
// the same client can never ride along.
func TestConflictReviewInbox_TitleFollowsExplicitCaseBindingNotClientRideAlong(t *testing.T) {
	t.Run("explicit_subject_case_hidden_behind_wall", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		caseName := "显式绑定隔离练习（虚构）"
		item := appointInboxReviewer(t, db, "check-reviewer", caseName)
		require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).Updates(map[string]interface{}{
			"client_id": 1, "conflict_check_id": "check-reviewer", "ethical_wall_enabled": true,
		}).Error)
		require.NoError(t, db.Model(&models.ConflictCheckRecord{}).Where("check_id = ?", "check-reviewer").
			Update("search_parameters", models.JSON{"intakeId": "intake-check-reviewer", "subjectCaseId": "100"}).Error)
		require.NoError(t, db.Create(&models.Case{ID: 101, CaseNumber: "VIS-B", Title: "同客户可见案件乙（虚构）",
			ClientID: 1, LawyerID: 10, CreatedBy: "10", SubjectState: models.SubjectStateEffective}).Error)
		svc := NewInboxService(repositories.NewInboxRepository(db), nil)
		resp, err := svc.GetInboxItemByID(context.Background(), item.ID, 20)
		require.NoError(t, err)
		require.NotNil(t, resp.WorkflowAction)
		require.Empty(t, resp.WorkflowAction.RelatedTitle, "the explicit case wall denies the title")
		list, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{Search: caseName})
		require.NoError(t, err)
		require.Zero(t, list.Pagination.Total, "title search must obey the same explicit binding")
	})
	t.Run("reverse_binding_hidden_behind_wall", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		caseName := "反向绑定隔离练习（虚构）"
		item := appointInboxReviewer(t, db, "check-reviewer", caseName)
		require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).Updates(map[string]interface{}{
			"client_id": 1, "conflict_check_id": "check-reviewer", "ethical_wall_enabled": true,
		}).Error)
		svc := NewInboxService(repositories.NewInboxRepository(db), nil)
		resp, err := svc.GetInboxItemByID(context.Background(), item.ID, 20)
		require.NoError(t, err)
		require.NotNil(t, resp.WorkflowAction)
		require.Empty(t, resp.WorkflowAction.RelatedTitle, "the reverse-bound case wall denies the title")
		list, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{Search: caseName})
		require.NoError(t, err)
		require.Zero(t, list.Pagination.Total)
	})
	t.Run("reverse_binding_with_visible_sibling_denied", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		caseName := "反向绑定混合练习（虚构）"
		item := appointInboxReviewer(t, db, "check-reviewer", caseName)
		// Exactly the audited leak shape: the check binds back to a walled
		// case through cases.conflict_check_id, while a second visible case
		// shares the client. The title must stay hidden.
		require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).Updates(map[string]interface{}{
			"client_id": 1, "conflict_check_id": "check-reviewer", "ethical_wall_enabled": true,
		}).Error)
		require.NoError(t, db.Create(&models.Case{ID: 101, CaseNumber: "VIS-B", Title: "同客户可见案件乙（虚构）",
			ClientID: 1, LawyerID: 10, CreatedBy: "10", SubjectState: models.SubjectStateEffective}).Error)
		svc := NewInboxService(repositories.NewInboxRepository(db), nil)
		resp, err := svc.GetInboxItemByID(context.Background(), item.ID, 20)
		require.NoError(t, err)
		require.NotNil(t, resp.WorkflowAction)
		require.Empty(t, resp.WorkflowAction.RelatedTitle, "a visible sibling case cannot unlock a walled bound case")
		list, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{Search: caseName})
		require.NoError(t, err)
		require.Zero(t, list.Pagination.Total)
	})
	t.Run("mixed_client_visibility_authorized_title_kept", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		caseName := "混合可见练习（虚构）"
		item := appointInboxReviewer(t, db, "check-reviewer", caseName)
		// The check has no explicit case binding; the client's only case is
		// fully visible, so the intake-branch authorization keeps the title.
		require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).
			Update("client_id", 1).Error)
		svc := NewInboxService(repositories.NewInboxRepository(db), nil)
		resp, err := svc.GetInboxItemByID(context.Background(), item.ID, 20)
		require.NoError(t, err)
		require.NotNil(t, resp.WorkflowAction)
		require.Equal(t, caseName, resp.WorkflowAction.RelatedTitle,
			"a visible client context must keep the frozen title")
		list, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{Search: caseName})
		require.NoError(t, err)
		require.Equal(t, int64(1), list.Pagination.Total)
		require.Equal(t, item.ID, list.Items[0].ID)
	})
	t.Run("reverse_binding_with_visible_case_grants_title", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		caseName := "反向绑定可见练习（虚构）"
		item := appointInboxReviewer(t, db, "check-reviewer", caseName)
		// The reverse binding is the REAL cases.conflict_check_id column,
		// not an implied client match: the frozen record carries no explicit
		// case reference, so the title must be authorized through the
		// exactly-one live reverse-bound case row. No wall: the reverse-bound
		// case is readable by the reviewer role, so the title stays
		// authorized through the case branch.
		require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).
			Updates(map[string]interface{}{"client_id": 1, "conflict_check_id": "check-reviewer"}).Error)
		resp, err := NewInboxService(repositories.NewInboxRepository(db), nil).
			GetInboxItemByID(context.Background(), item.ID, 20)
		require.NoError(t, err)
		require.NotNil(t, resp.WorkflowAction)
		require.Equal(t, caseName, resp.WorkflowAction.RelatedTitle)
	})

	t.Run("invalid_selected_binding_uses_reverse_fallback_not_next_candidate", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		caseName := "无效主绑定练习（虚构）"
		item := appointInboxReviewer(t, db, "check-reviewer", caseName)
		// The formal entry point (subjectCaseIDFromConflictParameters)
		// selects ONE candidate via firstNonEmpty and parses only that
		// value: subjectCaseId="abc" stays unbound, so the exactly-one
		// cases.conflict_check_id reverse binding to the WALLED case 100
		// governs and the visible sibling 101 on the same client must
		// never unlock the title. A parse-failure retry that fell through
		// to caseId="101" would authorize a case the real endpoint denies.
		require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).Updates(map[string]interface{}{
			"client_id": 1, "conflict_check_id": "check-reviewer", "ethical_wall_enabled": true,
		}).Error)
		require.NoError(t, db.Create(&models.Case{ID: 101, CaseNumber: "VIS-B", Title: "同客户可见案件乙（虚构）",
			ClientID: 1, LawyerID: 10, CreatedBy: "10", SubjectState: models.SubjectStateEffective}).Error)
		require.NoError(t, db.Model(&models.ConflictCheckRecord{}).Where("check_id = ?", "check-reviewer").
			Update("search_parameters", models.JSON{"intakeId": "intake-check-reviewer", "subjectCaseId": "abc", "caseId": "101"}).Error)
		svc := NewInboxService(repositories.NewInboxRepository(db), nil)
		resp, err := svc.GetInboxItemByID(context.Background(), item.ID, 20)
		require.NoError(t, err)
		require.NotNil(t, resp.WorkflowAction)
		require.Empty(t, resp.WorkflowAction.RelatedTitle, "the invalid selected binding must fall back like the endpoint, not to caseId=101")
		list, err := svc.ListInboxItems(context.Background(), 20, &ListInboxItemsRequest{Search: caseName})
		require.NoError(t, err)
		require.Zero(t, list.Pagination.Total, "title search must obey the same reverse-bound wall")
	})
}

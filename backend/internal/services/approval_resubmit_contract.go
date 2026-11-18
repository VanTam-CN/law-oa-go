package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"law-oa-go/internal/models"
)

// N3 / C09+C11: atomic structured resubmit with server-owned submission
// snapshots (plan 6.5). The legacy ResubmitApproval keeps its signature;
// intake approvals entering the structured contract must provide the
// expected intake revision, the return record they answer, and per-item
// replies. Every accepted resubmit persists a server-generated
// intake_submission snapshot and flips the active pointer inside the same
// transaction that revives the approval and rebinds the intake slot.

// SupplementReply is the applicant answer to one structured return item.
type SupplementReply struct {
	ItemID      string                   `json:"item_id"`
	Response    string                   `json:"response"`
	Attachments []map[string]interface{} `json:"attachments,omitempty"`
}

// ResubmitApprovalRequest carries the structured resubmit contract.
type ResubmitApprovalRequest struct {
	ExpectedIntakeRevision       string            `json:"expected_intake_revision"`
	ExpectedSubmissionSnapshotID string            `json:"expected_submission_snapshot_id"`
	ReturnRecordID               string            `json:"return_record_id"`
	RevisionNote                 string            `json:"revision_note"`
	Replies                      []SupplementReply `json:"replies"`
}

// ResubmitContractError is a typed 4xx rejection with a plan error code.
type ResubmitContractError struct {
	Code    string
	Message string
}

func (e *ResubmitContractError) Error() string {
	if e == nil {
		return "重提请求被拒绝"
	}
	return e.Message
}

func resubmitStale(message string) *ResubmitContractError {
	return &ResubmitContractError{Code: models.ErrCodeApprovalRoundStale, Message: message}
}

// snapshotUnresolved marks a structured request whose expected snapshot is
// missing or no longer active (plan 6.3: never guess from the newest row).
func snapshotUnresolved(message string) *ResubmitContractError {
	return &ResubmitContractError{Code: models.ErrCodeApprovalSnapshotUnresolved, Message: message}
}

// N3/C11 replay contract: a replayed structured resubmit (same return
// record anchor and the same normalized request) returns the original
// round result without creating a new snapshot, notification, or todo; a
// different-content replay of the same round is a 409.
type ResubmitReplayedError struct {
	SnapshotID    string
	SourceVersion int
}

func (e *ResubmitReplayedError) Error() string {
	if e == nil {
		return "重提请求已处理"
	}
	return fmt.Sprintf("重提请求已处理（第 %d 轮快照 %s），返回原轮次结果", e.SourceVersion, e.SnapshotID)
}

// resubmitSnapshotPayload is the slice of the intake_submission snapshot
// the replay detection reads back.
type resubmitSnapshotPayload struct {
	RequestDigest  string `json:"resubmit_request_digest"`
	ReturnRecordID string `json:"return_record_id"`
	IntakeRevision string `json:"intake_revision"`
}

type normalizedReplyKey struct {
	ItemID      string `json:"item_id"`
	Response    string `json:"response"`
	Attachments string `json:"attachments,omitempty"`
}

// resubmitRequestDigest normalizes a structured resubmit request for
// replay detection: trimmed text fields, replies sorted by item_id (order
// insensitive), JSON encoded, SHA-256.
func resubmitRequestDigest(req *ResubmitApprovalRequest) (string, error) {
	replies := make([]normalizedReplyKey, 0, len(req.Replies))
	for _, r := range req.Replies {
		attachments := ""
		if len(r.Attachments) > 0 {
			if data, err := json.Marshal(r.Attachments); err == nil {
				attachments = string(data)
			}
		}
		replies = append(replies, normalizedReplyKey{
			ItemID:      strings.TrimSpace(r.ItemID),
			Response:    strings.TrimSpace(r.Response),
			Attachments: attachments,
		})
	}
	sort.Slice(replies, func(i, j int) bool { return replies[i].ItemID < replies[j].ItemID })
	payload := struct {
		ExpectedIntakeRevision       string               `json:"expected_intake_revision"`
		ExpectedSubmissionSnapshotID string               `json:"expected_submission_snapshot_id"`
		ReturnRecordID               string               `json:"return_record_id"`
		RevisionNote                 string               `json:"revision_note"`
		Replies                      []normalizedReplyKey `json:"replies"`
	}{
		ExpectedIntakeRevision:       strings.TrimSpace(req.ExpectedIntakeRevision),
		ExpectedSubmissionSnapshotID: strings.TrimSpace(req.ExpectedSubmissionSnapshotID),
		ReturnRecordID:               strings.TrimSpace(req.ReturnRecordID),
		RevisionNote:                 strings.TrimSpace(req.RevisionNote),
		Replies:                      replies,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// detectResubmitReplayTx compares the request digest against the latest
// committed round under the intake lock. A matching digest anchored to the
// same return record and intake revision replays; different content falls
// through to the round unique index as the same-round 409 backstop.
func detectResubmitReplayTx(tx *gorm.DB, approvalID, digest, intakeRevision, returnRecordID string) error {
	var prior struct {
		SnapshotID    string
		SnapshotData  string
		SourceVersion int
	}
	err := tx.Table("approval_snapshots").
		Where("approval_request_id = ? AND snapshot_type = ?", approvalID, models.ApprovalSnapshotTypeIntakeSubmission).
		Order("source_version DESC").Limit(1).Take(&prior).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("读取最近提交快照失败：%w", err)
	}
	var p resubmitSnapshotPayload
	if json.Unmarshal([]byte(prior.SnapshotData), &p) != nil || p.RequestDigest == "" {
		return nil
	}
	if p.RequestDigest == digest && p.ReturnRecordID == returnRecordID && p.IntakeRevision == intakeRevision {
		return &ResubmitReplayedError{SnapshotID: prior.SnapshotID, SourceVersion: prior.SourceVersion}
	}
	return nil
}

// structuredResubmitRequested reports whether the caller entered the
// structured contract. Legacy callers (revision note only) keep the old path.
func (r *ResubmitApprovalRequest) structuredResubmitRequested() bool {
	if r == nil {
		return false
	}
	return strings.TrimSpace(r.ExpectedIntakeRevision) != "" ||
		strings.TrimSpace(r.ExpectedSubmissionSnapshotID) != "" ||
		strings.TrimSpace(r.ReturnRecordID) != "" ||
		len(r.Replies) > 0
}

// approvalIntakeScoped reports whether an approval is bound to an intake
// and therefore must be decided, resubmitted, and case-created strictly
// through its active submission snapshot. Malformed metadata fails closed.
func approvalIntakeScoped(approval *models.ApprovalRequest) bool {
	if approval == nil {
		return false
	}
	_, intakeID, err := approvalIntakeContext(approval)
	return err == nil && strings.TrimSpace(intakeID) != ""
}

// approvalIntakeContext parses approval metadata and reports the bound
// intake id when present.
func approvalIntakeContext(approval *models.ApprovalRequest) (map[string]interface{}, string, error) {
	metadata := map[string]interface{}{}
	if strings.TrimSpace(approval.Metadata) != "" {
		if err := json.Unmarshal([]byte(approval.Metadata), &metadata); err != nil {
			return nil, "", fmt.Errorf("解析审批元数据失败：%w", err)
		}
	}
	intakeID, _ := metadata["intake_id"].(string)
	return metadata, strings.TrimSpace(intakeID), nil
}

// currentIntakeRevision mirrors the handler-side derivation (draft_revision
// in intake metadata with legacy-<id> fallback) so both read paths agree.
// The intake metadata lives on case_intakes.metadata, not approval metadata,
// so this reads the locked row inside the transaction.
func lockedIntakeRevision(tx *gorm.DB, intakeID string) (string, error) {
	var row struct {
		Metadata []byte
	}
	if err := tx.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&row).Error; err != nil {
		return "", fmt.Errorf("读取接案版本失败：%w", err)
	}
	metadata := map[string]interface{}{}
	if len(row.Metadata) > 0 {
		if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
			return "", fmt.Errorf("解析接案元数据失败：%w", err)
		}
	}
	rev, _ := metadata["draft_revision"].(string)
	rev = strings.TrimSpace(rev)
	if rev != "" {
		return rev, nil
	}
	return "legacy-" + strings.TrimSpace(intakeID), nil
}

// verifySupplementReturnRecord ensures the referenced record exists, belongs
// to this approval, is a request_changes decision, and is still the latest
// return round.
func (s *ApprovalService) verifySupplementReturnRecord(approvalID, returnRecordID string) (*models.ApprovalRecord, error) {
	records, err := s.approvalRepo.FindRecordsByApprovalID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("读取审批记录失败：%w", err)
	}
	var referenced *models.ApprovalRecord
	latest := time.Time{}
	for i := range records {
		rec := &records[i]
		if rec.Decision != models.ApprovalDecisionRequestChanges {
			continue
		}
		if rec.ApprovalDate.After(latest) {
			latest = rec.ApprovalDate
		}
		if rec.ID == returnRecordID {
			referenced = rec
		}
	}
	if referenced == nil {
		return nil, resubmitStale("退回记录不存在或不属于该审批")
	}
	if referenced.ApprovalDate.Before(latest) {
		return nil, resubmitStale("已有更新的退回记录，原退回依据已过期")
	}
	return referenced, nil
}

// verifySupplementReturnRecordTx is the transaction-scoped variant used by
// the draft-save contract: callers already hold the approval row lock, so
// re-reading the records here guarantees the referenced round is still the
// latest request_changes round at write time.
func verifySupplementReturnRecordTx(tx *gorm.DB, approvalID, returnRecordID string) (*models.ApprovalRecord, error) {
	var records []models.ApprovalRecord
	if err := tx.Where("approval_request_id = ?", approvalID).
		Order("created_at ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("读取审批记录失败：%w", err)
	}
	var referenced *models.ApprovalRecord
	latest := time.Time{}
	for i := range records {
		rec := &records[i]
		if rec.Decision != models.ApprovalDecisionRequestChanges {
			continue
		}
		if rec.ApprovalDate.After(latest) {
			latest = rec.ApprovalDate
		}
		if rec.ID == returnRecordID {
			referenced = rec
		}
	}
	if referenced == nil {
		return nil, resubmitStale("退回记录不存在或不属于该审批")
	}
	if referenced.ApprovalDate.Before(latest) {
		return nil, resubmitStale("已有更新的退回记录，原退回依据已过期")
	}
	return referenced, nil
}

// validateSupplementReplies ensures each reply points at a real item from
// the referenced return record and passes the plain-text reply rule.
func validateSupplementReplies(record *models.ApprovalRecord, replies []SupplementReply) error {
	if record == nil {
		return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "退回记录缺失，不能保存补充回复"}
	}
	if len(replies) == 0 {
		return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "补充回复不能为空"}
	}
	items, err := parseStoredSupplementItems(record.ImposedRequirements)
	if err != nil {
		return err
	}
	valid := map[string]bool{}
	required := map[string]bool{}
	for _, it := range items {
		valid[it.ItemID] = true
		if it.Required {
			required[it.ItemID] = true
		}
	}
	answered := map[string]bool{}
	for _, reply := range replies {
		if strings.TrimSpace(reply.ItemID) == "" {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "补充回复缺少 item_id"}
		}
		if !valid[reply.ItemID] {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "补充回复指向不存在的补充项：" + reply.ItemID}
		}
		if strings.TrimSpace(reply.Response) == "" {
			continue
		}
		if len([]rune(reply.Response)) > supplementReplyDraftMaxLen {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "单条补充回复不能超过 2000 字"}
		}
		if dangerousMarkup.MatchString(reply.Response) {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "补充回复不能包含网页标记或外链"}
		}
		answered[reply.ItemID] = true
	}
	for itemID := range required {
		if !answered[itemID] {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "必答补充项尚未回复：" + itemID}
		}
	}
	return nil
}

// validateSupplementReplyDraft accepts partially answered drafts. It is the
// save-path contract: item ids must stay inside the frozen whitelist, and any
// non-empty answer must still obey the plain-text rule, but empty answers are
// allowed so the lawyer can fill items one at a time without losing the rest.
// The final resubmit path keeps the stricter validateSupplementReplies gate.
const supplementReplyDraftMaxLen = 2000

func validateSupplementReplyDraft(record *models.ApprovalRecord, replies []SupplementReply) error {
	if record == nil {
		return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "退回记录缺失，不能保存补充回复"}
	}
	items, err := parseStoredSupplementItems(record.ImposedRequirements)
	if err != nil {
		return err
	}
	valid := map[string]bool{}
	for _, it := range items {
		valid[it.ItemID] = true
	}
	seen := map[string]bool{}
	for _, reply := range replies {
		if strings.TrimSpace(reply.ItemID) == "" {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "补充回复缺少 item_id"}
		}
		if seen[reply.ItemID] {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "补充回复重复指向同一补充项：" + reply.ItemID}
		}
		seen[reply.ItemID] = true
		if !valid[reply.ItemID] {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "补充回复指向不存在的补充项：" + reply.ItemID}
		}
		if strings.TrimSpace(reply.Response) == "" {
			continue
		}
		if len([]rune(reply.Response)) > supplementReplyDraftMaxLen {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "单条补充回复不能超过 2000 字"}
		}
		if dangerousMarkup.MatchString(reply.Response) {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "补充回复不能包含网页标记或外链"}
		}
	}
	return nil
}

// parseStoredSupplementItems decodes the frozen item whitelist stored on a
// request_changes record. Malformed or empty payloads fail closed with the
// shared supplement error code so the revision UI cannot invent items.
func parseStoredSupplementItems(raw string) ([]SupplementItem, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "退回记录缺少结构化清单"}
	}
	var stored struct {
		Items []SupplementItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(trimmed), &stored); err != nil {
		return nil, &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "退回记录缺少结构化清单"}
	}
	if len(stored.Items) == 0 {
		return nil, &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "退回记录缺少结构化清单"}
	}
	return stored.Items, nil
}

// SaveSupplementReplies persists a draft of the per-item replies so a
// refresh or re-login can restore them (R4b-fix). It never rotates the
// intake revision and never writes intake facts: the intake row is read
// under its row lock only to reject stale drafts, mirroring the resubmit
// contract. The saved copy stays a draft; the authoritative copy remains
// the resubmit request body, which re-runs the same whitelist validation.
func (s *ApprovalService) SaveSupplementReplies(userID string, approvalID string, req *ResubmitApprovalRequest) (*models.ApprovalRecord, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败：%w", err)
	}
	if approval == nil {
		return nil, ErrApprovalRecordNotFound
	}
	if approval.ApplicantID != userID {
		return nil, ErrApprovalForbidden
	}
	// Lock-free fast path: reject obviously closed rounds without opening a
	// transaction. The authoritative status gate re-reads the row inside the
	// transaction below.
	if approval.Status != models.ApprovalStatusNeedsRevision {
		return nil, resubmitStale("当前审批状态不允许保存补充回复草稿")
	}
	var returnRecord *models.ApprovalRecord
	err = s.approvalRepo.Transaction(func(tx *gorm.DB) error {
		// Serialize against a concurrent resubmit on the approval row itself:
		// this conditional UPDATE matches zero rows once the round has left
		// needs_revision, and while it matches it holds the row lock until
		// commit, so a racing resubmit cannot flip the status between this
		// check and the draft write below.
		lockRes := tx.Model(&models.ApprovalRequest{}).
			Where("id = ? AND status = ?", approval.ID, models.ApprovalStatusNeedsRevision).
			Update("updated_at", time.Now().UTC())
		if lockRes.Error != nil {
			return fmt.Errorf("锁定审批记录失败：%w", lockRes.Error)
		}
		if lockRes.RowsAffected == 0 {
			return resubmitStale("当前审批状态不允许保存草稿")
		}
		locked, recErr := verifySupplementReturnRecordTx(tx, approval.ID, req.ReturnRecordID)
		if recErr != nil {
			return recErr
		}
		if err := validateSupplementReplyDraft(locked, req.Replies); err != nil {
			return err
		}
		if _, intakeID, ctxErr := approvalIntakeContext(approval); ctxErr == nil && intakeID != "" {
			if lockErr := lockIntakeRow(tx, intakeID); lockErr != nil {
				return fmt.Errorf("锁定接案记录失败：%w", lockErr)
			}
			currentRev, revErr := lockedIntakeRevision(tx, intakeID)
			if revErr != nil {
				return revErr
			}
			if expectedRev := strings.TrimSpace(req.ExpectedIntakeRevision); expectedRev == "" || expectedRev != currentRev {
				return resubmitStale("接案版本已变化，请刷新后重新填写补充回复")
			}
		}
		payload := map[string]interface{}{
			"return_record_id": req.ReturnRecordID,
			"replies":          req.Replies,
			"updated_at":       time.Now().UTC().Format(time.RFC3339),
			"updated_by":       userID,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("序列化补充回复失败：%w", err)
		}
		writeRes := tx.Model(&models.ApprovalRecord{}).
			Where("id = ? AND approval_request_id = ?", locked.ID, approval.ID).
			Update("supplement_replies", string(data))
		if writeRes.Error != nil {
			return fmt.Errorf("保存补充回复失败：%w", writeRes.Error)
		}
		if writeRes.RowsAffected == 0 {
			return resubmitStale("退回记录不存在或不属于该审批")
		}
		locked.SupplementReplies = ptrToJSONBString(data)
		returnRecord = locked
		return nil
	})
	if err != nil {
		return nil, err
	}
	return returnRecord, nil
}

// GetSupplementReplies returns the return items and the saved draft
// replies of the applicant for the referenced return round. Only the
// applicant may read them; approvers keep their record-history view.
func (s *ApprovalService) GetSupplementReplies(userID string, approvalID string, returnRecordID string) ([]SupplementItem, []SupplementReply, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, nil, fmt.Errorf("查找审批记录失败：%w", err)
	}
	if approval == nil {
		return nil, nil, ErrApprovalRecordNotFound
	}
	if approval.ApplicantID != userID {
		return nil, nil, ErrApprovalForbidden
	}
	returnRecord, err := s.verifySupplementReturnRecord(approval.ID, returnRecordID)
	if err != nil {
		return nil, nil, err
	}
	items, err := parseStoredSupplementItems(returnRecord.ImposedRequirements)
	if err != nil {
		return nil, nil, err
	}
	replies := []SupplementReply{}
	if returnRecord.SupplementReplies != nil && strings.TrimSpace(*returnRecord.SupplementReplies) != "" {
		var stored struct {
			ReturnRecordID string            `json:"return_record_id"`
			Replies        []SupplementReply `json:"replies"`
		}
		if err := json.Unmarshal([]byte(*returnRecord.SupplementReplies), &stored); err != nil {
			return nil, nil, &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "已存补充回复格式无效，请重新填写"}
		}
		if strings.TrimSpace(stored.ReturnRecordID) == strings.TrimSpace(returnRecordID) {
			replies = stored.Replies
		}
	}
	return items, replies, nil
}

// nextIntakeSubmissionRound computes round = max(existing)+1.
func nextIntakeSubmissionRound(tx *gorm.DB, approvalID string) (int, error) {
	var maxVer int
	if err := tx.Table("approval_snapshots").
		Where("approval_request_id = ? AND snapshot_type = ?", approvalID, models.ApprovalSnapshotTypeIntakeSubmission).
		Select("COALESCE(MAX(source_version), 0)").Scan(&maxVer).Error; err != nil {
		return 0, fmt.Errorf("读取提交轮次失败：%w", err)
	}
	return maxVer + 1, nil
}

// parseSearchParametersMap decodes the frozen search-parameters JSON stored
// on a conflict check record so gate helpers can reuse the models.JSON based
// extraction helper shared with the reviewer gate without a type cast.
func parseSearchParametersMap(raw string) models.JSON {
	var parsed models.JSON
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return models.JSON{}
	}
	return parsed
}

// persistIntakeSubmissionSnapshot writes the new snapshot row and returns
// its id. Caller must run inside the resubmit transaction.
func persistIntakeSubmissionSnapshot(tx *gorm.DB, approvalID string, roundNumber int, payload map[string]interface{}, snapshotID string) (string, error) {
	if snapshotID == "" {
		// The id is pre-minted by every caller so the metadata mirror, the
		// frozen config, and this row share one identifier. An empty id
		// would only surface as a driver-specific failure (empty-string
		// uuid insert fails on PG), so fail closed with a typed error here.
		return "", fmt.Errorf("提交快照缺少预生成 ID")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化提交快照失败：%w", err)
	}
	row := &models.ApprovalSnapshot{
		// Use a real UUID: approval_snapshots.id is a native uuid column on
		// PostgreSQL, so the legacy nanosecond-hex helper cannot be written.
		ID:                snapshotID,
		ApprovalRequestID: approvalID,
		SnapshotType:      models.ApprovalSnapshotTypeIntakeSubmission,
		SnapshotData:      string(data),
		SourceVersion:     roundNumber,
	}
	if err := tx.Create(row).Error; err != nil {
		if isUniqueViolationErr(err) {
			return "", resubmitStale("该轮次快照已存在，重提请求重复")
		}
		return "", fmt.Errorf("写入提交快照失败：%w", err)
	}
	return row.ID, nil
}

// newIntakeSnapshotID mints a snapshot id for a new submission round. The
// caller may pre-mint it to keep the metadata mirror, the frozen config, and
// the persisted row on one identifier inside one transaction.
func newIntakeSnapshotID() string {
	return uuid.NewString()
}

// isUniqueViolationErr normalizes sqlite and postgres unique violations.
func isUniqueViolationErr(err error) bool {
	if err == nil {
		return false
	}
	if asUniqueConstraintError(err) != nil {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "duplicate key value violates unique constraint")
}

// activeSnapshotIDTx reads the approval current active snapshot pointer.
func activeSnapshotIDTx(tx *gorm.DB, approvalID string) (string, error) {
	var row struct {
		ActiveSubmissionSnapshotID *string
	}
	if err := tx.Table("approval_requests").Select("active_submission_snapshot_id").
		Where("id = ?", approvalID).Take(&row).Error; err != nil {
		return "", fmt.Errorf("读取当前快照指针失败：%w", err)
	}
	if row.ActiveSubmissionSnapshotID == nil {
		return "", nil
	}
	return strings.TrimSpace(*row.ActiveSubmissionSnapshotID), nil
}

// lockIntakeRow takes the row lock inside the resubmit transaction so two
// racing resubmits serialize on the intake row.
func lockIntakeRow(tx *gorm.DB, intakeID string) error {
	if strings.TrimSpace(intakeID) == "" {
		return nil
	}
	var row struct{ ID string }
	return tx.Table("case_intakes").Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id").Where("id = ?", intakeID).Take(&row).Error
}

// intakeConflictBinding is the server-owned check/review binding the resubmit
// gate validates. The committed intake pointer (case_intakes.metadata.
// conflict_check_id, rotated by CommitIntakeConflictResult) decides whether
// the bound check still owns the current facts; the review must be the latest
// append-only conclusion. This mirrors validateConflictApprovalGate's
// evidence rules so a resubmit cannot outrun the case-creation gate.
type intakeConflictBinding struct {
	CheckID        string
	CheckStatus    string
	CoverageStatus string
	ReviewDecision string
	ReviewID       uint64
	ReviewerID     uint64
	// SnapshotID is the intake_submission snapshot this binding was frozen
	// into. Downstream readers must resolve it through the active pointer
	// instead of rebuilding facts from mutable rows.
	SnapshotID string
	// FrozenParameters carries the raw audit search_parameters of the bound
	// check record. Only the continuation reader consumes it, to re-prove
	// that the current protected subjects still match the identity anchors
	// frozen with the old check. Never serialized to any response.
	FrozenParameters string
}

// readIntakeConflictBindingTx resolves the check binding from the committed
// intake pointer alone (case_intakes.metadata.conflict_check_id, rotated by
// CommitIntakeConflictResult inside its commit transaction). Approval-side
// anchors (active approval_conflict_associations rows, approval metadata,
// conflict_check_id column) are NOT authoritative for resubmit: they lag
// behind a re-check until this resubmit's success-path sync refreshes them.
// Trusting the stale association here would both block legal resubmits and
// let an old check outrun re-detected facts. Client-supplied request fields
// are never trusted (plan 6.4: no client self-reported check IDs). The
// intake status column and parsed metadata are returned with the binding so
// the gate reads one consistent row snapshot.
func readIntakeConflictBindingTx(tx *gorm.DB, intakeID string) (*intakeConflictBinding, map[string]interface{}, string, error) {
	var intakeRow struct {
		Metadata []byte
		Status   string
	}
	if err := tx.Table("case_intakes").Select("metadata, status").
		Where("id = ?", intakeID).Take(&intakeRow).Error; err != nil {
		return nil, nil, "", fmt.Errorf("读取接案冲突绑定失败：%w", err)
	}
	metadata := map[string]interface{}{}
	if len(intakeRow.Metadata) > 0 {
		if err := json.Unmarshal(intakeRow.Metadata, &metadata); err != nil {
			return nil, nil, "", fmt.Errorf("解析接案元数据失败：%w", err)
		}
	}
	checkID := strings.TrimSpace(metadataString(metadata, "conflict_check_id"))
	return &intakeConflictBinding{CheckID: checkID}, metadata, strings.TrimSpace(intakeRow.Status), nil
}

// syncApprovalConflictBindingTx aligns the final-approval gate's authority
// set with the check the resubmit gate just verified.
// validateConflictApprovalGate reads the active approval_conflict_associations
// row first and falls back to approval metadata; both lag behind a re-check
// until this sync, so a proven resubmit supersedes stale active rows, inserts
// the new active association, and refreshes the metadata plus conflict_check_id
// column anchor — all in the same transaction. A failed gate never reaches
// this function, so failed-gate bindings, snapshots, and associations stay
// untouched.
func syncApprovalConflictBindingTx(tx *gorm.DB, approval *models.ApprovalRequest, binding *intakeConflictBinding) error {
	if tx.Migrator().HasTable("approval_conflict_associations") {
		if err := tx.Table("approval_conflict_associations").
			Where("approval_request_id = ? AND association_status = ?", approval.ID, "active").
			Update("association_status", "superseded").Error; err != nil {
			return fmt.Errorf("归档旧冲突关联失败：%w", err)
		}
		createdBy := strings.TrimSpace(approval.UpdatedBy)
		if createdBy == "" {
			createdBy = strings.TrimSpace(approval.CreatedBy)
		}
		if createdBy == "" {
			createdBy = strings.TrimSpace(approval.ApplicantID)
		}
		// Raw insert with explicit columns: nullable optional columns
		// (risk_level, association metadata) must stay NULL/DEFAULT because
		// the migration CHECK constraints reject empty strings and the
		// production gate only reads identity fields from this row.
		if err := tx.Exec(
			"INSERT INTO approval_conflict_associations (id, approval_request_id, conflict_check_id, association_status, created_by, created_at, updated_at) "+
				"VALUES (?, ?, ?, 'active', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			uuid.NewString(), approval.ID, binding.CheckID, createdBy).Error; err != nil {
			return fmt.Errorf("写入新冲突关联失败：%w", err)
		}
	}
	metadata, _, err := approvalIntakeContext(approval)
	if err != nil {
		return err
	}
	metadata["conflict_check_id"] = binding.CheckID
	if binding.SnapshotID != "" {
		// Mirror the active pointer into metadata and rewrite the frozen
		// case-creation config to source from this snapshot. Case creation
		// later reads only these keys, so a stale config or forged pointer
		// in approval metadata can never outrun the certified round.
		metadata["active_snapshot_id"] = binding.SnapshotID
		metadata["case_creation_config"] = map[string]interface{}{
			"source_snapshot_id": binding.SnapshotID,
			"source":             "intake_submission_snapshot",
		}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("序列化审批元数据失败：%w", err)
	}
	if err := tx.Model(approval).Update("metadata", string(raw)).Error; err != nil {
		return fmt.Errorf("同步审批冲突绑定失败：%w", err)
	}
	approval.Metadata = string(raw)
	// Keep the in-memory column anchor in sync: applyResubmitRevival calls
	// tx.Save(approval), which must not write the stale check id back.
	approval.ConflictCheckID = binding.CheckID
	return nil
}

// verifyIntakeConflictBindingTx implements the plan 6.4 resubmit gate: facts
// that changed require a new formal check plus an independent review before
// the old result can back a new submission. Read the binding server-side and
// fail closed when the check is missing, superseded by a newer commit (the
// intake metadata pointer moved), not completed, coverage unconfirmed, or the
// independent review has not reached an acceptable conclusion. Reuse is only
// allowed when the server can prove the binding is still valid; replies alone
// never satisfy it.
func (s *ApprovalService) verifyIntakeConflictBindingTx(tx *gorm.DB, approval *models.ApprovalRequest, intakeID string) (*intakeConflictBinding, error) {
	binding, metadata, intakeStatus, err := readIntakeConflictBindingTx(tx, intakeID)
	if err != nil {
		return nil, err
	}
	if binding.CheckID == "" {
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "审批缺少可验证的冲突检测绑定，请重新确认事实并完成利益冲突检测后重提",
		}
	}
	// R03: the pointer alone cannot prove the check still owns the current
	// facts. UpdateCaseIntake rolls a confirmed intake back to draft and
	// stamps conflict_invalidated_at on any facts change; ConfirmIntakeFacts
	// rotates lawyer_facts_version; only CommitIntakeConflictResult moves
	// the status back to conflict_ready, clears the invalidation stamp, and
	// records conflict_check_facts_version inside the commit transaction.
	// The resubmit gate therefore re-derives the whole chain server-side.
	if intakeStatus != "conflict_ready" {
		return nil, resubmitStale("接案事实已变化但尚未完成新的正式冲突检测，请重新确认事实并重新检测后重提")
	}
	if invalidatedAt, ok := metadata["conflict_invalidated_at"].(string); ok && strings.TrimSpace(invalidatedAt) != "" {
		return nil, resubmitStale("接案事实变更后尚未完成新的正式冲突检测，旧检测结论不能继续使用，请重新检测后重提")
	}
	committedFactsVersion, _ := metadata["conflict_check_facts_version"].(string)
	committedFactsVersion = strings.TrimSpace(committedFactsVersion)
	currentFactsVersion, _ := metadata["lawyer_facts_version"].(string)
	currentFactsVersion = strings.TrimSpace(currentFactsVersion)
	if committedFactsVersion == "" || currentFactsVersion == "" || committedFactsVersion != currentFactsVersion {
		return nil, resubmitStale("冲突检测绑定的事实版本与接案当前事实不一致，请重新确认事实并重新检测后重提")
	}
	var record struct {
		CheckStatus      string
		CheckResult      string
		SearchParameters string
	}
	if err := tx.Table("conflict_check_records").Select("check_status, check_result, search_parameters").
		Where("check_id = ?", binding.CheckID).Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ResubmitContractError{
				Code:    models.ErrCodeApprovalSnapshotUnresolved,
				Message: "绑定的冲突检测记录不存在，已阻止重提",
			}
		}
		return nil, fmt.Errorf("读取冲突检测记录失败：%w", err)
	}
	// The immutable audit record must name this intake: client-supplied
	// check IDs never enter the gate, and a record built for another intake
	// cannot back these facts even if a pointer points at it. Reuse the
	// frozen search-parameters helper shared with the reviewer gate.
	recordedIntakeID := conflictRecordIntakeID(parseSearchParametersMap(record.SearchParameters))
	if recordedIntakeID == "" || recordedIntakeID != intakeID {
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "冲突检测记录未绑定当前接案，无法证明检测证据归属，已阻止重提",
		}
	}
	binding.CheckStatus = strings.TrimSpace(record.CheckStatus)
	if !strings.EqualFold(binding.CheckStatus, "COMPLETED") {
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "绑定的冲突检测尚未完成，已阻止重提",
		}
	}
	binding.CoverageStatus = conflictCoverageStatusFromRecord(record.CheckResult, record.SearchParameters)
	if !strings.EqualFold(binding.CoverageStatus, "COMPLETE") {
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "冲突检测覆盖范围未确认完整，已阻止重提",
		}
	}
	// Read through the model, not raw SQL: ConflictReview hooks make the
	// review ledger append-only even if a future code path tried to rewrite
	// a conclusion.
	var review models.ConflictReview
	if err := tx.Where("check_id = ?", binding.CheckID).
		Order("created_at DESC").Take(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ResubmitContractError{
				Code:    models.ErrCodeApprovalSnapshotUnresolved,
				Message: "冲突检测尚未完成独立人工复核，不能重新提交",
			}
		}
		return nil, fmt.Errorf("读取冲突复核结论失败：%w", err)
	}
	binding.ReviewDecision = strings.ToLower(strings.TrimSpace(review.Decision))
	binding.ReviewID = uint64(review.ID)
	binding.ReviewerID = uint64(review.ReviewerID)
	switch binding.ReviewDecision {
	case "no_conflict", "false_positive":
		// Proven binding: the committed check still owns the facts and an
		// independent review accepted it.
		return binding, nil
	case "insufficient_information":
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "冲突复核结论为信息不足，请补充材料并重新复核后再重提",
		}
	case "confirmed_conflict":
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "独立复核已确认存在利益冲突，不能重新提交",
		}
	case "waiver_requested":
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "冲突复核已申请豁免，取得有效豁免批准前不能重新提交",
		}
	default:
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "冲突复核结论尚未达到可重提状态，不能重新提交",
		}
	}
}

// conflictCoverageStatusFromRecord mirrors conflictRecordCoverageComplete:
// coverage may live in the frozen response decision or in the original
// search parameters. Empty result means coverage was never confirmed.
func conflictCoverageStatusFromRecord(checkResult, searchParameters string) string {
	extract := func(raw string) string {
		var parsed map[string]interface{}
		if json.Unmarshal([]byte(raw), &parsed) != nil {
			return ""
		}
		if decision, ok := parsed["decision"].(map[string]interface{}); ok {
			if status, ok := decision["coverageStatus"].(string); ok && strings.TrimSpace(status) != "" {
				return strings.TrimSpace(status)
			}
			if status, ok := decision["coverage_status"].(string); ok && strings.TrimSpace(status) != "" {
				return strings.TrimSpace(status)
			}
		}
		return ""
	}
	if status := extract(checkResult); status != "" {
		return status
	}
	var params map[string]interface{}
	if json.Unmarshal([]byte(searchParameters), &params) != nil {
		return ""
	}
	for key, value := range params {
		if strings.EqualFold(strings.TrimSpace(key), "coverageStatus") || strings.EqualFold(strings.TrimSpace(key), "coverage_status") {
			return strings.TrimSpace(fmt.Sprint(value))
		}
	}
	return ""
}

// verifyResubmitStateTx re-reads the approval status inside the resubmit
// transaction. Racing resubmits serialize on the intake row lock; the second
// one must observe the winner's "resubmitted" status here and fail with a
// typed stale rejection instead of overwriting the winner's state.
func (s *ApprovalService) verifyResubmitStateTx(tx *gorm.DB, approval *models.ApprovalRequest) error {
	var row struct {
		Status string
	}
	if err := tx.Table("approval_requests").Select("status").
		Where("id = ?", approval.ID).Take(&row).Error; err != nil {
		return fmt.Errorf("读取审批状态失败：%w", err)
	}
	if !s.stateMachine.CanResubmit(row.Status) {
		return resubmitStale("该审批已被其他重提处理，请刷新后查看最新状态")
	}
	return nil
}

// verifyResubmitReplayTx implements the N3/C11 replay contract under the
// intake lock: a replayed structured request (same normalized content,
// same return-record anchor, same intake revision) resolves to the original
// round instead of creating a second revival, while different content is
// left to the later round-unique backstop as a 409. It must run before the
// resubmitted-state check because a replay arrives after the first attempt
// already succeeded. It returns the locked intake revision and the
// normalized request digest for reuse by the surrounding transaction.
func (s *ApprovalService) verifyResubmitReplayTx(tx *gorm.DB, approval *models.ApprovalRequest, req *ResubmitApprovalRequest) (string, string, error) {
	_, intakeID, err := approvalIntakeContext(approval)
	if err != nil || intakeID == "" {
		return "", "", err
	}
	currentRev, err := lockedIntakeRevision(tx, intakeID)
	if err != nil {
		return "", "", err
	}
	digest, err := resubmitRequestDigest(req)
	if err != nil {
		return "", "", fmt.Errorf("计算重提请求摘要失败：%w", err)
	}
	return currentRev, digest, detectResubmitReplayTx(tx, approval.ID, digest, currentRev, strings.TrimSpace(req.ReturnRecordID))
}

// structuredResubmitTx runs the C09/C11 contract inside one transaction.
func (s *ApprovalService) structuredResubmitTx(userID string, approval *models.ApprovalRequest, req *ResubmitApprovalRequest) error {
	return s.approvalRepo.Transaction(func(tx *gorm.DB) error {
		_, intakeID, err := approvalIntakeContext(approval)
		if err != nil {
			return err
		}
		if intakeID == "" {
			return snapshotUnresolved("审批未绑定接案记录，不能进入结构化重提")
		}
		if err := lockIntakeRow(tx, intakeID); err != nil {
			return fmt.Errorf("锁定接案记录失败：%w", err)
		}
		currentRev, digest, err := s.verifyResubmitReplayTx(tx, approval, req)
		if err != nil {
			return err
		}
		if err := s.verifyResubmitStateTx(tx, approval); err != nil {
			return err
		}
		returnRecord, err := s.verifySupplementReturnRecord(approval.ID, req.ReturnRecordID)
		if err != nil {
			return err
		}
		expectedRev := strings.TrimSpace(req.ExpectedIntakeRevision)
		if expectedRev == "" || expectedRev != currentRev {
			return resubmitStale("接案版本已变化，请刷新后重新提交")
		}
		activeSnapshotID, err := activeSnapshotIDTx(tx, approval.ID)
		if err != nil {
			return err
		}
		expectedSnap := strings.TrimSpace(req.ExpectedSubmissionSnapshotID)
		if expectedSnap != "" && activeSnapshotID != expectedSnap {
			return snapshotUnresolved("提交依据已更新，请刷新后基于当前快照重提")
		}
		if err := validateSupplementReplies(returnRecord, req.Replies); err != nil {
			return err
		}
		// Material return gate (review 2026-09-30, defect 7): every
		// material_request item must reference a material of THIS intake
		// that is genuinely uploaded (status=uploaded plus at least one
		// live attachment row), checked inside this transaction so a
		// concurrent upload cannot slip past the round's snapshot.
		if intakeID != "" {
			items, itemErr := parseStoredSupplementItems(returnRecord.ImposedRequirements)
			if itemErr == nil {
				if err := ValidateSupplementMaterialUploadsTx(tx, intakeID, items); err != nil {
					return err
				}
				// Correction evidence gate (review 2026-10-02): the legacy
				// status/live-count gate cannot distinguish a pre-return file
				// from a genuinely new current version. The frozen
				// material_fingerprints baseline of the returned round plus the
				// current storage_url pointer must prove NEW original-row
				// evidence; unknown baselines fail closed, never "unchanged".
				if err := enforceMaterialResubmitEvidenceTx(tx, approval.ID, items); err != nil {
					return err
				}
			}
		}
		// C09/plan 6.4 gate: the server, not the browser, decides whether
		// the current formal check plus independent review still bind these
		// facts. A replay of an already-accepted round never reaches this
		// point, and a failed gate leaves the whole transaction unopened
		// (no snapshot, no pointer, no revival).
		binding, err := s.verifyIntakeConflictBindingTx(tx, approval, intakeID)
		if err != nil {
			return err
		}
		if expectedSnap == "" {
			// Intake approvals are snapshot-scoped end to end: a resubmit
			// without the active snapshot id can no longer act on whichever
			// round happens to be newest (spec review 2026-09-27, HOLD item 1).
			// Checked after the binding gate so gate codes keep their
			// precedence; the rejection still precedes every write.
			return snapshotUnresolved("缺少当前提交快照标识，请刷新后基于当前快照重提")
		}
		roundNumber, err := nextIntakeSubmissionRound(tx, approval.ID)
		if err != nil {
			return err
		}
		// A round that follows a facts change must freeze the CURRENT
		// intake facts and parties itself (spec review 2026-09-27, HOLD
		// item 3b): a text-only slim round may chase facts back through
		// previous_snapshot_id, but the round that re-committed facts must
		// never let case creation resolve to the older party set. Legacy
		// thin test schemas without fact columns stay slim.
		frozen, ferr := s.freezeIntakeFactsForResubmitTx(tx, intakeID)
		if ferr != nil {
			return ferr
		}
		payload := map[string]interface{}{
			"approval_id":              approval.ID,
			"intake_id":                intakeID,
			"intake_revision":          currentRev,
			"round_number":             roundNumber,
			"previous_snapshot_id":     activeSnapshotID,
			"return_record_id":         req.ReturnRecordID,
			"replies":                  req.Replies,
			"revision_note":            req.RevisionNote,
			"submitted_by":             userID,
			"submitted_at":             time.Now().UTC().Format(time.RFC3339),
			"resubmit_request_digest":  digest,
			"conflict_check_id":        binding.CheckID,
			"conflict_review_state":    "independent_review_current",
			"conflict_review_id":       binding.ReviewID,
			"conflict_review_decision": binding.ReviewDecision,
			"conflict_reviewer_id":     binding.ReviewerID,
		}
		if frozen != nil {
			payload["structured_facts"] = frozen.Facts
			payload["client_id"] = frozen.ClientID
			payload["lawyer_facts_version"] = frozen.FactsVersion
		}
		// Freeze the material ledger state into the round snapshot: status
		// plus per-file digests. A later upload/delete cannot rewrite this
		// round, and tampering with stored bytes is detectable by re-hashing
		// against the fingerprint (contract 2026-09-30). Legacy test
		// schemas without the tables freeze an empty list.
		materialFingerprints, mfErr := freezeMaterialFingerprintsForSnapshot(tx, intakeID)
		if mfErr != nil {
			return mfErr
		}
		payload["material_fingerprints"] = materialFingerprints
		// Round diff (goal B, mixed return checklist): freeze the before/after
		// summary INTO this round's snapshot so approvers can audit exactly
		// what changed since the returned round. The returned-round snapshot
		// itself is never rewritten; the diff only lands in the new payload.
		if activeSnapshotID != "" {
			roundDiff, diffErr := buildIntakeRoundDiffTx(tx, activeSnapshotID, frozen, materialFingerprints)
			if diffErr != nil {
				return diffErr
			}
			payload["round_diff"] = roundDiff
		}
		// Pre-mint the round identifier so the final-gate sync, the frozen
		// case-creation config, and the snapshot row share one id inside this
		// transaction. The sync supersedes stale active associations and
		// mirrors metadata to this id before the row lands; row and pointer
		// updates follow in the same transaction.
		newSnapshotID := newIntakeSnapshotID()
		binding.SnapshotID = newSnapshotID
		// Final-gate sync (plan 6.5): a proven resubmit refreshes the active
		// conflict association plus metadata mirror and conflict_check_id
		// column anchor so the later approve step judges the new round.
		if err := syncApprovalConflictBindingTx(tx, approval, binding); err != nil {
			return err
		}
		payload["case_creation_config"] = map[string]interface{}{
			"source_snapshot_id": newSnapshotID,
			"source":             "intake_submission_snapshot",
		}
		snapshotID, err := persistIntakeSubmissionSnapshot(tx, approval.ID, roundNumber, payload, newSnapshotID)
		if err != nil {
			return err
		}
		if err := tx.Table("approval_requests").Where("id = ?", approval.ID).
			Update("active_submission_snapshot_id", snapshotID).Error; err != nil {
			return fmt.Errorf("更新提交快照指针失败：%w", err)
		}
		// Keep the in-memory struct in sync: applyResubmitRevival calls
		// tx.Save(approval), which would otherwise write the stale empty
		// pointer column back over the freshly activated snapshot.
		approval.ActiveSubmissionSnapshotID = &snapshotID
		return s.applyResubmitRevival(tx, approval, userID, req.RevisionNote)
	})
}

// buildIntakeRoundDiffTx compares the returned round's frozen facts and
// material fingerprints against the new round's values, producing a compact
// audit summary. Field diffs cover the client/title/type and conflict-relevant
// metadata keys; material diffs are anchored on material_id and distinguish
// genuinely new files from carried-over ones, so an approver can spot a
// replacement row masquerading as a re-upload. Identity material stays out:
// digests are compared as equal/changed booleans, never echoed.
func buildIntakeRoundDiffTx(tx *gorm.DB, activeSnapshotID string, frozen *frozenResubmitFacts, materialFingerprints []materialSnapshotFingerprint) (map[string]interface{}, error) {
	prev, err := loadRoundDiffBaselineSnapshotTx(tx, activeSnapshotID)
	if err != nil {
		return nil, err
	}
	fieldChanges := []map[string]interface{}{}
	baseline := map[string]interface{}{}
	if _, ok := prev["material_fingerprints"]; ok {
		baseline["material_baseline"] = "frozen"
	} else {
		baseline["material_baseline"] = "unknown"
	}
	if frozen != nil {
		factsBaseline, baselineErr := resolveRoundDiffFactsBaseline(prev)
		if baselineErr != nil {
			// A thin legacy round without frozen facts stays honestly
			// unknown: the diff must not reinvent a baseline from the live
			// intake row the returned round never froze.
			baseline["facts_baseline"] = "unknown"
		} else {
			baseline["facts_baseline"] = "frozen"
			fieldChanges = buildRoundDiffFieldChanges(factsBaseline, frozen)
		}
	}
	materialChanges := []map[string]interface{}{}
	prevMaterialsRaw, _ := prev["material_fingerprints"]
	prevMaterialsRawJSON, _ := json.Marshal(prevMaterialsRaw)
	var prevMaterials []materialSnapshotFingerprint
	if len(prevMaterialsRawJSON) > 0 {
		if err := json.Unmarshal(prevMaterialsRawJSON, &prevMaterials); err != nil {
			return nil, fmt.Errorf("解析上一轮材料指纹失败：%w", err)
		}
	}
	prevByMaterialID := map[uint]materialSnapshotFingerprint{}
	for _, item := range prevMaterials {
		prevByMaterialID[item.MaterialID] = item
	}
	for _, current := range materialFingerprints {
		change := map[string]interface{}{
			"material_id": current.MaterialID,
			"name":        current.Name,
			"material_type": current.MaterialType,
			"required":    current.Required,
			"status":      current.Status,
			"new_files":   []string{},
			"carried_over": []string{},
		}
		previous, existed := prevByMaterialID[current.MaterialID]
		if !existed {
			change["change_type"] = "new_material"
			names := []string{}
			for _, f := range current.Files {
				names = append(names, f.FileName)
			}
			change["new_files"] = names
		} else {
			prevFiles := map[uint]string{}
			for _, f := range previous.Files {
				prevFiles[f.FileID] = f.Digest
			}
			newFiles := []string{}
			carried := []string{}
			for _, f := range current.Files {
				if prevDigest, ok := prevFiles[f.FileID]; ok && prevDigest == f.Digest {
					carried = append(carried, f.FileName)
				} else {
					newFiles = append(newFiles, f.FileName)
				}
			}
			if current.Status != previous.Status {
				change["change_type"] = "status_and_files_changed"
				change["status_before"] = previous.Status
				change["status_after"] = current.Status
			} else if len(newFiles) > 0 {
				change["change_type"] = "files_changed"
			} else {
				change["change_type"] = "unchanged"
			}
			change["new_files"] = newFiles
			change["carried_over"] = carried
			change["file_count_before"] = len(previous.Files)
			change["file_count_after"] = len(current.Files)
			if previous.Name != current.Name {
				change["name_before"] = previous.Name
			}
			if previous.MaterialType != current.MaterialType {
				change["material_type_before"] = previous.MaterialType
			}
			if previous.Required != current.Required {
				change["required_before"] = previous.Required
			}
			if previous.CurrentFileID != current.CurrentFileID {
				change["current_file_id_before"] = previous.CurrentFileID
				change["current_file_id_after"] = current.CurrentFileID
			}
		}
		materialChanges = append(materialChanges, change)
	}
	// A material row deleted after the returned round can never appear in
	// the current list; listing only current rows would hide the deletion
	// behind a shorter list. Anchor on the frozen ids instead.
	for _, previous := range prevMaterials {
		if _, stillThere := currentMaterialIDs(materialFingerprints)[previous.MaterialID]; !stillThere {
			materialChanges = append(materialChanges, map[string]interface{}{
				"material_id":     previous.MaterialID,
				"name":            previous.Name,
				"material_type":   previous.MaterialType,
				"required":        previous.Required,
				"change_type":     "material_removed",
				"new_files":       []string{},
				"carried_over":    []string{},
				"file_count_before": len(previous.Files),
			})
		}
	}
	return map[string]interface{}{
		"round_diff_version": 2,
		"field_changes":      fieldChanges,
		"material_changes":   materialChanges,
		"baseline":           baseline,
	}, nil
}

// loadRoundDiffBaselineSnapshotTx loads the returned round itself (the
// active pointer passed by the caller). A material_baseline marks whether
// that round froze material fingerprints at all; a facts_baseline resolves
// only through legal previous_snapshot_id chains — never through a newest-row
// guess.
func loadRoundDiffBaselineSnapshotTx(tx *gorm.DB, activeSnapshotID string) (map[string]interface{}, error) {
	var prevRow struct {
		SnapshotData string
	}
	if err := tx.Table("approval_snapshots").
		Select("snapshot_data").
		Where("id = ? AND snapshot_type = ?", activeSnapshotID, models.ApprovalSnapshotTypeIntakeSubmission).
		Take(&prevRow).Error; err != nil {
		return nil, fmt.Errorf("读取上一轮提交快照失败：%w", err)
	}
	prev := map[string]interface{}{}
	if err := json.Unmarshal([]byte(prevRow.SnapshotData), &prev); err != nil {
		return nil, fmt.Errorf("解析上一轮提交快照失败：%w", err)
	}
	return prev, nil
}

// resolveRoundDiffFactsBaseline prefers the returned round's own frozen
// facts. Thin legacy payloads without facts resolve to unknown: the diff
// may not reinvent a baseline from a newest-row guess or the live intake.
func resolveRoundDiffFactsBaseline(prev map[string]interface{}) (*frozenResubmitFacts, error) {
	facts, _ := prev["structured_facts"].(map[string]interface{})
	if facts == nil {
		return nil, fmt.Errorf("上一轮快照未冻结事实")
	}
	clientID, _ := prev["client_id"].(string)
	version, _ := prev["lawyer_facts_version"].(string)
	return &frozenResubmitFacts{Facts: facts, ClientID: clientID, FactsVersion: version}, nil
}

// buildRoundDiffFieldChanges compares the frozen baseline facts against the
// new round fact by fact: scalars by value, parties by positional tuple
// (name/role/type/identity type), identity only as a changed boolean. The
// public summary never carries identity_digest or ciphertext material.
func buildRoundDiffFieldChanges(baseline *frozenResubmitFacts, current *frozenResubmitFacts) []map[string]interface{} {
	changes := []map[string]interface{}{}
	type fieldPair struct{ field, before, after string }
	pairs := []fieldPair{}
	appendPair := func(field string, beforeRaw, afterRaw interface{}) {
		before := strings.TrimSpace(fmt.Sprint(beforeRaw))
		after := strings.TrimSpace(fmt.Sprint(afterRaw))
		if before == "<nil>" {
			before = ""
		}
		if after == "<nil>" {
			after = ""
		}
		if before != after {
			pairs = append(pairs, fieldPair{field: field, before: before, after: after})
		}
	}
	appendPair("client_id", baseline.ClientID, current.ClientID)
	for _, key := range []string{"title", "case_type", "priority", "description"} {
		appendPair(key, baseline.Facts[key], current.Facts[key])
	}
	for _, key := range []string{"business_area", "responsible_lawyer_id"} {
		beforeValue, _ := baseline.Facts[key].(string)
		afterValue, _ := current.Facts[key].(string)
		appendPair(key, beforeValue, afterValue)
	}
	prevParties := roundDiffPartyRows(baseline.Facts["parties"])
	newParties := roundDiffPartyRows(current.Facts["parties"])
	for i := 0; i < len(prevParties) || i < len(newParties); i++ {
		prefix := fmt.Sprintf("party_%d", i+1)
		var before map[string]interface{}
		var after map[string]interface{}
		if i < len(prevParties) {
			before = prevParties[i]
		}
		if i < len(newParties) {
			after = newParties[i]
		}
		if before == nil {
			pairs = append(pairs, fieldPair{field: prefix + "_added", after: roundDiffPartyLabel(after)})
			continue
		}
		if after == nil {
			pairs = append(pairs, fieldPair{field: prefix + "_removed", before: roundDiffPartyLabel(before)})
			continue
		}
		appendPair(prefix+"_name", before["entity_name"], after["entity_name"])
		appendPair(prefix+"_role", before["party_role"], after["party_role"])
		appendPair(prefix+"_entity_type", before["entity_type"], after["entity_type"])
		appendPair(prefix+"_identity_type", before["identity_type"], after["identity_type"])
		beforeDigest := strings.TrimSpace(fmt.Sprint(before["identity_digest"]))
		afterDigest := strings.TrimSpace(fmt.Sprint(after["identity_digest"]))
		if beforeDigest != afterDigest {
			pairs = append(pairs, fieldPair{field: prefix + "_identity_changed", before: "true", after: "true"})
		}
	}
	for _, pair := range pairs {
		changes = append(changes, map[string]interface{}{
			"field":  pair.field,
			"before": pair.before,
			"after":  pair.after,
		})
	}
	return changes
}

// roundDiffPartyRows accepts both the typed []map freeze produced by
// freezeIntakeFactsForResubmitTx and the []interface{} shape that comes back
// from JSON round-trips of older snapshots.
func roundDiffPartyRows(raw interface{}) []map[string]interface{} {
	switch typed := raw.(type) {
	case []map[string]interface{}:
		return typed
	case []interface{}:
		rows := make([]map[string]interface{}, 0, len(typed))
		for _, item := range typed {
			row, _ := item.(map[string]interface{})
			if row != nil {
				rows = append(rows, row)
			}
		}
		return rows
	default:
		return nil
	}
}

func roundDiffPartyLabel(row map[string]interface{}) string {
	if row == nil {
		return ""
	}
	name := strings.TrimSpace(fmt.Sprint(row["entity_name"]))
	if name == "<nil>" {
		name = ""
	}
	return name
}

// roundDiffMetaString projects an intake metadata value into the same
// string shape used by scalar facts so both freeze paths compare equal
// when nothing changed.
func roundDiffMetaString(meta map[string]interface{}, key string) string {
	if meta == nil {
		return ""
	}
	value := strings.TrimSpace(fmt.Sprint(meta[key]))
	if value == "<nil>" {
		return ""
	}
	return value
}

func currentMaterialIDs(list []materialSnapshotFingerprint) map[uint]bool {
	ids := map[uint]bool{}
	for _, item := range list {
		ids[item.MaterialID] = true
	}
	return ids
}

// frozenResubmitFacts carries the current intake fact freeze for a
// fact-changing submission round.
type frozenResubmitFacts struct {
	Facts        map[string]interface{}
	ClientID     string
	FactsVersion string
}

// freezeIntakeFactsForResubmitTx reads the locked intake row's current
// facts, parties, and lawyer facts version inside the resubmit transaction.
// A nil result with nil error marks a legacy schema without fact columns:
// those worlds resolve frozen facts through the first-round snapshot chase.
func (s *ApprovalService) freezeIntakeFactsForResubmitTx(tx *gorm.DB, intakeID string) (*frozenResubmitFacts, error) {
	if !tx.Migrator().HasTable("case_intakes") || !tx.Migrator().HasColumn("case_intakes", "title") {
		return nil, nil
	}
	var row struct {
		Title       string
		CaseType    *string
		Priority    string
		Description *string
		ClientID    string
		Metadata    []byte
	}
	if err := tx.Table("case_intakes").
		Select("title, case_type, priority, description, COALESCE(CAST(client_id AS TEXT), '') AS client_id, metadata").
		Where("id = ?", intakeID).Take(&row).Error; err != nil {
		return nil, fmt.Errorf("冻结接案事实失败：%w", err)
	}
	intakeMeta := map[string]interface{}{}
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &intakeMeta)
	}
	factsVersion, _ := intakeMeta["lawyer_facts_version"].(string)
	facts := map[string]interface{}{
		"title":       row.Title,
		"case_type":   stringValueOrEmpty(row.CaseType),
		"priority":    row.Priority,
		"description": stringValueOrEmpty(row.Description),
	}
	facts["business_area"] = roundDiffMetaString(intakeMeta, "business_area")
	facts["responsible_lawyer_id"] = roundDiffMetaString(intakeMeta, "lawyer_id")
	parties := make([]map[string]interface{}, 0)
	if tx.Migrator().HasTable("case_intake_parties") {
		var partyRows []struct {
			EntityName     string
			EntityType     string
			PartyRole      string
			RelationDepth  int64
			IdentityType   *string
			IdentityDigest *string
			Aliases        *string
		}
		if err := tx.Table("case_intake_parties").
			Select("entity_name, entity_type, party_role, relation_depth, identity_type, identity_number_digest AS identity_digest, aliases").
			Where("intake_id = ?", intakeID).Order("id ASC").Find(&partyRows).Error; err != nil {
			return nil, fmt.Errorf("冻结接案当事方失败：%w", err)
		}
		for _, r := range partyRows {
			// Digest only: the protected identity secret never enters a
			// snapshot.
			parties = append(parties, map[string]interface{}{
				"entity_name":     r.EntityName,
				"entity_type":     r.EntityType,
				"party_role":      r.PartyRole,
				"relation_depth":  r.RelationDepth,
				"identity_type":   stringValueOrEmpty(r.IdentityType),
				"identity_digest": stringValueOrEmpty(r.IdentityDigest),
				"aliases":         stringValueOrEmpty(r.Aliases),
			})
		}
	}
	facts["parties"] = parties
	return &frozenResubmitFacts{
		Facts:        facts,
		ClientID:     strings.TrimSpace(row.ClientID),
		FactsVersion: strings.TrimSpace(factsVersion),
	}, nil
}

// applyResubmitRevival contains the existing atomic revival steps, shared
// with the legacy path.
func (s *ApprovalService) applyResubmitRevival(tx *gorm.DB, approval *models.ApprovalRequest, userID, revisionNote string) error {
	if err := s.terminalBinding.rebindAfterResubmitInTx(tx, approval); err != nil {
		if asUniqueConstraintError(err) != nil {
			return &ApprovalConflictGateError{
				Code:    "INTAKE_APPROVAL_ALREADY_IN_PROGRESS",
				Message: "该接案已有进行中的成案审批，不能重新提交；请在原审批单上继续处理",
			}
		}
		return err
	}
	if err := s.assigner.AssignApprover(approval); err != nil {
		return fmt.Errorf("重新分配审批人失败：%w", err)
	}
	approval.Status = models.ApprovalStatusResubmitted
	approval.UpdatedBy = userID
	if err := tx.Save(approval).Error; err != nil {
		return fmt.Errorf("重新提交失败：%w", err)
	}
	// N4/C05+C07: close the applicant's revision handoff and open the new
	// round's approver handoff inside the same revival transaction.
	if err := closeApprovalWorkflowItems(tx, approval.ID, inboxPurposeApplicantRevision); err != nil {
		return err
	}
	return s.openApprovalDecisionTask(tx, approval)
}

// StructuredResubmit is the public entry for the structured contract.
// Legacy callers (revision note only) fall through to ResubmitApproval.
func (s *ApprovalService) StructuredResubmit(userID string, approvalID string, req *ResubmitApprovalRequest) (*models.ApprovalRequest, error) {
	// Load first and scope the entry: an intake-bound approval can never
	// fall through to the legacy ResubmitApproval path (no snapshot, no
	// re-check gate, no atomic binding). Non-intake approvals keep the
	// legacy note-only entry for old callers (spec review 2026-09-27,
	// HOLD item 1).
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败：%w", err)
	}
	if approval == nil {
		return nil, ErrApprovalRecordNotFound
	}
	if req == nil || !req.structuredResubmitRequested() {
		if approvalIntakeScoped(approval) {
			return nil, snapshotUnresolved("接案审批必须基于当前提交快照进行结构化重提")
		}
		note := ""
		if req != nil {
			note = req.RevisionNote
		}
		return s.ResubmitApproval(userID, approvalID, note)
	}
	if approval.ApplicantID != userID {
		return nil, fmt.Errorf("无权重新提交此审批记录")
	}
	if !req.structuredResubmitRequested() {
		// Legacy callers keep the entry pre-check; structured callers are
		// judged inside the transaction so the C11 replay of an already
		// resubmitted round still resolves to the original result.
		if !s.stateMachine.CanResubmit(approval.Status) {
			return nil, fmt.Errorf("当前状态 %s 不允许重新提交", approval.Status)
		}
		if err := s.stateMachine.ValidateTransition(approval.Status, models.ApprovalStatusResubmitted); err != nil {
			return nil, err
		}
	}
	// Legacy text callers already wrote revision_note into metadata inside
	// ResubmitApproval; the structured path writes it into the snapshot and
	// metadata here before revival.
	metadata := map[string]interface{}{}
	if strings.TrimSpace(approval.Metadata) != "" {
		_ = json.Unmarshal([]byte(approval.Metadata), &metadata)
	}
	metadata["revision_note"] = req.RevisionNote
	metadata["resubmitted_at"] = time.Now().Format(time.RFC3339)
	if data, merr := json.Marshal(metadata); merr == nil {
		approval.Metadata = string(data)
	}
	if err := s.structuredResubmitTx(userID, approval, req); err != nil {
		var replayed *ResubmitReplayedError
		if errors.As(err, &replayed) {
			// Idempotent replay: return the original round result without
			// writing anything.
			return s.approvalRepo.FindByID(approvalID)
		}
		var gateErr *ApprovalConflictGateError
		if errors.As(err, &gateErr) {
			return nil, gateErr
		}
		return nil, err
	}
	_ = s.sendApprovalNotification(approval, "重新提交")
	return s.approvalRepo.FindByID(approvalID)
}

// ---------------------------------------------------------------------------
// C4 initial-submission binding: the first submit of an intake-scoped
// approval must persist the same server-owned authority set the resubmit
// path refreshes. The browser's conflict result is presentation data; the
// binding is derived from the committed intake pointer
// (case_intakes.metadata.conflict_check_id) and written — active
// approval_conflict_associations row, immutable intake_submission snapshot,
// active_submission_snapshot_id pointer, conflict_check_id column and
// metadata anchor — inside the caller's submit transaction. Any gate
// failure rolls the whole submission back, so an approval can never become
// visible to its approver without a provable binding.
// ---------------------------------------------------------------------------

func gateBindingError(code, message string) error {
	return &ApprovalConflictGateError{Code: code, Message: message}
}

// initialIntakeApprovalBinding runs inside the SubmitApproval commit
// transaction. Approvals whose metadata carries no intake_id keep the
// legacy behavior untouched; intake-scoped approvals must present a
// complete server-side chain (committed pointer, immutable completed check
// record owned by this intake, confirmed coverage, independent review) or
// the submission fails closed.
func (s *ApprovalService) initialIntakeApprovalBinding(tx *gorm.DB, approval *models.ApprovalRequest) error {
	metadata, intakeID, err := approvalIntakeContext(approval)
	if err != nil {
		return err
	}
	if intakeID == "" {
		return nil
	}
	// C4 second-round requirement: the intake facts the snapshot is about
	// must not move under this submission. Take the row lock BEFORE any
	// status/metadata/party read, serializing against fact edits and
	// re-check commits; a racing edit can only commit after this
	// transaction, and this submission then carries the exact facts it read.
	if err := lockIntakeRow(tx, intakeID); err != nil {
		return fmt.Errorf("锁定接案事实行失败：%w", err)
	}

	binding, intakeMeta, intakeStatus, err := readIntakeConflictBindingTx(tx, intakeID)
	if err != nil {
		return err
	}
	if binding.CheckID == "" {
		return gateBindingError("INTAKE_CONFLICT_LINK_MISMATCH", "接案记录缺少已提交的冲突检测绑定，已阻止提交")
	}
	// The integration handler's durable duplicate claim moves the intake
	// conflict_ready -> approval_submitting before CreateIntegratedApproval
	// runs, so exactly that claim state is accepted here; any other status
	// means the facts changed or another lifecycle owns the row.
	if intakeStatus != "conflict_ready" && intakeStatus != "approval_submitting" {
		return gateBindingError("INTAKE_CONFLICT_RESULT_STALE", "接案事实已变化或尚未完成冲突检测，已阻止提交")
	}
	if invalidatedAt, ok := intakeMeta["conflict_invalidated_at"].(string); ok && strings.TrimSpace(invalidatedAt) != "" {
		return gateBindingError("INTAKE_CONFLICT_RESULT_STALE", "接案事实变更后尚未完成新的正式冲突检测，已阻止提交")
	}
	committedFactsVersion, _ := intakeMeta["conflict_check_facts_version"].(string)
	committedFactsVersion = strings.TrimSpace(committedFactsVersion)
	currentFactsVersion, _ := intakeMeta["lawyer_facts_version"].(string)
	currentFactsVersion = strings.TrimSpace(currentFactsVersion)
	if committedFactsVersion == "" || currentFactsVersion == "" || committedFactsVersion != currentFactsVersion {
		return gateBindingError("INTAKE_CONFLICT_RESULT_STALE", "冲突检测绑定的事实版本与接案当前事实不一致，已阻止提交")
	}

	var record struct {
		CheckStatus      string
		CheckResult      string
		SearchParameters string
	}
	if err := tx.Table("conflict_check_records").Select("check_status, check_result, search_parameters").
		Where("check_id = ?", binding.CheckID).Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gateBindingError("CONFLICT_RESULT_NOT_COMPLETED", "绑定的冲突检测记录不存在，已阻止提交")
		}
		return fmt.Errorf("读取冲突检测记录失败：%w", err)
	}
	// The immutable audit record must name this intake: a pointer at a
	// check built for other facts cannot back this submission.
	recordedIntakeID := conflictRecordIntakeID(parseSearchParametersMap(record.SearchParameters))
	if recordedIntakeID == "" || recordedIntakeID != intakeID {
		return gateBindingError("CONFLICT_RESULT_NOT_OWNED", "冲突检测记录未绑定当前接案，已阻止提交")
	}
	binding.CheckStatus = strings.TrimSpace(record.CheckStatus)
	if !strings.EqualFold(binding.CheckStatus, "COMPLETED") {
		return gateBindingError("CONFLICT_RESULT_NOT_COMPLETED", "绑定的冲突检测尚未完成，已阻止提交")
	}
	binding.CoverageStatus = conflictCoverageStatusFromRecord(record.CheckResult, record.SearchParameters)
	if !strings.EqualFold(binding.CoverageStatus, "COMPLETE") {
		return gateBindingError("CONFLICT_COVERAGE_INCOMPLETE", "冲突检测覆盖范围未确认完整，已阻止提交")
	}

	var review models.ConflictReview
	if err := tx.Where("check_id = ?", binding.CheckID).
		Order("created_at DESC").Take(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gateBindingError("CONFLICT_REVIEW_REQUIRED", "冲突检测尚未完成独立人工复核，不能提交成案审批")
		}
		return fmt.Errorf("读取冲突复核结论失败：%w", err)
	}
	binding.ReviewDecision = strings.ToLower(strings.TrimSpace(review.Decision))
	switch binding.ReviewDecision {
	case "no_conflict", "false_positive":
		// Proven binding: the committed check still owns the facts and an
		// independent review accepted it.
	default:
		return gateBindingError("CONFLICT_REVIEW_DECISION_BLOCKED", "冲突复核结论不允许进入成案审批")
	}

	var existingRound int64
	if err := tx.Table("approval_snapshots").
		Where("approval_request_id = ? AND snapshot_type = ? AND source_version = ?", approval.ID, models.ApprovalSnapshotTypeIntakeSubmission, 1).
		Count(&existingRound).Error; err != nil {
		return fmt.Errorf("读取提交快照轮次失败：%w", err)
	}
	if existingRound > 0 {
		return gateBindingError("INTAKE_SUBMISSION_SNAPSHOT_EXISTS", "首轮提交快照已存在，不能重复绑定")
	}

	// Fact and party anchors come from the same committed intake row the
	// pointer lives on, so the snapshot can be replayed against later
	// revisions without trusting any client state.
	var intakeFacts struct {
		Title       string
		CaseType    *string
		Priority    string
		Description *string
		ClientID    string
		CreatedBy   string
	}
	if err := tx.Table("case_intakes").
		Select("title, case_type, priority, description, COALESCE(CAST(client_id AS TEXT), '') AS client_id, COALESCE(CAST(created_by AS TEXT), '') AS created_by").
		Where("id = ?", intakeID).Take(&intakeFacts).Error; err != nil {
		return fmt.Errorf("读取接案事实失败：%w", err)
	}
	structuredFacts := map[string]interface{}{
		"title":       intakeFacts.Title,
		"case_type":   stringValueOrEmpty(intakeFacts.CaseType),
		"priority":    intakeFacts.Priority,
		"description": stringValueOrEmpty(intakeFacts.Description),
	}
	structuredFacts["business_area"] = roundDiffMetaString(intakeMeta, "business_area")
	structuredFacts["responsible_lawyer_id"] = roundDiffMetaString(intakeMeta, "lawyer_id")
	partyRows := make([]map[string]interface{}, 0)
	if tx.Migrator().HasTable("case_intake_parties") {
		var rows []struct {
			EntityName     string
			EntityType     string
			PartyRole      string
			RelationDepth  int64
			IdentityType   *string
			IdentityDigest *string
			Aliases        *string
		}
		if err := tx.Table("case_intake_parties").
			Select("entity_name, entity_type, party_role, relation_depth, identity_type, identity_number_digest AS identity_digest, aliases").
			Where("intake_id = ?", intakeID).Order("id ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("读取接案当事方失败：%w", err)
		}
		for _, r := range rows {
			// No ciphertext in the snapshot: the digest identifies the
			// protected identity without carrying its secret.
			partyRows = append(partyRows, map[string]interface{}{
				"entity_name":     r.EntityName,
				"entity_type":     r.EntityType,
				"party_role":      r.PartyRole,
				"relation_depth":  r.RelationDepth,
				"identity_type":   stringValueOrEmpty(r.IdentityType),
				"identity_digest": stringValueOrEmpty(r.IdentityDigest),
				"aliases":         stringValueOrEmpty(r.Aliases),
			})
		}
	}
	structuredFacts["parties"] = partyRows
	rev, _ := intakeMeta["draft_revision"].(string)
	rev = strings.TrimSpace(rev)
	if rev == "" {
		rev = "legacy-" + intakeID
	}
	factsVersion, _ := intakeMeta["lawyer_facts_version"].(string)
	payload := map[string]interface{}{
		"approval_id":              approval.ID,
		"intake_id":                intakeID,
		"intake_revision":          rev,
		"intake_title":             intakeFacts.Title,
		"intake_status":            intakeStatus,
		"round_number":             1,
		"previous_snapshot_id":     "",
		"submitted_by":             approval.UpdatedBy,
		"submitted_at":             time.Now().UTC().Format(time.RFC3339),
		"conflict_check_id":        binding.CheckID,
		"conflict_check_status":    binding.CheckStatus,
		"conflict_coverage_status": binding.CoverageStatus,
		"conflict_review_id":       review.ID,
		"conflict_review_decision": binding.ReviewDecision,
		"conflict_reviewer_id":     review.ReviewerID,
		"lawyer_facts_version":     strings.TrimSpace(factsVersion),
		"client_id":                intakeFacts.ClientID,
		"intake_created_by":        intakeFacts.CreatedBy,
		"structured_facts":         structuredFacts,
		"case_creation_config":     metadata["case_creation_config"],
		"applicant_id":             approval.ApplicantID,
		"applicant_name":           approval.ApplicantName,
		"submission_source":        "initial_submit",
	}
	// Same material freeze as the resubmit round: the initial binding
	// snapshot carries a verifiable material reference for the round it
	// commits.
	materialFingerprints, mfErr := freezeMaterialFingerprintsForSnapshot(tx, intakeID)
	if mfErr != nil {
		return mfErr
	}
	payload["material_fingerprints"] = materialFingerprints
	snapshotID, err := persistIntakeSubmissionSnapshot(tx, approval.ID, 1, payload, newIntakeSnapshotID())
	if err != nil {
		return err
	}
	binding.SnapshotID = snapshotID
	if err := tx.Table("approval_requests").Where("id = ?", approval.ID).
		Update("active_submission_snapshot_id", snapshotID).Error; err != nil {
		return fmt.Errorf("更新提交快照指针失败：%w", err)
	}
	// Keep the in-memory struct in sync: the caller's SaveInTx must not
	// write the stale pointer/column anchors back over the fresh binding.
	approval.ActiveSubmissionSnapshotID = &snapshotID
	if err := syncApprovalConflictBindingTx(tx, approval, binding); err != nil {
		var gateErr *ApprovalConflictGateError
		if errors.As(err, &gateErr) {
			return err
		}
		return err
	}
	return nil
}

// stringValueOrEmpty dereferences an optional string column; nil and empty
// both collapse to "" so snapshot payloads stay JSON-stable across drivers.
func stringValueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

// ptrToJSONBString converts serialized supplement-replies JSON to a nullable
// string pointer so the JSONB column can stay NULL-safe across databases.
func ptrToJSONBString(data []byte) *string {
	v := string(data)
	return &v
}

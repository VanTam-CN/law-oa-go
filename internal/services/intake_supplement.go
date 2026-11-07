package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
)

// IntakeCollaborationError is safe to expose using the existing API envelope.
type IntakeCollaborationError struct {
	Code    string
	Message string
	Status  int
}

func (e *IntakeCollaborationError) Error() string { return e.Message }
func intakeCollaborationError(code, message string, status int) error {
	return &IntakeCollaborationError{Code: code, Message: message, Status: status}
}
func supplementInvalid(message string) error {
	return intakeCollaborationError("SUPPLEMENT_ITEM_INVALID", message, 400)
}
func submissionStale(message string) error {
	return intakeCollaborationError("APPROVAL_ROUND_STALE", message, 409)
}
func snapshotUnresolved(message string) error {
	return intakeCollaborationError("APPROVAL_SNAPSHOT_UNRESOLVED", message, 409)
}

var supplementFieldKeys = map[string]bool{
	"client_id": true, "title": true, "description": true, "case_type": true,
	"business_area": true, "sub_area": true, "lawyer_id": true,
	"parties": true, "party_name": true, "party_role": true, "party_identity": true,
	"opponent_name": true, "opponent_identity_number": true,
}

func plainSupplementText(value string, limit int) bool {
	lower := strings.ToLower(value)
	return strings.TrimSpace(value) != "" && utf8.RuneCountInString(value) <= limit &&
		!strings.ContainsAny(value, "<>\x00") && !strings.Contains(lower, "http:") &&
		!strings.Contains(lower, "https:") && !strings.Contains(lower, "javascript:")
}

// NormalizeSupplementRequirements accepts {schema_version:1,supplement_requests:[...]},
// rejects unknown keys, and generates stable item IDs before the record is saved.
// Legacy free-text returns are supported by the reader, never silently created here.
func NormalizeSupplementRequirements(input map[string]interface{}, reason string) (string, error) {
	if !plainSupplementText(reason, 5000) {
		return "", supplementInvalid("退回必须填写有效审批意见")
	}
	for k := range input {
		if k != "schema_version" && k != "supplement_requests" {
			return "", supplementInvalid("补充清单包含不支持的字段")
		}
	}
	if version, ok := input["schema_version"]; ok && fmt.Sprint(version) != "1" {
		return "", supplementInvalid("不支持的补充清单版本")
	}
	raw, err := json.Marshal(input["supplement_requests"])
	if err != nil {
		return "", supplementInvalid("补充清单无法解析")
	}
	var items []models.SupplementRequest
	if err := strictSupplementDecode(raw, &items); err != nil || len(items) < 1 || len(items) > 20 {
		return "", supplementInvalid("补充清单必须包含1至20项且字段有效")
	}
	for i := range items {
		item := &items[i]
		item.Category = strings.TrimSpace(item.Category)
		item.FieldKey = strings.TrimSpace(item.FieldKey)
		item.Description = strings.TrimSpace(item.Description)
		item.PartyLocalID = strings.TrimSpace(item.PartyLocalID)
		if item.ItemID != "" || item.Source != "" {
			return "", supplementInvalid("补充项编号与来源由服务端生成")
		}
		if item.Category != "field_correction" && item.Category != "identity_verification" && item.Category != "matter_explanation" {
			return "", supplementInvalid("不支持的补充类别")
		}
		if !supplementFieldKeys[item.FieldKey] || !plainSupplementText(item.Description, 1000) {
			return "", supplementInvalid("补充项字段或说明无效")
		}
		if item.PartyLocalID != "" && (!strings.HasPrefix(item.FieldKey, "party_") && item.FieldKey != "parties" || !plainSupplementText(item.PartyLocalID, 100)) {
			return "", supplementInvalid("当事人定位标识无效")
		}
		if strings.HasPrefix(item.FieldKey, "party_") && item.PartyLocalID == "" {
			return "", supplementInvalid("当事人补充项必须使用稳定定位编号")
		}
		// Identity requests deliberately carry no user-entered identity detail.
		if item.Category == "identity_verification" {
			item.Description = "需核验该主体身份"
		}
		item.ItemID = uuid.NewString()
	}
	raw, err = json.Marshal(map[string]interface{}{"schema_version": 1, "supplement_requests": items})
	return string(raw), err
}
func strictSupplementDecode(raw []byte, target interface{}) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if err := d.Decode(new(interface{})); err != io.EOF {
		return fmt.Errorf("unexpected trailing JSON")
	}
	return nil
}

// ReadSupplementRequests leaves historical record bytes untouched. A legacy
// free-text return has a deterministic virtual item so refreshes retain replies.
func ReadSupplementRequests(record *models.ApprovalRecord) ([]models.SupplementRequest, error) {
	if record == nil {
		return nil, submissionStale("退回记录不存在")
	}
	var envelope map[string]json.RawMessage
	raw := strings.TrimSpace(record.ImposedRequirements)
	if raw != "" && raw != "null" {
		if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
			return nil, supplementInvalid("退回清单无法解析")
		}
	}
	if value, ok := envelope["supplement_requests"]; ok {
		var items []models.SupplementRequest
		if string(envelope["schema_version"]) != "1" || strictSupplementDecode(value, &items) != nil || len(items) < 1 || len(items) > 20 {
			return nil, supplementInvalid("退回清单版本或内容无效")
		}
		seen := map[string]bool{}
		for _, item := range items {
			if item.ItemID == "" || seen[item.ItemID] || !supplementFieldKeys[item.FieldKey] {
				return nil, supplementInvalid("退回清单编号或字段无效")
			}
			seen[item.ItemID] = true
		}
		return items, nil
	}
	reason := strings.TrimSpace(record.DecisionReason)
	if reason == "" {
		return nil, supplementInvalid("历史退回缺少可回复的审批意见")
	}
	return []models.SupplementRequest{{ItemID: "legacy-" + record.ID, Category: "matter_explanation", FieldKey: "description", Description: reason, Required: true, Source: "legacy_text"}}, nil
}

func normalizeSupplementResponses(items []models.SupplementRequest, responses []models.SupplementResponse, complete bool) ([]models.SupplementResponse, error) {
	if len(responses) > 20 {
		return nil, supplementInvalid("补充回复不能超过20项")
	}
	allowed := map[string]models.SupplementRequest{}
	for _, item := range items {
		allowed[item.ItemID] = item
	}
	seen := map[string]bool{}
	result := make([]models.SupplementResponse, 0, len(responses))
	for _, response := range responses {
		response.ItemID = strings.TrimSpace(response.ItemID)
		response.Response = strings.TrimSpace(response.Response)
		if _, ok := allowed[response.ItemID]; !ok || response.ItemID == "" || seen[response.ItemID] || !plainSupplementText(response.Response, 2000) {
			return nil, supplementInvalid("补充回复包含未知项、重复项或无效说明")
		}
		seen[response.ItemID] = true
		result = append(result, response)
	}
	if complete {
		for _, item := range items {
			if item.Required && !seen[item.ItemID] {
				return nil, supplementInvalid("请逐项回复所有必填补充要求")
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ItemID < result[j].ItemID })
	return result, nil
}

func currentIntakeReturn(tx *gorm.DB, approval *models.ApprovalRequest, returnID string) (*models.ApprovalRecord, error) {
	if approval.Status != models.ApprovalStatusNeedsRevision {
		return nil, submissionStale("该审批当前不允许补正重提")
	}
	var records []models.ApprovalRecord
	if err := tx.Where("approval_request_id = ? AND decision = ?", approval.ID, "request_changes").Order("created_at DESC, id DESC").Limit(2).Find(&records).Error; err != nil {
		return nil, err
	}
	if len(records) == 0 || records[0].ID != returnID {
		return nil, submissionStale("退回轮次已变化，请重新加载原申请")
	}
	record := records[0]
	if len(records) > 1 && record.CreatedAt.Equal(records[1].CreatedAt) {
		return nil, submissionStale("退回记录顺序无法确认")
	}
	if approval.ActiveSubmissionSnapshotID != nil && (record.SubmissionSnapshotID == nil || *record.SubmissionSnapshotID != *approval.ActiveSubmissionSnapshotID) {
		return nil, submissionStale("退回记录不属于当前提交版本")
	}
	return &record, nil
}

// SaveIntakeSupplementResponsesInTx owns approval -> intake locks. It does not
// start/commit a transaction; ANY error must roll back the caller's transaction.
// It only changes three controlled metadata keys and updated_at, never facts/status.
func SaveIntakeSupplementResponsesInTx(tx *gorm.DB, intakeID, approvalID, userID, expectedRevision, returnRecordID string, responses []models.SupplementResponse) (string, error) {
	if tx == nil {
		return "", errors.New("补正事务未初始化")
	}
	var approval models.ApprovalRequest
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", approvalID).First(&approval).Error; err != nil {
		return "", err
	}
	if approval.ApplicantID != userID {
		return "", intakeCollaborationError("FORBIDDEN", "无权补正该申请", 403)
	}
	snapshot, err := ResolveApprovalSubmission(tx, &approval)
	if err != nil {
		return "", err
	}
	if submissionText(submissionMap(snapshot["metadata"])["intake_id"]) != intakeID || intakeID == "" {
		return "", intakeCollaborationError("FORBIDDEN", "审批与接案关联不一致", 403)
	}
	intake, metadata, err := lockSubmissionIntake(tx, intakeID, userID)
	if err != nil {
		return "", err
	}
	if err := AuthorizeApprovalResubmission(tx, userID, &approval); err != nil {
		return "", err
	}
	if expectedRevision == "" || expectedRevision != submissionRevision(intakeID, metadata) {
		return "", intakeCollaborationError("INTAKE_DRAFT_REVISION_CONFLICT", "草稿已更新，请刷新后再保存", 409)
	}
	if intake.Status == "case_created" || approval.CaseCreated || approval.CreatedCaseID != "" {
		return "", submissionStale("已成案申请不能补正")
	}
	record, err := currentIntakeReturn(tx, &approval, returnRecordID)
	if err != nil {
		return "", err
	}
	items, err := ReadSupplementRequests(record)
	if err != nil {
		return "", err
	}
	normalized, err := normalizeSupplementResponses(items, responses, false)
	if err != nil {
		return "", err
	}
	next := uuid.NewString()
	metadata["supplement_return_record_id"] = returnRecordID
	metadata["supplement_responses"] = normalized
	metadata["draft_revision"] = next
	raw, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	result := tx.Table("case_intakes").Where("id = ?", intakeID).Updates(map[string]interface{}{"metadata": string(raw), "updated_at": time.Now()})
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected != 1 {
		return "", submissionStale("接案记录已变化")
	}
	return next, nil
}

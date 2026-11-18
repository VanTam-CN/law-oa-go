package services

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// Author's isolated probes for the shared material-evidence algorithm.
// They complement, and never replace, the reviewer's independent
// full-chain probe in mvp_material_resubmit_independent_test.go.

func seedMaterialEvidenceGate(t *testing.T) (*gorm.DB, *models.ApprovalRequest) {
	t.Helper()
	db, approval, rec, intakeID, _ := seedPostCommitConflictWorld(t)
	require.NoError(t, db.AutoMigrate(&models.CaseMaterial{}, &models.MaterialFile{}))
	returnTime := time.Now().UTC().Add(-10 * time.Minute)
	rec.ApprovalDate, rec.CreatedAt = returnTime, returnTime
	rec.ImposedRequirements = `{"schema_version":1,"items":[{"item_id":"contract-fix","category":"material_request","field_key":"material","material_id":200,"description":"请在原合同材料行上传完整签章的新版本","required":true}]}`
	require.NoError(t, db.Save(rec).Error)
	oldFile := models.MaterialFile{ID: 10, MaterialID: 200, FileName: "old.pdf", ContentSHA256: strings.Repeat("a", 64), SizeBytes: 12, StoragePath: "isolated-fixture-old", UploadedBy: 1, CreatedAt: returnTime.Add(-time.Hour)}
	require.NoError(t, db.Create(&oldFile).Error)
	require.NoError(t, db.Create(&models.CaseMaterial{ID: 200, IntakeID: &intakeID, Name: "虚构签章合同", Required: true, Status: models.MaterialStatusUploaded, StorageURL: "/api/v1/case-materials/200/files/10/download", Metadata: "{}"}).Error)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	var returned models.ApprovalSnapshot
	require.NoError(t, db.Where("id = ?", initialSnap).Take(&returned).Error)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(returned.SnapshotData), &payload))
	payload["material_fingerprints"] = []map[string]interface{}{{
		"material_id": 200, "name": "虚构签章合同", "required": true, "status": "uploaded", "current_file_id": 10,
		"files": []map[string]interface{}{{"file_id": 10, "digest": strings.Repeat("a", 64)}},
	}}
	frozenJSON, err := json.Marshal(payload)
	require.NoError(t, err)
	require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Where("id = ?", initialSnap).Update("snapshot_data", string(frozenJSON)).Error)
	return db, approval
}

func newFileAfterReturn(t *testing.T, db *gorm.DB, returnTime time.Time, digest string) {
	t.Helper()
	newFile := models.MaterialFile{ID: 11, MaterialID: 200, FileName: "corrected.pdf", ContentSHA256: digest, SizeBytes: 18, StoragePath: "isolated-fixture-new", UploadedBy: 1, CreatedAt: returnTime.Add(time.Minute)}
	require.NoError(t, db.Create(&newFile).Error)
	require.NoError(t, db.Model(&models.CaseMaterial{}).Where("id = ?", 200).Update("storage_url", "/api/v1/case-materials/200/files/11/download").Error)
}

func gateReturnRecord(t *testing.T, db *gorm.DB, approvalID string) models.ApprovalRecord {
	t.Helper()
	var rec models.ApprovalRecord
	require.NoError(t, db.Where("approval_request_id = ? AND decision = ?", approvalID, models.ApprovalDecisionRequestChanges).Take(&rec).Error)
	return rec
}

func runGateEvidence(t *testing.T, db *gorm.DB, approvalID string) []MaterialEvidenceItem {
	t.Helper()
	items := []SupplementItem{{ItemID: "contract-fix", Category: "material_request", MaterialID: 200}}
	var got []MaterialEvidenceItem
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		got = EvaluateMaterialResubmitEvidence(tx, approvalID, items)
		return nil
	}))
	return got
}

func TestAuthorMaterialEvidenceNewFileAfterReturnCompletes(t *testing.T) {
	db, approval := seedMaterialEvidenceGate(t)
	rec := gateReturnRecord(t, db, approval.ID)
	newFileAfterReturn(t, db, rec.ApprovalDate, strings.Repeat("b", 64))
	evidence := runGateEvidence(t, db, approval.ID)
	require.Len(t, evidence, 1)
	require.Equal(t, MaterialEvidenceComplete, evidence[0].State)
	require.Equal(t, uint(11), evidence[0].CurrentFileID)
}

func TestAuthorMaterialEvidenceFileUploadedBeforeReturnRejected(t *testing.T) {
	db, approval := seedMaterialEvidenceGate(t)
	rec := gateReturnRecord(t, db, approval.ID)
	newFile := models.MaterialFile{ID: 11, MaterialID: 200, FileName: "pre-return.pdf", ContentSHA256: strings.Repeat("b", 64), SizeBytes: 18, StoragePath: "isolated-fixture-new", UploadedBy: 1, CreatedAt: rec.ApprovalDate.Add(-time.Minute)}
	require.NoError(t, db.Create(&newFile).Error)
	require.NoError(t, db.Model(&models.CaseMaterial{}).Where("id = ?", 200).Update("storage_url", "/api/v1/case-materials/200/files/11/download").Error)
	evidence := runGateEvidence(t, db, approval.ID)
	require.Len(t, evidence, 1)
	require.Equal(t, MaterialEvidenceMissing, evidence[0].State)
}

func TestAuthorMaterialEvidenceUnknownDigestUnknown(t *testing.T) {
	db, approval := seedMaterialEvidenceGate(t)
	rec := gateReturnRecord(t, db, approval.ID)
	newFileAfterReturn(t, db, rec.ApprovalDate, "")
	evidence := runGateEvidence(t, db, approval.ID)
	require.Len(t, evidence, 1)
	require.Equal(t, MaterialEvidenceUnknown, evidence[0].State)
}

func TestAuthorMaterialEvidenceCrossApprovalSnapshotUnknown(t *testing.T) {
	db, approval := seedMaterialEvidenceGate(t)
	require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Where("approval_request_id = ?", approval.ID).Update("approval_request_id", "other-approval").Error)
	evidence := runGateEvidence(t, db, approval.ID)
	require.Len(t, evidence, 1)
	require.Equal(t, MaterialEvidenceUnknown, evidence[0].State)
}

func TestAuthorMaterialEvidenceLegacyBlankPointerRefusedAsEvidence(t *testing.T) {
	db, approval := seedMaterialEvidenceGate(t)
	rec := gateReturnRecord(t, db, approval.ID)
	newFileAfterReturn(t, db, rec.ApprovalDate, strings.Repeat("b", 64))
	require.NoError(t, db.Model(&models.CaseMaterial{}).Where("id = ?", 200).Update("storage_url", "").Error)
	evidence := runGateEvidence(t, db, approval.ID)
	require.Len(t, evidence, 1)
	require.Equal(t, MaterialEvidenceMissing, evidence[0].State)
}

func TestAuthorMaterialEvidenceParsePointerCanonicalOnly(t *testing.T) {
	_, _, ok := parseMaterialPointerURL("/api/v1/case-materials/200/files/11/download")
	require.True(t, ok)
	for _, bad := range []string{"/api/v1/case-materials/0200/files/011/download", "/api/v1/case-materials/200/files/11", "/api/v1/case-materials/200/files/-1/download", "/api/v1/case-materials/200/files/01/download", "relative/path", ""} {
		_, _, ok := parseMaterialPointerURL(bad)
		require.False(t, ok)
	}
}

func TestAuthorMaterialEvidenceDTOJSONShape(t *testing.T) {
	raw, err := json.Marshal(MaterialEvidenceItem{ItemID: "contract-fix", MaterialID: 200, State: MaterialEvidenceComplete, Message: "已上传新版本附件并成为当前文件", CurrentFileID: 11})
	require.NoError(t, err)
	require.JSONEq(t, `{"item_id":"contract-fix","material_id":200,"state":"complete","message":"已上传新版本附件并成为当前文件","current_file_id":11}`, string(raw))
	rawUnknown, err := json.Marshal(MaterialEvidenceItem{ItemID: "contract-fix", MaterialID: 200, State: MaterialEvidenceUnknown, Message: "材料补正证据暂时无法读取，状态未知，不能视为已补正"})
	require.NoError(t, err)
	require.JSONEq(t, `{"item_id":"contract-fix","material_id":200,"state":"unknown","message":"材料补正证据暂时无法读取，状态未知，不能视为已补正"}`, string(rawUnknown))
}

// Both unknowns must agree on the read and full write paths and must leave
// the returned round byte-for-byte unchanged when resubmission is blocked.
func TestAuthorMaterialEvidenceUnknownSnapshotKindAndCreatedTimeFailClosed(t *testing.T) {
	for _, scenario := range []string{"wrong_snapshot_kind", "file_unknown_created_time"} {
		t.Run(scenario, func(t *testing.T) {
			db, approval := seedMaterialEvidenceGate(t)
			rec := gateReturnRecord(t, db, approval.ID)
			newFileAfterReturn(t, db, rec.ApprovalDate, strings.Repeat("b", 64))
			require.NotNil(t, approval.ActiveSubmissionSnapshotID)
			snapshotID := *approval.ActiveSubmissionSnapshotID
			if scenario == "wrong_snapshot_kind" {
				require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Where("id = ?", snapshotID).
					Update("snapshot_type", "unrelated_record").Error)
			} else {
				require.NoError(t, db.Exec("UPDATE material_files SET created_at = NULL, updated_at = ? WHERE id = ?",
					rec.ApprovalDate.Add(2*time.Minute), 11).Error)
			}
			var before models.ApprovalSnapshot
			require.NoError(t, db.Where("id = ?", snapshotID).Take(&before).Error)
			evidence := runGateEvidence(t, db, approval.ID)
			require.Len(t, evidence, 1)
			require.Equal(t, MaterialEvidenceUnknown, evidence[0].State)
			if scenario == "file_unknown_created_time" {
				require.Equal(t, uint(11), evidence[0].CurrentFileID)
				require.Contains(t, evidence[0].Message, "上传时间")
			}
			_, err := NewApprovalService(db).StructuredResubmit("u-lawyer", approval.ID, &ResubmitApprovalRequest{
				ExpectedIntakeRevision: "rev-1", ExpectedSubmissionSnapshotID: snapshotID,
				ReturnRecordID: rec.ID, RevisionNote: "原行合同材料补正",
				Replies: []SupplementReply{{ItemID: "contract-fix", Response: "已上传，请核对附件"}},
			})
			require.Error(t, err)
			var current models.ApprovalRequest
			require.NoError(t, db.Where("id = ?", approval.ID).Take(&current).Error)
			require.Equal(t, models.ApprovalStatusNeedsRevision, current.Status)
			require.NotNil(t, current.ActiveSubmissionSnapshotID)
			require.Equal(t, snapshotID, *current.ActiveSubmissionSnapshotID)
			var after models.ApprovalSnapshot
			require.NoError(t, db.Where("id = ?", snapshotID).Take(&after).Error)
			require.Equal(t, before.SnapshotData, after.SnapshotData)
			var rounds int64
			require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Where("approval_request_id = ?", approval.ID).Count(&rounds).Error)
			require.Equal(t, int64(1), rounds)
		})
	}
}

package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// This is the reviewer's full resubmit probe, not the author's helper test:
// replying to a request cannot turn an unchanged old attachment into new
// evidence. The positive case must revive the same approval and preserve
// its returned-round snapshot byte-for-byte.
func TestIndependentReturnedMaterialRequiresNewOriginalRowEvidence(t *testing.T) {
	for _, scenario := range []string{"old_file_and_text", "new_original_file", "new_file_wrong_row", "new_file_deleted", "new_file_bad_pointer", "new_file_old_pointer", "file_uploaded_before_return", "new_file_unknown_digest", "new_file_noncanonical_pointer", "different_approval_snapshot", "wrong_snapshot_kind", "file_unknown_created_time"} {
		t.Run(scenario, func(t *testing.T) {
			db, approval, rec, intakeID, _ := seedPostCommitConflictWorld(t)
			require.NoError(t, db.AutoMigrate(&models.CaseMaterial{}, &models.MaterialFile{}))
			returnTime := time.Now().UTC().Add(-10 * time.Minute)
			rec.ApprovalDate, rec.CreatedAt = returnTime, returnTime
			rec.ImposedRequirements = `{"schema_version":1,"items":[{"item_id":"contract-fix","category":"material_request","field_key":"material","material_id":200,"description":"请在原合同材料行上传完整签章的新版本","required":true}]}`
			require.NoError(t, db.Save(rec).Error)
			oldFile := models.MaterialFile{ID: 10, MaterialID: 200, FileName: "fictional-old.pdf", ContentSHA256: strings.Repeat("a", 64), SizeBytes: 12, StoragePath: "isolated-fixture-old", UploadedBy: 1, CreatedAt: returnTime.Add(-time.Hour)}
			require.NoError(t, db.Create(&oldFile).Error)
			material := models.CaseMaterial{ID: 200, IntakeID: &intakeID, Name: "虚构签章合同", Required: true, Status: models.MaterialStatusUploaded, StorageURL: "/api/v1/case-materials/200/files/10/download", Metadata: "{}"}
			require.NoError(t, db.Create(&material).Error)
			initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
			var returned models.ApprovalSnapshot
			require.NoError(t, db.Where("id = ?", initialSnap).Take(&returned).Error)
			var original map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(returned.SnapshotData), &original))
			original["material_fingerprints"] = []map[string]interface{}{{
				"material_id": 200, "name": material.Name, "required": true, "status": "uploaded", "current_file_id": 10,
				"files": []map[string]interface{}{{"file_id": 10, "digest": oldFile.ContentSHA256, "file_name": oldFile.FileName, "size_bytes": 12}},
			}}
			frozenJSON, err := json.Marshal(original)
			require.NoError(t, err)
			require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Where("id = ?", initialSnap).Update("snapshot_data", string(frozenJSON)).Error)
			if scenario == "different_approval_snapshot" {
				require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Where("id = ?", initialSnap).Update("approval_request_id", "some-other-approval").Error)
			}
			if scenario == "wrong_snapshot_kind" {
				require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Where("id = ?", initialSnap).Update("snapshot_type", "unrelated_record").Error)
			}
			if scenario != "old_file_and_text" {
				newFile := models.MaterialFile{ID: 11, MaterialID: 200, FileName: "fictional-corrected.pdf", ContentSHA256: strings.Repeat("b", 64), SizeBytes: 18, StoragePath: "isolated-fixture-new", UploadedBy: 1, CreatedAt: returnTime.Add(time.Minute)}
				if scenario == "new_file_wrong_row" {
					newFile.MaterialID = 201
				}
				if scenario == "file_uploaded_before_return" {
					newFile.CreatedAt = returnTime.Add(-time.Minute)
				}
				if scenario == "new_file_unknown_digest" {
					newFile.ContentSHA256 = ""
				}
				require.NoError(t, db.Create(&newFile).Error)
				if scenario == "file_unknown_created_time" {
					require.NoError(t, db.Exec("UPDATE material_files SET created_at = NULL, updated_at = ? WHERE id = ?", returnTime.Add(2*time.Minute), newFile.ID).Error)
				}
				pointer := "/api/v1/case-materials/200/files/11/download"
				if scenario == "new_file_bad_pointer" {
					pointer = "/api/v1/case-materials/201/files/11/download"
				}
				if scenario == "new_file_old_pointer" {
					pointer = material.StorageURL
				}
				if scenario == "new_file_noncanonical_pointer" {
					pointer = "/api/v1/case-materials/0200/files/011/download"
				}
				require.NoError(t, db.Model(&models.CaseMaterial{}).Where("id = ?", material.ID).Update("storage_url", pointer).Error)
				if scenario == "new_file_deleted" {
					require.NoError(t, db.Delete(&newFile).Error)
				}
			}
			items, parseErr := parseStoredSupplementItems(rec.ImposedRequirements)
			require.NoError(t, parseErr)
			readEvidence := EvaluateMaterialResubmitEvidence(db, approval.ID, items)
			require.Len(t, readEvidence, 1)
			require.Equal(t, "contract-fix", readEvidence[0].ItemID)
			require.Equal(t, uint(200), readEvidence[0].MaterialID)
			if scenario == "new_original_file" {
				require.Equal(t, MaterialEvidenceComplete, readEvidence[0].State)
			} else if readEvidence[0].State == MaterialEvidenceComplete {
				t.Errorf("read DTO incorrectly declared %s complete", scenario)
			}
			req := &ResubmitApprovalRequest{ExpectedIntakeRevision: "rev-1", ExpectedSubmissionSnapshotID: initialSnap, ReturnRecordID: rec.ID, RevisionNote: "按原行合同要求补正", Replies: []SupplementReply{{ItemID: "contract-fix", Response: "已补正，请核对原合同附件"}}}
			updated, err := NewApprovalService(db).StructuredResubmit("u-lawyer", approval.ID, req)
			var preserved models.ApprovalSnapshot
			require.NoError(t, db.Where("id = ?", initialSnap).Take(&preserved).Error)
			require.Equal(t, string(frozenJSON), preserved.SnapshotData, "the returned round must not be rewritten")
			if scenario == "new_original_file" {
				require.NoError(t, err)
				require.Equal(t, approval.ID, updated.ID)
				require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)
				require.NotNil(t, updated.ActiveSubmissionSnapshotID)
				require.NotEqual(t, initialSnap, *updated.ActiveSubmissionSnapshotID)
				var next models.ApprovalSnapshot
				require.NoError(t, db.Where("id = ?", *updated.ActiveSubmissionSnapshotID).Take(&next).Error)
				require.Equal(t, 2, next.SourceVersion)
				require.Contains(t, next.SnapshotData, "fictional-corrected.pdf")
			} else {
				require.Error(t, err, "old text or unrelated/invalid file must not satisfy material evidence")
				var current models.ApprovalRequest
				require.NoError(t, db.Where("id = ?", approval.ID).Take(&current).Error)
				require.Equal(t, models.ApprovalStatusNeedsRevision, current.Status)
				require.NotNil(t, current.ActiveSubmissionSnapshotID)
				require.Equal(t, initialSnap, *current.ActiveSubmissionSnapshotID)
				var rounds int64
				require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Where("approval_request_id = ?", approval.ID).Count(&rounds).Error)
				expectedRounds := int64(1)
				if scenario == "different_approval_snapshot" {
					expectedRounds = 0
				}
				require.Equal(t, expectedRounds, rounds, "failure must leave no new snapshot")
			}
		})
	}
}

func TestIndependentFrozenMaterialRespectsSelectedOriginalVersion(t *testing.T) {
	db := seedResubmitContractWorld(t)
	require.NoError(t, db.AutoMigrate(&models.CaseMaterial{}, &models.MaterialFile{}))
	intakeID := "independent-pointer-intake"
	row := models.CaseMaterial{ID: 200, IntakeID: &intakeID, Name: "虚构原合同", Status: models.MaterialStatusUploaded, StorageURL: "/api/v1/case-materials/200/files/10/download", Metadata: "{}"}
	require.NoError(t, db.Create(&row).Error)
	for _, id := range []uint{10, 11} {
		require.NoError(t, db.Create(&models.MaterialFile{ID: id, MaterialID: 200, FileName: fmt.Sprintf("fictional-%d.pdf", id), ContentSHA256: strings.Repeat("c", 64), SizeBytes: 9, StoragePath: "isolated-fixture", UploadedBy: 1}).Error)
	}
	var frozen []materialSnapshotFingerprint
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		var err error
		frozen, err = freezeMaterialFingerprintsForSnapshot(tx, intakeID)
		return err
	}))
	require.Len(t, frozen, 1)
	require.Equal(t, uint(10), frozen[0].CurrentFileID, "a selected old file must not be silently replaced by newest file")
}

package services

// Material correction evidence gate (review 2026-10-02): a material_request
// item in a return checklist is complete only when the CURRENT live state of
// the referenced material proves NEW correction evidence on top of the
// frozen returned-round baseline. Every condition is server-verified:
// the baseline must belong to THIS approval (snapshot row plus payload
// intake binding), the pointer must name a live file of THIS material row
// through the exact canonical download URL, that file must have been
// uploaded AFTER the latest legal return record, and its stored SHA-256
// must be verifiable and different from every frozen baseline digest.
// Text replies, pre-return files, other rows, soft-deleted files, broken or
// missing pointers, and read failures never count as completion, and a
// round without a frozen material baseline stays honestly unknown instead
// of being reported as unchanged.

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// Evidence states for the return-items read DTO and the resubmit gate.
const (
	MaterialEvidenceComplete = "complete"
	MaterialEvidenceMissing  = "missing"
	MaterialEvidenceUnknown  = "unknown"
)

// materialEvidencePointerPrefix is the canonical download path prefix the
// upload transaction stores in case_materials.storage_url; List enrichment
// matches on exactly this string, so the gate and the freeze helper must
// resolve the same contract. Parsing stays digit-exact: a leading-zero id
// such as /0200/files/011/download is NOT canonical and must fail here
// exactly like it fails the List renderer.
const materialEvidencePointerPrefix = "/api/v1/case-materials/"

// MaterialEvidenceItem is one material_request item's server-computed
// correction evidence for the return-items read DTO. current_file_id stays
// omitted while the pointer cannot be resolved.
type MaterialEvidenceItem struct {
	ItemID        string `json:"item_id"`
	MaterialID    uint   `json:"material_id"`
	State         string `json:"state"`
	Message       string `json:"message"`
	CurrentFileID uint   `json:"current_file_id,omitempty"`
}

// parseMaterialPointerURL parses the canonical download URL stored by the
// upload transaction into (materialID, fileID). Anything that is not
// exactly the canonical shape fails closed.
func parseMaterialPointerURL(raw string) (materialID, fileID uint, ok bool) {
	const suffix = "/download"
	const separator = "/files/"
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, materialEvidencePointerPrefix) {
		return 0, 0, false
	}
	rest := strings.TrimPrefix(trimmed, materialEvidencePointerPrefix)
	sep := strings.Index(rest, separator)
	if sep <= 0 || !strings.HasSuffix(rest, suffix) {
		return 0, 0, false
	}
	mid, midOK := parseCanonicalUint(rest[:sep])
	fid, fidOK := parseCanonicalUint(strings.TrimSuffix(rest[sep+len(separator):], suffix))
	if !midOK || !fidOK {
		return 0, 0, false
	}
	return uint(mid), uint(fid), true
}

// parseCanonicalUint accepts only plain decimal ids without leading zeros
// or signs, matching exactly what fmt.Sprint(uint) renders into the
// canonical download URL: "0200" and "011" are not canonical and must
// fail so the gate can never accept a pointer the List enrichment cannot
// render back.
func parseCanonicalUint(raw string) (uint, bool) {
	if raw == "" || len(raw) > 19 {
		return 0, false
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] < '0' || raw[i] > '9' {
			return 0, false
		}
	}
	if len(raw) > 1 && raw[0] == '0' {
		return 0, false
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		return 0, false
	}
	return uint(value), true
}

// resolveMaterialCurrentFilePointer applies the List enrichment pointer
// contract to one material row's live files: an explicit storage_url must
// resolve inside THIS row's live files, and an invalid explicit pointer
// never falls back. Only a completely blank pointer keeps the legacy
// newest-live-file fallback, and legacyFallback reports that case so
// callers can refuse to treat it as verified correction evidence.
func resolveMaterialCurrentFilePointer(material *models.CaseMaterial, liveFiles []models.MaterialFile) (fileID uint, pointed bool, legacyFallback bool) {
	if material == nil {
		return 0, false, false
	}
	pointer := strings.TrimSpace(material.StorageURL)
	if pointer == "" {
		if len(liveFiles) == 0 {
			return 0, false, false
		}
		// Legacy compatibility with the List enrichment fallback: freeze
		// the same head the list shows, but flag it so the evidence gate
		// can refuse to count it as proof of a completed correction.
		return liveFiles[len(liveFiles)-1].ID, true, true
	}
	pointerMaterialID, pointerFileID, ok := parseMaterialPointerURL(pointer)
	if !ok || pointerMaterialID != material.ID {
		return 0, false, false
	}
	for _, f := range liveFiles {
		if f.ID == pointerFileID {
			return pointerFileID, true, false
		}
	}
	// Soft-deleted, other-row, or nonexistent target: never guess.
	return 0, false, false
}

// snapshotMaterialBaseline loads the material fingerprints frozen into the
// approval's currently active submission round. The snapshot row must
// belong to THIS approval (approval_request_id column) and its payload
// intake_id must equal the approval's bound intake; anything else is
// treated as no baseline so another approval's frozen state can never
// vouch for this round. known reports whether that round actually froze
// the material_fingerprints key; a missing key stays unknown instead of
// being turned into an invented baseline.
func snapshotMaterialBaseline(tx *gorm.DB, approvalID string) (map[uint]materialSnapshotFingerprint, bool, error) {
	approvalID = strings.TrimSpace(approvalID)
	activeSnapshotID, err := activeSnapshotIDTx(tx, approvalID)
	if err != nil {
		return nil, false, err
	}
	if activeSnapshotID == "" {
		return nil, false, nil
	}
	var row struct {
		SnapshotData      string
		ApprovalRequestID string
		IntakeID          *string
	}
	if err := tx.Table("approval_snapshots").
		Select("approval_snapshots.snapshot_data, approval_snapshots.approval_request_id, approval_requests.metadata->>'intake_id' AS intake_id").
		Joins("JOIN approval_requests ON approval_requests.id = approval_snapshots.approval_request_id").
		Where("approval_snapshots.id = ? AND approval_snapshots.approval_request_id = ? AND approval_snapshots.snapshot_type = ?",
			activeSnapshotID, approvalID, models.ApprovalSnapshotTypeIntakeSubmission).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("读取退回轮提交快照失败：%w", err)
	}
	if strings.TrimSpace(row.ApprovalRequestID) != approvalID {
		return nil, false, nil
	}
	boundIntakeID := ""
	if row.IntakeID != nil {
		boundIntakeID = strings.TrimSpace(*row.IntakeID)
	}
	if boundIntakeID == "" {
		return nil, false, nil
	}
	payload := map[string]interface{}{}
	if err := json.Unmarshal([]byte(row.SnapshotData), &payload); err != nil {
		return nil, false, fmt.Errorf("解析退回轮提交快照失败：%w", err)
	}
	payloadIntakeID, _ := payload["intake_id"].(string)
	if strings.TrimSpace(payloadIntakeID) != boundIntakeID {
		return nil, false, nil
	}
	raw, exists := payload["material_fingerprints"]
	if !exists {
		return nil, false, nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, false, fmt.Errorf("解析退回轮材料基线失败：%w", err)
	}
	baseline := []materialSnapshotFingerprint{}
	if err := json.Unmarshal(encoded, &baseline); err != nil {
		return nil, false, fmt.Errorf("解析退回轮材料基线失败：%w", err)
	}
	byMaterialID := map[uint]materialSnapshotFingerprint{}
	for _, entry := range baseline {
		byMaterialID[entry.MaterialID] = entry
	}
	return byMaterialID, true, nil
}

// latestReturnRecordTx resolves the approval's newest server-owned
// request_changes record (same latest rule as the reply gate). Its
// approval_date, falling back to created_at, is the boundary separating
// "uploaded before the return" from "corrected after the return".
func latestReturnRecordTx(tx *gorm.DB, approvalID string) (*models.ApprovalRecord, error) {
	var row models.ApprovalRecord
	err := tx.Where("approval_request_id = ? AND decision = ?", strings.TrimSpace(approvalID), models.ApprovalDecisionRequestChanges).
		Order("approval_date DESC, created_at DESC, id DESC").
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取退回记录失败：%w", err)
	}
	return &row, nil
}

// evaluateMaterialResubmitEvidenceTx is the single algorithm shared by the
// resubmit gate and the return-items read DTO. Infrastructural read
// failures mark the affected item unknown AND surface the error so the
// write gate fails closed; the read surface keeps the unknown entries.
func evaluateMaterialResubmitEvidenceTx(tx *gorm.DB, approvalID string, items []SupplementItem) ([]MaterialEvidenceItem, error) {
	baseline, baselineKnown, err := snapshotMaterialBaseline(tx, approvalID)
	if err != nil {
		return nil, err
	}
	metadataRow := struct {
		Metadata string
	}{}
	if err := tx.Table("approval_requests").
		Select("metadata").
		Where("id = ?", approvalID).
		Take(&metadataRow).Error; err != nil {
		return nil, fmt.Errorf("读取审批绑定失败：%w", err)
	}
	metadata := map[string]interface{}{}
	if strings.TrimSpace(metadataRow.Metadata) != "" {
		if err := json.Unmarshal([]byte(metadataRow.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("解析审批绑定失败：%w", err)
		}
	}
	intakeID, _ := metadata["intake_id"].(string)
	intakeID = strings.TrimSpace(intakeID)

	evidence := []MaterialEvidenceItem{}
	var firstErr error
	for _, it := range items {
		if it.Category != "material_request" || it.MaterialID == 0 {
			continue
		}
		entry := MaterialEvidenceItem{
			ItemID:     it.ItemID,
			MaterialID: it.MaterialID,
			State:      MaterialEvidenceUnknown,
		}
		if !baselineKnown {
			entry.Message = "退回轮快照缺少材料指纹基线，无法证明补正，不能按未变化处理"
			evidence = append(evidence, entry)
			continue
		}
		baselineEntry, hadBaseline := baseline[it.MaterialID]
		var row models.CaseMaterial
		readErr := tx.Where("id = ?", it.MaterialID).Take(&row).Error
		if readErr != nil {
			if errors.Is(readErr, gorm.ErrRecordNotFound) {
				entry.State = MaterialEvidenceMissing
				entry.Message = "材料不存在"
			} else {
				entry.Message = "读取材料状态失败，不能证明补正"
				if firstErr == nil {
					firstErr = fmt.Errorf("读取材料 %d 状态失败：%w", it.MaterialID, readErr)
				}
			}
			evidence = append(evidence, entry)
			continue
		}
		if intakeID == "" || row.IntakeID == nil || *row.IntakeID != intakeID {
			entry.State = MaterialEvidenceMissing
			entry.Message = "材料不属于本接案申请，不能作为补正证据"
			evidence = append(evidence, entry)
			continue
		}
		if row.Status != models.MaterialStatusUploaded {
			entry.State = MaterialEvidenceMissing
			entry.Message = "材料状态不是已上传，不能算补正完成"
			evidence = append(evidence, entry)
			continue
		}
		returnRecord, returnErr := latestReturnRecordTx(tx, approvalID)
		if returnErr != nil {
			entry.Message = "读取退回记录失败，不能证明补正"
			if firstErr == nil {
				firstErr = returnErr
			}
			evidence = append(evidence, entry)
			continue
		}
		if returnRecord == nil || (returnRecord.ApprovalDate.IsZero() && returnRecord.CreatedAt.IsZero()) {
			entry.State = MaterialEvidenceMissing
			entry.Message = "退回时间不可确认，不能证明在退回后补正"
			evidence = append(evidence, entry)
			continue
		}
		returnTime := returnRecord.ApprovalDate
		if returnTime.IsZero() {
			returnTime = returnRecord.CreatedAt
		}
		liveFiles := []models.MaterialFile{}
		if err := tx.Where("material_id = ? AND deleted_at IS NULL", row.ID).
			Order("id ASC").Find(&liveFiles).Error; err != nil {
			entry.Message = "读取材料附件失败，不能证明补正"
			if firstErr == nil {
				firstErr = fmt.Errorf("读取材料 %d 附件失败：%w", row.ID, err)
			}
			evidence = append(evidence, entry)
			continue
		}
		fileID, pointed, legacyFallback := resolveMaterialCurrentFilePointer(&row, liveFiles)
		entry.CurrentFileID = fileID
		switch {
		case !pointed:
			entry.State = MaterialEvidenceMissing
			entry.Message = "当前附件指针无效（指向其他材料、已删除或不存在），不能回退到旧文件"
		case legacyFallback:
			entry.State = MaterialEvidenceMissing
			entry.Message = "材料缺少当前附件指针，不能证明补正完成"
		case !hadBaseline:
			entry.State = MaterialEvidenceMissing
			entry.Message = "退回轮未冻结该材料，不能证明补正"
		default:
			var current models.MaterialFile
			for _, f := range liveFiles {
				if f.ID == fileID {
					current = f
					break
				}
			}
			baselineDigestByID := map[uint]string{}
			baselineDigestSet := map[string]bool{}
			for _, bf := range baselineEntry.Files {
				baselineDigestByID[bf.FileID] = bf.Digest
				if d := strings.ToLower(strings.TrimSpace(bf.Digest)); d != "" {
					baselineDigestSet[d] = true
				}
			}
			uploadedAt := current.CreatedAt
			currentDigest := strings.ToLower(strings.TrimSpace(current.ContentSHA256))
			switch {
			case uploadedAt.IsZero():
				entry.State = MaterialEvidenceUnknown
				entry.Message = "当前附件上传时间未知，不能证明在退回后补正"
			case baselineDigestByID[fileID] != "":
				entry.State = MaterialEvidenceMissing
				if oldDigest := strings.ToLower(strings.TrimSpace(baselineDigestByID[fileID])); oldDigest != "" && currentDigest == oldDigest {
					entry.Message = "当前附件仍是退回前旧文件，不能算补正完成"
				} else {
					entry.Message = "当前文件与退回前登记内容不一致，不能算补正完成"
				}
			case !uploadedAt.After(returnTime):
				entry.State = MaterialEvidenceMissing
				entry.Message = "当前附件不是退回之后上传的新版本，不能证明在退回后补正"
			case !validMaterialDigest(currentDigest):
				entry.State = MaterialEvidenceUnknown
				entry.Message = "当前附件缺少可验证的内容摘要，状态未知，不能视为已补正"
			case baselineDigestSet[currentDigest]:
				entry.State = MaterialEvidenceMissing
				entry.Message = "当前附件与退回前附件内容一致，不能算补正完成"
			default:
				entry.State = MaterialEvidenceComplete
				entry.Message = "已上传新版本附件并成为当前文件"
			}
		}
		evidence = append(evidence, entry)
	}
	return evidence, firstErr
}

// materialDigestPattern accepts only the exact 64-hex SHA-256 form the
// upload transaction stores.
var materialDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// validMaterialDigest refuses empty or malformed digests: file content that
// cannot be proven can never count as verified correction evidence, so the
// caller reports unknown instead of complete.
func validMaterialDigest(digest string) bool {
	return materialDigestPattern.MatchString(digest)
}

// enforceMaterialResubmitEvidenceTx blocks a structured resubmit unless
// every material_request item proves complete correction evidence.
// Baseline/read failures fail closed as infrastructure errors; incomplete
// evidence is the typed 4xx contract.
func enforceMaterialResubmitEvidenceTx(tx *gorm.DB, approvalID string, items []SupplementItem) error {
	evidence, err := evaluateMaterialResubmitEvidenceTx(tx, approvalID, items)
	if err != nil {
		return fmt.Errorf("材料补正证据校验失败，已阻止重提：%w", err)
	}
	for _, entry := range evidence {
		if entry.State != MaterialEvidenceComplete {
			return &ResubmitContractError{
				Code: models.ErrCodeSupplementItemInvalid,
				Message: fmt.Sprintf("材料补正未完成（item %s / material %d）：%s",
					entry.ItemID, entry.MaterialID, entry.Message),
			}
		}
	}
	return nil
}

// unknownEvidenceForItems renders every material_request item as an honest
// unknown entry; it is the read DTO's degraded shape when evidence cannot
// be evaluated at all.
func unknownEvidenceForItems(items []SupplementItem) []MaterialEvidenceItem {
	evidence := make([]MaterialEvidenceItem, 0, len(items))
	for _, it := range items {
		if it.Category != "material_request" || it.MaterialID == 0 {
			continue
		}
		evidence = append(evidence, MaterialEvidenceItem{
			ItemID:     it.ItemID,
			MaterialID: it.MaterialID,
			State:      MaterialEvidenceUnknown,
			Message:    "材料补证证据暂时无法读取，状态未知，不能视为已补正",
		})
	}
	return evidence
}

// EvaluateMaterialResubmitEvidence is the read-side wrapper for the
// return-items DTO. It shares evaluateMaterialResubmitEvidenceTx with the
// resubmit gate so the UI and the write contract can never disagree. Any
// evaluation failure degrades to unknown entries instead of inventing a
// pass; callers keep serving items and replies.
func EvaluateMaterialResubmitEvidence(tx *gorm.DB, approvalID string, items []SupplementItem) []MaterialEvidenceItem {
	evidence, err := evaluateMaterialResubmitEvidenceTx(tx, approvalID, items)
	if err != nil {
		return unknownEvidenceForItems(items)
	}
	return evidence
}

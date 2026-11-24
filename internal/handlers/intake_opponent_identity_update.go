package handlers

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

// applyOpponentIdentityUpdate performs an in-place identity refresh for one
// existing opposing party. It never deletes or recreates the party row, so
// unknown columns and protected metadata survive narrow identity fixes.
func applyOpponentIdentityUpdate(tx *gorm.DB, intakeID string, input interface{}) error {
	payload, ok := input.(map[string]interface{})
	if !ok {
		return services.NewSubjectWorkflowError("INTAKE_PROTOCOL_INVALID", "opponent_identity_update 需要 JSON 对象")
	}
	for _, key := range []string{"entity_name", "entity_type", "identity_type", "identity_number"} {
		if value, valid := payload[key].(string); !valid || strings.TrimSpace(value) == "" {
			return services.NewSubjectWorkflowError("INTAKE_PROTOCOL_INVALID", fmt.Sprintf("opponent_identity_update 缺少 %s", key))
		}
	}
	aliasesSupplied := false
	for key := range payload {
		switch key {
		case "entity_name", "entity_type", "identity_type", "identity_number":
		case "aliases":
			aliasesSupplied = true
		default:
			return services.NewSubjectWorkflowError("INTAKE_PROTOCOL_INVALID", "opponent_identity_update 只允许 entity_name、entity_type、identity_type、identity_number、aliases")
		}
	}
	if _, ok := payload["aliases"].(string); aliasesSupplied && !ok {
		return services.NewSubjectWorkflowError("INTAKE_PROTOCOL_INVALID", "aliases 需要字符串")
	}
	entityName := strings.TrimSpace(stringValue(payload["entity_name"], ""))
	if entityName == "" {
		return services.NewSubjectWorkflowError("INTAKE_PARTY_NAME_REQUIRED", "接案当事人名称不能为空")
	}
	rawEntityType, ok := payload["entity_type"].(string)
	if !ok {
		return services.NewSubjectWorkflowError("INTAKE_PROTOCOL_INVALID", "entity_type 需要字符串")
	}
	suppliedType := strings.ToUpper(strings.TrimSpace(rawEntityType))
	switch suppliedType {
	case "PERSON", "INDIVIDUAL", "个人", "自然人", "ORGANIZATION", "组织", "其他组织", "LEGAL_PERSON", "COMPANY", "企业", "法人":
	default:
		return services.NewSubjectWorkflowError("INTAKE_PROTOCOL_INVALID", "未知的主体类型")
	}
	entityType := normalizeIntakeEntityType(suppliedType)
	identityType := strings.ToUpper(strings.TrimSpace(stringValue(payload["identity_type"], "")))
	identityNumber := security.NormalizeIdentityNumber(identityType, stringValue(payload["identity_number"], ""))
	if identityNumber == "" || !validIntakeIdentityType(entityType, identityType) || len([]rune(identityNumber)) < 4 {
		return services.NewSubjectWorkflowError("INTAKE_PARTY_IDENTITY_REQUIRED", "对方主体必须提供与主体类型匹配的可核验身份标识")
	}
	normalizedName := strings.ToLower(entityName)
	var rows []struct {
		ID           uint
		EntityType   string
		IdentityType string
		Cipher       string
		Digest       string
	}
	if err := tx.Table("case_intake_parties").
		Select("id, entity_type, identity_type, identity_number_ciphertext AS cipher, identity_number_digest AS digest").Where("intake_id = ?", intakeID).
		Where("LOWER(TRIM(party_role)) IN ?", []string{"opponent", "opposing_party"}).
		Where("LOWER(TRIM(entity_name)) = ?", normalizedName).
		Limit(2).Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return services.NewSubjectWorkflowError("INTAKE_PARTY_NOT_FOUND", "未找到匹配的对方主体")
	}
	if len(rows) > 1 {
		return services.NewSubjectWorkflowError("INTAKE_PARTY_AMBIGUOUS", "匹配到多个同名对方主体，已拒绝更新")
	}
	row := rows[0]
	if normalizeIntakeEntityType(row.EntityType) != entityType {
		return services.NewSubjectWorkflowError("INTAKE_CLIENT_TYPE_MISMATCH", "对方主体类型与保存值不一致，已拒绝更新")
	}
	nextDigest, digestErr := security.IdentityDigest(identityNumber)
	if digestErr != nil {
		return services.NewSubjectWorkflowError("SUBJECT_DATA_KEY_REQUIRED", "主体身份保护密钥不可用，已阻止保存对方主体")
	}
	// Preserve the existing ciphertext on an identical re-entry, including alias-only edits.
	sameIdentity := strings.EqualFold(strings.TrimSpace(row.IdentityType), identityType) && row.Cipher != "" && row.Digest == nextDigest
	updates := map[string]interface{}{}
	if !sameIdentity {
		ciphertext, digest, err := security.ProtectIdentityNumber(identityNumber)
		if err != nil {
			return services.NewSubjectWorkflowError("SUBJECT_DATA_KEY_REQUIRED", "主体身份保护密钥不可用，已阻止保存对方主体")
		}
		updates["identity_type"] = identityType
		updates["identity_number_ciphertext"] = ciphertext
		updates["identity_number_digest"] = digest
	}
	if aliasesSupplied {
		updates["aliases"] = payload["aliases"]
	}
	if len(updates) == 0 {
		return nil
	}
	result := tx.Table("case_intake_parties").Where("intake_id = ? AND id = ?", intakeID, row.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return services.NewSubjectWorkflowError("INTAKE_PARTY_NOT_FOUND", "对方主体发生变动，请重新打开草稿")
	}
	return nil
}

package handlers

import (
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"gorm.io/gorm"
)

// lookupCaseIntakePartyRows reads party rows for one intake whose stored
// name and role both match the given key. Name and role are bound as two
// separate parameters so the predicate stays portable across SQLite and
// MySQL: the previous "a || '|' || b" concatenation silently degraded to a
// logical OR on MySQL, where "||" is not string concatenation by default.
func (h *DemoAggregateHandler) lookupCaseIntakePartyRows(tx *gorm.DB, intakeID interface{}, entityName, partyRole string) ([]map[string]interface{}, error) {
	rows := []map[string]interface{}{}
	db := tx
	if db == nil {
		db = h.db
	}
	roles := []string{strings.ToLower(strings.TrimSpace(partyRole))}
	if roles[0] == "opponent" || roles[0] == "opposing_party" {
		roles = []string{"opponent", "opposing_party"}
	}
	err := db.Table("case_intake_parties").
		Where("intake_id = ?", intakeID).
		Where("LOWER(TRIM(entity_name)) = ?", strings.ToLower(strings.TrimSpace(entityName))).
		Where("LOWER(TRIM(party_role)) IN ?", roles).
		Find(&rows).Error
	return rows, err
}

// caseIntakePartyIdentityInputPresent reports whether the request payload
// carries an identity value. Draft saves clear the plaintext identity number
// before every write, so empty input means "unchanged" instead of "remove
// the protected identity". Only a payload with input is revalidated.
func caseIntakePartyIdentityInputPresent(party map[string]interface{}) bool {
	for _, key := range []string{"identity_number", "identityNumber"} {
		if value, ok := party[key]; ok && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
			return true
		}
	}
	return false
}

// Full replacement is also used by older native clients. Only fields actually
// supplied by that client may replace stored facts on the same party.
func (h *DemoAggregateHandler) preserveStoredParty(tx *gorm.DB, intakeID interface{}, row, party map[string]interface{}, requireStoredIdentity bool) error {
	invalid := func() error {
		return services.NewSubjectWorkflowError("INTAKE_PARTY_IDENTITY_REQUIRED", fmt.Sprintf("无法唯一核对当事人“%s”的身份，请重新核对", row["entity_name"]))
	}
	rows, err := h.lookupCaseIntakePartyRows(tx, intakeID, fmt.Sprint(row["entity_name"]), fmt.Sprint(row["party_role"]))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		if requireStoredIdentity {
			return invalid()
		}
		return nil
	}
	if len(rows) != 1 || normalizeIntakeEntityType(fmt.Sprint(rows[0]["entity_type"])) != fmt.Sprint(row["entity_type"]) {
		return invalid()
	}
	previous := rows[0]
	sameIdentity := valueString(previous, "identity_number_digest") == valueString(row, "identity_number_digest") &&
		strings.EqualFold(valueString(previous, "identity_type"), valueString(row, "identity_type")) && caseIntakePartyGuardedIdentityValid(previous)
	if requireStoredIdentity || sameIdentity {
		if !caseIntakePartyGuardedIdentityValid(previous) {
			return invalid()
		}
		for _, k := range []string{"identity_type", "identity_number_ciphertext", "identity_number_digest"} {
			row[k] = previous[k]
		}
	}
	return preserveIntakePartyFields(row, party, previous)
}

func preserveIntakePartyFields(row, party, previous map[string]interface{}) error {
	for _, k := range []string{"id", "case_id", "entity_id", "created_at", "party_role", "entity_type"} {
		if value, ok := previous[k]; ok {
			row[k] = value
		}
	}
	// Earlier native forms always sent 0 for clients/opponents and 1 for related parties.
	defaultDepth := "0"
	if row["party_role"] == "related_party" {
		defaultDepth = "1"
	}
	if depth, present := party["relation_depth"]; !present || fmt.Sprint(depth) == defaultDepth {
		row["relation_depth"] = previous["relation_depth"]
	}
	_, aliases := party["aliases"]
	_, legacyAlias := party["alias"]
	if !aliases && !legacyAlias {
		row["aliases"] = previous["aliases"]
	}
	if patch, present := party["metadata"]; present {
		incoming, ok := patch.(map[string]interface{})
		if !ok {
			return services.NewSubjectWorkflowError("INTAKE_PROTOCOL_INVALID", "主体元数据必须是 JSON 对象")
		}
		stored, err := decodeIntakeMetadata(previous["metadata"])
		if err != nil {
			return err
		}
		for key, value := range incoming {
			stored[key] = preserveBrowserNumbers(stored[key], value)
		}
		row["metadata"] = jsonStringValue(stored)
	} else {
		row["metadata"] = previous["metadata"]
	}
	return nil
}

// caseIntakePartyGuardedIdentityValid requires both halves of the protected
// identity pair; a row with only ciphertext or only digest cannot be reused.
func caseIntakePartyGuardedIdentityValid(row map[string]interface{}) bool {
	ciphertext := strings.TrimSpace(fmt.Sprint(row["identity_number_ciphertext"]))
	digest := strings.TrimSpace(fmt.Sprint(row["identity_number_digest"]))
	return ciphertext != "" && ciphertext != "<nil>" && digest != "" && digest != "<nil>"
}

// prepareCaseIntakePartyCreateRow is the create-path wrapper: a brand-new
// intake has no previous row to inherit from, so the payload must provide a
// fresh identity.
func (h *DemoAggregateHandler) prepareCaseIntakePartyCreateRow(party map[string]interface{}, intakeID interface{}, now time.Time) (map[string]interface{}, error) {
	row, err := prepareCaseIntakePartyRow(party, intakeID, now)
	if err != nil {
		return nil, err
	}
	if !caseIntakePartyGuardedIdentityValid(row) {
		return nil, services.NewSubjectWorkflowError("INTAKE_PARTY_IDENTITY_REQUIRED", fmt.Sprintf("当事人“%s”必须提供与主体类型匹配的可核验身份标识", row["entity_name"]))
	}
	return row, nil
}

// prepareCaseIntakeClientPartyRow resolves the client party from the
// PROTECTED client master record. The master-record read runs on the
// caller's transaction; authorization itself must have already passed
// authorizeIntakeClient for this client_id on this request. The row is
// only accepted when the guarded identity exists, the party name matches
// the selected client, and the payload's subject type agrees with the
// client master type. The identity never leaves the server as plaintext:
// it is copied in its guarded form (ciphertext + digest).
func (h *DemoAggregateHandler) prepareCaseIntakeClientPartyRow(tx *gorm.DB, clientIDInput interface{}, party map[string]interface{}, intakeID interface{}, now time.Time, sameClient bool) (map[string]interface{}, error) {
	if tx == nil {
		return nil, services.NewSubjectWorkflowError("INTAKE_PARTY_IDENTITY_REQUIRED", "客户方当事人缺少事务绑定，已阻止登记")
	}
	row, err := prepareCaseIntakePartyRow(party, intakeID, now)
	if err != nil {
		return nil, err
	}
	clientID := intakeClientID(clientIDInput)
	if clientID == 0 {
		return nil, services.NewSubjectWorkflowError("INTAKE_CLIENT_REQUIRED", "客户方当事人必须关联有效的客户档案")
	}
	// A full-save echo belongs to this intake's snapshot. Switching clients
	// or creating an intake instead resolves identity from the selected master.
	var previous map[string]interface{}
	if sameClient {
		var rows []map[string]interface{}
		if err := tx.Table("case_intake_parties").Where("intake_id = ? AND LOWER(TRIM(party_role)) = 'client'", intakeID).Limit(2).Find(&rows).Error; err != nil {
			return nil, err
		}
		if len(rows) > 1 {
			return nil, services.NewSubjectWorkflowError("INTAKE_CLIENT_MISMATCH", "客户方主体不唯一，请重新核对")
		}
		if len(rows) == 1 {
			previous = rows[0]
			if !strings.EqualFold(valueString(previous, "entity_name"), valueString(row, "entity_name")) || normalizeIntakeEntityType(valueString(previous, "entity_type")) != valueString(row, "entity_type") {
				return nil, services.NewSubjectWorkflowError("INTAKE_CLIENT_MISMATCH", "客户方主体与已保存事实不一致")
			}
			if caseIntakePartyGuardedIdentityValid(previous) && validIntakeIdentityType(valueString(row, "entity_type"), valueString(previous, "identity_type")) {
				for _, k := range []string{"identity_type", "identity_number_ciphertext", "identity_number_digest"} {
					row[k] = previous[k]
				}
				return row, preserveIntakePartyFields(row, party, previous)
			}
		}
	}
	var client models.Client
	if err := tx.First(&client, clientID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, services.NewSubjectWorkflowError("INTAKE_CLIENT_REQUIRED", "客户档案不存在，已阻止客户方当事人落库")
		}
		return nil, err
	}
	ciphertext := client.IdentityCiphertextValue()
	digest := client.IdentityDigestValue()
	if ciphertext == "" || digest == "" {
		return nil, services.NewSubjectWorkflowError("INTAKE_CLIENT_IDENTITY_REQUIRED", "客户主档案缺少受保护的可核验身份标识，不能登记客户方当事人")
	}
	clientName := strings.TrimSpace(client.Name)
	if clientName == "" || !strings.EqualFold(clientName, strings.TrimSpace(fmt.Sprint(row["entity_name"]))) {
		return nil, services.NewSubjectWorkflowError("INTAKE_CLIENT_MISMATCH", "客户方当事人名称与所选客户档案不一致，已阻止登记")
	}
	if !clientEntityTypeMatchesMaster(fmt.Sprint(row["entity_type"]), &client) {
		return nil, services.NewSubjectWorkflowError("INTAKE_CLIENT_TYPE_MISMATCH", "客户方当事人主体类型与客户档案不一致，已阻止登记")
	}
	row["identity_type"] = string(client.EffectiveIdentityType())
	row["identity_number_ciphertext"] = ciphertext
	row["identity_number_digest"] = digest
	if strings.TrimSpace(client.Aliases) != "" {
		row["aliases"] = client.Aliases
	}
	if previous != nil {
		return row, preserveIntakePartyFields(row, party, previous)
	}
	return row, nil
}

// clientEntityTypeMatchesMaster checks the payload subject type against the
// client master record's effective identity type, so a forged entity_type
// cannot reclassify a personal client as a legal person (or the reverse)
// on the server side.
func clientEntityTypeMatchesMaster(entityType string, client *models.Client) bool {
	normalized := normalizeIntakeEntityType(entityType)
	switch client.EffectiveIdentityType() {
	case models.IdentityTypeSocialCredit, models.IdentityTypeBusinessLicense, models.IdentityTypeOrgCode:
		return normalized == "LEGAL_PERSON"
	case models.IdentityTypeIDCard, models.IdentityTypePassport:
		return normalized == "INDIVIDUAL"
	}
	return false
}

// caseIntakePartyRoleValue is the single normalization point for the
// client-role decision: it accepts both party_role and the legacy role
// payload key so the two spellings cannot take different code paths.
func caseIntakePartyRoleValue(party map[string]interface{}) string {
	return strings.ToLower(strings.TrimSpace(stringValue(
		firstNonEmptyValue(party["party_role"], party["role"]),
		"related_party",
	)))
}

// caseIntakePartyIsClient reports whether the payload row is the client
// party, using the shared role normalization above.
func caseIntakePartyIsClient(party map[string]interface{}) bool {
	return caseIntakePartyRoleValue(party) == "client"
}

// prepareCaseIntakePartyUpdateRow is the update-path wrapper: with identity
// input it revalidates through prepareCaseIntakePartyRow; with empty input it
// inherits the protected identity from the matching previous row instead of
// wiping it.
func (h *DemoAggregateHandler) prepareCaseIntakePartyUpdateRow(tx *gorm.DB, intakeID interface{}, party map[string]interface{}, now time.Time, hasIdentityColumns bool) (map[string]interface{}, error) {
	row, err := prepareCaseIntakePartyRow(party, intakeID, now)
	if err != nil {
		return nil, err
	}
	fresh := caseIntakePartyIdentityInputPresent(party)
	if (fresh && !caseIntakePartyGuardedIdentityValid(row)) || (!fresh && !hasIdentityColumns) {
		return nil, services.NewSubjectWorkflowError("INTAKE_PARTY_IDENTITY_REQUIRED", "无法确认主体身份标识")
	}
	if err := h.preserveStoredParty(tx, intakeID, row, party, !fresh); err != nil {
		return nil, err
	}
	return row, nil
}

// replaceIntakePartyRoster reconciles an explicitly supplied full list. IDs
// in prepared rows came from matching the locked intake, never from the client.
func replaceIntakePartyRoster(tx *gorm.DB, intakeID string, rows []map[string]interface{}) error {
	retained := make(map[string]bool)
	var ids []string
	for _, row := range rows {
		if id := valueString(row, "id"); id != "" {
			if retained[id] {
				return services.NewSubjectWorkflowError("INTAKE_PARTY_AMBIGUOUS", "提交的主体清单重复指向同一条记录")
			}
			retained[id] = true
			ids = append(ids, id)
		}
	}
	removed := tx.Table("case_intake_parties").Where("intake_id = ?", intakeID)
	if len(ids) > 0 {
		removed = removed.Where("id NOT IN ?", ids)
	}
	if err := removed.Delete(map[string]interface{}{}).Error; err != nil {
		return err
	}
	for _, row := range rows {
		parties := tx.Table("case_intake_parties")
		if id := row["id"]; id != nil {
			result := parties.Where("intake_id = ? AND id = ?", intakeID, id).Omit("id", "intake_id", "case_id", "entity_id", "created_at").Updates(row)
			if result.Error != nil {
				return result.Error
			}
		} else if err := parties.Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

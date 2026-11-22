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
	err := db.Table("case_intake_parties").
		Where("intake_id = ?", intakeID).
		Where("LOWER(TRIM(entity_name)) = ?", strings.ToLower(strings.TrimSpace(entityName))).
		Where("LOWER(TRIM(party_role)) = ?", strings.ToLower(strings.TrimSpace(partyRole))).
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

// inheritCaseIntakePartyIdentity copies the protected identity of the matching
// previous row into the replacement row when the payload carries no identity.
// It fails closed unless exactly one same-name, same-role row remains, the
// entity type also matches, and that row carries both ciphertext and digest.
// Genuinely new parties still require a fresh identity.
func (h *DemoAggregateHandler) inheritCaseIntakePartyIdentity(tx *gorm.DB, intakeID interface{}, row map[string]interface{}, party map[string]interface{}, hasIdentityColumns bool) bool {
	if caseIntakePartyIdentityInputPresent(party) {
		return caseIntakePartyGuardedIdentityValid(row)
	}
	if !hasIdentityColumns {
		return false
	}
	rows, err := h.lookupCaseIntakePartyRows(tx, intakeID, fmt.Sprint(row["entity_name"]), fmt.Sprint(row["party_role"]))
	if err != nil {
		return false
	}
	if len(rows) != 1 {
		// Ambiguous multi-match inheritance would pick an arbitrary protected
		// identity, so the save must fail closed and ask for fresh input.
		return false
	}
	existing := rows[0]
	if fmt.Sprint(existing["entity_type"]) != fmt.Sprint(row["entity_type"]) || !caseIntakePartyGuardedIdentityValid(existing) {
		return false
	}
	row["identity_type"] = existing["identity_type"]
	row["identity_number_ciphertext"] = existing["identity_number_ciphertext"]
	row["identity_number_digest"] = existing["identity_number_digest"]
	if aliases := existing["aliases"]; aliases != nil {
		row["aliases"] = aliases
	}
	return true
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
	if !caseIntakePartyIdentityInputPresent(party) {
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
func (h *DemoAggregateHandler) prepareCaseIntakeClientPartyRow(tx *gorm.DB, clientIDInput interface{}, party map[string]interface{}, intakeID interface{}, now time.Time) (map[string]interface{}, error) {
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
	if !h.inheritCaseIntakePartyIdentity(tx, intakeID, row, party, hasIdentityColumns) {
		return nil, services.NewSubjectWorkflowError("INTAKE_PARTY_IDENTITY_REQUIRED", fmt.Sprintf("当事人“%s”必须提供与主体类型匹配的可核验身份标识", row["entity_name"]))
	}
	return row, nil
}

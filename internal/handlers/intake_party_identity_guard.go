package handlers

import (
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/services"

	"gorm.io/gorm"
)

// A party row is matched by the natural identity used by the intake page:
// name plus role within one intake. Update payloads replace the whole party
// set, so this triple identifies the previous protected row.
func caseIntakePartyIdentityKey(entityName, partyRole string) string {
	return strings.ToLower(strings.TrimSpace(entityName)) + "|" + strings.ToLower(strings.TrimSpace(partyRole))
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
// It returns true when the resulting row has a usable protected identity.
// Genuinely new parties still require a fresh identity.
func (h *DemoAggregateHandler) inheritCaseIntakePartyIdentity(tx *gorm.DB, intakeID interface{}, row map[string]interface{}, party map[string]interface{}, hasIdentityColumns bool) bool {
	if caseIntakePartyIdentityInputPresent(party) {
		ciphertext := strings.TrimSpace(fmt.Sprint(row["identity_number_ciphertext"]))
		return ciphertext != "" && ciphertext != "<nil>"
	}
	if !hasIdentityColumns {
		return false
	}
	key := caseIntakePartyIdentityKey(fmt.Sprint(row["entity_name"]), fmt.Sprint(row["party_role"]))
	rows := []map[string]interface{}{}
	if err := tx.Table("case_intake_parties").
		Where("intake_id = ?", intakeID).
		Where("LOWER(TRIM(entity_name)) || '|' || LOWER(TRIM(party_role)) = ?", key).
		Find(&rows).Error; err != nil {
		return false
	}
	for _, existing := range rows {
		ciphertext := strings.TrimSpace(fmt.Sprint(existing["identity_number_ciphertext"]))
		if ciphertext == "" || ciphertext == "<nil>" {
			continue
		}
		row["identity_type"] = existing["identity_type"]
		row["identity_number_ciphertext"] = existing["identity_number_ciphertext"]
		row["identity_number_digest"] = existing["identity_number_digest"]
		if aliases := existing["aliases"]; aliases != nil {
			row["aliases"] = aliases
		}
		return true
	}
	return false
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

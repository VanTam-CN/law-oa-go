package handlers

import (
	"strings"

	"gorm.io/gorm"
)

// The approval snapshot has its own allowlist. The editable detail view is paginated.
func approvalIntakeParties(tx *gorm.DB, intakeID string, clientID interface{}) ([]map[string]string, error) {
	var rows []struct {
		Role, Name, EntityType, IdentityType string
		Metadata                             []byte
	}
	err := tx.Table("case_intake_parties").Select("party_role AS role, entity_name AS name, entity_type, identity_type, metadata").Where("intake_id = ?", intakeID).Order("id").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	parties := make([]map[string]string, 0, len(rows)+1)
	hasClient := false
	for _, row := range rows {
		role := strings.ToLower(strings.TrimSpace(row.Role))
		if role == "opponent" {
			role = "opposing_party"
		}
		p := map[string]string{"role": role, "name": row.Name, "entity_type": row.EntityType, "identity_type": row.IdentityType}
		if role == "client" {
			hasClient = true
		}
		meta := integrationJSONMap(row.Metadata)
		for _, key := range []string{"party_type", "role"} {
			if label, ok := meta[key].(string); ok && strings.TrimSpace(label) != "" {
				p["party_type"] = label
				break
			}
		}
		parties = append(parties, p)
	}
	if !hasClient && clientID != nil {
		var client struct{ Name string }
		if err := tx.Table("clients").Select("name").Where("id = ?", clientID).Take(&client).Error; err != nil {
			return nil, err
		}
		parties = append([]map[string]string{{"role": "client", "name": client.Name}}, parties...)
	}
	return parties, nil
}

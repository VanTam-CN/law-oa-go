package handlers

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"

	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func TestUpdateIntakeNarrowIdentityRejectsUncertainTarget(t *testing.T) {
	for _, tc := range []struct {
		name      string
		mutate    func(map[string]any)
		duplicate bool
		status    int
	}{
		{"missing party", func(p map[string]any) { p["entity_name"] = "absent" }, false, 409},
		{"ambiguous party", nil, true, 409},
		{"wrong entity type", func(p map[string]any) { p["entity_type"] = "INDIVIDUAL"; p["identity_type"] = "PASSPORT" }, false, 409},
		{"non-string identity", func(p map[string]any) { p["identity_number"] = 1234 }, false, 409},
		{"non-string aliases", func(p map[string]any) { p["aliases"] = 1234 }, false, 409},
		{"unknown field", func(p map[string]any) { p["truly_unknown"] = "not allowed" }, false, 409},
		{"mixed list", nil, false, 400},
		{"mixed null", nil, false, 400},
		{"null update", nil, false, 409},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := newPartyIdentityGuardDB(t)
			if tc.duplicate {
				if err := db.Exec("INSERT INTO case_intake_parties (intake_id,entity_name,entity_type,party_role) VALUES ('intake-guarded','对手公司','LEGAL_PERSON','opponent')").Error; err != nil {
					t.Fatal(err)
				}
			}
			h := NewDemoAggregateHandler(db)
			h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
			g := gin.New()
			g.PUT("/api/v1/case-intakes/:id", func(c *gin.Context) { c.Set("role", "lawyer"); c.Set("user_id", "7"); h.UpdateCaseIntake(c) })
			operand := map[string]any{"entity_name": "对手公司", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": "TEST1234"}
			if tc.mutate != nil {
				tc.mutate(operand)
			}
			input := map[string]any{"opponent_identity_update": operand, "title": "must roll back"}
			switch tc.name {
			case "mixed list":
				input["parties"] = []any{}
			case "mixed null":
				input["materials"] = nil
			case "null update":
				input["opponent_identity_update"] = nil
			}
			body, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			status, _ := partyIdentityGuardRequest(t, g, string(body))
			if status != tc.status {
				t.Fatalf("native response %d, expected %d", status, tc.status)
			}
			var title string
			if err := db.Raw("SELECT title FROM case_intakes WHERE id='intake-guarded'").Scan(&title).Error; err != nil || title != "被保护草稿" {
				t.Fatalf("parent write survived rejection: %v", err)
			}
		})
	}
}

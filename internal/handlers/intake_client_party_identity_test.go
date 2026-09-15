package handlers

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

func newClientPartyIdentityFixture(t *testing.T) (*DemoAggregateHandler, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Client{}); err != nil {
		t.Fatalf("migrate client: %v", err)
	}
	return NewDemoAggregateHandler(db), db
}

func seedProtectedClient(t *testing.T, db *gorm.DB, id uint, name, clientType string, identityType models.IdentityType) models.Client {
	t.Helper()
	identityValue := "91310000MA1FL0000X"
	if identityType == models.IdentityTypeIDCard {
		identityValue = "310110199001010011"
	}
	ciphertext, digest, err := security.ProtectIdentityNumber(security.NormalizeIdentityNumber(string(identityType), identityValue))
	if err != nil {
		t.Fatalf("protect identity: %v", err)
	}
	client := models.Client{
		Name:                    name,
		Type:                    clientType,
		Email:                   "client@example.com",
		IdentityType:            identityType,
		IdentityNumberCiphertext: ciphertext,
		IdentityNumberDigest:    digest,
	}
	client.ID = id
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	return client
}

// The client party row inherits the guarded identity from the selected
// client master record; the browser never has to send it.
func TestClientPartyRowInheritsProtectedIdentity(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("s", 32))
	h, db := newClientPartyIdentityFixture(t)
	client := seedProtectedClient(t, db, 7, "星河智联有限公司", "企业", models.IdentityTypeSocialCredit)

	row, err := h.prepareCaseIntakeClientPartyRow(
		db,
		uint64(client.ID),
		map[string]interface{}{"entity_name": client.Name, "party_role": "client", "entity_type": "LEGAL_PERSON"},
		"intake-1", time.Now(),
	)
	if err != nil {
		t.Fatalf("prepare client party: %v", err)
	}
	if row["identity_number_ciphertext"] == "" || row["identity_number_digest"] == "" {
		t.Fatalf("client party row must carry the protected identity; row = %v", row)
	}
	if row["identity_number_digest"] != client.IdentityDigestValue() {
		t.Fatalf("digest mismatch: row = %v, client = %s", row["identity_number_digest"], client.IdentityDigestValue())
	}
}

// The party name must match the selected client: an arbitrary client_id
// cannot smuggle a different party name through the master-record path.
func TestClientPartyRowRejectsNameMismatch(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("s", 32))
	h, db := newClientPartyIdentityFixture(t)
	client := seedProtectedClient(t, db, 8, "星河智联有限公司", "企业", models.IdentityTypeSocialCredit)

	_, err := h.prepareCaseIntakeClientPartyRow(
		db,
		uint64(client.ID),
		map[string]interface{}{"entity_name": "别的主体有限公司", "party_role": "client", "entity_type": "LEGAL_PERSON"},
		"intake-2", time.Now(),
	)
	if err == nil {
		t.Fatalf("name mismatch must be rejected")
	}
	if !contains(err.Error(), "不一致") {
		t.Fatalf("expected mismatch error, got: %v", err)
	}
}

// A client master record without a guarded identity stays rejected.
func TestClientPartyRowRejectsClientWithoutIdentity(t *testing.T) {
	h, db := newClientPartyIdentityFixture(t)
	client := models.Client{Name: "无身份客户有限公司", Type: "企业", Email: "noid@example.com"}
	db.Create(&client)

	_, err := h.prepareCaseIntakeClientPartyRow(
		db,
		uint64(client.ID),
		map[string]interface{}{"entity_name": client.Name, "party_role": "client", "entity_type": "LEGAL_PERSON"},
		"intake-3", time.Now(),
	)
	if err == nil || !contains(err.Error(), "缺少受保护的可核验身份标识") {
		t.Fatalf("client without identity must be rejected, got: %v", err)
	}
}

// The create-path guard for non-client parties keeps requiring fresh
// identity input: the new client-inheritance path must not weaken it.
func TestOpposingPartyStillRequiresIdentity(t *testing.T) {
	h, _ := newClientPartyIdentityFixture(t)
	_, err := h.prepareCaseIntakePartyCreateRow(
		map[string]interface{}{"entity_name": "对手方有限公司", "party_role": "opposing_party", "entity_type": "LEGAL_PERSON"},
		"intake-4", time.Now(),
	)
	if err == nil || !contains(err.Error(), "必须提供与主体类型匹配的可核验身份标识") {
		t.Fatalf("opposing party without identity must still be rejected, got: %v", err)
	}
}

// A personal client whose payload forges entity_type LEGAL_PERSON must be
// rejected server-side; the frontend hint is not a trust boundary.
func TestClientPartyRowRejectsForgedEntityType(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("s", 32))
	h, db := newClientPartyIdentityFixture(t)
	client := seedProtectedClient(t, db, 11, "个人客户姓名", "个人", models.IdentityTypeIDCard)

	_, err := h.prepareCaseIntakeClientPartyRow(
		db,
		uint64(client.ID),
		map[string]interface{}{"entity_name": client.Name, "party_role": "client", "entity_type": "LEGAL_PERSON"},
		"intake-5", time.Now(),
	)
	if err == nil || !contains(err.Error(), "主体类型与客户档案不一致") {
		t.Fatalf("forged LEGAL_PERSON for a personal client must be rejected, got: %v", err)
	}
}

// A corporate client whose payload forges entity_type INDIVIDUAL must also
// be rejected server-side.
func TestClientPartyRowRejectsForgedIndividualTypeOnCompany(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("s", 32))
	h, db := newClientPartyIdentityFixture(t)
	client := seedProtectedClient(t, db, 12, "星河智联有限公司", "企业", models.IdentityTypeSocialCredit)

	_, err := h.prepareCaseIntakeClientPartyRow(
		db,
		uint64(client.ID),
		map[string]interface{}{"entity_name": client.Name, "party_role": "client", "entity_type": "INDIVIDUAL"},
		"intake-6", time.Now(),
	)
	if err == nil || !contains(err.Error(), "主体类型与客户档案不一致") {
		t.Fatalf("forged INDIVIDUAL for a corporate client must be rejected, got: %v", err)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

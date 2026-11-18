package services

import (
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func newMatterCollabServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	return newMatterCollabServiceDBWithTables(t, true)
}

// newMatterCollabServiceDBWithTables builds the service database with or
// without the material backing tables so the resubmit gate can be tested
// for both the uploaded-material contract and the fail-closed schema rule.
func newMatterCollabServiceDBWithTables(t *testing.T, withMaterialTables bool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "svc.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE approval_requests (id TEXT PRIMARY KEY, metadata TEXT, current_approver_id TEXT, status TEXT, applicant_id TEXT, created_at DATETIME, updated_at DATETIME)`,
		`INSERT INTO approval_requests (id, metadata) VALUES ('ap-m', '{"intake_id":"intake-own"}'), ('ap-other', '{"intake_id":"intake-other"}')`,
	}
	if withMaterialTables {
		stmts = append(stmts,
			`CREATE TABLE case_materials (id INTEGER PRIMARY KEY, case_id INTEGER, intake_id TEXT, name TEXT, material_type TEXT, status TEXT DEFAULT 'missing', required BOOLEAN DEFAULT 1, storage_url TEXT, owner_id INTEGER, uploaded_by INTEGER, uploaded_at DATETIME, metadata TEXT DEFAULT '{}', created_at DATETIME, updated_at DATETIME)`,
			`CREATE TABLE material_files (id INTEGER PRIMARY KEY, material_id INTEGER, content_sha256 TEXT, file_name TEXT, size_bytes INTEGER, content_type TEXT, storage_path TEXT, uploaded_by INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
			`INSERT INTO case_materials (id, intake_id, name, status) VALUES (11, 'intake-own', 'POA', 'uploaded'), (12, 'intake-own', 'ID-copy', 'missing'), (99, 'intake-other', 'Other-intake', 'uploaded')`,
			`INSERT INTO material_files (id, material_id, content_sha256, file_name, size_bytes, storage_path, uploaded_by) VALUES (5, 11, 'digest-5', 'poa.pdf', 10, '/x/poa.pdf', 1)`,
		)
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("seed: %v: %v", s, err)
		}
	}
	return db
}

// Defect 7 regression: a material_request in a resubmit round must point
// at a material of the SAME intake that is genuinely uploaded (status plus
// at least one live attachment). Ownership alone no longer passes, and a
// soft-deleted file row no longer counts as uploaded.
func TestValidateSupplementMaterialUploadsTx(t *testing.T) {
	db := newMatterCollabServiceDB(t)
	uploaded := []SupplementItem{{Category: "material_request", MaterialID: 11}}
	missing := []SupplementItem{{Category: "material_request", MaterialID: 12}}
	if err := ValidateSupplementMaterialUploadsTx(db, "intake-own", uploaded); err != nil {
		t.Fatalf("uploaded own-intake material rejected: %v", err)
	}
	if err := ValidateSupplementMaterialUploadsTx(db, "intake-own", missing); err == nil {
		t.Fatalf("not-yet-uploaded material accepted")
	}
	// Cross-intake ownership still fails even when that material IS
	// uploaded (material 99 belongs to intake-other).
	other := []SupplementItem{{Category: "material_request", MaterialID: 99}}
	if err := ValidateSupplementMaterialUploadsTx(db, "intake-own", other); err == nil {
		t.Fatalf("cross-intake uploaded material accepted")
	}
	// A soft-deleted attachment no longer satisfies the gate.
	db.Exec("UPDATE material_files SET deleted_at = CURRENT_TIMESTAMP WHERE material_id = 11")
	if err := ValidateSupplementMaterialUploadsTx(db, "intake-own", uploaded); err == nil {
		t.Fatalf("soft-deleted attachment still satisfied upload gate")
	}
}

func TestFreezeMaterialFingerprintsForSnapshot(t *testing.T) {
	db := newMatterCollabServiceDB(t)
	fps, err := freezeMaterialFingerprintsForSnapshot(db, "intake-own")
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if len(fps) != 2 {
		t.Fatalf("got %d fingerprints, want 2", len(fps))
	}
	if fps[0].MaterialID != 11 || fps[0].Status != "uploaded" || len(fps[0].Files) != 1 || fps[0].Files[0].Digest != "digest-5" {
		t.Fatalf("fingerprint[0] = %+v", fps[0])
	}
	if fps[1].MaterialID != 12 || len(fps[1].Files) != 0 {
		t.Fatalf("fingerprint[1] = %+v", fps[1])
	}
}

func TestValidateSupplementMaterialItemsOwnership(t *testing.T) {
	db := newMatterCollabServiceDB(t)

	// Own-intake reference passes.
	own := map[string]interface{}{"items": []interface{}{map[string]interface{}{"category": "material_request", "field_key": "material", "material_id": 11}}}
	if err := ValidateSupplementMaterialItems(db, "ap-m", own); err != nil {
		t.Fatalf("own intake ref rejected: %v", err)
	}
	// Cross-application reference (material of another intake) rejected.
	other := map[string]interface{}{"items": []interface{}{map[string]interface{}{"category": "material_request", "field_key": "material", "material_id": 99}}}
	if err := ValidateSupplementMaterialItems(db, "ap-m", other); err == nil {
		t.Fatalf("cross-intake ref accepted")
	}
	// Missing material_id rejected.
	noid := map[string]interface{}{"items": []interface{}{map[string]interface{}{"category": "material_request", "field_key": "material"}}}
	if err := ValidateSupplementMaterialItems(db, "ap-m", noid); err == nil {
		t.Fatalf("missing material_id accepted")
	}
	// A non-intake approval cannot carry material_request at all.
	bad := map[string]interface{}{"items": []interface{}{map[string]interface{}{"category": "material_request", "field_key": "material", "material_id": 11}}}
	if err := ValidateSupplementMaterialItems(db, "missing-approval", bad); err == nil {
		t.Fatalf("material_request without approval accepted")
	}
}

func TestSupplementMaterialItemContractValidation(t *testing.T) {
	// material_request without field_key=material rejected.
	if _, err := ValidateSupplementItems(suppRaw("material_request", "case_summary", 11)); err == nil {
		t.Fatalf("material_request with wrong field_key accepted")
	}
	// material_request without material_id rejected.
	if _, err := ValidateSupplementItems(suppRaw("material_request", "material", 0)); err == nil {
		t.Fatalf("material_request without material_id accepted")
	}
	// field_key=material on a non-material category rejected.
	if _, err := ValidateSupplementItems(suppRaw("matter_explanation", "material", 11)); err == nil {
		t.Fatalf("material field_key on other category accepted")
	}
	// Valid material_request passes and preserves the id.
	items, err := ValidateSupplementItems(suppRaw("material_request", "material", 11))
	if err != nil {
		t.Fatalf("valid material_request rejected: %v", err)
	}
	if items[0].MaterialID != 11 {
		t.Fatalf("MaterialID = %d, want 11", items[0].MaterialID)
	}
}

func suppRaw(category, fieldKey string, materialID uint) map[string]interface{} {
	item := map[string]interface{}{
		"category":    category,
		"field_key":   fieldKey,
		"description": "请补充材料",
	}
	if materialID > 0 {
		item["material_id"] = materialID
	}
	return map[string]interface{}{
		"schema_version": 1,
		"items":          []interface{}{item},
	}
}

// Fail-closed review 2026-10-01: a material_request with the material
// backing tables missing must be REJECTED, not silently waived. Only a
// round with no material_request items keeps the text-only legacy
// contract (nil) when the schema is absent.
func TestValidateSupplementMaterialUploadsMissingTablesFailsClosed(t *testing.T) {
	db := newMatterCollabServiceDBWithTables(t, false)
	requested := []SupplementItem{{Category: "material_request", MaterialID: 11}}
	err := ValidateSupplementMaterialUploadsTx(db, "intake-own", requested)
	if err == nil {
		t.Fatalf("material_request accepted while case_materials/material_files tables are missing")
	}
	var contractErr *ResubmitContractError
	if !errors.As(err, &contractErr) {
		t.Fatalf("expected ResubmitContractError, got %T: %v", err, err)
	}
	if contractErr.Code != models.ErrCodeSupplementItemInvalid {
		t.Fatalf("expected code %s, got %s", models.ErrCodeSupplementItemInvalid, contractErr.Code)
	}
	// Text-only rounds (no material_request) still pass on the slim schema.
	if err := ValidateSupplementMaterialUploadsTx(db, "intake-own", []SupplementItem{{Category: "matter_explanation", MaterialID: 0}}); err != nil {
		t.Fatalf("text-only round rejected on slim schema: %v", err)
	}
}

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// setupMaterialFileHistoryRouter builds a handler with its own DB and
// storage root and wires ONLY the material endpoints this file exercises,
// so task or router regressions can never mask material history results.
func setupMaterialFileHistoryRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := newMatterCollabDB(t)
	h := NewMatterCollaborationHandlerWithDB(db)
	h.StorageDir = t.TempDir()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.GET("/api/v1/case-intakes/:id/materials", h.ListIntakeMaterials)
	r.POST("/api/v1/case-intakes/:id/materials", h.CreateIntakeMaterial)
	r.GET("/api/v1/cases/:id/materials", h.ListCaseMaterials)
	r.POST("/api/v1/cases/:id/materials", h.CreateCaseMaterial)
	r.PUT("/api/v1/case-materials/:materialId", h.UpdateCaseMaterial)
	r.GET("/api/v1/case-materials/:materialId/files", h.ListCaseMaterialFiles)
	r.POST("/api/v1/case-materials/:materialId/files", h.UploadCaseMaterial)
	r.GET("/api/v1/case-materials/:materialId/files/:fileId/download", h.DownloadCaseMaterialFile)
	return r, db
}

func decodeMaterialFilesBody(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var payload struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v: %s", err, body)
	}
	return payload.Data
}

func TestMaterialFileHistoryListAndVersions(t *testing.T) {
	r, db := setupMaterialFileHistoryRouter(t)

	// Two uploads on the same material row: both files must survive and
	// the newer file becomes the current version.
	for i, content := range []string{"OLD", "NEW"} {
		w := doMultipartFile(r, "/api/v1/case-materials/200/files", "1", "lawyer", "v"+string(rune('1'+i))+".txt", content)
		if w.Code != http.StatusOK {
			t.Fatalf("upload v%d: got %d: %s", i+1, w.Code, w.Body.String())
		}
	}
	var fileCount int64
	db.Table("material_files").Where("material_id = ?", 200).Count(&fileCount)
	if fileCount != 2 {
		t.Fatalf("history rows = %d, want 2 (old files must not be overwritten)", fileCount)
	}

	// The dedicated history endpoint lists both, oldest first, with
	// 1-based versions and a safe download URL per file.
	w := doJSON(r, http.MethodGet, "/api/v1/case-materials/200/files", "1", "lawyer", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list history: got %d: %s", w.Code, w.Body.String())
	}
	data := decodeMaterialFilesBody(t, w.Body.Bytes())
	files, _ := data["files"].([]interface{})
	if len(files) != 2 {
		t.Fatalf("history files = %d, want 2", len(files))
	}
	first, _ := files[0].(map[string]interface{})
	second, _ := files[1].(map[string]interface{})
	if first["version"].(float64) != 1 || second["version"].(float64) != 2 {
		t.Fatalf("versions = %v, %v; want 1, 2", first["version"], second["version"])
	}
	if first["download_url"] != "/api/v1/case-materials/200/files/1/download" {
		t.Fatalf("download url = %v", first["download_url"])
	}
	if bytes.Contains(w.Body.Bytes(), []byte("storage_path")) {
		t.Fatalf("history response leaked storage_path: %s", w.Body.String())
	}

	// Case list enrichment exposes the current version metadata for the
	// same row without extra client requests.
	lw := doJSON(r, http.MethodGet, "/api/v1/cases/8/materials", "1", "lawyer", nil)
	if lw.Code != http.StatusOK {
		t.Fatalf("list case materials: got %d: %s", lw.Code, lw.Body.String())
	}
	ldata := decodeMaterialFilesBody(t, lw.Body.Bytes())
	materials, _ := ldata["materials"].([]interface{})
	if len(materials) != 1 {
		t.Fatalf("materials = %d, want 1", len(materials))
	}
	mat, _ := materials[0].(map[string]interface{})
	if mat["file_version"].(float64) != 2 || mat["file_count"].(float64) != 2 {
		t.Fatalf("file metadata = %v/%v, want version 2 count 2", mat["file_version"], mat["file_count"])
	}
	current, _ := mat["current_file"].(map[string]interface{})
	if current == nil || current["file_name"] != "v2.txt" {
		t.Fatalf("current file = %v, want v2.txt", mat["current_file"])
	}
}

func TestMaterialFileHistoryPermission(t *testing.T) {
	r, _ := setupMaterialFileHistoryRouter(t)
	// lawyerB (id 2) owns case 1 only: history of case 8 material fails.
	if w := doJSON(r, http.MethodGet, "/api/v1/case-materials/200/files", "2", "lawyer", nil); w.Code != http.StatusForbidden {
		t.Fatalf("cross-case history: got %d want 403", w.Code)
	}
	// Intake history follows the intake boundary: unrelated lawyer 2 is
	// rejected; creator 1 reads the same endpoint.
	intake := "aaaa1111-0000-0000-0000-000000000001"
	if w := doJSON(r, http.MethodGet, "/api/v1/case-materials/100/files", "2", "lawyer", nil); w.Code != http.StatusForbidden {
		t.Fatalf("unrelated intake history: got %d want 403", w.Code)
	}
	if w := doJSON(r, http.MethodGet, "/api/v1/case-materials/100/files", "1", "lawyer", nil); w.Code != http.StatusOK {
		t.Fatalf("creator intake history: got %d: %s", w.Code, w.Body.String())
	}
	_ = intake
}

func TestCreateCaseMaterialOwnerMustBeCaseMember(t *testing.T) {
	r, db := setupMaterialFileHistoryRouter(t)
	// lawyerB (id 2) is outside case 8: assigning him as owner is refused
	// server-side even though the writer is a legal case manager.
	w := doJSON(r, http.MethodPost, "/api/v1/cases/8/materials", "1", "lawyer", map[string]interface{}{
		"name": "责任材料", "owner_id": 2,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("invalid owner create: got %d want 403: %s", w.Code, w.Body.String())
	}
	var count int64
	db.Table("case_materials").Where("case_id = ? AND name = ?", 8, "责任材料").Count(&count)
	if count != 0 {
		t.Fatalf("invalid owner row was created")
	}
	// The case lawyer (id 1) is a legal owner and the row is created with
	// the assignment applied.
	okW := doJSON(r, http.MethodPost, "/api/v1/cases/8/materials", "1", "lawyer", map[string]interface{}{
		"name": "责任材料", "owner_id": 1,
	})
	if okW.Code != http.StatusOK {
		t.Fatalf("valid owner create: got %d: %s", okW.Code, okW.Body.String())
	}
	var ownerID int
	db.Raw("SELECT owner_id FROM case_materials WHERE case_id = ? AND name = ?", 8, "责任材料").Scan(&ownerID)
	if ownerID != 1 {
		t.Fatalf("owner_id = %d, want 1", ownerID)
	}
}

func TestIntakeMaterialListEnrichmentAndVersionUpload(t *testing.T) {
	r, db := setupMaterialFileHistoryRouter(t)
	w := doMultipartFile(r, "/api/v1/case-materials/100/files", "1", "lawyer", "intake-v1.txt", "A")
	if w.Code != http.StatusOK {
		t.Fatalf("intake upload: got %d: %s", w.Code, w.Body.String())
	}
	// A second upload on the same intake row is a new version; the old
	// file row survives.
	w2 := doMultipartFile(r, "/api/v1/case-materials/100/files", "1", "lawyer", "intake-v2.txt", "B")
	if w2.Code != http.StatusOK {
		t.Fatalf("intake second upload: got %d: %s", w2.Code, w2.Body.String())
	}
	var count int64
	db.Model(&models.MaterialFile{}).Where("material_id = ?", 100).Count(&count)
	if count != 2 {
		t.Fatalf("intake history rows = %d, want 2", count)
	}
	lw := doJSON(r, http.MethodGet, "/api/v1/case-intakes/aaaa1111-0000-0000-0000-000000000001/materials", "1", "lawyer", nil)
	if lw.Code != http.StatusOK {
		t.Fatalf("intake list: got %d: %s", lw.Code, lw.Body.String())
	}
	ldata := decodeMaterialFilesBody(t, lw.Body.Bytes())
	materials, _ := ldata["materials"].([]interface{})
	if len(materials) != 1 {
		t.Fatalf("intake materials = %d, want 1", len(materials))
	}
	mat, _ := materials[0].(map[string]interface{})
	if mat["file_version"].(float64) != 2 {
		t.Fatalf("intake file_version = %v, want 2", mat["file_version"])
	}
}

func TestMaterialListCurrentPointerMustMatchOwnFile(t *testing.T) {
	tests := []struct {
		name        string
		pointer     string
		deletedFile bool
		wantID      uint
		wantVersion int
		wantCount   int
	}{
		{"missing pointer may use newest", "", false, 2, 2, 2},
		{"blank pointer may use newest", " \t ", false, 2, 2, 2},
		{"valid older pointer stays older", "/api/v1/case-materials/200/files/1/download", false, 1, 1, 2},
		{"valid current pointer", "/api/v1/case-materials/200/files/2/download", false, 2, 2, 2},
		{"pointer to another material", "/api/v1/case-materials/100/files/3/download", false, 0, 0, 2},
		{"pointer to another material file under this row", "/api/v1/case-materials/200/files/3/download", false, 0, 0, 2},
		{"pointer to nonexistent file", "/api/v1/case-materials/200/files/999/download", false, 0, 0, 2},
		{"pointer with unexpected query", "/api/v1/case-materials/200/files/2/download?other=1", false, 0, 0, 2},
		{"pointer to unavailable record", "/api/v1/case-materials/200/files/2/download", true, 0, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, db := setupMaterialFileHistoryRouter(t)
			for _, path := range []string{
				"/api/v1/case-materials/200/files",
				"/api/v1/case-materials/200/files",
				"/api/v1/case-materials/100/files",
			} {
				if w := doMultipartFile(r, path, "1", "lawyer", "fictional.txt", "FICTIONAL"); w.Code != http.StatusOK {
					t.Fatalf("upload fixture: got %d: %s", w.Code, w.Body.String())
				}
			}
			if tt.deletedFile {
				if err := db.Exec("UPDATE material_files SET deleted_at = CURRENT_TIMESTAMP WHERE id = 2").Error; err != nil {
					t.Fatal(err)
				}
			}
			if err := db.Model(&models.CaseMaterial{}).Where("id = 200").Update("storage_url", tt.pointer).Error; err != nil {
				t.Fatal(err)
			}
			w := doJSON(r, http.MethodGet, "/api/v1/cases/8/materials", "1", "lawyer", nil)
			if w.Code != http.StatusOK {
				t.Fatalf("material list: got %d: %s", w.Code, w.Body.String())
			}
			data := decodeMaterialFilesBody(t, w.Body.Bytes())
			rows, ok := data["materials"].([]interface{})
			if !ok || len(rows) != 1 {
				t.Fatalf("expected one material, got %v", data)
			}
			row := rows[0].(map[string]interface{})
			if row["file_version"] != float64(tt.wantVersion) || row["file_count"] != float64(tt.wantCount) {
				t.Fatalf("version/count = %v/%v, want %d/%d", row["file_version"], row["file_count"], tt.wantVersion, tt.wantCount)
			}
			current, present := row["current_file"]
			if tt.wantID == 0 {
				if present {
					t.Fatalf("invalid explicit pointer must omit current_file, got %v", current)
				}
			} else {
				file, ok := current.(map[string]interface{})
				if !ok || file["id"] != float64(tt.wantID) {
					t.Fatalf("current_file = %v, want id %d", current, tt.wantID)
				}
			}
			var stored models.CaseMaterial
			if err := db.Take(&stored, 200).Error; err != nil {
				t.Fatal(err)
			}
			if stored.StorageURL != tt.pointer {
				t.Fatalf("list mutated pointer: %q, want %q", stored.StorageURL, tt.pointer)
			}
		})
	}
}

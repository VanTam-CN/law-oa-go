package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

// Independent review covers byte retrieval, not just list metadata: a new
// current pointer must not make the former attachment disappear or bypass
// the same case-level read boundary.
func TestIndependentMaterialVersionsDownloadOriginalBytes(t *testing.T) {
	router, _ := setupMaterialFileHistoryRouter(t)
	contents := []string{"fictional contract before correction", "fictional contract after correction"}
	for version, content := range contents {
		response := doMultipartFile(router, "/api/v1/case-materials/200/files", "1", "lawyer", fmt.Sprintf("contract-v%d.txt", version+1), content)
		if response.Code != http.StatusOK {
			t.Fatalf("upload version %d: %d %s", version+1, response.Code, response.Body.String())
		}
	}
	history := doJSON(router, http.MethodGet, "/api/v1/case-materials/200/files", "1", "lawyer", nil)
	if history.Code != http.StatusOK {
		t.Fatalf("history: %d %s", history.Code, history.Body.String())
	}
	files, ok := decodeMaterialFilesBody(t, history.Body.Bytes())["files"].([]interface{})
	if !ok || len(files) != len(contents) {
		t.Fatalf("expected two historical versions, got %v", files)
	}
	for index, raw := range files {
		file := raw.(map[string]interface{})
		path := file["download_url"].(string)
		response := doJSON(router, http.MethodGet, path, "1", "lawyer", nil)
		if response.Code != http.StatusOK || response.Body.String() != contents[index] {
			t.Fatalf("version %d download: status=%d bytes=%q", index+1, response.Code, response.Body.String())
		}
		if response.Header().Get("Content-Disposition") == "" {
			t.Fatalf("version %d has no attachment disposition", index+1)
		}
		denied := doJSON(router, http.MethodGet, path, "2", "lawyer", nil)
		if denied.Code != http.StatusForbidden {
			t.Fatalf("unrelated lawyer historical download: %d, expected 403", denied.Code)
		}
	}
}

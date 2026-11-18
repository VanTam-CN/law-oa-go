package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Tests in this file cover the three P4 revision-contract regressions found
// in independent review of 5ddbdba:
//  1. the confirmed/conflict_ready invalidation branch must keep the freshly
//     stamped revision (never write the pre-save token back);
//  2. legacy drafts without a stored draft_revision must load and save once
//     through a server-derived stable token, then behave like any draft;
//  3. revision checks on facts confirmation / conflict start must run after
//     object authorization (403 wins over the revision hint) and confirmation
//     must map its locked re-check failure to 409, not 500.

func intakeContractRequest(t *testing.T, r interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}, method, path, body, revisionHeader, userID string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", "lawyer")
	req.Header.Set("X-Test-User-Id", userID)
	if revisionHeader != "" {
		req.Header.Set("X-Intake-Revision", revisionHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var payload map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload
}

func setIntakeStatus(t *testing.T, db interface {
	Exec(string, ...interface{}) interface{ Error() error }
}, status string) {
	t.Helper()
	_ = status
}

func TestUpdateIntakeInvalidatesConfirmedSnapshotAndRotatesRevision(t *testing.T) {
	for _, status := range []string{"lawyer_facts_confirmed", "conflict_ready"} {
		t.Run(status, func(t *testing.T) {
			h, db := newDraftRevisionDB(t)
			r := newDraftRevisionRouter(h)
			stampDraftRevision(t, db, "rev-1")
			if err := db.Table("case_intakes").Where("id = ?", "intake-rev").Update("status", status).Error; err != nil {
				t.Fatalf("set status %s: %v", status, err)
			}

			// Tab A: first facts change on a confirmed snapshot must rotate
			// the revision token, and the STORED token must equal the response
			// token (the 5ddbdba regression wrote the pre-save token back).
			codeA, savedA := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", "{\"expected_revision\":\"rev-1\",\"title\":\"A确认后改事实\"}", "")
			if codeA != http.StatusOK {
				t.Fatalf("first facts change: status=%d payload=%v", codeA, savedA)
			}
			dataA, _ := savedA["data"].(map[string]interface{})
			revA, _ := dataA["revision"].(string)
			if revA == "" || revA == "rev-1" {
				t.Fatalf("revision must rotate after confirmed facts change, got %q", revA)
			}
			storedRev, clientDigest, opponentDigest, storedTitle := draftRevisionServerState(t, db)
			if storedRev != revA {
				t.Fatalf("stored revision %q must equal response %q (stale token written back)", storedRev, revA)
			}
			if storedTitle != "A确认后改事实" {
				t.Fatalf("tab A title must persist, got %q", storedTitle)
			}

			// Tab B: the pre-save token must now be dead, with zero mutation.
			codeB, conflictB := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", "{\"expected_revision\":\"rev-1\",\"description\":\"B旧页覆盖\"}", "")
			if codeB != http.StatusConflict {
				t.Fatalf("stale tab save after invalidation: status=%d want 409 payload=%v", codeB, conflictB)
			}
			errObjB, _ := conflictB["error"].(map[string]interface{})
			if errObjB == nil || errObjB["code"] != "INTAKE_DRAFT_REVISION_CONFLICT" {
				t.Fatalf("want INTAKE_DRAFT_REVISION_CONFLICT, got %v", conflictB)
			}
			ctxB, _ := errObjB["context"].(map[string]interface{})
			if ctxB == nil || ctxB["current_revision"] != revA {
				t.Fatalf("conflict must carry current revision %q, got %v", revA, conflictB)
			}
			revAfter, clientAfter, opponentAfter, titleAfter := draftRevisionServerState(t, db)
			if revAfter != revA || titleAfter != "A确认后改事实" || clientAfter != clientDigest || opponentAfter != opponentDigest {
				t.Fatalf("rejected stale save must not mutate: rev=%q title=%q client=%q opponent=%q", revAfter, titleAfter, clientAfter, opponentAfter)
			}

			// The invalidation branch must have recorded its audit markers.
			var row struct {
				Metadata []byte
			}
			if err := db.Table("case_intakes").Select("metadata").Where("id = ?", "intake-rev").Take(&row).Error; err != nil {
				t.Fatalf("reload metadata: %v", err)
			}
			metadata := objectValue(row.Metadata)
			if metadata["conflict_invalidated_by"] != "7" || fmt.Sprint(metadata["conflict_invalidation_reason"]) == "" {
				t.Fatalf("invalidation markers missing: %v", metadata)
			}
		})
	}
}

func TestLegacyDraftWithoutStoredRevisionLoadsAndSavesOnce(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	// newDraftRevisionDB seeds metadata '{}': no draft_revision token at all.
	r := newDraftRevisionRouter(h)

	code, listPayload := draftRevisionRequest(t, r, http.MethodGet, "/api/v1/case-intakes", "", "")
	if code != http.StatusOK {
		t.Fatalf("legacy list: status=%d payload=%v", code, listPayload)
	}
	data, _ := listPayload["data"].(map[string]interface{})
	items, _ := data["items"].([]interface{})
	legacyToken := "legacy-intake-rev"
	found := false
	for _, raw := range items {
		row, _ := raw.(map[string]interface{})
		if fmt.Sprint(row["id"]) == "intake-rev" {
			found = true
			if row["revision"] != legacyToken {
				t.Fatalf("legacy list row must expose derived token %q, got %v", legacyToken, row["revision"])
			}
		}
	}
	if !found {
		t.Fatalf("legacy draft missing from list: %v", items)
	}

	code, wb := draftRevisionRequest(t, r, http.MethodGet, "/api/v1/case-intakes/intake-rev", "", "")
	if code != http.StatusOK {
		t.Fatalf("legacy workbench: status=%d payload=%v", code, wb)
	}
	wbData, _ := wb["data"].(map[string]interface{})
	if wbData == nil || wbData["revision"] != legacyToken {
		t.Fatalf("legacy workbench must expose derived token, got %v", wb)
	}

	// One save through the derived token must succeed and materialize a
	// fresh server-owned revision.
	code, saved := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", "{\"expected_revision\":\"legacy-intake-rev\",\"title\":\"旧草稿复活\"}", "")
	if code != http.StatusOK {
		t.Fatalf("legacy save: status=%d payload=%v", code, saved)
	}
	savedData, _ := saved["data"].(map[string]interface{})
	materialized, _ := savedData["revision"].(string)
	if materialized == "" || materialized == legacyToken {
		t.Fatalf("legacy save must materialize a fresh revision, got %q", materialized)
	}
	rev, _, _, title := draftRevisionServerState(t, db)
	if rev != materialized || title != "旧草稿复活" {
		t.Fatalf("legacy save state: rev=%q title=%q want %q", rev, title, materialized)
	}

	// Once materialized, the derived legacy token is dead: a page that kept
	// it must get 409, not a silent overwrite.
	code, second := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", "{\"expected_revision\":\"legacy-intake-rev\",\"title\":\"旧token复用\"}", "")
	if code != http.StatusConflict {
		t.Fatalf("stale legacy token must 409 after materialization: status=%d payload=%v", code, second)
	}
	errObj, _ := second["error"].(map[string]interface{})
	if errObj == nil || errObj["code"] != "INTAKE_DRAFT_REVISION_CONFLICT" {
		t.Fatalf("want INTAKE_DRAFT_REVISION_CONFLICT, got %v", second)
	}
}

func TestConfirmAndConflictCheckAuthorizeBeforeRevisionHint(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	r.POST("/api/v1/case-intakes/:id/conflict-check", h.StartIntakeConflictCheck)
	stampDraftRevision(t, db, "rev-1")

	// A foreign user with a STALE revision header must see 403 (object
	// authorization), never 409 (revision hint) and never a write.
	code, payload := intakeContractRequest(t, r, http.MethodPost, "/api/v1/case-intakes/intake-rev/facts-confirmation", "{}", "rev-0", "999")
	if code != http.StatusForbidden {
		t.Fatalf("foreign confirm with stale header: status=%d want 403 payload=%v", code, payload)
	}
	code, payload = intakeContractRequest(t, r, http.MethodPost, "/api/v1/case-intakes/intake-rev/conflict-check", "{}", "rev-0", "999")
	if code != http.StatusForbidden {
		t.Fatalf("foreign conflict-check with stale header: status=%d want 403 payload=%v", code, payload)
	}
	rev, _, _, _ := draftRevisionServerState(t, db)
	if rev != "rev-1" {
		t.Fatalf("rejected foreign calls must not mutate revision: %q", rev)
	}

	// The owner with a stale header still gets the revision guard: 409 with
	// the current server token in the conflict context.
	code, payload = intakeContractRequest(t, r, http.MethodPost, "/api/v1/case-intakes/intake-rev/facts-confirmation", "{}", "rev-0", "7")
	if code != http.StatusConflict {
		t.Fatalf("owner stale confirm: status=%d want 409 payload=%v", code, payload)
	}
	errObj, _ := payload["error"].(map[string]interface{})
	if errObj == nil || errObj["code"] != "INTAKE_DRAFT_REVISION_CONFLICT" {
		t.Fatalf("owner stale confirm must carry revision conflict code, got %v", payload)
	}
	ctx, _ := errObj["context"].(map[string]interface{})
	if ctx == nil || ctx["current_revision"] != "rev-1" {
		t.Fatalf("conflict context must carry current revision, got %v", payload)
	}
}

// Regression for the independent 221d468 finding (real API + Postgres): a
// partial PUT WITHOUT a metadata key (title only, correct expected_revision)
// returned 200 but erased every stored business metadata field, because
// objectValue(nil) returns an empty map and the nil fallback never fired.
func TestUpdateIntakeWithoutMetadataKeepsStoredBusinessFields(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	businessMetadata := "{\"draft_revision\":\"rev-1\",\"business_area\":\"争议解决\",\"sub_area\":\"商事诉讼\",\"dispute_amount\":\"120000\",\"source_channel\":\"客户转介\"}"
	if err := db.Table("case_intakes").Where("id = ?", "intake-rev").
		Update("metadata", businessMetadata).Error; err != nil {
		t.Fatalf("seed business metadata: %v", err)
	}

	code, saved := draftRevisionRequest(t, r, http.MethodPut,
		"/api/v1/case-intakes/intake-rev",
		"{\"expected_revision\":\"rev-1\",\"title\":\"仅更新标题\"}",
		"")
	if code != http.StatusOK {
		t.Fatalf("partial update: status=%d payload=%v", code, saved)
	}
	data, _ := saved["data"].(map[string]interface{})
	responseRevision, _ := data["revision"].(string)
	if responseRevision == "" || responseRevision == "rev-1" {
		t.Fatalf("revision must rotate on partial save, got %q", responseRevision)
	}

	var row struct {
		Title    string
		Metadata []byte
	}
	if err := db.Table("case_intakes").Select("title, metadata").
		Where("id = ?", "intake-rev").Take(&row).Error; err != nil {
		t.Fatalf("reload intake: %v", err)
	}
	metadata := objectValue(row.Metadata)
	if intakeDraftRevision(row.Metadata) != responseRevision {
		t.Fatalf("stored revision %q must equal response %q",
			intakeDraftRevision(row.Metadata), responseRevision)
	}
	if row.Title != "仅更新标题" {
		t.Fatalf("title must update, got %q", row.Title)
	}
	for key, want := range map[string]string{
		"business_area":  "争议解决",
		"sub_area":       "商事诉讼",
		"dispute_amount": "120000",
		"source_channel": "客户转介",
	} {
		if got := fmt.Sprint(metadata[key]); got != want {
			t.Fatalf("partial save must keep stored metadata %s: got %q want %q (metadata=%v)",
				key, got, want, metadata)
		}
	}
}

// A save that DOES provide metadata merges onto the stored metadata: fields
// the client sends win, fields it omits stay untouched.
func TestUpdateIntakeMergesProvidedMetadataWithoutErasingStoredFields(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	businessMetadata := "{\"draft_revision\":\"rev-1\",\"business_area\":\"争议解决\",\"dispute_amount\":\"120000\",\"fee_base\":\"固定费\"}"
	if err := db.Table("case_intakes").Where("id = ?", "intake-rev").
		Update("metadata", businessMetadata).Error; err != nil {
		t.Fatalf("seed business metadata: %v", err)
	}

	code, saved := draftRevisionRequest(t, r, http.MethodPut,
		"/api/v1/case-intakes/intake-rev",
		"{\"expected_revision\":\"rev-1\",\"metadata\":{\"fee_base\":\"风险代理\"}}",
		"")
	if code != http.StatusOK {
		t.Fatalf("metadata merge update: status=%d payload=%v", code, saved)
	}

	var row struct {
		Metadata []byte
	}
	if err := db.Table("case_intakes").Select("metadata").
		Where("id = ?", "intake-rev").Take(&row).Error; err != nil {
		t.Fatalf("reload intake: %v", err)
	}
	metadata := objectValue(row.Metadata)
	if got := fmt.Sprint(metadata["fee_base"]); got != "风险代理" {
		t.Fatalf("provided metadata field must win, got fee_base=%q", got)
	}
	for key, want := range map[string]string{
		"business_area":  "争议解决",
		"dispute_amount": "120000",
	} {
		if got := fmt.Sprint(metadata[key]); got != want {
			t.Fatalf("merge must keep stored metadata %s: got %q want %q (metadata=%v)",
				key, got, want, metadata)
		}
	}
}

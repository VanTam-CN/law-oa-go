package handlers

import (
	"testing"

	"github.com/google/uuid"
)

func TestNPPostgresNativeCannotForgeSourcePointer(t *testing.T) {
	f := npPostgresFixture(t)
	assertNoPointer := func(id string) {
		t.Helper()
		var present bool
		if err := f.db.Raw("SELECT jsonb_exists(metadata, ?) FROM case_intakes WHERE id=?", npDescriptionSourceRevisionKey, id).Scan(&present).Error; err != nil || present {
			t.Fatalf("client-authored source pointer persisted: present=%v err=%v", present, err)
		}
	}
	code, out := f.request(t, "POST", "/api/v1/case-intakes", "", map[string]interface{}{
		"title": "虚构新增草稿", "metadata": map[string]interface{}{npDescriptionSourceRevisionKey: uuid.NewString(), "ordinary_note": "keep"},
	})
	if code != 201 {
		t.Fatalf("create failed: %d %v", code, out)
	}
	created := out["data"].(map[string]interface{})
	assertNoPointer(created["id"].(string))
	for _, status := range []string{"draft", "lawyer_facts_confirmed", "conflict_ready"} {
		t.Run(status, func(t *testing.T) {
			if err := f.db.Table("case_intakes").Where("id=?", f.id).Update("status", status).Error; err != nil {
				t.Fatal(err)
			}
			code, out := f.request(t, "PUT", "/api/v1/case-intakes/"+f.id, "", map[string]interface{}{
				"expected_revision": f.revision, "title": "虚构标题-" + status,
				"metadata": map[string]interface{}{npDescriptionSourceRevisionKey: uuid.NewString(), "ordinary_note": "keep"},
			})
			npData(t, code, out)
			assertNoPointer(f.id)
			var row map[string]interface{}
			if err := f.db.Table("case_intakes").Where("id=?", f.id).Take(&row).Error; err != nil {
				t.Fatal(err)
			}
			metadata := objectValue(row["metadata"])
			if metadata["ordinary_note"] != "keep" || metadata["np_unrelated"] == nil {
				t.Fatal("ordinary metadata was lost")
			}
			f.revision = intakeDraftRevisionFor(f.id, row["metadata"])
		})
	}
}

func TestNPPostgresLegacyNullMetadataIsInitialized(t *testing.T) {
	for _, source := range []bool{false, true} {
		name := "manual"
		if source {
			name = "source"
		}
		t.Run(name, func(t *testing.T) {
			f := npPostgresFixture(t)
			// The frozen schema is NOT NULL; only this synthetic legacy fixture relaxes it.
			if err := f.db.Exec("ALTER TABLE case_intakes ALTER COLUMN metadata DROP NOT NULL").Error; err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := f.db.Exec("UPDATE case_intakes SET metadata='{}'::jsonb WHERE id=? AND metadata IS NULL", f.id).Error; err != nil {
					t.Error(err)
				}
				if err := f.db.Exec("ALTER TABLE case_intakes ALTER COLUMN metadata SET NOT NULL").Error; err != nil {
					t.Error(err)
				}
			})
			if err := f.db.Table("case_intakes").Where("id=?", f.id).Update("metadata", nil).Error; err != nil {
				t.Fatal(err)
			}
			f.revision = "legacy-" + f.id
			text := "虚构遗留手工摘要"
			var selection interface{}
			if source {
				text, selection = f.candidate(t)
			}
			path := "/api/v1/case-intakes/" + f.id
			code, out := f.request(t, "PUT", path, uuid.NewString(), f.body(text, selection))
			result := npData(t, code, out)
			var row map[string]interface{}
			if err := f.db.Table("case_intakes").Where("id=?", f.id).Take(&row).Error; err != nil {
				t.Fatal(err)
			}
			metadata := objectValue(row["metadata"])
			if metadata[intakeDraftRevisionKey] != result["revision"] || result["revision"] == f.revision {
				t.Fatal("successful save did not persist its new revision", metadata, result)
			}
			if source && metadata[npDescriptionSourceRevisionKey] != result["revision"] {
				t.Fatal("source pointer not persisted", metadata)
			}
			code, _ = f.request(t, "PUT", path, uuid.NewString(), f.body("迟到的旧版本", nil))
			if code != 409 {
				t.Fatalf("stale legacy revision accepted: %d", code)
			}
		})
	}
}

func TestNPPostgresMalformedIntakeIDIsBadRequest(t *testing.T) {
	f := npPostgresFixture(t)
	for _, body := range []map[string]interface{}{{"expected_revision": f.revision, "title": "虚构标题"}, f.body("虚构摘要", nil)} {
		code, out := f.request(t, "PUT", "/api/v1/case-intakes/not-a-uuid", uuid.NewString(), body)
		errorBody, _ := out["error"].(map[string]interface{})
		if code != 400 || errorBody["code"] != "INVALID_REQUEST" {
			t.Fatalf("invalid identifier was not rejected at boundary: %d %v", code, out)
		}
	}
}

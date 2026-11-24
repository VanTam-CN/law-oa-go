//go:build np_m1_pg && np_m1_browser_pg

package m1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"law-oa-go/internal/models"
)

// Chrome exercises local recovery with synthetic workflow responses. The approval
// POST goes unchanged to the registered router and writes to the disposable PG.
func TestM1R4WorkbenchApprovalBrowser(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	g, token := nativePartyRouter(t, db, f)
	check := uuid.NewString()
	insertConflictPGRecord(t, db, f, check, f.revision)
	if err := db.Exec(`UPDATE case_intakes SET title='示例科技服务合同纠纷案',status='conflict_ready',metadata=metadata||jsonb_build_object('draft_revision','checked','conflict_check_id',?::text) WHERE id=?`, check, f.intakeID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO case_intake_parties(intake_id,party_role,entity_type,entity_name)
 SELECT ?,CASE n WHEN 0 THEN 'client' WHEN 1 THEN 'opponent' ELSE 'related_party' END,'LEGAL_PERSON',
 CASE n WHEN 0 THEN '上海示例科技有限公司' WHEN 1 THEN '上海华信建设集团有限公司' WHEN 20 THEN '投影外合成关联方' ELSE '虚构关联主体'||n END FROM generate_series(0,20) n ORDER BY CASE n WHEN 1 THEN 2 WHEN 2 THEN 1 ELSE n END`, f.intakeID).Error; err != nil {
		t.Fatal(err)
	}
	admin := models.User{Username: "browser-" + check, Name: "Fixture approver", Email: check + "@np.invalid", Password: "disabled", Role: "admin", Status: "active"}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	var expectedRoster string
	if err := db.Raw(`SELECT jsonb_agg(jsonb_build_array(CASE party_role WHEN 'opponent' THEN 'opposing_party' ELSE party_role END,entity_name)
        ORDER BY CASE party_role WHEN 'client' THEN 0 WHEN 'opponent' THEN 1 ELSE 2 END,id)::text FROM case_intake_parties WHERE intake_id=?`, f.intakeID).Scan(&expectedRoster).Error; err != nil {
		t.Fatal(err)
	}
	const path = "/api/v1/integration/approvals/with-conflict"
	submitted := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		r.Header.Set("Authorization", "Bearer "+token)
		g.ServeHTTP(w, r)
		submitted <- struct{}{}
	}))
	defer server.Close()
	fixture, err := json.Marshal(map[string]any{"url": server.URL + path, "intake": f.intakeID, "client": json.Number(f.clientID), "check": check})
	if err != nil {
		t.Fatal(err)
	}
	dir := os.Getenv("NP_M1_FIXTURE_DIR")
	if dir == "" {
		t.Fatal("run scripts/np-first-slice/verify-m1-browser-approval.sh")
	}
	tmp := filepath.Join(dir, "pending.json")
	if err := os.WriteFile(tmp, fixture, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmp, filepath.Join(dir, "ready.json")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-submitted:
	case <-time.After(90 * time.Second):
		t.Fatal("no browser approval received")
	}
	var count int
	err = db.Raw(`SELECT count(*) FROM approval_requests a WHERE metadata::jsonb->>'intake_id'=? AND metadata::jsonb->>'conflict_check_id'=?
 AND metadata::jsonb->>'intake_revision'='checked'
 AND metadata::jsonb->'parties'=metadata::jsonb#>'{approval_snapshot,metadata,parties}'
 AND (SELECT jsonb_agg(jsonb_build_array(p->>'role',p->>'name') ORDER BY ord) FROM jsonb_array_elements(metadata::jsonb->'parties') WITH ORDINALITY AS row(p,ord))=
 ?::jsonb`, f.intakeID, check, expectedRoster).Scan(&count).Error
	if err != nil || count != 1 {
		t.Fatalf("expected one persisted matching roster and frozen snapshot: count=%d err=%v", count, err)
	}
}

#!/usr/bin/python3
"""Read only, fixed synthetic local container: bind browser results to PG facts."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[2]
REPORT = ROOT / "docs/np-evidence/browser-pg-comparison.json"
mode = sys.argv[1]
assert mode in ("before", "after")
container = "np-first-slice-db-1"
project = subprocess.check_output(["docker", "inspect", "--format", '{{index .Config.Labels "com.docker.compose.project"}}', container], text=True).strip()
assert project == "np-first-slice"
query = """
SELECT jsonb_build_object(
 'protected', jsonb_build_object('title',i.title,'client',i.client_id,'status',i.status,
   'metadata',i.metadata - ARRAY['draft_revision','np_description_source_revision'],
   'parties',(SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM case_intake_parties p WHERE intake_id=i.id),
   'materials',(SELECT jsonb_agg(to_jsonb(m) ORDER BY id) FROM case_materials m WHERE intake_id=i.id),
   'files',(SELECT jsonb_agg(to_jsonb(f) ORDER BY f.id) FROM material_files f JOIN case_materials m ON f.material_id=m.id WHERE m.intake_id=i.id),
   'control_d2',(SELECT to_jsonb(d) FROM case_intakes d WHERE id='00000000-0000-4000-8000-000000009802'))::text,
 'revision',i.metadata->>'draft_revision','source_revision',i.metadata->>'np_description_source_revision',
 'sources',(SELECT count(*) FROM case_intake_source_snapshots WHERE intake_id=i.id),
 'receipts',(SELECT count(*) FROM case_intake_write_requests WHERE intake_id=i.id),
 'files',(SELECT count(*) FROM material_files f JOIN case_materials m ON f.material_id=m.id WHERE m.intake_id=i.id)
) FROM case_intakes i WHERE id='00000000-0000-4000-8000-000000009801';
"""
raw = subprocess.check_output(["docker", "exec", "-i", container, "sh", "-c", 'exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d np_first_slice_dev -At'], input=query, text=True)
data = json.loads(raw)
data["protected_sha256"] = hashlib.sha256(data.pop("protected").encode()).hexdigest()
assert data["files"] == 2
if mode == "before":
    REPORT.parent.mkdir(parents=True, exist_ok=True)
    REPORT.write_text(json.dumps({"before": data}, ensure_ascii=False, indent=2) + "\n")
else:
    result = json.loads(REPORT.read_text())
    before = result["before"]
    assert before["protected_sha256"] == data["protected_sha256"], "non-target data changed"
    assert data["sources"] - before["sources"] == 3, "expected exactly three source-bound writes"
    assert data["receipts"] - before["receipts"] == 4, "expected exactly four successful writes"
    assert data["revision"] != before["revision"]
    assert not data["source_revision"], "last explicitly manual save must be unlinked"
    result["after"] = data
    result["assertions"] = {"protected_fields_children_files_and_D2_unchanged": True, "source_snapshot_delta": 3, "receipt_delta": 4, "manual_source_unlinked": True}
    REPORT.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
print("BROWSER_PG_" + mode.upper() + "_OK")

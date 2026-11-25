"""I05 read-only check against the four-scenario target after its final manual save."""
import http.client
import json
import os
from pathlib import Path
import runpy

if not __debug__ or os.environ.get("NP_LOCAL_CHECK") != "synthetic-only":
    raise SystemExit("Synthetic fixture with assertions enabled required")

seed = json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
compare = runpy.run_path(str(Path(__file__).resolve().parents[1] / "np-first-slice/compare-main-browser-db.py"))
before = compare["snapshot"](seed)
assert before["sources"] == 3 and before["receipts"] == 4 and before["manual"] == 1
token = None

def call(method, path, body=None):
    client = http.client.HTTPConnection("127.0.0.1", 18789, timeout=10)
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    try:
        client.request(method, "/api/v1/" + path, json.dumps(body) if body else None, headers)
        response = client.getresponse()
        assert response.status == 200, (method, response.status)
        return json.load(response)["data"]
    finally:
        client.close()

token = call("POST", "auth/login", {"account": seed["username"], "password": os.environ["NP_BROWSER_PASSWORD"]})["token"]
rows = [row for row in call("GET", "case-intakes")["items"] if row["id"] == seed["target"]]
assert len(rows) == 1
detail = call("GET", "case-intakes/" + seed["target"])
alias = call("GET", "cases/intake-workbench/" + seed["target"])
assert before["revision"] == rows[0]["revision"] == detail["revision"] == alias["revision"]
assert rows[0]["id"] == detail["intake"]["id"] == alias["intake"]["id"]
assert detail["intake"]["description"] == alias["intake"]["description"]
assert before == compare["snapshot"](seed)
print(json.dumps({"gate": "I05", "input": "route B after final manual save",
                  "status": "PASS",
                  "list_detail_alias_pg_revision_equal": True, "canonical_and_alias_description_equal": True,
                  "source_count": before["sources"], "receipt_count": before["receipts"],
                  "whole_target_and_protected_records_unchanged": True, "protected_sha256": before["protected_sha256"]}, indent=2))

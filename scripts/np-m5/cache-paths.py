"""Observe Redis by exact synthetic route; keep credentials, keys and response bodies private."""
import collections
import http.client
import json
import os
from pathlib import Path
import re
import runpy
import subprocess
import uuid

if not __debug__ or os.environ.get("NP_LOCAL_CHECK") != "synthetic-only":
    raise SystemExit("Synthetic fixture with assertions enabled required")
work = Path(os.environ["NP_M5_WORK"])
compare = runpy.run_path(str(Path(__file__).resolve().parents[1] / "np-first-slice/compare-main-browser-db.py"))
own = json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
other = json.loads(Path(os.environ["NP_BROWSER_OTHER_SEED"]).read_text())
before = compare["snapshot"](own)
assert (before["sources"], before["receipts"]) == (3, 4)
saved = compare["query"]("""SELECT jsonb_build_array(request_id,source_revision)
    FROM case_intake_write_requests WHERE intake_id=:'target' AND actor_id=:'actor_id'
    AND NULLIF(source_revision,'') IS NOT NULL ORDER BY created_at DESC LIMIT 1;""", own)
rid, rev = (str(uuid.UUID(item)) for item in saved)
redis = os.environ["NP_REDIS_CID"]
assert subprocess.check_output(["docker","inspect","-f",'{{index .Config.Labels "law-oa-m5"}}',redis],text=True).strip() == work.name
cli = ["docker", "exec", redis, "redis-cli", "--raw"]
nonce = uuid.uuid4().hex
marker = "m5-old-" + nonce
token = None

def call(method, path, body=None, auth=True):
    headers = {"Content-Type": "application/json"}
    if auth and token:
        headers["Authorization"] = "Bearer " + token
    conn = http.client.HTTPConnection("127.0.0.1", 18789, timeout=10)
    try:
        conn.request(method, path, json.dumps(body) if body else None, headers)
        r = conn.getresponse()
        data = r.read()
        assert marker.encode() not in data
        return r.status, r.getheader("Cache-Control", ""), json.loads(data)
    finally:
        conn.close()

prefix = "/api/v1/"
s, _, obj = call("POST", prefix+"auth/login", {"account":own["username"],"password":os.environ["NP_BROWSER_PASSWORD"]})
assert s == 200
token = obj["data"]["token"]
contexts = prefix+"case-intake-assist/contexts"
s, h, obj = call("POST", contexts, {"intake_id":own["target"],"expected_revision":before["revision"]})
assert s == 200 and "no-store" in h
ctx = str(uuid.UUID(obj["data"]["id"]))
context_version = obj["data"]["context_version"]
base = prefix+"case-intakes"
cases = [
    ("neighbor",prefix+"clients/stats",403),("capabilities",prefix+"case-intake-assist/capabilities",200),
    ("context",contexts+"/"+ctx,200),("list",base,200),("detail",base+"/"+own["target"],200),
    ("workbench_alias",prefix+"cases/intake-workbench/"+own["target"],200),
    ("historical_source",base+"/"+own["target"]+"/source-snapshots/"+rev,200),
    ("committed_receipt",base+"/write-results/"+rid,200),
    ("other_draft",base+"/"+other["target"],403),("unknown_draft",base+"/"+str(uuid.uuid4()),404)]
evidence = {}
try:
    for label, path, expected in cases:
        # An otherwise unused query parameter keeps each old key exclusive to this observation.
        query = "m5_probe=" + nonce
        key = "lawoa::"+path+":"+query
        ident = uuid.uuid4().hex
        sentinel = json.dumps({"stale":marker})
        assert subprocess.check_output(cli+["SET",key,sentinel,"EX","120","NX"],text=True).strip()=="OK"
        mon = subprocess.Popen(cli+["MONITOR"],stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,text=True)
        counts = collections.Counter()
        sensitive_operations = 0
        try:
            assert mon.stdout.readline().strip()=="OK"
            s, h, obj = call("GET",path+"?"+query)
            subprocess.run(cli+["ECHO",ident],check=True,stdout=subprocess.DEVNULL)
            # Do not persist raw MONITOR traffic, even temporarily.
            for line in mon.stdout:
                if '"echo" "'+ident+'"' in line.lower():
                    break
                cmd = re.search(r'\] "([^"]+)" "([^"]+)"',line)
                if cmd:
                    operation, actual = cmd.groups()
                    if any(prefix in line for prefix in ('"lawoa::/api/v1/case-intakes', '"lawoa::/api/v1/case-intake-assist', '"lawoa::/api/v1/cases/intake-workbench')):
                        sensitive_operations += 1
                    if actual == key or actual.startswith(key+":"):
                        counts[operation.lower()]+=1
            else:
                raise AssertionError("Redis observation ended early")
            assert s == expected, (label,s)
            if label=="neighbor":
                assert counts["get"]>0, "positive control did not observe neighboring cache GET"
            else:
                assert sensitive_operations == 0, (label, sensitive_operations)
                assert not counts and {"private","no-store"}.issubset({v.strip() for v in h.lower().split(",")}), label
                data = obj.get("data") or {}
                if label=="historical_source":
                    assert data["source"] and data["source_revision"]==rev
                elif label=="committed_receipt":
                    assert data["state"]=="committed" and data["request_id"]==rid and data["id"]==own["target"]
                elif label in ("detail","workbench_alias"):
                    assert data["revision"]==before["revision"]
            evidence[label] = {"http":s,"exact_key_operations":dict(counts),"sensitive_prefix_operations":sensitive_operations,"cache_control":h}
        finally:
            mon.terminate()
            mon.communicate(timeout=4)
            subprocess.run(cli+["DEL",key],check=True,stdout=subprocess.DEVNULL)
finally:
    s, _, obj = call("DELETE", contexts+"/"+ctx, {"expected_context_version":context_version})
    assert s == 200 and obj["data"]["ended"] is True
assert compare["snapshot"](own)==before
print(json.dumps({"scope":"I10/C10 route-by-route on isolated root and Redis", "routes":evidence,
                  "preseeded_old_value_returned":False,"pg_before_after_equal":True},indent=2))

"""Read the identical historical source and receipt with one token across a wall change."""

import http.client
import json
import os
from pathlib import Path
import re
import subprocess
import uuid

if not __debug__ or os.environ.get("NP_LOCAL_CHECK") != "synthetic-only":
    raise SystemExit("Synthetic fixture required")
work = Path(os.environ["NP_M5_WORK"])
window = work / "history-window"
db = (window / "db").read_text().strip()
pg = (work / "pg").read_text().strip()
redis = (window / "redis").read_text().strip()
assert re.fullmatch("np_m5_history_[a-f0-9]{12}", db)
for container in (pg, redis):
    assert (
        subprocess.check_output(
            [
                "docker",
                "inspect",
                "-f",
                '{{index .Config.Labels "law-oa-m5"}}',
                container,
            ],
            text=True,
        ).strip()
        == work.name
    )
seed = json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
target = str(uuid.UUID(seed["target"]))
actor = int(seed["actor_id"])


def sql(q):
    return subprocess.check_output(
        [
            "docker",
            "exec",
            pg,
            "psql",
            "-XqAtw",
            "-v",
            "ON_ERROR_STOP=1",
            "-U",
            "np_m1",
            "-d",
            db,
            "-c",
            q,
        ],
        text=True,
    ).strip()


assert sql("SELECT current_database()") == db
known = json.loads(
    sql(f"""SELECT jsonb_build_array(w.request_id,w.source_revision) FROM case_intake_write_requests w
   JOIN case_intake_source_snapshots s ON s.intake_id=w.intake_id AND s.saved_revision=w.source_revision
   WHERE w.intake_id='{target}' AND w.actor_id='{actor}' AND w.state='committed' ORDER BY w.created_at DESC LIMIT 1""")
)
key, revision = [str(uuid.UUID(x)) for x in known]
token = None


def http_call(path, body=None):
    c = http.client.HTTPConnection("127.0.0.1", 18791, timeout=10)
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    try:
        c.request(
            "POST" if body else "GET", path, json.dumps(body) if body else None, headers
        )
        r = c.getresponse()
        return r.status, r.getheader("Cache-Control", ""), json.load(r)
    finally:
        c.close()


status, _, doc = http_call(
    "/api/v1/auth/login",
    {"account": seed["username"], "password": os.environ["NP_BROWSER_PASSWORD"]},
)
assert status == 200
token = doc["data"]["token"]
paths = {
    "source": f"/api/v1/case-intakes/{target}/source-snapshots/{revision}",
    "receipt": "/api/v1/case-intakes/write-results/" + key,
}
prior = {}
for label, path in paths.items():
    status, h, out = http_call(path)
    assert status == 200 and h == "private, no-store"
    prior[label] = out["data"]
assert (
    prior["source"]["source_revision"] == revision
    and prior["source"]["source"]
    and prior["receipt"]["request_id"] == key
)
assert (
    prior["receipt"]["id"] == target
    and prior["receipt"]["state"] == "committed"
    and prior["receipt"]["revision"] == revision
)
source_pack = prior["source"]["source"]
old_contents = [
    source_pack["description"],
    *(entry["text"] for entry in source_pack["sources"]),
]
assert len(old_contents) >= 2 and all(
    isinstance(text, str) and len(text) > 3 for text in old_contents
)
wall = int(
    sql(f"""INSERT INTO cases(case_number,title,client_id,lawyer_id,case_type,status,ethical_wall_enabled)
  VALUES('M5-{uuid.uuid4().hex[:16]}','M5 historical read',{int(seed['client_id'])},{actor},'civil','pending',true) RETURNING id""")
)


def row_digest():
    return sql(f"""SELECT encode(sha256(convert_to(jsonb_build_array(to_jsonb(i),to_jsonb(s),to_jsonb(w))::text,'UTF8')),'hex')
       FROM case_intakes i JOIN case_intake_source_snapshots s ON s.intake_id=i.id JOIN case_intake_write_requests w ON w.intake_id=i.id
       WHERE i.id='{target}' AND s.saved_revision='{revision}' AND w.request_id='{key}' AND w.actor_id='{actor}'""")


before = row_digest()
assert re.fullmatch("[a-f0-9]{64}", before)
try:
    sql(
        f"INSERT INTO case_ethical_wall_whitelist(case_id,user_id,granted_by,granted_at,reason) VALUES({wall},{actor},{actor},now(),'M5 synthetic')"
    )
    for label, path in paths.items():
        assert http_call(path)[2]["data"] == prior[label]
    sql(
        f"DELETE FROM case_ethical_wall_whitelist WHERE case_id={wall} AND user_id={actor}"
    )
    for label, path in paths.items():
        cache_key = "lawoa::" + path
        cli = ["docker", "exec", redis, "redis-cli", "--raw"]
        # Preload this disposable cache with the authorized projection before testing denial.
        put = subprocess.run(
            [
                "docker",
                "exec",
                "-i",
                redis,
                "redis-cli",
                "--raw",
                "-x",
                "SETEX",
                cache_key,
                "30",
            ],
            input=json.dumps(prior[label]),
            capture_output=True,
            text=True,
            check=True,
        )
        assert put.stdout.strip() == "OK"
        try:
            status, h, out = http_call(path)
            assert (
                status == 403
                and h == "private, no-store"
                and out["error"]["code"] == "OBJECT_ACCESS_DENIED"
            )
            denial_body = json.dumps(out, ensure_ascii=False)
            assert all(text not in denial_body for text in old_contents)
            assert not out.get("data") and all(
                v not in json.dumps(out) for v in (revision, key, target)
            )
            assert not any(
                k in out
                for k in (
                    "source",
                    "state",
                    "request_id",
                    "source_revision",
                    "candidate",
                )
            )
        finally:
            subprocess.run(
                cli + ["DEL", cache_key], check=True, stdout=subprocess.DEVNULL
            )
finally:
    sql(
        f"DELETE FROM case_ethical_wall_whitelist WHERE case_id={wall}; DELETE FROM cases WHERE id={wall}"
    )
for label, path in paths.items():
    assert http_call(path)[2]["data"] == prior[label]
assert row_digest() == before
print(
    json.dumps(
        {
            "gates": ["I10", "C10"],
            "same_registered_token": True,
            "same_paths_and_objects": True,
            "source": {
                "before": 200,
                "granted": 200,
                "withdrawn": 403,
                "restored": 200,
            },
            "committed_receipt": {
                "before": 200,
                "granted": 200,
                "withdrawn": 403,
                "restored": 200,
            },
            "preloaded_authorized_projection_redis": True,
            "prior_summary_and_source_texts_absent_from_entire_denials": True,
            "denials": "private, no-store; OBJECT_ACCESS_DENIED; no data",
            "intake_source_receipt_digest_unchanged": before,
        },
        indent=2,
    )
)

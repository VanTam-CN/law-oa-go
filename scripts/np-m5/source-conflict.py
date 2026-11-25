"""I15: hold the real detector at audit insertion, save reviewed source through root, release."""

import http.client
import json
import os
from pathlib import Path
import runpy
import select
import shutil
import subprocess
import time
import sys
import uuid
from concurrent.futures import ThreadPoolExecutor

if not __debug__ or os.environ.get("NP_LOCAL_CHECK") != "synthetic-only":
    raise SystemExit("Synthetic fixture with assertions enabled required")
mode = sys.argv[1] if len(sys.argv) == 2 else "source"
if mode not in ("source", "native", "matching") or len(sys.argv) > 2:
    raise SystemExit("expected source, native or matching")
compare = runpy.run_path(
    str(
        Path(__file__).resolve().parents[1]
        / "np-first-slice/compare-main-browser-db.py"
    )
)
query = compare["query"]
actor = json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
nonce = uuid.uuid4().hex
token = None


def call(method, path, body=None, headers=None, expected=(200, 201)):
    head = {"Content-Type": "application/json", **(headers or {})}
    if token:
        head["Authorization"] = "Bearer " + token
    conn = http.client.HTTPConnection("127.0.0.1", 18789, timeout=40)
    try:
        conn.request(
            method,
            "/api/v1/" + path,
            json.dumps(body) if body is not None else None,
            head,
        )
        r = conn.getresponse()
        payload = json.load(r)
        code = (payload.get("error") or {}).get("code")
        assert r.status in expected, (method, path.split("/")[0], r.status, code)
        return payload["data"] if r.status < 300 else code
    finally:
        conn.close()


token = call(
    "POST",
    "auth/login",
    {"account": actor["username"], "password": os.environ["NP_BROWSER_PASSWORD"]},
)["token"]
client = call(
    "POST",
    "clients",
    {
        "name": "M5 client " + nonce,
        "type": "企业",
        "identity_type": "SOCIAL_CREDIT_CODE",
        "identity_number": "M5-" + nonce,
    },
)
client_id = int(client["id"])
owner = int(actor["actor_id"])
case_number = "M5-" + nonce[:14]
case_id = query(
    f"""INSERT INTO cases(case_number,title,client_id,lawyer_id,case_type,status) VALUES
  ('{case_number}','M5 synthetic case',{client_id},{owner},'civil','pending') RETURNING to_jsonb(id);""",
    write=True,
)
did = call(
    "POST",
    "case-intakes",
    {
        "title": "M5 source " + nonce,
        "case_type": "civil",
        "client_id": client_id,
        "metadata": {
            "subject_case_id": str(case_id),
            "subject_case_number": case_number,
        },
        "parties": [
            {
                "entity_name": "M5 opposing " + nonce,
                "entity_type": "LEGAL_PERSON",
                "party_role": "opposing_party",
                "identity_type": "SOCIAL_CREDIT_CODE",
                "identity_number": "M5-OP-" + nonce,
            }
        ],
    },
)
target = str(uuid.UUID(did["id"]))
path = "case-intakes/" + target
confirmed = call(
    "POST", path + "/facts-confirmation", {}, {"X-Intake-Revision": did["revision"]}
)["revision"]


def state():
    return query(f"""SELECT jsonb_build_object('intake',encode(sha256(convert_to(to_jsonb(i)::text,'UTF8')),'hex'),
      'case',encode(sha256(convert_to(to_jsonb(c)::text,'UTF8')),'hex'),'revision',i.metadata->>'draft_revision',
      'status',i.status,'case_linked',i.metadata->>'subject_case_id'=c.id::text,
      'case_check',c.conflict_check_id,'intake_check',i.metadata->>'conflict_check_id',
      'sources_receipts',encode(sha256(convert_to(jsonb_build_array(
          (SELECT jsonb_agg(to_jsonb(s) ORDER BY saved_revision) FROM case_intake_source_snapshots s WHERE s.intake_id=i.id),
          (SELECT jsonb_agg(to_jsonb(w) ORDER BY request_id) FROM case_intake_write_requests w WHERE w.intake_id=i.id)
          )::text,'UTF8')),'hex'),
      'source_linked',EXISTS(SELECT FROM case_intake_source_snapshots s JOIN case_intake_write_requests w
         ON w.intake_id=s.intake_id AND w.source_revision=s.saved_revision AND w.result_revision=s.saved_revision
         WHERE s.intake_id=i.id AND w.actor_id='{owner}' AND w.state='committed'
           AND s.saved_revision=i.metadata->>'np_description_source_revision' AND s.saved_revision=i.metadata->>'draft_revision'))
      FROM case_intakes i JOIN cases c ON c.id={int(case_id)} WHERE i.id='{target}';""")


before = state()
assert (
    before["case_linked"]
    and before["status"] == "lawyer_facts_confirmed"
    and before["revision"] == confirmed
)
if mode == "matching":

    def record_count():
        return query(
            f"SELECT to_jsonb(count(*)) FROM conflict_check_records WHERE client_id='{client_id}';"
        )

    assert record_count() == 0
    applied = call(
        "POST", path + "/conflict-check", {}, {"X-Intake-Revision": confirmed}
    )
    final = state()
    assert applied["revision"] == final["revision"] and confirmed != final["revision"]
    assert (
        final["status"] == "conflict_ready"
        and applied["checkId"] == final["case_check"] == final["intake_check"]
    )
    assert record_count() == 1
    code = call(
        "POST", path + "/conflict-check", {}, {"X-Intake-Revision": confirmed}, (409,)
    )
    assert (
        code == "INTAKE_DRAFT_REVISION_CONFLICT"
        and state() == final
        and record_count() == 1
    )
    print(
        json.dumps(
            {
                "gate": "I15 matching",
                "root_http": 200,
                "nonempty_case_and_intake_link_same_check": True,
                "revision_rotated": True,
                "stale_repeat_http": 409,
                "one_persisted_result": True,
                "case_and_intake_after_repeated_request_equal": True,
            },
            indent=2,
        )
    )
    sys.exit(0)

psql = shutil.which("psql") or "/usr/local/opt/libpq/bin/psql"
with ThreadPoolExecutor(1) as pool:
    lock = subprocess.Popen(
        [psql, "-XqAtw", "-v", "ON_ERROR_STOP=1", os.environ["NP_TEST_DSN"]],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        text=True,
    )
    try:
        lock.stdin.write(
            "BEGIN; SET LOCAL lock_timeout='5s'; SET LOCAL idle_in_transaction_session_timeout='30s'; "
            "LOCK TABLE conflict_check_records IN SHARE MODE; SELECT pg_backend_pid();\n"
        )
        lock.stdin.flush()
        assert select.select([lock.stdout], [], [], 8)[0]
        holder = int(lock.stdout.readline().strip())
        result = pool.submit(
            call,
            "POST",
            path + "/conflict-check",
            {},
            {"X-Intake-Revision": confirmed},
            (409,),
        )
        deadline = time.monotonic() + 12
        while not query(f"""SELECT to_jsonb(EXISTS(SELECT FROM pg_stat_activity WHERE {holder}=ANY(pg_blocking_pids(pid))
               AND query ILIKE '%insert%conflict_check_records%'));"""):
            if result.done():
                result.result()
            assert time.monotonic() < deadline, "detector failed to reach PG insert"
            time.sleep(0.05)
        assert state() == before
        changed = call(
            "PUT",
            path,
            {"expected_revision": confirmed, "title": "M5 revised " + nonce},
        )
        assert changed["revision"] != confirmed and changed["status"] == "draft"
        if mode == "native":
            edited = state()
            assert (
                edited["revision"] == changed["revision"]
                and edited["status"] == "draft"
                and edited["case_linked"]
            )
            assert not edited["source_linked"]
        else:
            context = call(
                "POST",
                "case-intake-assist/contexts",
                {"intake_id": target, "expected_revision": changed["revision"]},
            )
            turn = call(
                "POST",
                "case-intake-assist/contexts/" + context["id"] + "/turns",
                {
                    "request_id": str(uuid.uuid4()),
                    "expected_context_version": context["context_version"],
                    "text": "Synthetic corrected delivery facts.",
                },
            )
            cand = turn["candidate"]
            selection = {
                "context_id": context["id"],
                "expected_context_version": turn["context_version"],
                "fields": {
                    "description": {
                        "candidate_id": cand["id"],
                        "requested_mode": "candidate_sources",
                        "reviewed": True,
                    }
                },
            }
            key = str(uuid.uuid4())
            saved = call(
                "PUT",
                path,
                {
                    "expected_revision": changed["revision"],
                    "description": cand["description"],
                    "changed_fields": ["description"],
                    "source_selection": selection,
                },
                {"Idempotency-Key": key},
            )
            edited = state()
            assert (
                saved["request_id"] == key
                and saved["id"] == target
                and edited["revision"] == saved["revision"]
            )
            assert (
                edited["source_linked"]
                and edited["status"] == "draft"
                and edited["case_linked"]
            )
    finally:
        lock.communicate("ROLLBACK;\n", timeout=8)
    assert lock.returncode == 0 and result.result() == "INTAKE_DRAFT_REVISION_CONFLICT"
assert state() == edited
print(
    json.dumps(
        {
            "gate": "I15-NP" if mode == "source" else "I15-NATIVE",
            "root_http": "synchronous production detector",
            "matched_case_rows": 1,
            "confirmed_detector_wait_before_edit": True,
            "start_preserved_revision": True,
            "native_edit_then_reviewed_source_save": mode == "source",
            "source_receipt_intake_pg_linked": edited["source_linked"],
            "sources_and_receipts_digest_unchanged": True,
            "post_edit_sources_receipts_sha256": edited["sources_receipts"],
            "late_completion_http": 409,
            "code": "INTAKE_DRAFT_REVISION_CONFLICT",
            "intake_and_case_whole_rows_equal_to_post_edit": True,
            "post_edit_intake_sha256": edited["intake"],
            "post_edit_case_sha256": edited["case"],
        },
        indent=2,
    )
)

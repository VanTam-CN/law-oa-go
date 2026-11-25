"""I08 on a disposable actor: read-only query, explicit abandon, retry and replay."""
import hashlib
import http.client
import json
import os
from pathlib import Path
import runpy
import select
import shutil
import subprocess
import time
import uuid

if not __debug__:
    raise SystemExit("M5 assertions must be enabled")

from_root = Path(__file__).resolve().parents[1]
compare = runpy.run_path(str(from_root / "np-first-slice/compare-main-browser-db.py"))
seed = json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
assert os.environ["NP_LOCAL_CHECK"] == "synthetic-only"
base = "/api/v1/case-intakes"
target = base + "/" + seed["target"]
token = ""


def call(method, path, payload=None, key=None, drop=None):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    if key:
        headers["Idempotency-Key"] = key
    if drop:
        headers["X-NP-Drop-Response"] = drop
    conn = http.client.HTTPConnection("127.0.0.1", 3004 if drop else 18789, timeout=12)
    try:
        conn.request(method, path, json.dumps(payload) if payload is not None else None, headers)
        response = conn.getresponse()
        data = json.loads(response.read())
        assert not drop, "expected a genuine socket close"
        return response.status, data.get("data"), (data.get("error") or {}).get("code")
    except (http.client.RemoteDisconnected, ConnectionResetError):
        assert drop, "unexpected transport failure on the ordinary root route"
    finally:
        conn.close()


def receipt(key, state, payload=None):
    path = base + "/write-results/" + key
    if payload:
        status, data, _ = call("POST", path + "/abandon", payload)
    else:
        status, data, _ = call("GET", path)
    assert status == 200 and data["request_id"] == key and data["state"] == state
    if state != "unknown":
        assert data["id"] == seed["target"]
    if state == "committed":
        uuid.UUID(data["revision"])
    return data


def snapshot():
    return compare["snapshot"](seed)


status, auth, _ = call("POST", "/api/v1/auth/login", {"account": seed["username"], "password": os.environ["NP_BROWSER_PASSWORD"]})
assert status == 200, ("login", status)
token = auth["token"]
before = snapshot()
assert before["receipts"] == before["sources"] == 0
body = {"expected_revision": before["revision"], "description": "M5-disconnect-" + str(uuid.uuid4()), "changed_fields": ["description"]}
abandon = {"action": "update", "intake_id": seed["target"]}
early, late = str(uuid.uuid4()), str(uuid.uuid4())
call("PUT", target, body, early, "before-forward")
receipt(early, "unknown")
assert snapshot() == before
receipt(early, "abandoned", abandon)
stopped = snapshot()
assert stopped["receipts"] == 1 and stopped["target_sha256"] == before["target_sha256"]
assert call("PUT", target, body, early)[::2] == (409, "SAVE_REQUEST_ABANDONED")
assert snapshot() == stopped

call("PUT", target, body, late, "after-commit")
done = receipt(late, "committed")
after = snapshot()
assert after["receipts"] == 2 and after["sources"] == 0
assert after["revision"] == done["revision"] != before["revision"]
assert after["target_sha256"] != before["target_sha256"]
assert after["protected_sha256"] == before["protected_sha256"]
status, replay, _ = call("PUT", target, body, late)
assert status == 200 and replay == done
status, _, code = call("PUT", target, dict(body, description="different synthetic input"), late)
assert (status, code) == (409, "IDEMPOTENCY_PAYLOAD_MISMATCH")
assert receipt(late, "committed", abandon) == done and snapshot() == after
print(json.dumps({"gate": "I08", "before_forward": "unknown", "late_replay_after_abandon": 409,
                  "after_commit": "committed", "identical_replay": 200, "different_replay": 409,
                  "receipt_delta": 2, "protected_sha256": after["protected_sha256"]}))

# Receipt lookup alone cannot distinguish "not sent" from an in-flight writer.
# Observe the real writer waiting on the draft row before closing its TCP socket.
psql = shutil.which("psql") or "/usr/local/opt/libpq/bin/psql"
assert uuid.UUID(seed["target"])
lock = subprocess.Popen([psql, "-XqAtw", "-v", "ON_ERROR_STOP=1", os.environ["NP_TEST_DSN"]],
                        stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True)
attempt = str(uuid.uuid4())
pending_body = dict(body, expected_revision=after["revision"], description="M5-in-flight-" + attempt)
wire = http.client.HTTPConnection("127.0.0.1", 18789, timeout=5)
try:
    lock.stdin.write("BEGIN; SET LOCAL idle_in_transaction_session_timeout='30s';\n"
                     "SELECT pg_backend_pid() FROM case_intakes WHERE id='" + seed["target"] + "' FOR UPDATE;\n")
    lock.stdin.flush()
    assert select.select([lock.stdout], [], [], 5)[0], "row lock was not acquired"
    holder = int(lock.stdout.readline().strip())
    pending_headers = {"Content-Type": "application/json", "Authorization": "Bearer " + token, "Idempotency-Key": attempt}
    wire.request("PUT", target, json.dumps(pending_body), pending_headers)
    deadline = time.monotonic() + 6
    while not compare["query"]("SELECT to_jsonb(EXISTS(SELECT 1 FROM pg_stat_activity WHERE "
                               + str(holder) + " = ANY(pg_blocking_pids(pid))));"):
        if time.monotonic() >= deadline:
            raise RuntimeError("no PG waiter observed for the isolated holder")
        time.sleep(.05)
    receipt(attempt, "unknown")
finally:
    wire.close()
    lock.communicate("ROLLBACK;\n", timeout=5)
    assert lock.returncode == 0
deadline = time.monotonic() + 8
while True:
    status, outcome, code = call("POST", base + "/write-results/" + attempt + "/abandon", abandon)
    if status == 200:
        break
    assert (status, code) == (409, "SAVE_GUARD_BUSY") and time.monotonic() < deadline
    time.sleep(.1)
state = outcome["state"]
assert state in ("abandoned", "committed") and receipt(attempt, state) == outcome
settled = snapshot()
assert settled["receipts"] == after["receipts"] + 1 and settled["protected_sha256"] == after["protected_sha256"]
status, replay, code = call("PUT", target, pending_body, attempt)
if state == "abandoned":
    assert (status, code) == (409, "SAVE_REQUEST_ABANDONED") and settled["target_sha256"] == after["target_sha256"]
else:
    assert status == 200 and replay == outcome and settled["revision"] == outcome["revision"] != after["revision"]
assert snapshot() == settled
print(json.dumps({"gate": "I08-in-flight", "pg_row_waiter_before_disconnect": True,
                  "early_query": "unknown", "resolved_state": state, "replay_status": status, "receipt_delta": 1}))

events_file = Path(os.environ["NP_M5_PROXY_EVENT_FILE"])
events = [json.loads(line) for line in events_file.read_text().splitlines()]
assert [e["stage"] for e in events] == ["before-forward", "after-commit"]
keys = compare["query"]("""SELECT jsonb_object_agg(encode(sha256(convert_to(request_id::text,'UTF8')),'hex'),state)
    FROM case_intake_write_requests WHERE intake_id=:'target' AND actor_id=:'actor_id';""", seed)
assert len(keys) == 3
for event, issued, expected, upstream in zip(events, (early, late), ("abandoned", "committed"), (None, 200)):
    digest = hashlib.sha256(issued.encode()).hexdigest()
    assert event["key_sha256"] == digest and keys[digest] == expected
    assert event["upstream_status"] == upstream and event["client_dropped"] is True
    assert event["receipt_status"] == 200 and event["receipt_key_match"] is True
    assert event["receipt_committed"] == (expected == "committed")
    if expected == "committed":
        assert event["receipt_id_match"] is True
assert keys[hashlib.sha256(attempt.encode()).hexdigest()] == state
print(json.dumps({"gate": "I08-correlation", "pg_receipt_states": keys, "proxy_events_checked": len(events)}))

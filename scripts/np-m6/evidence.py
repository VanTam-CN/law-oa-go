"""Export per-run counts and fingerprints. Raw events, rows and HTTP bodies stay private."""

import hashlib
import json
import os
from pathlib import Path
import re

if not __debug__:
    raise SystemExit("run with assertions enabled")
work = Path(os.environ["NP_M5_WORK"])
assert work.name.startswith("np-m5.") and work.stat().st_mode & 0o077 == 0


def read(name):
    return (work / name).read_text()


assert read("passed").strip() == "M6_I12_REHEARSAL_PASS"


def read_json(name):
    return json.loads(read(name))


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


before = read_json("seeded.pg.json")
assert (
    before["source_rows"] == 1
    and before["committed_rows"] == 2
    and before["abandoned_rows"] == 1
)
assert (
    before["current_pointer_empty"]
    and before["receipt_rows"] == before["committed_rows"] + before["abandoned_rows"]
)
assert all(
    re.fullmatch("[0-9a-f]{64}", before[k])
    for k in ("draft_sha256", "sources_sha256", "receipts_sha256", "tombstones_sha256")
)
for step in ["repeated", "restarted", "rolled-back"]:
    assert read_json(step + ".pg.json") == before
matrix = {
    "partial": "partial 000086 state",
    "permission": "permission denied for table case_intakes",
    "append": "first apply: appending frozen 000086",
    "mismatch": "shared catalog check returned not-ready",
    "repeat": "repeat: verify-only mode",
}
for name, marker in matrix.items():
    assert marker in read(name + ".log")
for name in ("append", "repeat"):
    assert "PASS: NP schema verified" in read(name + ".log")
runtime = []
for line in read("runtime-stamps").splitlines():
    stage, ref, digest = line.split()
    assert re.fullmatch("[0-9a-f]{40}", ref) and re.fullmatch("[0-9a-f]{64}", digest)
    runtime.append(
        {"stage": stage, "embedded_commit": ref, "running_binary_sha256": digest}
    )
assert [r["stage"] for r in runtime] == ["legacy", "candidate", "restart", "rollback"]
assert (
    runtime[1]["running_binary_sha256"]
    == runtime[2]["running_binary_sha256"]
    == "213bb27532851e03bcc17ee630665e6fd2edd914cc1b711b477269decfae48d0"
)
assert runtime[-1]["running_binary_sha256"] != runtime[1]["running_binary_sha256"]
api = [json.loads(line) for line in read("boundary.jsonl").splitlines()]
assert [row["step"] for row in api] == [
    "legacy-create",
    "legacy-verify",
    "open",
    "restarted",
    "disabled",
    "legacy-final",
]
assert (
    api[-1]["valid_version_save_on_unprotected_draft"]
    and not api[4]["assistant_enabled"]
)
timeline = read_json("switch-result.json")
assert timeline["all_checks_pass"] and timeline["events_after_dedupe"] == 8
assert [x["stage"] for x in timeline["generations"]] == [x["stage"] for x in runtime]
assert len(read("steps").splitlines()) == 4
base = read("baseline-tables.sha256").strip()
assert re.fullmatch("[0-9a-f]{64}", base)
assert re.fullmatch(r"[0-9]+:[0-9]+", read("verified-table-oids").strip())
protocol = {}
for name in ("seed", "repeat", "restart", "rollback"):
    raw = work / (name + "-result.json")
    result = json.loads(raw.read_text())
    assert result["phase"] == ("seed" if name == "seed" else "verify")
    assert result["failed"] == [] and result["checks"]
    checks = [item["name"] for item in result["checks"] if item["ok"] is True]
    assert len(set(checks)) == len(result["checks"])
    assert all(
        re.fullmatch(r"[SV][0-9]{2}[a-z]? [A-Za-z0-9 /()_-]{2,90}", check)
        for check in checks
    )
    protocol[name] = {"passed_checks": checks, "raw_sha256": sha(raw)}
print(
    json.dumps(
        dict(
            gate="I12",
            scope="synthetic root upgrade / restart / protected rollback",
            status="PASS",
            app_runtime=runtime,
            migration_observations=matrix,
            frozen_migration_sha256=sha(Path("migrations/000086_narrow_pilot.up.sql")),
            shared_catalog_check_sha256=sha(
                Path("internal/database/narrow_pilot_schema.sql")
            ),
            baseline_public_table_names_and_oids_sha256=base,
            np_table_oids_retained=True,
            business_api=api,
            lifecycle=timeline,
            stopped_root_empty_http_windows=4,
            database_snapshots_equal=[
                "seeded",
                "verify-only",
                "restarted",
                "rolled-back",
            ],
            ledger=before,
            protocol=protocol,
            container_events_raw_sha256=sha(work / "events.jsonl"),
        ),
        indent=2,
    )
)

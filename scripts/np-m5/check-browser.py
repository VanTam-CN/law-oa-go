"""Check the frozen named browser cases and their PG window, emit only evidence fields."""
import argparse
import json
import os
from pathlib import Path
import runpy
import subprocess
import sys
from tempfile import TemporaryDirectory

if not __debug__:
    raise SystemExit("M5 assertions must be enabled")

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("suite", choices=["route-b", "storage", "prod", "navigation", "selection", "tabs", "legacy", "editing"])
parser.add_argument("seed", type=Path)
parser.add_argument("before", type=Path)
parser.add_argument("reports", nargs="+", type=Path)
args = parser.parse_args()
scripts = Path(__file__).resolve().parents[1]
root = scripts.parent
merged = {"suites": [], "errors": []}
runs = []
for file in args.reports:
    report = json.loads(file.read_text())
    stats = report["stats"]
    assert stats["expected"] and not any(stats[k] for k in ("skipped", "unexpected", "flaky"))
    merged["suites"].extend(report["suites"])
    merged["errors"].extend(report["errors"])
    runs.append({k: stats[k] for k in ("startTime", "duration", "expected")})

with TemporaryDirectory() as tmp:
    merged_path = Path(tmp) / "report.json"
    merged_path.write_text(json.dumps(merged))
    subprocess.run([sys.executable, scripts / "np-ci/check-i11-browser.py", args.suite, merged_path],
                   check=True, stdout=sys.stderr)

comparator = scripts / "np-first-slice/compare-main-browser-db.py"
compare = runpy.run_path(str(comparator))
before = json.loads(args.before.read_text())
row = compare["snapshot"](json.loads(args.seed.read_text()))
assert row["target"] == before["target"] and row["control"] == before["control"]
assert row["protected_sha256"] == before["protected_sha256"]
assert before["receipts"] == before["sources"] == 0
if args.suite in ("storage", "prod", "navigation"):
    assert row == before
if args.suite == "route-b":
    subprocess.run([sys.executable, comparator, "after", args.seed, args.before], check=True, stdout=sys.stderr)
    env = dict(os.environ, NP_GO_SHA_MANIFEST=str(args.before.resolve()))
    go = subprocess.run(["go", "test", "-count=1", "-json", "./internal/handlers",
                         "-run", "^TestNPGoDigestMatchesBrowserManifest$"], cwd=root, env=env, capture_output=True)
    go.check_returncode()
    events = [json.loads(line) for line in go.stdout.splitlines()]
    assert [e["Action"] for e in events if e.get("Test") == "TestNPGoDigestMatchesBrowserManifest"
            and e["Action"] in ("run", "pass", "fail", "skip")] == ["run", "pass"]
print(json.dumps({"suite": args.suite, "runs": runs, "protected_sha256": row["protected_sha256"],
                  "target_whole_row_unchanged": row["target_sha256"] == before["target_sha256"],
                  "target_materials": row["target_materials"], "protected_parties": len(row["parties"]),
                  "sources": row["sources"], "receipts": row["receipts"], "linked": row["linked"],
                  "manual": row["manual"], "go_digest_matched": True if args.suite == "route-b" else None}))

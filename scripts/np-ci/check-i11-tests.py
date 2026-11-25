"""Verify the frozen per-package Go and per-file Jest cases against actual runner output."""
import json
from collections import Counter
from pathlib import Path
import sys

root = Path(__file__).resolve().parents[2]
mode, filename = sys.argv[1:]
if mode == "jest":
    manifest = json.loads(Path(__file__).with_name("i11-jest-cases.json").read_text())
    expected = Counter((file, name) for file, names in manifest.items() for name in names)
    actual = Counter()
    report = json.loads(Path(filename).read_text())
    if report["success"] is not True:
        sys.exit("I11_JEST_RUN_FAILED")
    for suite in report["testResults"]:
        rel = Path(suite["name"]).relative_to(root / "frontend").as_posix()
        if suite["status"] != "passed":
            sys.exit("I11_JEST_SUITE_NOT_PASSED")
        for case in suite["assertionResults"]:
            if case["status"] != "passed":
                sys.exit("I11_JEST_CASE_NOT_PASSED")
            actual[rel, case["fullName"]] += 1
    if not expected or actual != expected:
        sys.exit("I11_JEST_REQUIRED_SET_MISMATCH")
elif mode in {"go", "tagged"}:
    if mode == "go":
        names = set(json.loads((root / "docs/np-evidence/m3/independent-review/verification.json").read_text())["go"]["expected"])
        names.remove("TestNPGoDigestMatchesBrowserManifest")  # verified in the browser's live PG window
        overrides = {
            "TestOptimizedDatabaseSlowQueryLogDoesNotLeakParameters": "law-oa-go/internal/database",
            "TestNarrowPilotAuditSkipsSensitiveBodyBeforeRead": "law-oa-go/internal/middleware",
            "TestNarrowPilotCacheZeroReadWrite": "law-oa-go/internal/middleware",
            "TestIntakeEarlyResponsesNoStore": "law-oa-go",
        }
        expected = {(overrides.get(n, "law-oa-go/internal/handlers"), n) for n in names.union(overrides)}
    else:
        names = ["TestNativeConcurrentDraftAndHistory", "TestAsyncPGCompletionAfterNativeEdit", "TestAsyncPGFormalDetectorCompletesOnce",
                 "TestConflictPGLinkAdvancesRevisionOnce", "TestM3ConflictCompletionAfterSourceSave",
                 "TestM3ConflictCompletionAfterSourceSave/sync", "TestM3ConflictCompletionAfterSourceSave/async",
                 "TestSchemaRequiresFrozenKeysAndTypes"]
        names += ["TestSchemaRequiresFrozenKeysAndTypes/" + s for s in ("wrong-index-columns", "missing-fk", "wrong-column-type")]
        expected = {("law-oa-go/tests/integration/m1", n) for n in names}
    ran, passed = Counter(), Counter()
    packages = set()
    with open(filename) as events:
        for line in events:
            event = json.loads(line)
            action, pkg, name = event["Action"], event.get("Package"), event.get("Test")
            if action in {"fail", "skip", "build-fail"}:
                sys.exit("I11_GO_FAILED_OR_SKIPPED")
            if action == "pass" and not name:
                packages.add(pkg)
            if name and action in {"run", "pass"}:
                (ran if action == "run" else passed)[pkg, name] += 1
    if not expected or any(ran[k] != 1 or passed[k] != 1 or k[0] not in packages for k in expected):
        sys.exit("I11_GO_REQUIRED_SET_MISMATCH")
else:
    sys.exit("I11_UNKNOWN_VALIDATION_MODE")
print(f"I11_TEST_PASS group={mode} required={len(expected)}")

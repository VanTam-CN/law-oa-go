"""Require real test events for the frozen M2 early PG checks."""
import json
import sys

schema = "TestSchemaRequiresFrozenKeysAndTypes"
required = {"TestNativeConcurrentDraftAndHistory", schema}
required.update(schema + "/" + variant for variant in
                ("wrong-index-columns", "missing-fk", "wrong-column-type"))

ran, passed, skipped = set(), set(), set()
package_passed = False
failed = False
with open(sys.argv[1]) as events:
    for line in events:
        event = json.loads(line)
        if event.get("Package") != "law-oa-go/tests/integration/m1":
            continue
        name, action = event.get("Test"), event["Action"]
        if name and action == "run":
            ran.add(name)
        elif name and action == "pass":
            passed.add(name)
        elif action == "skip":
            skipped.add(name)
        elif action == "fail":
            failed = True
        elif not name and action == "pass":
            package_passed = True

if int(sys.argv[2]) or failed:
    sys.exit("M2_GO_FAILED")
if skipped:
    sys.exit("M2_REQUIRED_TEST_SKIPPED")
if not ran:
    sys.exit("M2_NO_TESTS")
if not package_passed or not required <= (ran & passed):
    sys.exit("M2_REQUIRED_TEST_MISSING")
print(f"M2_PG_PASS required={len(required)} passed={len(passed)}")

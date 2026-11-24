"""Require one actual expected pass per named browser scenario."""
from collections import Counter
import json
from pathlib import Path
import sys

manifest = json.loads(Path(__file__).with_name("i11-browser-cases.json").read_text())
key, report = sys.argv[1:]
data = json.loads(Path(report).read_text())
passed = []


def inspect(suite):
    for spec in suite.get("specs", []):
        title = spec["title"]
        for test in spec["tests"]:
            results = test["results"]
            if (test["expectedStatus"] != "passed" or test["status"] != "expected"
                    or len(results) != 1 or results[0]["status"] != "passed"):
                sys.exit("I11_BROWSER_CASE_NOT_PASSED")
            passed.append(title)
    for item in suite.get("suites", []):
        inspect(item)


inspect(data)
if data.get("errors") or not passed or Counter(passed) != Counter(manifest[key]):
    sys.exit("I11_BROWSER_REQUIRED_SET_MISMATCH")
print(f"I11_BROWSER_PASS suite={key} cases={len(passed)}")

"""Exercise the frozen HTTP receipt and source contracts on the same owned draft."""

import json
import os
import subprocess
import sys
import uuid
from pathlib import Path

from boundary import login, pg, psql, request, work

phase = sys.argv[1]
assert phase in ("seed", "verify")
seed = json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
target = str(uuid.UUID(seed["target"]))
path = "case-intakes/" + target
token = login(seed["username"])
private = work / "private-ledger.json"
checks = []


def check(name, condition):
    assert condition, name
    checks.append({"name": name, "ok": True})


def api(method, route, body=None, status=200, key=None):
    response = request(method, route, token, body, status, key)
    return response["data"] if status == 200 else response["error"]["code"]


def write(key, body, status=200):
    return api("PUT", path, body, status, key)


def receipt(key, abandon=False):
    route = path + "/write-results/" + key
    if abandon:
        return api(
            "POST", route + "/abandon", {"action": "update", "intake_id": target}
        )
    return api("GET", route)


def matches(result, key, state):
    return (
        result["state"] == state
        and result["request_id"] == key
        and result["id"] == target
    )


if phase == "seed":
    probe = subprocess.check_output(
        ["docker", "exec", "-i", pg, *psql, "-v", "target=" + target],
        input=Path("scripts/np-m6/snapshot.sql").read_text(),
        text=True,
    )
    baseline = json.loads(probe)
    check(
        "S01 pristine ledger",
        baseline["source_rows"] == baseline["receipt_rows"] == 0
        and baseline["current_pointer_empty"],
    )
    revision = api("GET", path)["revision"]
    context = api(
        "POST",
        "case-intake-assist/contexts",
        {"intake_id": target, "expected_revision": revision},
    )
    turn = api(
        "POST",
        "case-intake-assist/contexts/" + context["id"] + "/turns",
        {
            "request_id": str(uuid.uuid4()),
            "expected_context_version": context["context_version"],
            "text": "Synthetic M6 intake facts for source review.",
        },
    )
    candidate = turn["candidate"]
    check(
        "S02 reviewed candidate with source",
        len(candidate["sources"]) == 1 and bool(candidate["description"]),
    )
    selection = {
        "context_id": context["id"],
        "expected_context_version": turn["context_version"],
        "fields": {
            "description": {
                "candidate_id": candidate["id"],
                "requested_mode": "candidate_sources",
                "reviewed": True,
            }
        },
    }
    source_body = {
        "expected_revision": revision,
        "description": candidate["description"],
        "changed_fields": ["description"],
        "source_selection": selection,
    }
    source_key, hand_key, tomb_key = (str(uuid.uuid4()) for _ in range(3))
    first = write(source_key, source_body)
    check(
        "S03 source commit identity",
        matches(first, source_key, "committed")
        and first["source_revision"] == first["revision"] != revision,
    )
    source_path = path + "/source-snapshots/" + first["revision"]
    stored = api("GET", source_path)
    pack = stored["source"]
    check(
        "S04 current reviewed source",
        stored["current"]
        and pack["context_id"] == context["id"]
        and pack["candidate_id"] == candidate["id"]
        and pack["original_revision"] == revision
        and pack["sources"] == candidate["sources"],
    )
    edited = "M6 manually reviewed text"
    hand = write(
        hand_key,
        {
            "expected_revision": first["revision"],
            "description": edited,
            "changed_fields": ["description"],
        },
    )
    check(
        "S05 manual commit clears current source",
        matches(hand, hand_key, "committed")
        and hand["revision"] != first["revision"]
        and not hand["source_revision"],
    )
    unknown = receipt(tomb_key)
    check(
        "S06 unknown before explicit abandon",
        unknown["request_id"] == tomb_key and unknown["state"] == "unknown",
    )
    tomb = receipt(tomb_key, True)
    check(
        "S07 persisted abandon identity",
        matches(tomb, tomb_key, "abandoned") and not tomb["revision"],
    )
    private.write_text(
        json.dumps(
            dict(
                target=target,
                first=first,
                hand=hand,
                tomb=tomb,
                source_body=source_body,
                edited=edited,
                pack=pack,
            )
        )
    )
else:
    saved = json.loads(private.read_text())
    assert saved["target"] == target
    first, hand, tomb = (saved[k] for k in ("first", "hand", "tomb"))
    source_key = first["request_id"]
    current = api("GET", path)
    check(
        "V01 current draft after manual save",
        current["revision"] == hand["revision"]
        and current["intake"]["description"] == saved["edited"]
        and not current["intake"]["metadata"].get("np_description_source_revision"),
    )
    now = api("GET", path + "/source-snapshots/" + hand["revision"])
    previous = api("GET", path + "/source-snapshots/" + first["revision"])
    check(
        "V02 historical source exact",
        previous["current"] is False
        and previous["source"] == saved["pack"]
        and previous["source_revision"] == first["revision"]
        and now["current"] is True
        and now["source"] is None,
    )
    check(
        "V03 original receipt identities and revisions",
        all(receipt(row["request_id"]) == row for row in (first, hand, tomb)),
    )
    check(
        "V04 replay committed result",
        write(source_key, saved["source_body"]) == first
        and receipt(source_key, True) == first,
    )
    check(
        "V05 changed payload on same key rejected",
        write(source_key, dict(saved["source_body"], description="Different"), 409)
        == "IDEMPOTENCY_PAYLOAD_MISMATCH",
    )
    late = {
        "expected_revision": hand["revision"],
        "description": "Late synthetic write",
        "changed_fields": ["description"],
    }
    check(
        "V06 tombstone blocks late original",
        receipt(tomb["request_id"], True) == tomb
        and write(tomb["request_id"], late, 409) == "SAVE_REQUEST_ABANDONED",
    )
    check(
        "V07 legacy save blocked by historical source",
        api(
            "PUT",
            path,
            {"expected_revision": hand["revision"], "title": "M6 legacy form"},
            409,
        )
        == "UPDATE_PROTOCOL_REQUIRED",
    )
print(json.dumps({"phase": phase, "checks": checks, "failed": []}))

#!/usr/bin/env python3
"""Verify M6 Docker lifecycle evidence from docker events and roots.txt without running anything."""

import json
import os
import sys
from pathlib import Path

STAGES = ("legacy", "candidate", "restart", "rollback")
ACTIONS = ("start", "die")
ROOT_IMAGE = "alpine:3.20"
BASE_IMAGES = ("postgres", "redis", "node")


class EvidenceError(Exception):
    def __init__(self, reason):
        super().__init__(reason)
        self.reason = reason


def parse_roots(text):
    rows = []
    for raw in text.splitlines():
        line = raw.strip()
        if not line:
            continue
        parts = line.split()
        if len(parts) != 2:
            raise EvidenceError("roots_line_format")
        stage, identity = parts
        if len(identity) != 64 or any(c not in "0123456789abcdef" for c in identity):
            raise EvidenceError("roots_id_format")
        rows.append((stage, identity))
    if tuple(row[0] for row in rows) != STAGES:
        raise EvidenceError("roots_stage_order")
    legacy, candidate, restart, rollback = (row[1] for row in rows)
    if candidate != restart:
        raise EvidenceError("restart_reuse_mismatch")
    if len({legacy, candidate, rollback}) != 3:
        raise EvidenceError("roots_id_collision")
    return rows


def event_image(event):
    attrs = event.get("Actor", {}).get("Attributes", {})
    return attrs.get("image") or event.get("from") or ""


def parse_events(lines, expected_label=None):
    seen = {}
    duplicates = 0
    events = []
    for raw in lines:
        line = raw.strip()
        if not line:
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            raise EvidenceError("event_not_json")
        if not isinstance(event, dict):
            raise EvidenceError("event_not_json")
        if event.get("Type") != "container":
            raise EvidenceError("event_type")
        action = event.get("Action") or event.get("status")
        if action not in ACTIONS:
            raise EvidenceError("event_action")
        identity = event.get("Actor", {}).get("ID") or event.get("id") or ""
        if len(identity) != 64 or any(c not in "0123456789abcdef" for c in identity):
            raise EvidenceError("event_id_format")
        nano = event.get("timeNano")
        if not isinstance(nano, int) or nano <= 0:
            raise EvidenceError("event_time_nano")
        label = event.get("Actor", {}).get("Attributes", {}).get("law-oa-m5")
        if not label:
            raise EvidenceError("event_label_missing")
        image = event_image(event)
        if not image:
            raise EvidenceError("event_image_missing")
        if any(image.startswith(base) for base in BASE_IMAGES):
            raise EvidenceError("root_image_is_base")
        key = (nano, action, identity)
        prior = seen.get(nano)
        if prior is not None:
            if prior == key:
                duplicates += 1
                continue
            raise EvidenceError("ambiguous_time_nano")
        seen[nano] = key
        events.append({"nano": nano, "action": action, "id": identity, "image": image, "label": label})
    if not events:
        raise EvidenceError("no_events")
    events.sort(key=lambda item: item["nano"])
    labels = {item["label"] for item in events}
    if len(labels) != 1:
        raise EvidenceError("event_label_mismatch")
    if expected_label is not None and labels != {expected_label}:
        raise EvidenceError("event_label_mismatch")
    return events, duplicates


def check(events, roots, duplicates):
    expected_ids = [roots[0][1], roots[1][1], roots[3][1]]
    known = set(expected_ids)
    unknown = {item["id"] for item in events} - known
    if unknown:
        raise EvidenceError("extra_root")
    images = {item["image"] for item in events}
    if images != {ROOT_IMAGE}:
        raise EvidenceError("root_image_mismatch")
    gen_of = {roots[0][1]: 0, roots[1][1]: 1, roots[3][1]: 3}
    sequence = [(gen_of[item["id"]], item["action"]) for item in events]
    expected = []
    for gen, count in ((0, 1), (1, 2), (3, 1)):
        for _ in range(count):
            expected.append((gen, "start"))
            expected.append((gen, "die"))
    if len(sequence) != len(expected):
        limit = min(len(sequence), len(expected))
        position = next((i for i in range(limit) if sequence[i] != expected[i]), limit)
        if len(sequence) < len(expected) and position == limit:
            raise EvidenceError(f"missing_event:{STAGES[expected[position][0]]}:{expected[position][1]}" if position < len(expected) else "missing_event")
        raise EvidenceError("unexpected_event_count")
    position = next((i for i in range(len(expected)) if sequence[i] != expected[i]), None)
    if position is not None:
        gen, action = expected[position]
        actual_gen, actual_action = sequence[position]
        if actual_gen == gen:
            raise EvidenceError(f"wrong_action:{STAGES[gen]}:{actual_action}")
        raise EvidenceError(f"wrong_generation:{STAGES[actual_gen]}:{actual_action}")
    base = events[0]["nano"]
    generations = []
    intervals = []
    cursor = 0
    for gen, count in ((0, 1), (1, 2), (3, 1)):
        for occurrence in range(count):
            start_event = events[cursor]
            die_event = events[cursor + 1]
            cursor += 2
            start_ms = (start_event["nano"] - base) / 1e6
            end_ms = (die_event["nano"] - base) / 1e6
            generations.append({
                "stage": STAGES[len(generations)],
                "occurrence": occurrence,
                "start_ms": round(start_ms, 3),
                "end_ms": round(end_ms, 3),
                "duration_ms": round(end_ms - start_ms, 3),
            })
            intervals.append((start_event["nano"], die_event["nano"]))
    peak = 0
    for index, (start, die) in enumerate(intervals):
        concurrent = sum(1 for other in intervals if other[0] <= start < other[1])
        peak = max(peak, concurrent)
        if die <= start:
            raise EvidenceError("nonpositive_lifecycle")
        if index and intervals[index - 1][1] > start:
            raise EvidenceError("overlapping_windows")
    if peak != 1:
        raise EvidenceError("concurrent_roots")
    gaps = {}
    for index in range(1, len(intervals)):
        prior_end = intervals[index - 1][1]
        start = intervals[index][0]
        gaps[f"{generations[index - 1]['stage']}:{generations[index]['stage']}"] = round((start - prior_end) / 1e6, 3)
    return {
        "gate": "M6-I12-switch-evidence",
        "stages": list(STAGES),
        "lifecycles": len(intervals),
        "distinct_ids": len(known),
        "restart_reuses_candidate_id": roots[1][1] == roots[2][1],
        "events_after_dedupe": len(events),
        "duplicates_removed": duplicates,
        "same_image_all_generations": True,
        "image": ROOT_IMAGE,
        "peak_concurrent_roots": peak,
        "generations": generations,
        "transition_gaps_ms": gaps,
        "all_checks_pass": True,
    }


def run_real(work):
    events_path = work / "events.jsonl"
    roots_path = work / "roots.txt"
    try:
        events, duplicates = parse_events(events_path.read_text(encoding="utf8").splitlines(), expected_label=work.name)
        roots = parse_roots(roots_path.read_text(encoding="utf8"))
        return check(events, roots, duplicates)
    except EvidenceError as error:
        print(json.dumps({"gate": "M6-I12-switch-evidence", "all_checks_pass": False, "reason": error.reason}), file=sys.stderr)
        raise SystemExit(1)
    except OSError:
        print(json.dumps({"gate": "M6-I12-switch-evidence", "all_checks_pass": False, "reason": "evidence_unreadable"}), file=sys.stderr)
        raise SystemExit(1)


def synth_events(roots, mutate=None):
    nano = 1_000_000_000
    lines = []
    plan = [(0, 1), (1, 2), (3, 1)]
    for gen, count in plan:
        for _ in range(count):
            for action in ACTIONS:
                nano += 500_000_000
                identity = roots[0] if gen == 0 else (roots[1] if gen == 1 else roots[3])
                lines.append(json.dumps({
                    "Type": "container",
                    "Action": action,
                    "Actor": {"ID": identity, "Attributes": {
                        "image": ROOT_IMAGE, "law-oa-m5": "np-m5.selftest", "name": "root"}},
                    "timeNano": nano,
                }))
    if mutate == "duplicate":
        lines.insert(4, lines[4])
    if mutate == "missing_die":
        lines.pop()
    if mutate == "overlap":
        payload = json.loads(lines[2])
        payload["timeNano"] = json.loads(lines[1])["timeNano"] - 1
        lines[2] = json.dumps(payload)
    if mutate == "extra_root":
        nano += 500_000_000
        lines.append(json.dumps({"Type": "container", "Action": "start",
            "Actor": {"ID": "f" * 64, "Attributes": {"image": ROOT_IMAGE, "law-oa-m5": "np-m5.selftest"}}, "timeNano": nano}))
    if mutate == "image_mismatch":
        for index, line in enumerate(lines):
            payload = json.loads(line)
            if payload["Actor"]["ID"] == roots[1]:
                payload["Actor"]["Attributes"]["image"] = "alpine:3.19"
                lines[index] = json.dumps(payload)
    if mutate == "base_image_root":
        payload = json.loads(lines[0])
        payload["Actor"]["Attributes"]["image"] = "redis:7-alpine"
        lines[0] = json.dumps(payload)
    if mutate == "label_mismatch":
        payload = json.loads(lines[0])
        payload["Actor"]["Attributes"]["law-oa-m5"] = "np-m5.other-window"
        lines[0] = json.dumps(payload)
    return lines


def self_test():
    roots = ["a" * 64, "b" * 64, "b" * 64, "c" * 64]
    roots_text = "\n".join(f"{stage} {identity}" for stage, identity in zip(STAGES, roots))
    roots_rows = parse_roots(roots_text)
    cases = {}
    events, duplicates = parse_events(synth_events(roots, "duplicate"), expected_label="np-m5.selftest")
    cases["dedupe_positive"] = duplicates == 1
    valid = check(events, roots_rows, duplicates)
    cases["valid_positive"] = bool(valid)
    cases["four_stage_names"] = [item["stage"] for item in valid["generations"]] == list(STAGES)
    cases["restart_not_candidate"] = valid["generations"][2]["stage"] == "restart"
    negative = {
        "missing_die": "missing_event",
        "overlap": "wrong_generation",
        "extra_root": "extra_root",
        "image_mismatch": "root_image_mismatch",
        "base_image_root": "root_image_is_base",
        "label_mismatch": "event_label_mismatch",
    }
    for name, reason in negative.items():
        try:
            events, duplicates = parse_events(synth_events(roots, name))
            check(events, roots_rows, duplicates)
            cases[name] = False
        except EvidenceError as error:
            cases[name] = error.reason.startswith(reason)
    try:
        parse_roots(roots_text.replace("restart b", "restart d"))
        cases["restart_reuse_mismatch"] = False
    except EvidenceError as error:
        cases["restart_reuse_mismatch"] = error.reason == "restart_reuse_mismatch"
    ok = all(cases.values())
    print(json.dumps({"gate": "M6-I12-switch-self-test", "self_test_pass": ok, "cases": cases}, indent=2))
    if not ok:
        raise SystemExit(1)


def main():
    if not __debug__ or os.environ.get("NP_LOCAL_CHECK") != "synthetic-only":
        raise SystemExit("synthetic fixture required")
    if len(sys.argv) == 2 and sys.argv[1] == "--self-test":
        self_test()
        return
    if len(sys.argv) != 1:
        raise SystemExit("usage: switch-check.py [--self-test]")
    work = Path(os.environ["NP_M5_WORK"])
    print(json.dumps(run_real(work), indent=2))


if __name__ == "__main__":
    main()

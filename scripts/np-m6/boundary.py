"""Check the same synthetic old tab, and actual assistant access on each side of a cut."""

import http.client
import json
import os
from pathlib import Path
import subprocess
import sys
import uuid

if not __debug__ or os.environ.get("NP_LOCAL_CHECK") != "synthetic-only":
    raise SystemExit("synthetic fixture required")
work = Path(os.environ["NP_M5_WORK"])
assert work.name.startswith("np-m5.") and work.stat().st_mode & 0o077 == 0
pg = (work / "pg").read_text().strip()
assert (
    subprocess.check_output(
        ["docker", "inspect", "-f", '{{index .Config.Labels "law-oa-m5"}}', pg],
        text=True,
    ).strip()
    == work.name
)
psql = [
    "psql",
    "-XqAtw",
    "-v",
    "ON_ERROR_STOP=1",
    "-U",
    "np_m1",
    "-d",
    "np_first_slice_dev",
]


def request(method, path, token=None, body=None, status=200, key=None):
    conn = http.client.HTTPConnection("127.0.0.1", 18789, timeout=10)
    headers = {"Content-Type": "application/json"}
    if key:
        headers["Idempotency-Key"] = key
    if token:
        headers["Authorization"] = "Bearer " + token
    try:
        conn.request(
            method,
            "/api/v1/" + path,
            json.dumps(body) if body is not None else None,
            headers,
        )
        response = conn.getresponse()
        assert response.status in ((status,) if isinstance(status, int) else status), (
            method,
            response.status,
        )
        return json.load(response)
    finally:
        conn.close()


def login(username):
    return request(
        "POST",
        "auth/login",
        body={"account": username, "password": os.environ["NP_BROWSER_PASSWORD"]},
    )["data"]["token"]


def row_digest(target):
    target = str(uuid.UUID(target))
    sql = f"SELECT encode(sha256(convert_to(to_jsonb(i)::text,'UTF8')),'hex') FROM case_intakes i WHERE i.id='{target}'"
    digest = subprocess.check_output(
        ["docker", "exec", pg, *psql, "-c", sql], text=True
    ).strip()
    assert len(digest) == 64
    return digest


def record(step, facts):
    with (work / "boundary.jsonl").open("a") as f:
        print(json.dumps({"step": step, **facts}), file=f)


def main(mode):
    if mode == "legacy-create":
        subprocess.run(
            ["docker", "exec", "-i", "-e", "NP_BROWSER_PASSWORD", pg, *psql],
            check=True,
            text=True,
            input=r"""\getenv pw NP_BROWSER_PASSWORD
    INSERT INTO users(username,email,name,password,role,status) VALUES
    ('m6_legacy','m6_legacy@np.invalid','M6 synthetic',crypt(:'pw',gen_salt('bf')),'lawyer','active');
    """,
        )
        token = login("m6_legacy")
        draft = request(
            "POST",
            "case-intakes",
            token,
            {"case_type": "civil", "title": "Synthetic M6", "description": "first"},
            (200, 201),
        )["data"]
        target = str(uuid.UUID(draft["id"]))
        path = "case-intakes/" + target
        form = {"title": "Synthetic M6 revised", "description": "reviewed legacy draft"}
        request("PUT", path, token, form)
        assert (
            request("GET", "cases/intake-workbench/" + target, token)["data"]["intake"][
                "description"
            ]
            == form["description"]
        )
        (work / "legacy-tab.json").write_text(
            json.dumps(
                dict(token=token, target=target, form=form, digest=row_digest(target))
            )
        )
        record(
            mode,
            {"old_native_save_and_reopen": True, "pg_row_sha256": row_digest(target)},
        )
    elif mode in ("legacy-verify", "legacy-final"):
        old = json.loads((work / "legacy-tab.json").read_text())
        path = "case-intakes/" + old["target"]
        assert (
            request("GET", path, old["token"])["data"]["intake"]["description"]
            == old["form"]["description"]
        )
        assert (
            request("PUT", path, old["token"], old["form"], 409)["error"]["code"]
            == "INTAKE_DRAFT_REVISION_REQUIRED"
        )
        assert row_digest(old["target"]) == old["digest"]
        revision = request("GET", path, old["token"])["data"]["revision"]
        stale = dict(old["form"], expected_revision=str(uuid.uuid4()))
        assert (
            request("PUT", path, old["token"], stale, 409)["error"]["code"]
            == "INTAKE_DRAFT_REVISION_CONFLICT"
        )
        assert row_digest(old["target"]) == old["digest"]
        if mode == "legacy-final":
            body = dict(
                old["form"],
                expected_revision=revision,
                description="M6 reopened and reviewed",
            )
            updated = request("PUT", path, old["token"], body)["data"]
            reopened = request("GET", path, old["token"])["data"]
            assert (
                reopened["intake"]["description"] == body["description"]
                and updated["revision"] == reopened["revision"] != revision
            )
        record(
            mode,
            {
                "stale_revision": 409,
                "valid_version_save_on_unprotected_draft": mode == "legacy-final",
                "old_tab_native_get": 200,
                "old_full_save": 409,
                "whole_pg_row_unchanged_after_rejected_requests": True,
            },
        )
    elif mode in ("open", "restarted", "disabled"):
        seed = json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
        token = (
            login(seed["username"])
            if mode == "open"
            else json.loads((work / "context-private.json").read_text())["token"]
        )
        context = "case-intake-assist/contexts"
        enabled = request("GET", "case-intake-assist/capabilities", token)["data"][
            "enabled"
        ]
        assert enabled is (mode != "disabled")
        revision = request("GET", "case-intakes/" + seed["target"], token)["data"][
            "revision"
        ]
        form = {"intake_id": seed["target"], "expected_revision": revision}
        if mode != "open":
            previous = json.loads((work / "context-private.json").read_text())
            assert (
                request("GET", context + "/" + previous["id"], token, status=409)[
                    "error"
                ]["code"]
                == "CONTEXT_UNAVAILABLE"
            )
        if enabled:
            got = request("POST", context, token, form)["data"]
            assert (
                request("GET", context + "/" + got["id"], token)["data"]["id"]
                == got["id"]
            )
            (work / "context-private.json").write_text(
                json.dumps({"token": token, "id": got["id"]})
            )
        else:
            assert (
                request("POST", context, token, form, 403)["error"]["code"]
                == "ASSISTANT_DISABLED"
            )
        record(
            mode,
            {
                "assistant_enabled": enabled,
                "fresh_context_http": 200 if enabled else 403,
                "previous_context_http": None if mode == "open" else 409,
            },
        )
    else:
        raise SystemExit("unknown boundary step")


if __name__ == "__main__":
    main(sys.argv[1])

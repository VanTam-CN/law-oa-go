from pathlib import Path
import datetime, hashlib, json, os, subprocess, time
root = Path.cwd()
evidence = root / "reports/evidence/intake-collaboration-mvp/20261107/backend-final-fixes"

def manifest():
    files = sorted(set(root.glob("internal/**/*.go")) | set(root.glob("migrations/*.sql")) | {root / "go.mod", root / "go.sum", root / "main.go"})
    return {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest() for p in files if p.is_file()}

before = manifest()
(evidence / "31-final-source-before.json").write_text(json.dumps(before, indent=2)+"\n")
env = os.environ.copy()
env.update(GOCACHE="/private/tmp/lawoa-intake-collaboration-runtime/go-cache", GOPROXY="off")
dsn = "host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable"
env.update(LAW_OA_COLLAB_PG_DSN=dsn, INTAKE_COLLAB_PG_DSN=dsn, CLIENT_IDEMPOTENCY_TEST_DSN=dsn.replace("dbname=lawoa_intake_mvp_test", "dbname=lawoa_client_idempotency_20261107"))
pattern = "^Test(BackendFinalReviewPostgres|ApprovalCollaborationPostgres|Inbox|ClientCreation|ClientParty|OpposingParty|IntakeSubmission|IntakeSupplement|ApprovalSubmission|CoreReview|CreateCaseIntake|UpdateIntake|ConfirmIntake|StartIntake|DraftRestore|DraftList|ListIntake|F04B|LegacyDraft|ConfirmAndConflict|PrepareCaseIntake|IntakeWrite|DenyIntake|ResubmitAtomic|CaseTxGuard|ConflictApproval|ConflictOfficer|IntakeDuplicateApprovalGuard|IntakeDelayedInterleaveDuplicateApproval|IntakeConcurrentDistinctIntakesAllSucceed|IntakeResidualDraftRetryAfterFailedSubmit)"
commands = [
    ("32-final-race.jsonl", ["go", "test", "-race", "-count=1", "-timeout=180s", "-json", "./internal/services", "./internal/handlers", "./internal/repositories", "./internal/models", "-run", pattern]),
    ("33-final-vet.log", ["go", "vet", "./internal/services", "./internal/handlers", "./internal/repositories", "./internal/models"]),
    ("34-final-build.log", ["go", "build", "./..."]),
    ("35-final-diff-check.log", ["git", "diff", "--check", "--", "internal"]),
]
results=[]
for filename, command in commands:
    start=time.monotonic()
    with (evidence / filename).open("w") as output:
        result = subprocess.run(command, stdout=output, stderr=subprocess.STDOUT, env=env)
    results.append(dict(log=filename, command=command, exit_code=result.returncode, seconds=time.monotonic()-start))
    print(filename, result.returncode, flush=True)
rows=[]
for line in (evidence / "32-final-race.jsonl").read_text().splitlines():
    try: rows.append(json.loads(line))
    except ValueError: pass
terminal=[r for r in rows if r.get("Action") in ("pass", "fail", "skip")]
tests=[r for r in terminal if "Test" in r]
leaves=[r for r in tests if not any(x["Test"].startswith(r["Test"]+"/") and x["Package"]==r["Package"] for x in tests)]
after=manifest()
(evidence / "36-final-source-after.json").write_text(json.dumps(after, indent=2)+"\n")
summary={"time":datetime.datetime.now().astimezone().isoformat(), "head":subprocess.check_output(["git","rev-parse","HEAD"],text=True).strip(), "commands":results, "source_count":len(before), "changed_during_run":[k for k in before.keys()|after.keys() if before.get(k)!=after.get(k)], "packages":[r for r in terminal if "Test" not in r], "leaf_counts":{a:sum(r["Action"]==a for r in leaves) for a in ("pass","fail","skip")}, "pg_leaves":[r for r in leaves if "Postgres" in r["Test"] or "postgres" in r["Test"]], "leaves":leaves}
(evidence / "37-final-results.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2)+"\n")
(evidence / "32-final-race.log").write_text("".join(r.get("Output", "") for r in rows))
print(json.dumps({k:summary[k] for k in ["leaf_counts","source_count","changed_during_run"]},ensure_ascii=False),flush=True)

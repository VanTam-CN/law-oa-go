"""I14: exercise only the isolated clone and its own root, preserving the stored evidence."""
import http.client
import json
import os
from pathlib import Path
import re
import subprocess
import uuid

if not __debug__ or os.environ.get("NP_LOCAL_CHECK")!="synthetic-only":
    raise SystemExit("Synthetic fixture required")
work=Path(os.environ["NP_M5_WORK"])
window=work/"history-window"
db=(window/"db").read_text().strip()
pg=(work/"pg").read_text().strip()
assert re.fullmatch("np_m5_history_[a-f0-9]{12}",db) and db!=os.environ["DB_DATABASE"]
for resource in [pg,(window/"app").read_text().strip()]:
    assert subprocess.check_output(["docker","inspect","-f",'{{index .Config.Labels "law-oa-m5"}}',resource],text=True).strip()==work.name
seed=json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
target=str(uuid.UUID(seed["target"]))
owner=int(seed["actor_id"])
controller=Path(__file__).with_name("history-window.sh")

def sql(command):
    return subprocess.check_output(["docker","exec",pg,"psql","-XqAtw","-v","ON_ERROR_STOP=1","-U","np_m1","-d",db,"-c",command],text=True).strip()

assert sql("SELECT current_database()")==db

def state():
    return json.loads(sql(f"""SELECT jsonb_build_object('revision',i.metadata->>'draft_revision',
     'pointer',i.metadata->>'np_description_source_revision','sources',jsonb_array_length(s.rows),
     'receipts',jsonb_array_length(w.rows),'digest',encode(sha256(convert_to(jsonb_build_array(
     to_jsonb(i),s.rows,w.rows,(SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM case_intake_parties p WHERE intake_id=i.id),
     (SELECT jsonb_agg(to_jsonb(m) ORDER BY id) FROM case_materials m WHERE intake_id=i.id))::text,'UTF8')),'hex'))
     FROM case_intakes i CROSS JOIN LATERAL (SELECT jsonb_agg(to_jsonb(s) ORDER BY saved_revision) AS rows
     FROM case_intake_source_snapshots s WHERE intake_id=i.id) s CROSS JOIN LATERAL
     (SELECT jsonb_agg(to_jsonb(w) ORDER BY request_id) AS rows FROM case_intake_write_requests w WHERE intake_id=i.id) w
     WHERE i.id='{target}' AND i.created_by={owner}"""))

def call(method,path,body=None,token=None,key=None):
    headers={"Content-Type":"application/json"}
    if token: headers["Authorization"]="Bearer "+token
    if key: headers["Idempotency-Key"]=key
    conn=http.client.HTTPConnection("127.0.0.1",18791,timeout=10)
    try:
        conn.request(method,"/api/v1/"+path,json.dumps(body) if body is not None else None,headers)
        r=conn.getresponse()
        return r.status,json.load(r)
    finally:
        conn.close()

def login():
    status,r=call("POST","auth/login",{"account":seed["username"],"password":os.environ["NP_BROWSER_PASSWORD"]})
    assert status==200
    return r["data"]["token"]

before=state()
assert before["sources"]==3 and before["receipts"]==4 and before["pointer"] is None
token=login()
out={}

def reject(name,expected,strict=False):
    body={"expected_revision":before["revision"],"description":"Synthetic history guard edit"}
    key=None
    if strict:
        body["changed_fields"]=["description"]
        key=str(uuid.uuid4())
    else:
        body["title"]="Synthetic legacy full edit"
    s,r=call("PUT","case-intakes/"+target,body,token,key)
    outcome=[s,(r.get("error") or {}).get("code")]
    assert outcome==expected,(name,outcome)
    out[name]=outcome

def capabilities(value):
    status,response=call("GET","case-intake-assist/capabilities",token=token)
    assert status==200 and response["data"]["enabled"] is value

def root(actors):
    subprocess.run(["bash",str(controller),"root"],env=dict(os.environ,NP_ASSISTANT_ACTORS=actors),
                   check=True,stdout=subprocess.DEVNULL)

capabilities(True)
reject("legacy_enabled",[409,"UPDATE_PROTOCOL_REQUIRED"])
assert state()==before
try:
    root("")
    token=login()
    capabilities(False)
    reject("legacy_empty_allowlist",[409,"UPDATE_PROTOCOL_REQUIRED"])
    assert state()==before
    for label,alter,restore in [
        ("missing_table","ALTER TABLE case_intake_source_snapshots RENAME TO m5_history_parked",
         "ALTER TABLE m5_history_parked RENAME TO case_intake_source_snapshots"),
        ("query_error","ALTER TABLE case_intake_source_snapshots RENAME COLUMN intake_id TO m5_parked_id",
         "ALTER TABLE case_intake_source_snapshots RENAME COLUMN m5_parked_id TO intake_id")]:
        sql(alter)
        try:
            reject(label+"_legacy",[503,"SAVE_RESULT_UNKNOWN"])
            reject(label+"_strict",[503,"SAVE_SCHEMA_REQUIRED"],strict=True)
        finally:
            sql(restore)
        assert state()==before,label
finally:
    root(str(owner))
token=login()
capabilities(True)
assert state()==before
print(json.dumps({"gate":"I14 migrated history","dedicated_cloned_database":True,
    "before":{"sources":3,"receipts":4,"current_source_pointer":None},
    "outcomes":out,"intake_sources_receipts_parties_materials_before_after_equal":True,
    "restored_schema_and_capability":True,"protected_digest":before["digest"]},indent=2))

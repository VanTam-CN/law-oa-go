"""Verify role, status, registered session and identity redaction in the owned history clone."""
import base64
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
db=(work/"history-window/db").read_text().strip()
pg=(work/"pg").read_text().strip()
assert re.fullmatch("np_m5_history_[a-f0-9]{12}",db)
assert subprocess.check_output(["docker","inspect","-f",'{{index .Config.Labels "law-oa-m5"}}',pg],text=True).strip()==work.name
seed=json.loads(Path(os.environ["NP_BROWSER_SEED"]).read_text())
actor=int(seed["actor_id"])
token=None
def sql(q):
    return subprocess.check_output(["docker","exec",pg,"psql","-XqAtw","-v","ON_ERROR_STOP=1","-U","np_m1","-d",db,"-c",q],text=True).strip()

def call(method,path,body=None,expected=200,key=None):
    headers={"Content-Type":"application/json"}
    if key: headers["Idempotency-Key"]=key
    if token: headers["Authorization"]="Bearer "+token
    conn=http.client.HTTPConnection("127.0.0.1",18791,timeout=10)
    try:
        conn.request(method,"/api/v1/"+path,json.dumps(body) if body is not None else None,headers)
        r=conn.getresponse()
        doc=json.load(r)
        code=(doc.get("error") or {}).get("code")
        assert r.status in (expected if isinstance(expected,tuple) else (expected,)),(method,path.split("/")[0],r.status,code)
        return doc
    finally:
        conn.close()

assert sql("SELECT current_database()")==db and sql(f"SELECT role||':'||status FROM users WHERE id={actor}")=="lawyer:active"
token=call("POST","auth/login",{"account":seed["username"],"password":os.environ["NP_BROWSER_PASSWORD"]})["data"]["token"]
assert call("GET","case-intake-assist/capabilities")["data"]["enabled"] is True
nonce=uuid.uuid4().hex
client_value="M5CL"+nonce
party_value="M5OP"+nonce
client=call("POST","clients",{"type":"企业","name":"M5 redaction "+nonce,
    "identity_type":"SOCIAL_CREDIT_CODE","identity_number":client_value})["data"]
draft=call("POST","case-intakes",{"title":"M5 synthetic privacy fixture","client_id":client["id"],"case_type":"civil",
    "parties":[{"entity_name":"M5 opponent","entity_type":"LEGAL_PERSON","party_role":"opposing_party",
                "identity_type":"SOCIAL_CREDIT_CODE","identity_number":party_value}]},(200,201))["data"]
target=str(uuid.UUID(draft["id"]))
columns=json.loads(sql(f"""SELECT jsonb_build_array(c.id_card_ciphertext,c.id_card_digest,p.identity_number_ciphertext,p.identity_number_digest)
    FROM case_intake_parties p JOIN case_intakes i ON i.id=p.intake_id JOIN clients c ON c.id=i.client_id WHERE i.id='{target}';"""))
markers=[client_value,party_value,*columns]
assert len(set(markers))==6 and all(isinstance(x,str) and len(x)>=12 for x in markers)
for path in ("case-intakes/"+target,"cases/intake-workbench/"+target):
    full=call("GET",path)
    assert full["data"]["intake"]["id"]==target and full["data"]["parties"]
    blob=json.dumps(full)
    assert all(marker not in blob for marker in markers)

context="case-intake-assist/contexts"
body={"intake_id":target,"expected_revision":draft["revision"]}
opened=call("POST",context,body)["data"]["id"]
role_results={}
# Test each transition against a session known to have opened the very same context.
for label,field,mutant,original in (("role_changed","role","assistant","lawyer"),("inactive","status","inactive","active")):
    assert sql(f"UPDATE users SET {field}='{mutant}' WHERE id={actor} AND {field}='{original}' RETURNING id")==str(actor)
    try:
        codes=[]
        for method,path,payload in (("GET",context+"/"+opened,None),("POST",context,body)):
            codes.append(call(method,path,payload,403)["error"]["code"])
        assert len(set(codes))==1
        role_results[label]={"context_read":403,"context_create":403,"code":codes[0]}
    finally:
        assert sql(f"UPDATE users SET {field}='{original}' WHERE id={actor} AND {field}='{mutant}' RETURNING id")==str(actor)
    assert call("GET","case-intake-assist/capabilities")["data"]["enabled"] is True

# The token remains the unmodified value from real login. Only its test registration is temporarily detached.
claim=token.split(".")[1]
key=json.loads(base64.urlsafe_b64decode(claim+"="*(-len(claim)%4)))["uuid"]
assert re.fullmatch("[a-f0-9]{32}",key)
redis=(work/"history-window/redis").read_text().strip()
def redis_exists(k):
    return subprocess.check_output(["docker","exec",redis,"redis-cli","--raw","EXISTS",k],text=True).strip()=="1"
assert redis_exists("access_token:"+key)
replacement=uuid.uuid4().hex
assert sql(f"UPDATE auth_token_sessions SET access_token_uuid='{replacement}' WHERE user_id={actor} AND access_token_uuid='{key}' RETURNING user_id")==str(actor)
try:
    assert call("GET","case-intakes/"+target)["data"]["intake"]["id"]==target
    code=call("POST",context,body,401)["error"]["code"]
    assert code=="ACCESS_SESSION_REQUIRED"
finally:
    assert sql(f"UPDATE auth_token_sessions SET access_token_uuid='{key}' WHERE user_id={actor} AND access_token_uuid='{replacement}' RETURNING user_id")==str(actor)
restored=call("POST",context,body)["data"]
assert restored["id"]
# Use a nonempty source candidate issued before a real client's access changes.
turn=call("POST",context+"/"+restored["id"]+"/turns",{"expected_context_version":restored["context_version"],
    "request_id":str(uuid.uuid4()),"text":"Synthetic reviewed facts for permission transition."})["data"]
candidate=turn["candidate"]
wall=int(sql(f"""INSERT INTO cases(case_number,title,client_id,lawyer_id,case_type,status,ethical_wall_enabled)
    VALUES('M5-{nonce[:16]}','M5 synthetic wall',{int(client["id"])},{actor},'civil','pending',true) RETURNING id"""))
def guard_state():
    return sql(f"""SELECT encode(sha256(convert_to(to_jsonb(i)::text,'UTF8')),'hex'),
      (SELECT count(*) FROM case_intake_source_snapshots s WHERE s.intake_id=i.id),
      (SELECT count(*) FROM case_intake_write_requests w WHERE w.intake_id=i.id)
      FROM case_intakes i WHERE i.id='{target}'""")
initial=guard_state()
try:
    call("GET","case-intakes/"+target,expected=403)
    sql(f"INSERT INTO case_ethical_wall_whitelist(case_id,user_id,granted_by,granted_at,reason) VALUES({wall},{actor},{actor},now(),'M5 synthetic')")
    assert call("GET","case-intakes/"+target)["data"]["intake"]["id"]==target
    sql(f"DELETE FROM case_ethical_wall_whitelist WHERE case_id={wall} AND user_id={actor}")
    call("GET","cases/intake-workbench/"+target,expected=403)
    call("GET",context+"/"+restored["id"],expected=403)
    sel={"context_id":restored["id"],"expected_context_version":turn["context_version"],"fields":{
        "description":{"candidate_id":candidate["id"],"reviewed":True,"requested_mode":"candidate_sources"}}}
    denial=call("PUT","case-intakes/"+target,{"expected_revision":draft["revision"],"description":candidate["description"],
        "changed_fields":["description"],"source_selection":sel},403,key=str(uuid.uuid4()))
    assert denial["error"]["code"]=="OBJECT_ACCESS_DENIED" and guard_state()==initial
finally:
    sql(f"DELETE FROM case_ethical_wall_whitelist WHERE case_id={wall}; DELETE FROM cases WHERE id={wall}")
assert call("GET","case-intakes/"+target)["data"]["intake"]["id"]==target
call("POST","auth/logout",{"token":token})
call("POST",context,body,401)
logout_revocation_recorded=redis_exists("blacklist:"+key) or not redis_exists("access_token:"+key)
assert logout_revocation_recorded
# Repeat the worker's expiry/revocation on the clone using the approved actor.
def signin(user):
    return call("POST","auth/login",{"account":user["username"],"password":os.environ["NP_BROWSER_PASSWORD"]})["data"]["token"]
token=signin(seed)
claim=token.split(".")[1]
current_key=json.loads(base64.urlsafe_b64decode(claim+"="*(-len(claim)%4)))["uuid"]
assert re.fullmatch("[a-f0-9]{32}", current_key)
assert sql(f"UPDATE auth_token_sessions SET access_token_expires=now()-interval '1 second' WHERE user_id={actor} AND access_token_uuid='{current_key}' RETURNING user_id")==str(actor)
call("POST",context,body,401)
token=signin(seed)
call("POST","auth/revoke/user",{"user_id":actor})
call("POST",context,body,401)
other=json.loads(Path(os.environ["NP_BROWSER_OTHER_SEED"]).read_text())
assert int(other["actor_id"])!=actor
token=signin(other)
assert call("GET","case-intake-assist/capabilities")["data"]["enabled"] is False
other_detail=call("GET","case-intakes/"+other["target"])["data"]
assert other_detail["intake"]["id"]==other["target"]
call("POST",context,{"intake_id":other["target"],"expected_revision":other_detail["revision"]},403)
call("GET","case-intakes/"+target,expected=403)
call("GET","case-intakes/"+str(uuid.uuid4()),expected=404)
print(json.dumps({"gates":["I03","I04","I13"],"root":"dedicated migrated clone, same verified binary",
    "identity_values":{"actual_nonempty_plaintexts":2,"actual_pg_ciphertexts":2,"actual_pg_digests":2,
                       "routes_checked":["canonical","workbench_alias"],"matches_in_full_response":0},
    "transitions":role_results,"unregistered_normal_login_token":{"native_detail":200,"assistant_create":401,"code":code},
    "ethical_wall":{"closed":403,"explicit_grant":200,"withdrawn_alias":403,"existing_context":403,"reviewed_source_save":403,
                    "intake_row_sources_and_receipts_unchanged":True,"restored_native_access":200},
    "auth_redis_registration_and_logout_observed":True,"sensitive_create_after_logout":401,
    "sensitive_create_expired":401,"sensitive_create_after_revoke":401,
    "second_active_lawyer":{"assistant_enabled":False,"self":200,"other":403,"missing":404,"create_context":403},
    "restored_registration_and_active_role_verified":True},indent=2))

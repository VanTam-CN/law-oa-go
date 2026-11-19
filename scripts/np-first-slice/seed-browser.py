#!/usr/bin/python3
"""Add only synthetic material versions through the actual loopback API."""
import json
import os
import urllib.request
import uuid

BASE = "http://127.0.0.1:18789/api/v1"
INTAKE = "00000000-0000-4000-8000-000000009801"

def call(path, body=None, token=None, content_type="application/json"):
    headers = {"Content-Type": content_type}
    if token:
        headers["Authorization"] = "Bearer " + token
    if body is not None and not isinstance(body, bytes):
        body = json.dumps(body).encode()
    request = urllib.request.Request(BASE + path, data=body, headers=headers)
    with urllib.request.urlopen(request, timeout=15) as response:
        result = json.load(response)
    assert result["success"], "synthetic API rejected the operation"
    return result["data"]

password = os.environ["NP_FIXTURE_PASSWORD"]
assert len(password) >= 12
token = call("/auth/login", {"account": "np.fixture.lawyer", "password": password})["token"]
listing = call(f"/case-intakes/{INTAKE}/materials", token=token)
materials = listing["materials"]
matching = [m for m in materials if m["name"] == "虚构回归材料"]
if not matching:
    call(f"/case-intakes/{INTAKE}/materials", {"name": "虚构回归材料", "material_type": "other", "required": True}, token)
    materials = call(f"/case-intakes/{INTAKE}/materials", token=token)["materials"]
    matching = [m for m in materials if m["name"] == "虚构回归材料"]
assert len(matching) == 1
path = f'/case-materials/{matching[0]["id"]}/files'
files = call(path, token=token)["files"]
assert len(files) <= 2, "unexpected synthetic material history"
texts = ["fictional browser original bytes", "fictional browser corrected bytes"]
for index in range(len(files), 2):
    boundary = "np-" + uuid.uuid4().hex
    body = (f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="np-v{index+1}.txt"\r\n'
            f'Content-Type: text/plain\r\n\r\n{texts[index]}\r\n--{boundary}--\r\n').encode()
    call(path, body, token, f"multipart/form-data; boundary={boundary}")
assert len(call(path, token=token)["files"]) == 2
print("SYNTHETIC_BROWSER_MATERIALS_OK versions=2")

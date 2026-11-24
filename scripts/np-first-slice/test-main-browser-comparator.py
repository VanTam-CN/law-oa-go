#!/usr/bin/env python3
"""Live PG checks of the comparison tool only. Each case owns a fresh synthetic window."""
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile

TOOL = Path(__file__).with_name('compare-main-browser-db.py')
spec = importlib.util.spec_from_file_location('comparator', TOOL)
tool = importlib.util.module_from_spec(spec)
spec.loader.exec_module(tool)

# Model only the four commits consumed by the comparator. No handler runs in this check.
FOUR_COMMITS = """
CREATE TEMP TABLE revisions AS SELECT n, gen_random_uuid()::text AS rev FROM generate_series(1,4) n;
INSERT INTO case_intake_source_snapshots(intake_id,saved_revision,serverpack)
SELECT :'target',rev,'{}'::jsonb FROM revisions WHERE n<4;
INSERT INTO case_intake_write_requests(actor_id,request_id,action,intake_id,state,request_digest,result_revision,source_revision)
SELECT :'actor_id',gen_random_uuid(),'update',:'target','committed','fixture',rev,CASE WHEN n<4 THEN rev END FROM revisions;
UPDATE case_intakes SET description='虚构人工摘要',updated_at=now(),
       metadata=jsonb_set(metadata,'{draft_revision}',to_jsonb((SELECT rev FROM revisions WHERE n=4)))
WHERE id=:'target';
"""


def check(args, expected):
    p = subprocess.run([sys.executable, str(TOOL), *map(str, args)], text=True, capture_output=True)
    if expected:
        if p.returncode != 1 or expected not in p.stderr:
            raise RuntimeError('expected comparator rejection: ' + expected)
    elif p.returncode:
        raise RuntimeError('unexpected comparator failure: ' + p.stderr.strip())


def run_case(root, name, change='', expected='', empty_history=False):
    manifest, baseline = root / (name + '.json'), root / (name + '.before')
    check(['seed', manifest], '')
    seed = json.loads(manifest.read_text())
    check(['before', manifest, baseline], '')
    sql = ('' if empty_history else FOUR_COMMITS) + change + "\nSELECT 'true'::json;"
    tool.query(sql, seed, write=True)
    check(['after', manifest, baseline], expected)
    print('COMPARATOR_CASE_OK ' + name)


if __name__ == '__main__':
    protected = 'protected rows changed'
    membership = 'invalid nonempty fixture'
    with tempfile.TemporaryDirectory() as directory:
        root = Path(directory)
        run_case(root, 'four-linked-and-manual')
        run_case(root, 'legacy-unlinked-receipt', """UPDATE case_intake_write_requests SET source_revision=''
            WHERE intake_id=:'target' AND source_revision IS NULL;""")
        run_case(root, 'no-business-writes', expected='expected exactly +3', empty_history=True)
        run_case(root, 'target-title', "UPDATE case_intakes SET title='changed' WHERE id=:'target';", protected)
        run_case(root, 'control-summary', "UPDATE case_intakes SET description='' WHERE id=:'control';", protected)
        run_case(root, 'unknown-big-rounding', """UPDATE case_intakes SET metadata=jsonb_set(metadata,
            '{np_unrelated,big}','9007199254740992') WHERE id=:'target';""", membership)
        run_case(root, 'client-row', """UPDATE clients SET version=version+1
            WHERE id=(SELECT client_id FROM case_intakes WHERE id=:'target');""", protected)
        run_case(root, 'party-row', "UPDATE case_intake_parties SET metadata= '{\"x\":1}' WHERE intake_id=:'target';", protected)
        run_case(root, 'material-row', "UPDATE case_materials SET required=NOT required WHERE intake_id=:'target';", protected)
        run_case(root, 'empty-materials', "DELETE FROM case_materials WHERE intake_id=:'target';", membership)
        run_case(root, 'wrong-receipt-actor', """UPDATE case_intake_write_requests SET actor_id='wrong'
            WHERE intake_id=:'target' AND source_revision IS NOT NULL;""", 'expected three linked commits')
        run_case(root, 'final-still-linked', """UPDATE case_intakes SET metadata=
            jsonb_set(metadata,'{np_description_source_revision}','"old"') WHERE id=:'target';""", 'expected three linked commits')

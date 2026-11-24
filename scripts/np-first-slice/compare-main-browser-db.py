#!/usr/bin/env python3
"""Run seed/before/after against the dedicated v14+86 fixture. Emit no row contents."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

HERE = Path(__file__).resolve().parent
GUARD = """
SET LOCAL search_path = public, pg_catalog;
SET LOCAL TIME ZONE 'UTC';
DO $$ BEGIN
 IF current_database() <> 'np_first_slice_dev' OR current_user <> 'np_m1'
    OR (SELECT version FROM schema_bootstrap_state WHERE id=1)
       IS DISTINCT FROM 'postgres-mvp-2026-08-27-v14'
    OR to_regclass('case_intake_source_snapshots') IS NULL
    OR to_regclass('case_intake_write_requests') IS NULL THEN
  RAISE EXCEPTION 'not the isolated v14+86 fixture';
 END IF;
END $$;
"""
SNAPSHOT = """
WITH t AS (SELECT * FROM case_intakes WHERE id=:'target'), c AS (SELECT * FROM case_intakes WHERE id=:'control'),
p AS (SELECT * FROM case_intake_parties WHERE intake_id IN (:'target',:'control')),
m AS (SELECT * FROM case_materials WHERE intake_id IN (:'target',:'control')),
client AS (SELECT * FROM clients WHERE id IN (SELECT client_id FROM t UNION SELECT client_id FROM c)),
s AS (SELECT * FROM case_intake_source_snapshots WHERE intake_id=:'target'),
w AS (SELECT * FROM case_intake_write_requests WHERE intake_id=:'target')
SELECT jsonb_build_object(
 'target', t.id, 'control', c.id, 'revision', t.metadata->>'draft_revision',
 'source_revision', t.metadata->>'np_description_source_revision',
 'large_number', t.metadata#>'{np_unrelated,big}' = '9007199254740993'::jsonb,
 'owned', t.created_by=:'actor_id'::bigint AND c.created_by=t.created_by,
 'target_materials',(SELECT count(*) FROM m WHERE intake_id=t.id),
 'parties',(SELECT jsonb_agg(id ORDER BY id) FROM p),
 'materials',(SELECT jsonb_agg(id ORDER BY id) FROM m),
 'clients',(SELECT jsonb_agg(id::text ORDER BY id) FROM client),
 'protected',jsonb_build_array(
    (to_jsonb(t)-ARRAY['description','updated_at','metadata']) ||
       jsonb_build_object('metadata',t.metadata-ARRAY['draft_revision','np_description_source_revision']),
    to_jsonb(c),(SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM p),
    (SELECT jsonb_agg(to_jsonb(m) ORDER BY id) FROM m),
    (SELECT jsonb_agg(to_jsonb(client) ORDER BY id) FROM client))::text,
 'sources',(SELECT count(*) FROM s), 'receipts',(SELECT count(*) FROM w),
 'linked',(SELECT count(DISTINCT s.saved_revision) FROM w JOIN s ON s.saved_revision=w.source_revision
        WHERE w.result_revision=s.saved_revision AND w.state='committed' AND w.actor_id=:'actor_id'),
 'manual',(SELECT count(*) FROM w WHERE state='committed' AND actor_id=:'actor_id'
                                     AND result_revision=t.metadata->>'draft_revision' AND NULLIF(source_revision,'') IS NULL)
) FROM t CROSS JOIN c;
"""


def query(sql, fields=None, write=False):
    env = {k: v for k, v in os.environ.items() if not k.startswith('PG')}
    env.update(PGPASSWORD=os.environ['PGPASSWORD'], PGCONNECT_TIMEOUT='3', PGSSLMODE='disable')
    cmd = [shutil.which('psql') or '/usr/local/opt/libpq/bin/psql', '-XqAtw', '-v', 'ON_ERROR_STOP=1',
           '-h', '127.0.0.1', '-p', '18790', '-d', 'np_first_slice_dev', '-U', 'np_m1']
    for key in ('target', 'control', 'actor_id') if fields else ():
        cmd += ['-v', key + '=' + str(fields[key])]
    begin = 'BEGIN;' if write else 'BEGIN ISOLATION LEVEL REPEATABLE READ READ ONLY;'
    proc = subprocess.run(cmd, input=begin + GUARD + sql + '\nCOMMIT;\n', env=env,
                          text=True, capture_output=True)
    if proc.returncode:
        raise ValueError(f'isolated PG query failed (psql exit {proc.returncode})')
    if not proc.stdout.strip():
        raise ValueError('target or control missing')
    return json.loads(proc.stdout)


def write_new(path, obj):
    with path.open('x', encoding='utf8') as f:
        json.dump(obj, f, ensure_ascii=False, indent=2)
        f.write('\n')


def snapshot(seed):
    row = query(SNAPSHOT, seed)
    if (seed['target'] == seed['control'] or not row['owned'] or not row['large_number']
            or row['clients'] != [seed['client_id']] or not row['parties'] or row['parties'] != seed['parties']
            or row['target_materials'] < 2 or row['materials'] != seed['materials'] or not row['revision']):
        raise ValueError('invalid nonempty fixture or protected membership')
    row['protected_sha256'] = hashlib.sha256(row.pop('protected').encode()).hexdigest()
    return row


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['seed', 'before', 'after'])
    parser.add_argument('manifest', type=Path)
    parser.add_argument('baseline', nargs='?', type=Path)
    args = parser.parse_args()
    os.umask(0o077)
    if args.mode == 'seed':
        if args.baseline or args.manifest.exists() or len(os.environ.get('NP_BROWSER_PASSWORD', '')) < 12:
            raise ValueError('seed requires fresh manifest and synthetic password of at least 12 characters')
        write_new(args.manifest, query((HERE / 'seed-main-browser.sql').read_text(), write=True))
    else:
        if not args.baseline:
            parser.error('supply the baseline file')
        row = snapshot(json.loads(args.manifest.read_text()))
        if args.mode == 'before':
            if row['sources'] or row['receipts'] or row['source_revision'] is not None:
                raise ValueError('dirty four-scenario window')
            write_new(args.baseline, row)
        else:
            before = json.loads(args.baseline.read_text())
            if any(before[k] != row[k] for k in ('target', 'control', 'protected_sha256')):
                raise ValueError('protected rows changed or baseline belongs to another window')
            if (row['sources'] - before['sources'], row['receipts'] - before['receipts']) != (3, 4):
                raise ValueError('expected exactly +3 sources and +4 receipts')
            if (row['revision'] == before['revision'] or row['source_revision'] is not None
                    or row['linked'] != 3 or row['manual'] != 1):
                raise ValueError('expected three linked commits and final unlinked manual revision')
            print(json.dumps({k: row[k] for k in ('protected_sha256', 'sources', 'receipts', 'linked', 'manual')}))
    print('BROWSER_PG_' + args.mode.upper() + '_OK')


if __name__ == '__main__':
    try:
        main()
    except (KeyError, OSError, ValueError) as exc:
        sys.exit(str(exc))

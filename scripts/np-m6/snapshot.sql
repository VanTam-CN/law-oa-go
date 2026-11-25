-- One statement and one MVCC snapshot. Only cardinalities and digests leave PostgreSQL.
WITH intake AS (SELECT * FROM case_intakes WHERE id = :'target'),
sources AS (SELECT * FROM case_intake_source_snapshots WHERE intake_id = :'target'),
receipts AS (SELECT * FROM case_intake_write_requests WHERE intake_id = :'target'),
digests AS (
  SELECT to_jsonb(i) AS draft,
    (SELECT jsonb_agg(to_jsonb(s) ORDER BY s.id) FROM sources s) AS sources,
    (SELECT jsonb_agg(to_jsonb(r) ORDER BY r.id) FROM receipts r) AS receipts,
    (SELECT jsonb_agg(to_jsonb(r) ORDER BY r.id) FROM receipts r WHERE state='abandoned') AS tombstones
  FROM intake i
)
SELECT jsonb_build_object(
  'source_rows', (SELECT count(*) FROM sources),
  'committed_rows', (SELECT count(*) FROM receipts WHERE state='committed'),
  'abandoned_rows', (SELECT count(*) FROM receipts WHERE state='abandoned'),
  'receipt_rows', (SELECT count(*) FROM receipts),
  'current_pointer_empty', nullif(draft->'metadata'->>'np_description_source_revision','') IS NULL,
  'draft_sha256', encode(sha256(convert_to(draft::text,'UTF8')),'hex'),
  'sources_sha256', encode(sha256(convert_to(sources::text,'UTF8')),'hex'),
  'receipts_sha256', encode(sha256(convert_to(receipts::text,'UTF8')),'hex'),
  'tombstones_sha256', encode(sha256(convert_to(tombstones::text,'UTF8')),'hex')
) FROM digests;

WITH approved AS (
 SELECT a.id, a.created_case_id::bigint case_id, s.snapshot_data
 FROM approval_requests a JOIN approval_snapshots s ON s.id=a.active_submission_snapshot_id
 WHERE a.id IN ('18e5ba60a968c0a0','18e5bad3447ea150')
), expected AS (
 SELECT a.id, p->>'entity_name' name, p->>'identity_number_digest' digest,
 CASE WHEN p->>'party_role'='client' OR p->'metadata'->>'party_type'='CO_CLIENT' THEN 'PLAINTIFF'
      WHEN p->>'party_role'='opposing_party' OR p->'metadata'->>'party_type'='OPPOSING_PARTY' THEN 'DEFENDANT'
      WHEN p->'metadata'->>'party_type'='GUARANTOR' THEN 'INTERESTED_PARTY' END role
 FROM approved a, jsonb_array_elements(a.snapshot_data->'metadata'->'intake_parties') p
), actual AS (
 SELECT a.id,e.name,e.identity_number_digest digest,cp.role
 FROM approved a JOIN case_parties cp ON cp.case_id=a.case_id JOIN entities e ON e.id=cp.entity_id
), missing AS (SELECT * FROM expected EXCEPT ALL SELECT * FROM actual),
extra AS (SELECT * FROM actual EXCEPT ALL SELECT * FROM expected)
SELECT json_build_object('approval_id',a.id,'case_id',a.case_id,
 'case_count_for_unique_fixture_title',(SELECT count(*) FROM cases WHERE title=(SELECT title FROM cases WHERE id=a.case_id)),
 'missing_approved_party_roles',(SELECT count(*) FROM missing m WHERE m.id=a.id),
 'extra_case_party_roles',(SELECT count(*) FROM extra e WHERE e.id=a.id),
 'canonical_slot_count',(SELECT count(*) FROM approval_case_creation_tracking t WHERE t.intake_id=a.snapshot_data->>'intake_id'),
 'canonical_slot_completed',(SELECT count(*) FROM approval_case_creation_tracking t WHERE t.intake_id=a.snapshot_data->>'intake_id' AND t.creation_status='completed' AND t.case_id=a.case_id::text))
FROM approved a ORDER BY a.id;

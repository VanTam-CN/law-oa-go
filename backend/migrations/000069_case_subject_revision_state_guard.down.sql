DROP TRIGGER IF EXISTS trg_case_subject_revisions_state_guard ON case_subject_revisions;
DROP FUNCTION IF EXISTS law_oa_guard_case_subject_revision();

CREATE TRIGGER trg_case_subject_revisions_append_only
BEFORE UPDATE OR DELETE ON case_subject_revisions
FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();

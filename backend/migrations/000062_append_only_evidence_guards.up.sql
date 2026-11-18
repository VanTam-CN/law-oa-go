-- PostgreSQL database guard for compliance evidence. Application hooks are
-- useful but cannot protect against a direct SQL mutation by an over-privileged
-- service account, so the evidence tables also reject UPDATE and DELETE.

CREATE OR REPLACE FUNCTION law_oa_reject_append_only_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'append-only evidence cannot be updated or deleted: %', TG_TABLE_NAME;
END;
$$;

DROP TRIGGER IF EXISTS trg_compliance_audit_events_append_only ON compliance_audit_events;
CREATE TRIGGER trg_compliance_audit_events_append_only
BEFORE UPDATE OR DELETE ON compliance_audit_events
FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();

DROP TRIGGER IF EXISTS trg_conflict_reviews_append_only ON conflict_reviews;
CREATE TRIGGER trg_conflict_reviews_append_only
BEFORE UPDATE OR DELETE ON conflict_reviews
FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();

DROP TRIGGER IF EXISTS trg_case_subject_revisions_append_only ON case_subject_revisions;
CREATE TRIGGER trg_case_subject_revisions_append_only
BEFORE UPDATE OR DELETE ON case_subject_revisions
FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();

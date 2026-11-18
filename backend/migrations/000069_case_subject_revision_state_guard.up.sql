-- Case subject revisions are stateful workflow records, not immutable evidence
-- rows. Preserve their identity and transition contract while allowing the
-- application to advance recheck and registration states.

DROP TRIGGER IF EXISTS trg_case_subject_revisions_append_only ON case_subject_revisions;
DROP TRIGGER IF EXISTS trg_case_subject_revisions_state_guard ON case_subject_revisions;

CREATE OR REPLACE FUNCTION law_oa_guard_case_subject_revision()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'case subject revisions cannot be deleted';
    END IF;

    IF NEW.id IS DISTINCT FROM OLD.id
       OR NEW.case_id IS DISTINCT FROM OLD.case_id
       OR NEW.base_subject_version IS DISTINCT FROM OLD.base_subject_version
       OR NEW.change_type IS DISTINCT FROM OLD.change_type
       OR NEW.requested_by IS DISTINCT FROM OLD.requested_by
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.reason IS DISTINCT FROM OLD.reason THEN
        RAISE EXCEPTION 'immutable case subject revision fields cannot be changed';
    END IF;

    IF NEW.payload IS DISTINCT FROM OLD.payload
       AND NOT (OLD.status = 'ENTITY_REGISTRATION_PENDING' AND NEW.status = 'CHANGE_PROPOSED') THEN
        RAISE EXCEPTION 'case subject revision payload may only resolve a pending entity registration';
    END IF;

    IF NEW.status IS DISTINCT FROM OLD.status AND NOT (
        (OLD.status = 'ENTITY_REGISTRATION_PENDING' AND NEW.status IN ('CHANGE_PROPOSED', 'CHANGE_REJECTED'))
        OR (OLD.status = 'CHANGE_PROPOSED' AND NEW.status = 'RECHECK_RUNNING')
        OR (OLD.status = 'RECHECK_REQUIRED' AND NEW.status IN ('RECHECK_RUNNING', 'CHANGE_APPROVED_AND_EFFECTIVE', 'CHANGE_REJECTED'))
        OR (OLD.status = 'RECHECK_RUNNING' AND NEW.status = 'RECHECK_REQUIRED')
    ) THEN
        RAISE EXCEPTION 'invalid case subject revision transition: % -> %', OLD.status, NEW.status;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_case_subject_revisions_state_guard
BEFORE UPDATE OR DELETE ON case_subject_revisions
FOR EACH ROW EXECUTE FUNCTION law_oa_guard_case_subject_revision();

DROP TRIGGER IF EXISTS trg_law_firm_policy_endorsements_append_only
    ON law_firm_compliance_policy_endorsements;
DROP TRIGGER IF EXISTS trg_law_firm_policy_packages_append_only
    ON law_firm_compliance_policy_packages;
DROP TABLE IF EXISTS law_firm_compliance_policy_endorsements;
DROP TABLE IF EXISTS law_firm_compliance_policy_packages;

# Production evidence gate plan

## Problem

QA health reports healthy because the production conflict gate is not enforced there. Users may confuse QA readiness with production readiness.

## Goal

Separate QA readiness from production evidence readiness and make missing production evidence explicit.

## Acceptance

- Production readiness cannot be healthy without complete required governance and operational evidence.
- QA clearly reports which production gates are skipped and why.
- Missing evidence is actionable.
- Relevant tests pass.

# Data Protection Impact Assessment (DPIA) — Template

> **TASK-P0-COMP-001 (PR-G).**
> Use this template for every processing activity that meets the GDPR
> Article 35 threshold (high risk to data subjects). Concrete DPIAs
> are filed in `compliance/dpia/dpia-<scope>.md` and linked from
> `compliance/dpia/README.md`. Reviewed and signed by the DPO before
> the processing goes live (TASK-P5-COMP-001 / TASK-P6-COMP-002).

## 1. Identification

| Field | Value |
|---|---|
| DPIA ID                | DPIA-NNNN |
| Scope (system / activity) | … |
| Date drafted           | YYYY-MM-DD |
| Author                 | … (role + name) |
| DPO reviewer           | … |
| Status                 | draft / under-review / approved / superseded |
| Linked requirements    | REQ-COMP-GDPR-003 |
| Linked tasks           | TASK-…, TASK-… |

## 2. Description of the processing

### 2.1 Nature

What personal data is collected, stored, transformed, transmitted, or
deleted. Reference the relevant `services/<svc>/db/schema/*.sql`
columns + Kafka topics + S3 prefixes by name.

### 2.2 Scope

- **Data categories** — name, email, IP address, device fingerprint, …
- **Data subjects** — internal users / external customers / data
  subjects whose imagery is processed.
- **Volume estimates** — rows/day, records total, retention horizon.
- **Geography** — region(s) of processing per
  [services/packages/region/region.go](../../services/packages/region/region.go).

### 2.3 Context

- Lawful basis under GDPR Art. 6 (consent / contract / legal
  obligation / vital interests / public task / legitimate interests).
- Purpose limitation — what the data is used for and what it is NOT.
- Relationship between controller and data subject.

### 2.4 Purposes

Plain-English statement of why the processing happens; cross-reference
the user story / requirement that authorises it.

## 3. Necessity and proportionality

- Why this data, this volume, this retention?
- What alternatives were considered (and rejected)?
- Data-minimisation choices already in place.
- Pseudonymisation / anonymisation / aggregation applied where the
  source data permits.

## 4. Risks to data subjects

For each risk:

| ID | Risk | Likelihood (L/M/H) | Severity (L/M/H) | Mitigations | Residual |
|---|---|---|---|---|---|
| R1 | Unauthorised disclosure due to misconfigured ABAC | M | H | TASK-P1-AUTHZ-001 deny-wins eval; audit-log of every allow/deny | L |
| R2 | …                                                       |   |   | …                                                                    |   |

## 5. Measures to address risks

- Technical (encryption at rest + in transit, key management, RBAC +
  ABAC, audit-log integrity, classification labelling).
- Organisational (training, access reviews, supplier agreements,
  incident response).
- Cross-references: relevant rows in `compliance/controls/*.csv`.

## 6. Consultation

- Internal stakeholders consulted (Engineering, Security, Legal).
- External consultations (supervisory authority where required by
  Art. 36 — high residual risk).

## 7. Approval

| Role | Name | Signature | Date |
|---|---|---|---|
| Author              | … | … | YYYY-MM-DD |
| DPO                 | … | … | YYYY-MM-DD |
| Engineering owner   | … | … | YYYY-MM-DD |
| Compliance officer  | … | … | YYYY-MM-DD |

## 8. Review schedule

This DPIA is reviewed:

- on change of scope (new field / new region / new purpose);
- on incident touching the in-scope data;
- at least annually as part of the management-review cycle
  (ISO 27001 A.5.4).

## 9. Change history

| Date | Author | Change |
|---|---|---|
| YYYY-MM-DD | …      | Initial draft. |

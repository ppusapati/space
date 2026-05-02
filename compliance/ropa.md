# Records of Processing Activities (ROPA)

> **TASK-P0-COMP-001 (PR-G).**
> GDPR Article 30 register of every processing activity touching
> personal data. Skeleton seeded in this PR; per-activity rows are
> appended as services land in Phase 1+. Referenced from
> [`controls/gdpr.csv`](controls/gdpr.csv) row `Art.30`.

| Field | Source-of-truth |
|---|---|
| Controller    | Chetana Inc. (legal entity TBD; finalised in TASK-P6-COMP-002) |
| DPO contact   | See [`dpo.md`](dpo.md) — populated in TASK-P6-COMP-002 (blocked on OQ-009) |
| EU Rep contact | See [`eu-representative.md`](eu-representative.md) — populated in TASK-P6-COMP-002 (blocked on OQ-010) |
| Last updated  | 2026-05-02 |

## Processing activities

The columns follow the GDPR Article 30(1) requirements. New rows are
added when a service that processes personal data goes live.

| ID | Name | Purpose | Categories of subjects | Categories of personal data | Recipients | Transfers outside EU | Retention | Security measures | Linked DPIA |
|---|---|---|---|---|---|---|---|---|---|
| ROPA-0001 | User authentication & session management | Authenticate platform users; maintain sessions; enforce ABAC | Internal users (employees, contractors); external IaaS customers (Phase 5) | email; hashed password; MFA secret; WebAuthn credentials; session metadata; nationality (for ITAR ABAC); is_us_person flag | Internal services (audit, notify, realtime-gw, export); IAM service replicas | None until v1.x EU rollout. v1 deploys exclusively to AWS GovCloud (US). | Active sessions: idle 1h / absolute 24h. Password hashes + MFA secrets: lifetime of account. Audit log of auth events: 5y online + 7y cold (per `services/audit/migrations/0002_retention.sql`). | TLS 1.3 (FIPS-validated); Argon2id password hashing; WebAuthn clone detection; rate-limited login (10/min/IP, 5 fails/account); audit chain integrity. | DPIA-0002 |
| ROPA-0002 | Audit log of platform actions | Maintain forensically-defensible record of every authenticated action (REQ-FUNC-PLT-AUDIT-*) | Internal users; external customers; system service accounts | actor_id; action; target resource; classification of target; before/after state for config changes; client IP | Audit service replicas; compliance reviewers; auditors | None | 5y online + 7y cold | Append-only with SHA-256 hash chain; tampering detected by chain verifier; DB role isolation. | DPIA-0002 |
| ROPA-0003 | Notification delivery (email, SMS, in-app) | Send platform notifications (security alerts, exports ready, AOI matches) | Internal users; external customers | email; SMS phone; notification preferences; rendered notification body (may include PII per template) | AWS SES (FIPS endpoint); AWS SNS (FIPS endpoint); realtime-gw for in-app | None until v1.x rollout | Notification audit: 5y online. Render artefact: 30 days. | TLS to SES/SNS via FIPS endpoints; per-template variable validation; mandatory-template guard for security messages. | DPIA-0002 |
| ROPA-0004 | External customer registration (IaaS public surface) | Allow external customers to sign up for the imagery API; manage AOI subscriptions; deliver presigned URLs | External customers (organisations + individual researchers) | name; email; org name; API key (hashed); usage metering counters | Public API gateway; subscription matcher; notify | None until v1.x rollout | Account: lifetime + 90 days. Usage logs: 5y. | API key stored hashed; per-key rate limiting; usage metering recorded against key id (not PII). | DPIA-0001 |
| ROPA-0005 | GDPR Subject Access / Erasure / Portability requests | Honour data-subject rights under GDPR Articles 15/17/20 | Any data subject of the platform | All data the requesting subject has in any controller-side store | Requestor (via signed presigned URL); compliance reviewer | None | Request artefact: 7 years (audit-log retention applies). | SAR routed through export service with 24h URL expiry; erasure preserves audit chain via hash anonymisation. | DPIA-0002 |

## Cross-references

- Each row's data flows are diagrammed in `plan/design.md` §1.
- The classification of each data category aligns with
  [`classification.yaml`](classification.yaml).
- Audit retention numbers come from
  [`services/audit/migrations/0002_retention.sql`](../services/audit/migrations/0002_retention.sql)
  (created in TASK-P1-AUDIT-002).
- Cross-border transfer position is recorded under each row's
  "Transfers outside EU" column. Until the EU region cluster lands
  (v1.x), no production data leaves AWS GovCloud (US-East), which
  removes the EU→US transfer question entirely for v1.

## Change history

| Date | Author | Change |
|---|---|---|
| 2026-05-02 | Compliance | Initial skeleton with 5 ROPA rows seeded for Phase 1+ services. |
